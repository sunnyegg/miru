package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/mpv"
	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/paths"
	"github.com/sunnyegg/miru/internal/secrets"
	"github.com/sunnyegg/miru/internal/storage"
	"github.com/sunnyegg/miru/internal/torrentx"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	apiCacheTTL      = 7 * 24 * time.Hour
	watchingCacheKey = "watching"
)

type App struct {
	ctx      context.Context
	initErr  error
	dirs     paths.Dirs
	store    *storage.Store
	tokens   secrets.Store
	player   *mpv.Player
	torrents *torrentx.Manager

	playMu sync.Mutex
	play   *playSession

	loginMu     sync.Mutex
	loginSrv    *http.Server
	loginCancel context.CancelFunc
}

type playSession struct {
	episodeID     int64
	anilistID     int
	episodeNum    int
	synced        bool
	episodeMapped bool
	mapFailed     bool
	lastProgress  mpv.Progress
}

func NewApp() *App {
	return &App{
		player: &mpv.Player{},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.init(); err != nil {
		a.initErr = err
		runtime.LogError(ctx, err.Error())
	}
}

func (a *App) shutdown(_ context.Context) {
	a.stopLoginServer()
	if a.player != nil {
		a.player.Stop()
	}
	if a.torrents != nil {
		a.torrents.Close()
	}
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) init() error {
	loadDotEnv()
	dirs, err := paths.Resolve()
	if err != nil {
		return fmt.Errorf("resolve dirs: %w", err)
	}
	a.dirs = dirs

	store, err := storage.Open(dirs.DatabaseFile())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	a.store = store
	a.tokens = secrets.New(dirs.TokenFile())
	a.torrents = torrentx.NewManager(store)
	a.torrents.SetCallbacks(a.emitTorrent, a.ingestTorrentFiles)

	if err := store.FailInterruptedDownloads(); err != nil {
		return err
	}
	return a.ensureDefaults()
}

func (a *App) ensureDefaults() error {
	if a.settingMissing("sync_threshold") {
		if err := a.store.SetSetting("sync_threshold", "85"); err != nil {
			return err
		}
	}
	if a.settingMissing("download_dir") {
		dir, err := paths.DefaultDownloadDir()
		if err != nil {
			return err
		}
		if err := a.store.SetSetting("download_dir", dir); err != nil {
			return err
		}
	}
	if !a.settingMissing("mpv_path") {
		return nil
	}
	detected, err := mpv.Detect("")
	if err != nil {
		return nil
	}
	return a.store.SetSetting("mpv_path", detected)
}

func (a *App) settingMissing(key string) bool {
	_, err := a.store.GetSetting(key)
	return errors.Is(err, storage.ErrNotFound)
}

func (a *App) ready() error {
	if a.initErr != nil {
		return a.initErr
	}
	return nil
}

func (a *App) InitError() string {
	if a.initErr == nil {
		return ""
	}
	return a.initErr.Error()
}

func (a *App) networkHTTPClient() (*http.Client, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return nil, err
	}
	return (networking.Config{
		Mode:    settings.NetworkMode,
		Address: settings.Socks5Address,
	}).HTTPClient()
}

func (a *App) anilistClient() (*anilist.Client, error) {
	token, _ := a.tokens.Get()
	httpClient, err := a.networkHTTPClient()
	if err != nil {
		return nil, err
	}
	return anilist.NewWithHTTP(token, httpClient), nil
}

func (a *App) newAnilist(token string) (*anilist.Client, error) {
	httpClient, err := a.networkHTTPClient()
	if err != nil {
		return nil, err
	}
	return anilist.NewWithHTTP(token, httpClient), nil
}
