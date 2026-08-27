package torrentx

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

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
