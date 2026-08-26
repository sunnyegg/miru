package torrentx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sunnyegg/miru/internal/storage"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
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
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	BytesCompleted int64   `json:"bytesCompleted"`
	BytesTotal     int64   `json:"bytesTotal"`
	Percent        float64 `json:"percent"`
	Error          string  `json:"error"`
	Source         string  `json:"source"`
}

type Manager struct {
	store      *storage.Store
	mu         sync.Mutex
	client     *torrent.Client
	dataDir    string
	current    *torrent.Torrent
	job        storage.TorrentJob
	onProgress func(JobView)
	onComplete func([]string)
}

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
		Percent:        JobPercent(job.BytesCompleted, job.BytesTotal),
		Error:          job.Error,
		Source:         job.Source,
	}
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

func (m *Manager) Busy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.job.ID != 0 && m.job.Status == "DOWNLOADING"
}

func (m *Manager) Start(source, destDir string) error {
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

	client, err := m.ensureClient(destDir)
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		_ = m.store.UpdateTorrentJob(job)
		return err
	}

	t, err := addSource(client, source)
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
	if job.ID != 0 && job.Status == "DOWNLOADING" {
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

func (m *Manager) ensureClient(dataDir string) (*torrent.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil && m.dataDir == dataDir {
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
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	m.client = client
	m.dataDir = dataDir
	return client, nil
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
	t.DownloadAll()

	m.update(jobID, func(job *storage.TorrentJob) {
		job.Name = name
		job.BytesTotal = total
		job.InfoHash = hash
	})

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			if m.current != t || m.job.ID != jobID || m.job.Status != "DOWNLOADING" {
				m.mu.Unlock()
				return
			}
			completed := t.BytesCompleted()
			m.job.BytesCompleted = completed
			m.job.BytesTotal = total
			if info != nil && m.job.Name == "" {
				m.job.Name = name
			}
			view := ToView(m.job)
			cb := m.onProgress
			done := completed >= total && total > 0
			m.mu.Unlock()

			_ = m.store.UpdateTorrentJob(m.snapshot(jobID))
			if cb != nil {
				cb(view)
			}
			if done {
				m.finish(t, jobID)
				return
			}
		}
	}
}

func (m *Manager) finish(t *torrent.Torrent, jobID int64) {
	t.DisallowDataUpload()
	files := videoFiles(t)

	m.mu.Lock()
	if m.job.ID == jobID {
		m.job.Status = "COMPLETED"
		m.job.BytesCompleted = t.BytesCompleted()
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

func addSource(client *torrent.Client, source string) (*torrent.Torrent, error) {
	if strings.HasPrefix(strings.ToLower(source), "magnet:") {
		return client.AddMagnet(source)
	}
	mi, err := metainfo.LoadFromFile(source)
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
