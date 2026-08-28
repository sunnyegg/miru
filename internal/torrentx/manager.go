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

const (
	minConcurrentDownloads = 1
	maxConcurrentDownloads = 8
	DefaultSeedRatio       = 0.5
	maxSeedRatio           = 10.0
)

type JobView struct {
	ID                        int64      `json:"id"`
	Name                      string     `json:"name"`
	Status                    string     `json:"status"`
	BytesCompleted            int64      `json:"bytesCompleted"`
	BytesTotal                int64      `json:"bytesTotal"`
	BytesUploaded             int64      `json:"bytesUploaded"`
	Percent                   float64    `json:"percent"`
	UploadRatio               float64    `json:"uploadRatio"`
	SpeedBytesPerSecond       int64      `json:"speedBytesPerSecond"`
	UploadSpeedBytesPerSecond int64      `json:"uploadSpeedBytesPerSecond"`
	Error                     string     `json:"error"`
	Source                    string     `json:"source"`
	Live                      bool       `json:"live"`
	Files                     []FileView `json:"files"`
}

type Manager struct {
	store         *storage.Store
	mu            sync.Mutex
	client        *torrent.Client
	dataDir       string
	networkKey    string
	uploadRate    *rate.Limiter
	downloadRate  *rate.Limiter
	sessions      map[int64]*session
	maxConcurrent int
	limits        RateLimits
	seedRatio     float64
	networkConfig networking.Config
	onProgress    func(JobView)
	onComplete    func([]string)
	onError       func(string, error)
	persistLogged map[int64]struct{}
	userStarted   map[int64]struct{}
}

type RateLimits struct {
	Download int64
	Upload   int64
}

func NewManager(store *storage.Store) *Manager {
	return &Manager{
		store:         store,
		sessions:      make(map[int64]*session),
		persistLogged: make(map[int64]struct{}),
		userStarted:   make(map[int64]struct{}),
		maxConcurrent: minConcurrentDownloads,
		seedRatio:     DefaultSeedRatio,
	}
}

func (m *Manager) UserStarted(id int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.userStarted[id]
	return ok
}

func (m *Manager) SetCallbacks(onProgress func(JobView), onComplete func([]string), onError func(string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onProgress = onProgress
	m.onComplete = onComplete
	m.onError = onError
}

func (m *Manager) SetQueueConfig(limits RateLimits, networkConfig networking.Config) {
	m.rememberConfig(limits, networkConfig)
}

func (m *Manager) SetMaxConcurrent(max int) {
	if max < minConcurrentDownloads {
		max = minConcurrentDownloads
	}
	if max > maxConcurrentDownloads {
		max = maxConcurrentDownloads
	}
	m.mu.Lock()
	m.maxConcurrent = max
	m.mu.Unlock()
	m.PumpQueue()
}

func ClampSeedRatio(ratio float64) float64 {
	if ratio < 0 || ratio > maxSeedRatio {
		return DefaultSeedRatio
	}
	return ratio
}

func (m *Manager) SetSeedRatio(ratio float64) {
	m.mu.Lock()
	m.seedRatio = ClampSeedRatio(ratio)
	m.mu.Unlock()
}

func (m *Manager) seedRatioLocked() float64 {
	if m.seedRatio < 0 || m.seedRatio > maxSeedRatio {
		return DefaultSeedRatio
	}
	return m.seedRatio
}

func ClampMaxConcurrent(max int) int {
	if max < minConcurrentDownloads {
		return minConcurrentDownloads
	}
	if max > maxConcurrentDownloads {
		return maxConcurrentDownloads
	}
	return max
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
	files := decodeFiles(job.FilesJSON)
	if files == nil {
		files = []FileView{}
	}
	bytesTotal := job.BytesTotal
	if bytesTotal <= 0 {
		bytesTotal = selectedBytesTotal(files)
	}
	return JobView{
		ID:             job.ID,
		Name:           job.Name,
		Status:         job.Status,
		BytesCompleted: job.BytesCompleted,
		BytesTotal:     bytesTotal,
		BytesUploaded:  job.BytesUploaded,
		Percent:        JobPercent(job.BytesCompleted, bytesTotal),
		UploadRatio:    UploadRatio(job.BytesUploaded, bytesTotal),
		Error:          job.Error,
		Source:         job.Source,
		Files:          files,
	}
}

func liveView(job storage.TorrentJob) JobView {
	view := ToView(job)
	view.Live = true
	return view
}

func liveViewWithTorrent(job storage.TorrentJob, t *torrent.Torrent) JobView {
	view := liveView(job)
	if t == nil || t.Info() == nil {
		return view
	}
	files := fileViewsFromTorrent(t, decodeFiles(job.FilesJSON))
	view.Files = files
	completed, total := completedAndTotal(files)
	if total > 0 {
		view.BytesTotal = total
		view.BytesCompleted = completed
		view.Percent = JobPercent(completed, total)
		view.UploadRatio = UploadRatio(job.BytesUploaded, total)
	}
	return view
}

func completedAndTotal(files []FileView) (int64, int64) {
	var completed, total int64
	for _, file := range files {
		completed += file.BytesCompleted
		total += file.Length
	}
	return completed, total
}

func UploadRatio(uploaded, total int64) float64 {
	if uploaded <= 0 || total <= 0 {
		return 0
	}
	return float64(uploaded) / float64(total)
}

func (m *Manager) Status() (JobView, error) {
	m.mu.Lock()
	for _, session := range m.sessions {
		view := liveView(session.job)
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
	liveIDs := make(map[int64]struct{}, len(m.sessions))
	for jobID := range m.sessions {
		liveIDs[jobID] = struct{}{}
	}
	m.mu.Unlock()

	views := make([]JobView, 0, len(jobs))
	for _, job := range jobs {
		view := ToView(job)
		_, view.Live = liveIDs[job.ID]
		views = append(views, view)
	}
	return views, nil
}

func (m *Manager) Busy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions) > 0
}

func isActiveStatus(status string) bool {
	return status == "DOWNLOADING" || status == "PAUSED" || status == "SEEDING"
}

func (m *Manager) Start(source, destDir string, files []FileView, limits RateLimits, networkConfig networking.Config) error {
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

	m.rememberConfig(limits, networkConfig)

	job := storage.TorrentJob{
		Source:     source,
		DestDir:    destDir,
		Name:       displaySource(source),
		Status:     "QUEUED",
		BytesTotal: selectedBytesTotal(files),
		FilesJSON:  encodeFiles(files),
	}
	id, err := m.store.InsertTorrentJob(job)
	if err != nil {
		return err
	}
	job.ID = id
	m.mu.Lock()
	m.userStarted[id] = struct{}{}
	m.mu.Unlock()
	m.emitProgress(ToView(job))
	m.PumpQueue()
	return nil
}

func (m *Manager) activateJob(job storage.TorrentJob, limits RateLimits, networkConfig networking.Config) error {
	m.rememberConfig(limits, networkConfig)

	if job.DestDir == "" {
		return errors.New("download folder is empty")
	}
	if err := os.MkdirAll(job.DestDir, 0o755); err != nil {
		return err
	}

	client, err := m.ensureClient(job.DestDir, limits, networkConfig)
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		m.persistJob(job, "torrent persist failed job")
		m.emitProgress(ToView(job))
		m.PumpQueue()
		return err
	}

	sourceHTTP, err := networkConfig.HTTPClient()
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		m.persistJob(job, "torrent persist failed job")
		m.emitProgress(ToView(job))
		m.PumpQueue()
		return err
	}
	torrentHandle, err := addSource(client, job.Source, sourceHTTP)
	if err != nil {
		job.Status = "FAILED"
		job.Error = err.Error()
		m.persistJob(job, "torrent persist failed job")
		m.emitProgress(ToView(job))
		m.PumpQueue()
		return err
	}

	job.Status = "DOWNLOADING"
	job.Error = ""
	if err := m.store.UpdateTorrentJob(job); err != nil {
		torrentHandle.Drop()
		return err
	}

	m.mu.Lock()
	m.sessions[job.ID] = &session{
		torrent: torrentHandle,
		job:     job,
	}
	m.mu.Unlock()

	go m.run(torrentHandle, job.ID)
	m.emitProgress(liveView(job))
	return nil
}

func (m *Manager) PumpQueue() {
	for {
		m.mu.Lock()
		if m.downloadingCountLocked() >= m.maxConcurrent {
			m.mu.Unlock()
			return
		}
		limits := m.limits
		networkConfig := m.networkConfig
		m.mu.Unlock()

		job, err := m.store.NextQueuedTorrentJob()
		if errors.Is(err, storage.ErrNotFound) {
			return
		}
		if err != nil {
			m.reportError("torrent queue read", err)
			return
		}

		if activateErr := m.activateJob(job, limits, networkConfig); activateErr != nil {
			continue
		}
	}
}

func (m *Manager) Cancel(id int64) error {
	job, err := m.store.TorrentJobByID(id)
	if err != nil {
		return err
	}

	if job.Status == "QUEUED" {
		job.Status = "CANCELLED"
		job.Error = "cancelled"
		if err := m.store.UpdateTorrentJob(job); err != nil {
			return err
		}
		m.emitProgress(ToView(job))
		return nil
	}

	m.mu.Lock()
	session, ok := m.sessionByID(id)
	if !ok {
		m.mu.Unlock()
		return errors.New("no active download to cancel")
	}
	torrentHandle := session.torrent
	if isActiveStatus(session.job.Status) {
		session.job.Status = "CANCELLED"
		session.job.Error = "cancelled"
	}
	job = session.job
	delete(m.sessions, id)
	m.mu.Unlock()

	if torrentHandle != nil {
		torrentHandle.Drop()
	}
	if err := m.store.UpdateTorrentJob(job); err != nil {
		return err
	}
	m.emitProgress(ToView(job))
	m.PumpQueue()
	return nil
}

func (m *Manager) Remove(id int64, deleteFiles bool) error {
	job, err := m.store.TorrentJobByID(id)
	if err != nil {
		return err
	}

	m.mu.Lock()
	torrentHandle := m.removeSessionLocked(id)
	m.mu.Unlock()

	if torrentHandle != nil {
		torrentHandle.Drop()
	}

	if !deleteFiles {
		if err := m.store.DeleteTorrentJob(id); err != nil {
			return err
		}
		if job.Status == "DOWNLOADING" || job.Status == "QUEUED" {
			m.PumpQueue()
		}
		return nil
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

	if err := m.store.DeleteTorrentJob(id); err != nil {
		return err
	}
	if job.Status == "DOWNLOADING" || job.Status == "QUEUED" {
		m.PumpQueue()
	}
	return nil
}

func (m *Manager) Pause(id int64) error {
	m.mu.Lock()
	session, ok := m.sessionByID(id)
	if !ok {
		m.mu.Unlock()
		return errors.New("no active download to pause")
	}
	torrentHandle := session.torrent
	job := session.job
	if torrentHandle == nil || job.Status != "DOWNLOADING" && job.Status != "SEEDING" {
		m.mu.Unlock()
		return errors.New("no active download to pause")
	}
	previous := job.Status
	session.pausedFrom = previous
	job.Status = "PAUSED"
	session.job = job
	m.mu.Unlock()

	if previous == "SEEDING" {
		torrentHandle.DisallowDataUpload()
	} else if torrentHandle.Info() != nil {
		torrentHandle.CancelPieces(0, torrentHandle.NumPieces())
	}
	if err := m.persistAndEmit(job.ID); err != nil {
		return err
	}
	m.PumpQueue()
	return nil
}

func (m *Manager) Resume(id int64) error {
	m.mu.Lock()
	session, ok := m.sessionByID(id)
	if !ok {
		m.mu.Unlock()
		return errors.New("no paused download to resume")
	}
	torrentHandle := session.torrent
	job := session.job
	previous := session.pausedFrom
	if torrentHandle == nil || job.Status != "PAUSED" {
		m.mu.Unlock()
		return errors.New("no paused download to resume")
	}
	if previous == "" {
		previous = "DOWNLOADING"
	}
	if previous == "DOWNLOADING" && m.downloadingCountLocked() >= m.maxConcurrent {
		m.mu.Unlock()
		return errors.New("all download slots are in use")
	}
	job.Status = previous
	session.job = job
	session.pausedFrom = ""
	m.mu.Unlock()

	if previous == "SEEDING" {
		torrentHandle.AllowDataUpload()
		return m.persistAndEmit(job.ID)
	}
	if torrentHandle.Info() == nil {
		return m.persistAndEmit(job.ID)
	}
	if _, err := applyFileSelection(torrentHandle, decodeFiles(job.FilesJSON)); err != nil {
		return err
	}
	return m.persistAndEmit(job.ID)
}

func (m *Manager) persistAndEmit(jobID int64) error {
	job := m.snapshot(jobID)
	if err := m.store.UpdateTorrentJob(job); err != nil {
		return err
	}
	m.mu.Lock()
	_, live := m.sessionByID(jobID)
	m.mu.Unlock()
	if live {
		m.emitProgress(liveView(job))
		return nil
	}
	m.emitProgress(ToView(job))
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
	if job.DestDir == "" {
		return errors.New("download folder is empty")
	}
	if err := os.MkdirAll(job.DestDir, 0o755); err != nil {
		return err
	}

	m.rememberConfig(limits, networkConfig)

	client, err := m.ensureClient(job.DestDir, limits, networkConfig)
	if err != nil {
		return err
	}
	sourceHTTP, err := networkConfig.HTTPClient()
	if err != nil {
		return err
	}
	torrentHandle, err := addSource(client, job.Source, sourceHTTP)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.sessions[job.ID] = &session{
		torrent:      torrentHandle,
		job:          job,
		uploadOffset: job.BytesUploaded,
	}
	m.mu.Unlock()

	go m.run(torrentHandle, job.ID)
	return nil
}

func (m *Manager) Finish(id int64) error {
	m.mu.Lock()
	session, live := m.sessionByID(id)
	var torrentHandle *torrent.Torrent
	if live && session.job.Status == "SEEDING" {
		torrentHandle = session.torrent
	}
	m.mu.Unlock()
	if live && torrentHandle != nil {
		m.finish(torrentHandle, id)
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
	complete := m.onComplete
	m.mu.Unlock()
	m.emitProgress(ToView(job))
	if complete != nil {
		complete(videoFilesFromDisk(job.DestDir, job.Name))
	}
	return nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	sessions := make([]*session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	client := m.client
	m.sessions = make(map[int64]*session)
	m.client = nil
	m.dataDir = ""
	m.networkKey = ""
	m.mu.Unlock()

	for _, session := range sessions {
		if session.torrent != nil {
			session.torrent.Drop()
		}
		job := session.job
		if job.ID == 0 || !isActiveStatus(job.Status) || job.Status == "SEEDING" {
			continue
		}
		if job.Status == "PAUSED" {
			job.Status = "FAILED"
			job.Error = "interrupted by restart"
		} else {
			job.Status = "QUEUED"
			job.Error = ""
		}
		m.persistJob(job, "torrent persist on close")
	}
	if client != nil {
		client.Close()
	}
}

func (m *Manager) persistJob(job storage.TorrentJob, operation string) {
	if err := m.store.UpdateTorrentJob(job); err != nil {
		m.reportError(operation, err)
	}
}

func (m *Manager) persistProgress(jobID int64, job storage.TorrentJob) {
	if err := m.store.UpdateTorrentJob(job); err != nil {
		m.reportPersistOnce(jobID, err)
		return
	}
	m.clearPersistOnce(jobID)
}

func (m *Manager) reportPersistOnce(jobID int64, err error) {
	m.mu.Lock()
	_, already := m.persistLogged[jobID]
	if !already {
		if m.persistLogged == nil {
			m.persistLogged = make(map[int64]struct{})
		}
		m.persistLogged[jobID] = struct{}{}
	}
	reporter := m.onError
	m.mu.Unlock()
	if already || reporter == nil || err == nil {
		return
	}
	reporter("torrent persist progress", err)
}

func (m *Manager) clearPersistOnce(jobID int64) {
	m.mu.Lock()
	delete(m.persistLogged, jobID)
	m.mu.Unlock()
}

func (m *Manager) reportError(operation string, err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	reporter := m.onError
	m.mu.Unlock()
	if reporter != nil {
		reporter(operation, err)
	}
}

func (m *Manager) update(jobID int64, fn func(*storage.TorrentJob)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessionByID(jobID)
	if ok {
		fn(&session.job)
	}
}

func (m *Manager) snapshot(jobID int64) storage.TorrentJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessionByID(jobID)
	if ok {
		return session.job
	}
	return storage.TorrentJob{ID: jobID}
}
