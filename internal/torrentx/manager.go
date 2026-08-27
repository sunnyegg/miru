package torrentx

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sunnyegg/miru/internal/storage"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/sunnyegg/miru/internal/networking"
	"golang.org/x/time/rate"
)

var videoExt = map[string]bool{
	".mkv":  true,
	".mp4":  true,
	".avi":  true,
	".webm": true,
	".mov":  true,
	".m4v":  true,
}

type JobView struct {
	ID                        int64   `json:"id"`
	Name                      string  `json:"name"`
	Status                    string  `json:"status"`
	BytesCompleted            int64   `json:"bytesCompleted"`
	BytesTotal                int64   `json:"bytesTotal"`
	BytesUploaded             int64   `json:"bytesUploaded"`
	Percent                   float64 `json:"percent"`
	UploadRatio               float64 `json:"uploadRatio"`
	SpeedBytesPerSecond       int64   `json:"speedBytesPerSecond"`
	UploadSpeedBytesPerSecond int64   `json:"uploadSpeedBytesPerSecond"`
	Error                     string  `json:"error"`
	Source                    string  `json:"source"`
}

type Manager struct {
	store        *storage.Store
	mu           sync.Mutex
	client       *torrent.Client
	dataDir      string
	networkKey   string
	uploadRate   *rate.Limiter
	downloadRate *rate.Limiter
	current      *torrent.Torrent
	job          storage.TorrentJob
	pausedFrom   string
	onProgress   func(JobView)
	onComplete   func([]string)
}

type RateLimits struct {
	Download int64
	Upload   int64
}

const rateLimiterBurst = 16 * 1024

func NewManager(store *storage.Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) SetCallbacks(onProgress func(JobView), onComplete func([]string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onProgress = onProgress
	m.onComplete = onComplete
}

func JobPercent(completed, total int64) float64 {
	if total <= 0 {
		return 0
	}
	p := float64(completed) / float64(total) * 100
	if p > 100 {
		return 100
	}
	return p
}

func ToView(job storage.TorrentJob) JobView {
	return JobView{
		ID:             job.ID,
		Name:           job.Name,
		Status:         job.Status,
		BytesCompleted: job.BytesCompleted,
		BytesTotal:     job.BytesTotal,
		BytesUploaded:  job.BytesUploaded,
		Percent:        JobPercent(job.BytesCompleted, job.BytesTotal),
		UploadRatio:    UploadRatio(job.BytesUploaded, job.BytesTotal),
		Error:          job.Error,
		Source:         job.Source,
	}
}

func UploadRatio(uploaded, total int64) float64 {
	if uploaded <= 0 || total <= 0 {
		return 0
	}
	return float64(uploaded) / float64(total)
}

func (m *Manager) Status() (JobView, error) {
	m.mu.Lock()
	if m.job.ID != 0 {
		view := ToView(m.job)
		m.mu.Unlock()
		return view, nil
	}
	m.mu.Unlock()

	job, err := m.store.LatestTorrentJob()
	if err != nil {
		return JobView{}, err
	}
	return ToView(job), nil
}

func (m *Manager) History() ([]JobView, error) {
	jobs, err := m.store.ListTorrentJobs()
	if err != nil {
		return nil, err
	}
	views := make([]JobView, 0, len(jobs))
	for _, job := range jobs {
		views = append(views, ToView(job))
	}
	return views, nil
}

func (m *Manager) Busy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.job.ID != 0 && isActiveStatus(m.job.Status)
}

func isActiveStatus(status string) bool {
	return status == "DOWNLOADING" || status == "PAUSED" || status == "SEEDING"
}

func (m *Manager) Start(source, destDir string, limits RateLimits, networkConfig networking.Config) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return errors.New("empty torrent source")
	}
	if destDir == "" {
		return errors.New("download folder is empty")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	if m.Busy() {
		return errors.New("a download is already running")
	}

	job := storage.TorrentJob{
		Source:  source,
		DestDir: destDir,
		Name:    displaySource(source),
		Status:  "DOWNLOADING",
	}
	id, err := m.store.InsertTorrentJob(job)
	if err != nil {
		return err
	}
	job.ID = id

	client, err := m.ensureClient(destDir, limits, networkConfig)
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		_ = m.store.UpdateTorrentJob(job)
		return err
	}

	sourceHTTP, err := networkConfig.HTTPClient()
	if err != nil {
		return err
	}
	t, err := addSource(client, source, sourceHTTP)
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		_ = m.store.UpdateTorrentJob(job)
		return err
	}

	m.mu.Lock()
	m.current = t
	m.job = job
	m.mu.Unlock()

	go m.run(t, job.ID)
	return nil
}

func (m *Manager) Cancel() error {
	m.mu.Lock()
	t := m.current
	job := m.job
	m.current = nil
	if job.ID != 0 && isActiveStatus(job.Status) {
		job.Status = "CANCELLED"
		job.Error = "cancelled"
		m.job = job
	}
	m.mu.Unlock()

	if t != nil {
		t.Drop()
	}
	if job.ID != 0 {
		return m.store.UpdateTorrentJob(job)
	}
	return nil
}

func (m *Manager) Pause() error {
	m.mu.Lock()
	t := m.current
	job := m.job
	if t == nil || job.ID == 0 || (job.Status != "DOWNLOADING" && job.Status != "SEEDING") {
		m.mu.Unlock()
		return errors.New("no active download to pause")
	}
	previous := job.Status
	m.pausedFrom = previous
	job.Status = "PAUSED"
	m.job = job
	m.mu.Unlock()

	if previous == "SEEDING" {
		t.DisallowDataUpload()
	} else if t.Info() != nil {
		t.CancelPieces(0, t.NumPieces())
	}
	return m.persistAndEmit(job.ID)
}

func (m *Manager) Resume() error {
	m.mu.Lock()
	t := m.current
	job := m.job
	previous := m.pausedFrom
	if t == nil || job.ID == 0 || job.Status != "PAUSED" {
		m.mu.Unlock()
		return errors.New("no paused download to resume")
	}
	if previous == "" {
		previous = "DOWNLOADING"
	}
	job.Status = previous
	m.job = job
	m.pausedFrom = ""
	m.mu.Unlock()

	if previous == "SEEDING" {
		t.AllowDataUpload()
	} else if t.Info() != nil {
		t.DownloadAll()
	}
	return m.persistAndEmit(job.ID)
}

func (m *Manager) persistAndEmit(jobID int64) error {
	job := m.snapshot(jobID)
	if err := m.store.UpdateTorrentJob(job); err != nil {
		return err
	}
	m.mu.Lock()
	cb := m.onProgress
	m.mu.Unlock()
	if cb != nil {
		cb(ToView(job))
	}
	return nil
}

func (m *Manager) Close() {
	_ = m.Cancel()
	m.mu.Lock()
	client := m.client
	m.client = nil
	m.mu.Unlock()
	if client != nil {
		client.Close()
	}
}

func (m *Manager) ensureClient(dataDir string, limits RateLimits, networkConfig networking.Config) (*torrent.Client, error) {
	normalizedNetwork, err := networkConfig.Normalized()
	if err != nil {
		return nil, err
	}
	networkKey := normalizedNetwork.Mode + ":" + normalizedNetwork.Address
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil && m.dataDir == dataDir && m.networkKey == networkKey {
		m.applyRateLimitsLocked(limits)
		return m.client, nil
	}
	if m.client != nil {
		m.client.Close()
		m.client = nil
	}
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = dataDir
	cfg.ListenPort = 0
	cfg.Seed = true
	cfg.NoUpload = false
	uploadRate, downloadRate := newRateLimiters(limits)
	cfg.UploadRateLimiter = uploadRate
	cfg.DownloadRateLimiter = downloadRate
	if normalizedNetwork.Mode == networking.ModeSystem {
		cfg.HTTPProxy = http.ProxyFromEnvironment
	} else if normalizedNetwork.Mode == networking.ModeSOCKS5 {
		cfg.HTTPDialContext = normalizedNetwork.DialContext
		cfg.TrackerDialContext = normalizedNetwork.DialContext
	}
	if normalizedNetwork.Mode == networking.ModeSOCKS5 {
		cfg.DialForPeerConns = false
		cfg.DisableUTP = true
		cfg.NoDHT = true
		cfg.NoDefaultPortForwarding = true
		cfg.DisablePEX = true
		cfg.DisableWebtorrent = true
		cfg.DisableIPv6 = true
		cfg.AcceptPeerConnections = false
		cfg.TrackerListenPacket = func(network, _ string) (net.PacketConn, error) {
			return nil, errors.New("UDP trackers are disabled when using SOCKS5")
		}
		client, err := torrent.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		client.AddDialer(torrent.NetworkDialer{
			Network: "tcp4",
			Dialer:  normalizedNetwork,
		})
		m.client = client
		m.dataDir = dataDir
		m.networkKey = networkKey
		m.uploadRate = uploadRate
		m.downloadRate = downloadRate
		return client, nil
	}
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	m.client = client
	m.dataDir = dataDir
	m.networkKey = networkKey
	m.uploadRate = uploadRate
	m.downloadRate = downloadRate
	return client, nil
}

func newRateLimiters(limits RateLimits) (*rate.Limiter, *rate.Limiter) {
	return newRateLimiter(limits.Upload), newRateLimiter(limits.Download)
}

func newRateLimiter(bytesPerSecond int64) *rate.Limiter {
	if bytesPerSecond <= 0 {
		return rate.NewLimiter(rate.Inf, rateLimiterBurst)
	}
	return rate.NewLimiter(rate.Limit(bytesPerSecond), rateLimiterBurst)
}

func (m *Manager) ApplyRateLimits(limits RateLimits) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.applyRateLimitsLocked(limits)
}

func (m *Manager) applyRateLimitsLocked(limits RateLimits) {
	if m.uploadRate != nil {
		m.uploadRate.SetLimit(rateLimit(limits.Upload))
	}
	if m.downloadRate != nil {
		m.downloadRate.SetLimit(rateLimit(limits.Download))
	}
}

func rateLimit(bytesPerSecond int64) rate.Limit {
	if bytesPerSecond <= 0 {
		return rate.Inf
	}
	return rate.Limit(bytesPerSecond)
}

func (m *Manager) run(t *torrent.Torrent, jobID int64) {
	select {
	case <-t.GotInfo():
	case <-time.After(2 * time.Minute):
		m.fail(jobID, errors.New("timed out waiting for torrent metadata"))
		return
	}

	info := t.Info()
	name := t.Name()
	total := t.Length()
	hash := t.InfoHash().HexString()
	m.mu.Lock()
	status := m.job.Status
	m.mu.Unlock()
	if status == "DOWNLOADING" {
		t.DownloadAll()
	}

	m.update(jobID, func(job *storage.TorrentJob) {
		job.Name = name
		job.BytesTotal = total
		job.InfoHash = hash
	})

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	lastBytes := t.BytesCompleted()
	lastUploaded := uploadedBytes(t)
	lastAt := time.Now()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			if m.current != t || m.job.ID != jobID || !isActiveStatus(m.job.Status) {
				m.mu.Unlock()
				return
			}
			completed := t.BytesCompleted()
			uploaded := uploadedBytes(t)
			now := time.Now()
			speed := bytesPerSecond(completed-lastBytes, now.Sub(lastAt))
			uploadSpeed := bytesPerSecond(uploaded-lastUploaded, now.Sub(lastAt))
			lastBytes = completed
			lastUploaded = uploaded
			lastAt = now
			m.job.BytesCompleted = completed
			m.job.BytesTotal = total
			m.job.BytesUploaded = uploaded
			if info != nil && m.job.Name == "" {
				m.job.Name = name
			}
			view := ToView(m.job)
			view.SpeedBytesPerSecond = speed
			view.UploadSpeedBytesPerSecond = uploadSpeed
			cb := m.onProgress
			downloadDone := m.job.Status == "DOWNLOADING" && completed >= total && total > 0
			seedingDone := m.job.Status == "SEEDING" && seedingComplete(uploaded, total)
			m.mu.Unlock()

			_ = m.store.UpdateTorrentJob(m.snapshot(jobID))
			if cb != nil {
				cb(view)
			}
			if downloadDone {
				m.startSeeding(jobID)
				continue
			}
			if seedingDone {
				m.finish(t, jobID)
				return
			}
		}
	}
}

func seedingComplete(uploaded, total int64) bool {
	return total > 0 && uploaded >= (total+1)/2
}

func (m *Manager) startSeeding(jobID int64) {
	m.mu.Lock()
	if m.job.ID == jobID && m.job.Status == "DOWNLOADING" {
		m.job.Status = "SEEDING"
	}
	job := m.job
	cb := m.onProgress
	m.mu.Unlock()
	_ = m.store.UpdateTorrentJob(job)
	if cb != nil {
		cb(ToView(job))
	}
}

func bytesPerSecond(bytes int64, elapsed time.Duration) int64 {
	if bytes <= 0 || elapsed <= 0 {
		return 0
	}
	return int64(float64(bytes) / elapsed.Seconds())
}

func uploadedBytes(t *torrent.Torrent) int64 {
	stats := t.Stats()
	return stats.BytesWrittenData.Int64()
}

func (m *Manager) finish(t *torrent.Torrent, jobID int64) {
	t.DisallowDataUpload()
	files := videoFiles(t)

	m.mu.Lock()
	if m.job.ID == jobID {
		m.job.Status = "COMPLETED"
		m.job.BytesCompleted = t.BytesCompleted()
		m.job.BytesUploaded = uploadedBytes(t)
		m.job.Error = ""
		m.current = nil
	}
	job := m.job
	complete := m.onComplete
	progress := m.onProgress
	m.mu.Unlock()

	_ = m.store.UpdateTorrentJob(job)
	if progress != nil {
		progress(ToView(job))
	}
	if complete != nil {
		complete(files)
	}
}

func (m *Manager) fail(jobID int64, err error) {
	m.mu.Lock()
	if m.current != nil {
		m.current.Drop()
		m.current = nil
	}
	if m.job.ID == jobID {
		m.job.Status = "FAILED"
		m.job.Error = err.Error()
	}
	job := m.job
	cb := m.onProgress
	m.mu.Unlock()
	_ = m.store.UpdateTorrentJob(job)
	if cb != nil {
		cb(ToView(job))
	}
}

func (m *Manager) update(jobID int64, fn func(*storage.TorrentJob)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.job.ID == jobID {
		fn(&m.job)
	}
}

func (m *Manager) snapshot(jobID int64) storage.TorrentJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.job.ID == jobID {
		return m.job
	}
	return storage.TorrentJob{ID: jobID}
}

func addSource(client *torrent.Client, source string, httpClient *http.Client) (*torrent.Torrent, error) {
	if strings.HasPrefix(strings.ToLower(source), "magnet:") {
		return client.AddMagnet(source)
	}

	parsed, err := url.Parse(source)
	isHTTP := err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https")
	if !isHTTP {
		mi, err := metainfo.LoadFromFile(source)
		if err != nil {
			return nil, err
		}
		return client.AddTorrent(mi)
	}

	if httpClient == nil {
		return nil, errors.New("HTTP client is unavailable")
	}
	response, err := httpClient.Get(parsed.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New("torrent URL returned an HTTP error")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	mi, err := metainfo.Load(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return client.AddTorrent(mi)
}

func videoFiles(t *torrent.Torrent) []string {
	var out []string
	for _, f := range t.Files() {
		ext := strings.ToLower(filepath.Ext(f.DisplayPath()))
		if !videoExt[ext] {
			continue
		}
		out = append(out, f.Path())
	}
	return out
}

func displaySource(source string) string {
	if strings.HasPrefix(strings.ToLower(source), "magnet:") {
		return "Magnet download"
	}
	return filepath.Base(source)
}

func ResolveDataPath(destDir, torrentPath string) string {
	if filepath.IsAbs(torrentPath) {
		return torrentPath
	}
	return filepath.Join(destDir, torrentPath)
}
