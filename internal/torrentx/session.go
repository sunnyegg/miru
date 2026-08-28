package torrentx

import (
	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/anacrolix/torrent"
)

type session struct {
	torrent      *torrent.Torrent
	job          storage.TorrentJob
	pausedFrom   string
	uploadOffset int64
}

func (m *Manager) sessionByID(jobID int64) (*session, bool) {
	session, ok := m.sessions[jobID]
	return session, ok
}

func (m *Manager) downloadingCountLocked() int {
	count := 0
	for _, session := range m.sessions {
		if session.job.Status == "DOWNLOADING" {
			count++
		}
	}
	return count
}

func (m *Manager) rememberConfig(limits RateLimits, networkConfig networking.Config) {
	m.limits = limits
	m.networkConfig = networkConfig
}

func (m *Manager) emitProgress(view JobView) {
	if m.onProgress != nil {
		m.onProgress(view)
	}
}

func (m *Manager) removeSessionLocked(jobID int64) *torrent.Torrent {
	session, ok := m.sessions[jobID]
	if !ok {
		return nil
	}
	delete(m.sessions, jobID)
	return session.torrent
}
