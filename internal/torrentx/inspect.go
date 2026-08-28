package torrentx

import (
	"errors"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/networking"

	"github.com/anacrolix/torrent"
)

func (m *Manager) Inspect(source, destDir string, limits RateLimits, networkConfig networking.Config) (ContentsView, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return ContentsView{}, errors.New("empty torrent source")
	}
	if strings.HasPrefix(strings.ToLower(source), "magnet:") {
		return m.inspectMagnet(source, destDir, limits, networkConfig)
	}

	sourceHTTP, err := networkConfig.HTTPClient()
	if err != nil {
		return ContentsView{}, err
	}
	metainfoFile, err := loadMetaInfo(source, sourceHTTP)
	if err != nil {
		return ContentsView{}, err
	}
	info, err := metainfoFile.UnmarshalInfo()
	if err != nil {
		return ContentsView{}, err
	}
	return contentsFromInfo(&info), nil
}

func (m *Manager) inspectMagnet(source, destDir string, limits RateLimits, networkConfig networking.Config) (ContentsView, error) {
	if destDir == "" {
		return ContentsView{}, errors.New("download folder is empty")
	}
	m.rememberConfig(limits, networkConfig)

	client, err := m.ensureClient(destDir, limits, networkConfig)
	if err != nil {
		return ContentsView{}, err
	}
	sourceHTTP, err := networkConfig.HTTPClient()
	if err != nil {
		return ContentsView{}, err
	}
	torrentHandle, err := addSource(client, source, sourceHTTP)
	if err != nil {
		return ContentsView{}, err
	}

	select {
	case <-torrentHandle.GotInfo():
	case <-time.After(2 * time.Minute):
		m.dropIfUnowned(torrentHandle)
		return ContentsView{}, errors.New("timed out waiting for torrent metadata")
	}

	contents := contentsFromTorrent(torrentHandle)
	m.dropIfUnowned(torrentHandle)
	return contents, nil
}

func (m *Manager) dropIfUnowned(torrentHandle *torrent.Torrent) {
	if m.sessionOwnsTorrent(torrentHandle) {
		return
	}
	torrentHandle.Drop()
}

func (m *Manager) sessionOwnsTorrent(torrentHandle *torrent.Torrent) bool {
	if torrentHandle == nil {
		return false
	}
	hash := torrentHandle.InfoHash()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, session := range m.sessions {
		if session.torrent == torrentHandle {
			return true
		}
		if session.torrent == nil {
			continue
		}
		if session.torrent.InfoHash() == hash {
			return true
		}
	}
	return false
}
