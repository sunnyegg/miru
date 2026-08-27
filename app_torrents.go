package main

import (
	"errors"

	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/storage"
	"github.com/sunnyegg/miru/internal/torrentx"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) StartMagnet(magnet string) error {
	return a.startTorrent(magnet)
}

func (a *App) StartTorrentURL(source string) error {
	return a.startTorrent(source)
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

func (a *App) CancelDownload() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Cancel()
}

func (a *App) PauseDownload() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Pause()
}

func (a *App) ResumeDownload() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.torrents.Resume()
}

func (a *App) DownloadHistory() ([]torrentx.JobView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.torrents.History()
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
		return
	}
	for _, rel := range files {
		path := torrentx.ResolveDataPath(settings.DownloadDir, rel)
		if _, err := a.importPath(path); err != nil {
			runtime.LogError(a.ctx, err.Error())
		}
	}
	runtime.EventsEmit(a.ctx, "library:changed", true)
}

func (a *App) emitTorrent(view torrentx.JobView) {
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
