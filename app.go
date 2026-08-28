package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/discordrpc"
	"github.com/sunnyegg/miru/internal/mpv"
	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/paths"
	"github.com/sunnyegg/miru/internal/secrets"
	"github.com/sunnyegg/miru/internal/storage"
	"github.com/sunnyegg/miru/internal/torrentx"
	"github.com/sunnyegg/miru/internal/update"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	apiCacheTTL         = 7 * 24 * time.Hour
	currentListCacheTTL = time.Hour
	watchingCacheKey    = "watching"
	completedCacheKey   = "completed"
)

func animeListCacheKey(status string) string {
	return "anilist:list:v2:" + strings.ToLower(strings.TrimSpace(status))
}

type App struct {
	ctx      context.Context
	initErr  error
	dirs     paths.Dirs
	store    *storage.Store
	tokens   secrets.Store
	player   *mpv.Player
	discord  *discordrpc.Client
	torrents *torrentx.Manager

	playMu sync.Mutex
	play   *playSession

	loginMu     sync.Mutex
	loginSrv    *http.Server
	loginCancel context.CancelFunc

	feedPollerMu sync.Mutex
	feedPoller   *feedPoller
}

type playSession struct {
	episodeID     int64
	anilistID     int
	episodeNum    int
	animeTitle    string
	synced        bool
	episodeMapped bool
	mapFailed     bool
	loggedAnilist bool
	lastProgress  mpv.Progress
}

func NewApp() *App {
	return &App{
		player:  &mpv.Player{},
		discord: discordrpc.New(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.init(); err != nil {
		a.initErr = err
		runtime.LogError(ctx, redactError(err))
	}
}

func (a *App) shutdown(_ context.Context) {
	a.feedPollerMu.Lock()
	if a.feedPoller != nil {
		a.feedPoller.stop()
		a.feedPoller = nil
	}
	a.feedPollerMu.Unlock()
	a.stopLoginServer()
	if a.discord != nil {
		a.discord.Clear()
	}
	if a.player != nil {
		a.player.Stop()
	}
	if a.torrents != nil {
		a.torrents.Close()
		a.torrents = nil
	}
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			a.logDebugErr("close database", err)
		}
		a.store = nil
	}
}

func (a *App) init() error {
	if exe, err := os.Executable(); err == nil {
		update.CleanupOld(exe)
	}
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
	a.torrents.SetCallbacks(a.emitTorrent, a.ingestTorrentFiles, a.logDebugErr)

	if err := store.RecoverInterruptedDownloads(); err != nil {
		return err
	}
	if err := a.ensureDefaults(); err != nil {
		return err
	}
	if err := a.configureTorrents(); err != nil {
		return err
	}
	a.startFeedPoller()
	return nil
}

func (a *App) configureTorrents() error {
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	limits := torrentRateLimits(settings)
	a.torrents.SetQueueConfig(limits, networkConfig(settings))
	a.torrents.ApplyRateLimits(limits)
	a.torrents.SetMaxConcurrent(settings.MaxConcurrentDownloads)
	a.torrents.SetSeedRatio(settings.SeedRatio)
	a.torrents.PumpQueue()
	return nil
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
	if a.settingMissing("max_concurrent_downloads") {
		if err := a.store.SetSetting("max_concurrent_downloads", "1"); err != nil {
			return err
		}
	}
	if a.settingMissing("seed_ratio") {
		if err := a.store.SetSetting("seed_ratio", "0.5"); err != nil {
			return err
		}
	}
	if a.settingMissing("rss_poll_interval_minutes") {
		if err := a.store.SetSetting("rss_poll_interval_minutes", "30"); err != nil {
			return err
		}
	}
	if !a.settingMissing("mpv_path") {
		return nil
	}
	detected, err := mpv.Detect("")
	if err != nil {
		a.logDebugErr("detect mpv", err)
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
		Mode:     settings.NetworkMode,
		Address:  settings.Socks5Address,
		ProxyURL: settings.HttpProxyURL,
	}).HTTPClient()
}

func (a *App) newAnilist(token string) (*anilist.Client, error) {
	httpClient, err := a.networkHTTPClient()
	if err != nil {
		return nil, err
	}
	return anilist.NewWithHTTP(token, httpClient), nil
}
