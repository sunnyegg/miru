package torrentx

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/storage"

	"github.com/anacrolix/torrent"
)

var videoExt = map[string]bool{
	".mkv":  true,
	".mp4":  true,
	".avi":  true,
	".webm": true,
	".mov":  true,
	".m4v":  true,
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
	hash := t.InfoHash().HexString()
	m.mu.Lock()
	session, ok := m.sessionByID(jobID)
	status := ""
	selectedFiles := []FileView{}
	if ok {
		status = session.job.Status
		selectedFiles = decodeFiles(session.job.FilesJSON)
	}
	m.mu.Unlock()
	if !ok {
		return
	}

	total := selectedBytesTotal(selectedFiles)
	if total == 0 {
		total = t.Length()
	}
	if status == "DOWNLOADING" {
		selectedTotal, err := applyFileSelection(t, selectedFiles)
		if err != nil {
			m.fail(jobID, err)
			return
		}
		if selectedTotal > 0 {
			total = selectedTotal
		}
	}

	if len(selectedFiles) == 0 {
		selectedFiles = fileViewsFromTorrent(t, nil)
	}

	m.update(jobID, func(job *storage.TorrentJob) {
		job.Name = name
		job.BytesTotal = total
		job.InfoHash = hash
		if job.FilesJSON == "" {
			job.FilesJSON = encodeFiles(selectedFiles)
		}
	})

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	lastBytes := t.BytesCompleted()
	lastUploaded := uploadedBytes(t)
	lastAt := time.Now()

	for range ticker.C {
		m.mu.Lock()
		session, ok = m.sessionByID(jobID)
		if !ok || session.torrent != t || !isActiveStatus(session.job.Status) {
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
		session.job.BytesCompleted = completed
		session.job.BytesTotal = total
		session.job.BytesUploaded = session.uploadOffset + uploaded
		if info != nil && session.job.Name == "" {
			session.job.Name = name
		}
		view := liveViewWithTorrent(session.job, t)
		view.SpeedBytesPerSecond = speed
		view.UploadSpeedBytesPerSecond = uploadSpeed
		if view.BytesTotal > 0 {
			total = view.BytesTotal
			completed = view.BytesCompleted
			session.job.BytesCompleted = completed
			session.job.BytesTotal = total
		}
		downloadDone := session.job.Status == "DOWNLOADING" && completed >= total && total > 0
		seedingDone := session.job.Status == "SEEDING" && seedingComplete(session.job.BytesUploaded, total, m.seedRatioLocked())
		m.mu.Unlock()

		m.persistProgress(jobID, m.snapshot(jobID))
		m.emitProgress(view)
		if downloadDone {
			m.startSeeding(jobID)
			m.PumpQueue()
			continue
		}
		if seedingDone {
			m.finish(t, jobID)
			return
		}
	}
}

func seedingComplete(uploaded, total int64, ratio float64) bool {
	if total <= 0 {
		return false
	}
	if ratio <= 0 {
		return true
	}
	required := int64(float64(total) * ratio)
	return uploaded >= required
}

func (m *Manager) startSeeding(jobID int64) {
	m.mu.Lock()
	session, ok := m.sessionByID(jobID)
	if !ok {
		m.mu.Unlock()
		return
	}
	if session.job.Status == "DOWNLOADING" {
		session.job.Status = "SEEDING"
	}
	job := session.job
	m.mu.Unlock()
	m.persistJob(job, "torrent persist seeding")
	m.emitProgress(liveView(job))
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
	selected := decodeFiles(m.snapshot(jobID).FilesJSON)
	files := videoFiles(t, selected)

	m.mu.Lock()
	session, ok := m.sessionByID(jobID)
	if ok {
		session.job.Status = "COMPLETED"
		session.job.BytesCompleted = t.BytesCompleted()
		session.job.BytesUploaded = session.uploadOffset + uploadedBytes(t)
		session.job.Error = ""
		delete(m.sessions, jobID)
	}
	job := storage.TorrentJob{ID: jobID}
	if ok {
		job = session.job
	}
	complete := m.onComplete
	m.mu.Unlock()

	m.persistJob(job, "torrent persist complete")
	m.clearPersistOnce(jobID)
	m.emitProgress(ToView(job))
	if complete != nil {
		complete(files)
	}
	m.PumpQueue()
}

func (m *Manager) fail(jobID int64, err error) {
	m.reportError(fmt.Sprintf("torrent job %d failed", jobID), err)
	m.mu.Lock()
	session, ok := m.sessionByID(jobID)
	var torrentHandle *torrent.Torrent
	if ok {
		torrentHandle = session.torrent
		session.job.Status = "FAILED"
		session.job.Error = err.Error()
		job := session.job
		delete(m.sessions, jobID)
		m.mu.Unlock()
		if torrentHandle != nil {
			torrentHandle.Drop()
		}
		m.persistJob(job, "torrent persist failed job")
		m.clearPersistOnce(jobID)
		m.emitProgress(ToView(job))
		m.PumpQueue()
		return
	}
	m.mu.Unlock()
}

func videoFiles(t *torrent.Torrent, selected []FileView) []string {
	wanted := selectedPathSet(selected)
	var out []string
	for _, f := range t.Files() {
		if !fileWanted(f, wanted) {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.DisplayPath()))
		if !videoExt[ext] {
			continue
		}
		out = append(out, f.Path())
	}
	return out
}

func videoFilesFromDisk(destDir, name string) []string {
	destDir = filepath.Clean(destDir)
	root := filepath.Clean(filepath.Join(destDir, name))
	relative, err := filepath.Rel(destDir, root)
	if name == "" || name == "Magnet download" || err != nil || !filepath.IsLocal(relative) {
		return nil
	}

	var out []string
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if videoExt[strings.ToLower(filepath.Ext(path))] {
			out = append(out, path)
		}
		return nil
	})
	return out
}
