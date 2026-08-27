package torrentx

import (
	"errors"
	"os"
	"path/filepath"
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
	Live                      bool    `json:"live"`
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
	uploadOffset int64
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

func liveView(job storage.TorrentJob) JobView {
	view := ToView(job)
	view.Live = true
	return view
}

func UploadRatio(uploaded, total int64) float64 {
	if uploaded <= 0 || total <= 0 {
		return 0
	}
	return float64(uploaded) / float64(total)
}

func (m *Manager) Status() (JobView, error) {
	m.mu.Lock()
	if m.job.ID != 0 && m.current != nil {
		view := liveView(m.job)
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
	m.mu.Lock()
	liveID := int64(0)
	if m.current != nil {
		liveID = m.job.ID
	}
	m.mu.Unlock()

	views := make([]JobView, 0, len(jobs))
	for _, job := range jobs {
		view := ToView(job)
		view.Live = job.ID == liveID
		views = append(views, view)
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
	m.uploadOffset = 0
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

func (m *Manager) Remove(id int64, deleteFiles bool) error {
	job, err := m.store.TorrentJobByID(id)
	if err != nil {
		return err
	}

	m.mu.Lock()
	active := m.current
	if m.job.ID == id {
		m.current = nil
		m.job = storage.TorrentJob{}
	} else {
		active = nil
	}
	m.mu.Unlock()

	if active != nil {
		active.Drop()
	}

	if !deleteFiles {
		return m.store.DeleteTorrentJob(id)
	}

	destDir := filepath.Clean(job.DestDir)
	target := filepath.Clean(filepath.Join(destDir, job.Name))
	relative, relErr := filepath.Rel(destDir, target)
	safeName := job.Name != "" && job.Name != "Magnet download"
	if safeName && relErr == nil && filepath.IsLocal(relative) {
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	}

	return m.store.DeleteTorrentJob(id)
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
		cb(liveView(job))
	}
	return nil
}

func (m *Manager) ResumeSeeding(id int64, limits RateLimits, networkConfig networking.Config) error {
	job, err := m.store.TorrentJobByID(id)
	if err != nil {
		return err
	}
	if job.Status != "SEEDING" {
		return errors.New("download is not waiting to seed")
	}
	if m.Busy() {
		return errors.New("a download is already running")
	}
	if job.DestDir == "" {
		return errors.New("download folder is empty")
	}
	if err := os.MkdirAll(job.DestDir, 0o755); err != nil {
		return err
	}

	client, err := m.ensureClient(job.DestDir, limits, networkConfig)
	if err != nil {
		return err
	}
	sourceHTTP, err := networkConfig.HTTPClient()
	if err != nil {
		return err
	}
	t, err := addSource(client, job.Source, sourceHTTP)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.current = t
	m.job = job
	m.uploadOffset = job.BytesUploaded
	m.mu.Unlock()

	go m.run(t, job.ID)
	return nil
}

func (m *Manager) Finish(id int64) error {
	m.mu.Lock()
	live := m.job.ID == id && m.current != nil && m.job.Status == "SEEDING"
	t := m.current
	m.mu.Unlock()
	if live {
		m.finish(t, id)
		return nil
	}

	job, err := m.store.TorrentJobByID(id)
	if err != nil {
		return err
	}
	if job.Status != "SEEDING" {
		return errors.New("download is not seeding")
	}

	job.Status = "COMPLETED"
	job.Error = ""
	if err := m.store.UpdateTorrentJob(job); err != nil {
		return err
	}

	m.mu.Lock()
	progress := m.onProgress
	complete := m.onComplete
	m.mu.Unlock()
	if progress != nil {
		progress(ToView(job))
	}
	if complete != nil {
		complete(videoFilesFromDisk(job.DestDir, job.Name))
	}
	return nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	t := m.current
	job := m.job
	client := m.client
	m.current = nil
	m.job = storage.TorrentJob{}
	m.client = nil
	m.pausedFrom = ""
	m.uploadOffset = 0
	if job.ID != 0 && isActiveStatus(job.Status) && job.Status != "SEEDING" {
		job.Status = "CANCELLED"
		job.Error = "cancelled"
	}
	m.mu.Unlock()

	if t != nil {
		t.Drop()
	}
	if job.ID != 0 {
		_ = m.store.UpdateTorrentJob(job)
	}
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
