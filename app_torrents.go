package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/nyaa"
	"github.com/sunnyegg/miru/internal/storage"
	"github.com/sunnyegg/miru/internal/tokyotosho"
	"github.com/sunnyegg/miru/internal/torrentx"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) StartMagnet(magnet string) error {
	return a.startTorrent(magnet)
}

func (a *App) StartTorrentURL(source string) error {
	return a.startTorrent(source)
}

func (a *App) SearchNyaa(query string) ([]NyaaResultView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("nyaa search query is empty")
	}
	if len(query) > 200 {
		return nil, errors.New("nyaa search query is too long")
	}
	httpClient, err := a.networkHTTPClient()
	if err != nil {
		return nil, err
	}
	results, err := nyaa.NewWithHTTP(httpClient).Search(query)
	if err != nil {
		return nil, err
	}
	out := make([]NyaaResultView, 0, len(results))
	for _, result := range results {
		out = append(out, NyaaResultView{
			Title:     result.Title,
			Link:      result.Link,
			Magnet:    result.Magnet(),
			Published: result.Published.Format(time.RFC3339),
			Size:      result.Size,
			Seeders:   result.Seeders,
			Leechers:  result.Leechers,
			Downloads: result.Downloads,
			Trusted:   result.Trusted,
			Remake:    result.Remake,
		})
	}
	return out, nil
}

func (a *App) SearchTokyoToshokan(query string) ([]NyaaResultView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("tokyo toshokan search query is empty")
	}
	if len(query) > 200 {
		return nil, errors.New("tokyo toshokan search query is too long")
	}
	httpClient, err := a.networkHTTPClient()
	if err != nil {
		return nil, err
	}
	results, err := tokyotosho.NewWithHTTP(httpClient).Search(query)
	if err != nil {
		return nil, err
	}
	out := make([]NyaaResultView, 0, len(results))
	for _, result := range results {
		out = append(out, NyaaResultView{
			Title:     result.Title,
			Link:      result.Link,
			Magnet:    result.Magnet,
			Published: result.Published.Format(time.RFC3339),
			Size:      result.Size,
			Seeders:   result.Seeders,
			Leechers:  result.Leechers,
			Downloads: result.Downloads,
			Trusted:   result.Trusted,
			Remake:    result.Remake,
		})
	}
	return out, nil
}

func (a *App) startTorrent(source string) error {
	if err := a.ready(); err != nil {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	return a.torrents.Start(source, settings.DownloadDir, torrentRateLimits(settings), networkConfig(settings))
}

func (a *App) StartTorrentFile() error {
	if err := a.ready(); err != nil {
		return err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open torrent file",
		Filters: []runtime.FileFilter{{
			DisplayName: "Torrent",
			Pattern:     "*.torrent",
		}},
	})
	if err != nil || path == "" {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	return a.torrents.Start(path, settings.DownloadDir, torrentRateLimits(settings), networkConfig(settings))
}

func (a *App) DownloadStatus() (*torrentx.JobView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	view, err := a.torrents.Status()
	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &view, nil
}

func (a *App) CancelDownload(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Cancel(id)
}

func (a *App) PauseDownload(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Pause(id)
}

func (a *App) ResumeDownload(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Resume(id)
}

func (a *App) ResumeSeeding(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	return a.torrents.ResumeSeeding(id, torrentRateLimits(settings), networkConfig(settings))
}

func (a *App) FinishDownload(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Finish(id)
}

func (a *App) DownloadHistory() ([]torrentx.JobView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.torrents.History()
}

func (a *App) RemoveDownload(id int64, deleteFiles bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	job, err := a.store.TorrentJobByID(id)
	if err != nil {
		return err
	}
	if err := a.torrents.Remove(id, deleteFiles); err != nil {
		return err
	}
	if !deleteFiles {
		return nil
	}

	destDir := filepath.Clean(job.DestDir)
	prefix := filepath.Clean(filepath.Join(destDir, job.Name))
	relative, relErr := filepath.Rel(destDir, prefix)
	if relErr != nil || !filepath.IsLocal(relative) {
		return nil
	}
	if err := a.store.DeleteEpisodesByFilePrefix(prefix); err != nil {
		return err
	}
	runtime.EventsEmit(a.ctx, "library:changed", true)
	return nil
}

func (a *App) OpenDownloadFolder() error {
	if err := a.ready(); err != nil {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	if settings.DownloadDir == "" {
		return errors.New("download folder is empty")
	}
	runtime.BrowserOpenURL(a.ctx, "file://"+settings.DownloadDir)
	return nil
}

func (a *App) ingestTorrentFiles(files []string) {
	settings, err := a.loadSettings()
	if err != nil {
		a.logDebugErr("ingest torrent settings", err)
		return
	}
	for _, rel := range files {
		path := torrentx.ResolveDataPath(settings.DownloadDir, rel)
		if _, err := a.importPath(path); err != nil {
			a.logDebugErr("ingest torrent file", err)
		}
	}
	runtime.EventsEmit(a.ctx, "library:changed", true)
}

func (a *App) emitTorrent(view torrentx.JobView) {
	if view.Status == "FAILED" && view.Error != "" {
		a.logErr(fmt.Sprintf("torrent job %d failed", view.ID), errors.New(view.Error))
	}
	runtime.EventsEmit(a.ctx, "torrent:progress", view)
}

func networkConfig(settings SettingsView) networking.Config {
	return networking.Config{
		Mode:    settings.NetworkMode,
		Address: settings.Socks5Address,
	}
}

func torrentRateLimits(settings SettingsView) torrentx.RateLimits {
	return torrentx.RateLimits{
		Download: normalizeRateLimit(settings.DownloadRateLimit),
		Upload:   normalizeRateLimit(settings.UploadRateLimit),
	}
}
