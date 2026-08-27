package torrentx

import (
	"errors"
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

	for range ticker.C {
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
		m.job.BytesUploaded = m.uploadOffset + uploaded
		if info != nil && m.job.Name == "" {
			m.job.Name = name
		}
		view := liveView(m.job)
		view.SpeedBytesPerSecond = speed
		view.UploadSpeedBytesPerSecond = uploadSpeed
		cb := m.onProgress
		downloadDone := m.job.Status == "DOWNLOADING" && completed >= total && total > 0
		seedingDone := m.job.Status == "SEEDING" && seedingComplete(m.job.BytesUploaded, total)
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
		cb(liveView(job))
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
		m.job.BytesUploaded = m.uploadOffset + uploadedBytes(t)
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
