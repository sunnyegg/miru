package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/media"
	"github.com/sunnyegg/miru/internal/mpv"
	"github.com/sunnyegg/miru/internal/paths"
	"github.com/sunnyegg/miru/internal/secrets"
	"github.com/sunnyegg/miru/internal/storage"
	syncprogress "github.com/sunnyegg/miru/internal/syncprogress"
	"github.com/sunnyegg/miru/internal/torrentx"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	episodeID  int64
	anilistID  int
	episodeNum int
	synced     bool
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
	if _, err := a.store.GetSetting("sync_threshold"); errors.Is(err, storage.ErrNotFound) {
		if err := a.store.SetSetting("sync_threshold", "85"); err != nil {
			return err
		}
	}
	if _, err := a.store.GetSetting("download_dir"); errors.Is(err, storage.ErrNotFound) {
		dir, err := paths.DefaultDownloadDir()
		if err != nil {
			return err
		}
		if err := a.store.SetSetting("download_dir", dir); err != nil {
			return err
		}
	}
	if _, err := a.store.GetSetting("mpv_path"); errors.Is(err, storage.ErrNotFound) {
		if detected, err := mpv.Detect(""); err == nil {
			if err := a.store.SetSetting("mpv_path", detected); err != nil {
				return err
			}
		}
	}
	return nil
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

func (a *App) GetSettings() (SettingsView, error) {
	if err := a.ready(); err != nil {
		return SettingsView{}, err
	}
	return a.loadSettings()
}

func (a *App) SaveSettings(view SettingsView) error {
	if err := a.ready(); err != nil {
		return err
	}
	threshold := view.SyncThreshold
	if threshold <= 0 || threshold > 100 {
		threshold = 85
	}
	view.DownloadRateLimit = normalizeRateLimit(view.DownloadRateLimit)
	view.UploadRateLimit = normalizeRateLimit(view.UploadRateLimit)
	pairs := map[string]string{
		"mpv_path":            strings.TrimSpace(view.MpvPath),
		"download_dir":        strings.TrimSpace(view.DownloadDir),
		"sync_threshold":      formatFloat(threshold),
		"anilist_client_id":   strings.TrimSpace(view.AnilistClientId),
		"download_rate_limit": formatInt64(view.DownloadRateLimit),
		"upload_rate_limit":   formatInt64(view.UploadRateLimit),
	}
	for key, value := range pairs {
		if err := a.store.SetSetting(key, value); err != nil {
			return err
		}
	}
	if a.torrents != nil {
		a.torrents.ApplyRateLimits(torrentx.RateLimits{
			Download: normalizeRateLimit(view.DownloadRateLimit),
			Upload:   normalizeRateLimit(view.UploadRateLimit),
		})
	}
	return nil
}

func (a *App) PickMpvPath() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select MPV binary",
	})
	return path, err
}

func (a *App) PickDownloadDir() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select download folder",
	})
}

func (a *App) DetectMpv() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return "", err
	}
	return mpv.Detect(settings.MpvPath)
}

func (a *App) ListEpisodes() ([]EpisodeView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	rows, err := a.store.ListEpisodes()
	if err != nil {
		return nil, err
	}
	out := make([]EpisodeView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toEpisodeView(row))
	}
	return out, nil
}

func (a *App) ImportLocalFile() (ImportResult, error) {
	if err := a.ready(); err != nil {
		return ImportResult{}, err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Import video",
		Filters: []runtime.FileFilter{{
			DisplayName: "Video",
			Pattern:     "*.mkv;*.mp4;*.avi;*.webm;*.mov;*.m4v",
		}},
	})
	if err != nil || path == "" {
		return ImportResult{}, err
	}
	return a.importPath(path)
}

func (a *App) SearchAnime(query string) ([]AnimeView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	client := anilist.New("")
	results, err := client.Search(query)
	if err != nil {
		return nil, err
	}
	return toAnimeViews(results), nil
}

func (a *App) BindEpisode(episodeID int64, anilistID int) error {
	if err := a.ready(); err != nil {
		return err
	}
	ep, err := a.store.GetEpisode(episodeID)
	if err != nil {
		return err
	}
	client := a.anilistClient()
	anime, err := client.GetAnime(anilistID)
	if err != nil {
		return err
	}
	if err := a.store.UpsertAnime(toStoredAnime(anime)); err != nil {
		return err
	}
	episodeNum := int(ep.EpisodeNumber.Int64)
	if !ep.EpisodeNumber.Valid {
		parsed := media.ParseFilename(ep.FilePath)
		if parsed.HasEpisode {
			episodeNum = parsed.Episode
		}
	}
	return a.store.BindEpisode(episodeID, anilistID, episodeNum)
}

func (a *App) PlayEpisode(episodeID int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	ep, err := a.store.GetEpisode(episodeID)
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(ep.FilePath); statErr != nil {
		return fmt.Errorf("file missing: %s", ep.FilePath)
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	mpvPath, err := mpv.Detect(settings.MpvPath)
	if err != nil {
		return err
	}

	session := &playSession{episodeID: episodeID}
	if ep.AnilistID.Valid {
		session.anilistID = int(ep.AnilistID.Int64)
	}
	if ep.EpisodeNumber.Valid {
		session.episodeNum = int(ep.EpisodeNumber.Int64)
	}

	a.playMu.Lock()
	a.play = session
	a.playMu.Unlock()

	return a.player.Play(mpvPath, ep.FilePath, func(p mpv.Progress) {
		runtime.EventsEmit(a.ctx, "mpv:progress", PlaybackEvent{
			EpisodeID: episodeID,
			Percent:   p.Percent,
		})
		a.maybeSync(session, p.Percent, settings.SyncThreshold)
	}, func(exitErr error) {
		msg := ""
		if exitErr != nil && !strings.Contains(exitErr.Error(), "signal: killed") {
			msg = exitErr.Error()
		}
		runtime.EventsEmit(a.ctx, "mpv:ended", SyncEvent{EpisodeID: episodeID, OK: true, Message: msg})
	})
}

func (a *App) AnilistStatus() (AnilistStatus, error) {
	if err := a.ready(); err != nil {
		return AnilistStatus{}, err
	}
	token, err := a.tokens.Get()
	if err != nil {
		return AnilistStatus{Connected: false}, nil
	}
	name, err := anilist.New(token).ViewerName()
	if err != nil {
		return AnilistStatus{Connected: false}, nil
	}
	return AnilistStatus{Connected: true, Username: name}, nil
}

func (a *App) OpenAnilistLogin() error {
	if err := a.ready(); err != nil {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	loginURL, err := anilist.LoginURL(settings.AnilistClientId)
	if err != nil {
		return err
	}
	if err := a.startLoginServer(); err != nil {
		return err
	}
	runtime.BrowserOpenURL(a.ctx, loginURL)
	return nil
}

func (a *App) startLoginServer() error {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()
	if a.loginSrv != nil {
		return nil
	}

	ln, err := net.Listen("tcp", anilist.ListenAddr)
	if err != nil {
		return fmt.Errorf("login port %s is in use; paste the token in Settings instead: %w", anilist.ListenAddr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	settings, err := a.loadSettings()
	if err != nil {
		_ = ln.Close()
		return err
	}
	secret := envTrim("ANILIST_CLIENT_SECRET")
	if secret == "" {
		secret, _ = a.store.GetSetting("anilist_client_secret")
	}
	clientID := settings.AnilistClientId
	mux := anilist.NewMux(anilist.MuxConfig{
		ExchangeCode: func(code string) (string, error) {
			return anilist.ExchangeCode(nil, anilist.TokenURL, clientID, secret, code)
		},
		OnToken: func(token string) error {
			if err := a.SaveAnilistToken(token); err != nil {
				return err
			}
			runtime.EventsEmit(a.ctx, "anilist:connected", true)
			cancel()
			return nil
		},
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	a.loginSrv = srv
	a.loginCancel = cancel

	go func() {
		_ = srv.Serve(ln)
	}()
	go func() {
		<-ctx.Done()
		a.stopLoginServer()
	}()
	return nil
}

func (a *App) stopLoginServer() {
	a.loginMu.Lock()
	srv := a.loginSrv
	cancel := a.loginCancel
	a.loginSrv = nil
	a.loginCancel = nil
	a.loginMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if srv != nil {
		ctx, done := context.WithTimeout(context.Background(), 2*time.Second)
		defer done()
		_ = srv.Shutdown(ctx)
	}
}

func (a *App) SaveAnilistToken(token string) error {
	if err := a.ready(); err != nil {
		return err
	}
	token = strings.TrimSpace(token)
	name, err := anilist.New(token).ViewerName()
	if err != nil {
		return err
	}
	if err := a.tokens.Set(token); err != nil {
		return err
	}
	runtime.LogInfo(a.ctx, "AniList connected as "+name)
	return nil
}

func (a *App) LogoutAnilist() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.tokens.Delete()
}

func (a *App) StartMagnet(magnet string) error {
	if err := a.ready(); err != nil {
		return err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	return a.torrents.Start(magnet, settings.DownloadDir, torrentRateLimits(settings))
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
	return a.torrents.Start(path, settings.DownloadDir, torrentRateLimits(settings))
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

func (a *App) loadSettings() (SettingsView, error) {
	view := SettingsView{SyncThreshold: 85}
	view.MpvPath, _ = a.store.GetSetting("mpv_path")
	view.DownloadDir, _ = a.store.GetSetting("download_dir")
	view.AnilistClientId, _ = a.store.GetSetting("anilist_client_id")
	if strings.TrimSpace(view.AnilistClientId) == "" {
		view.AnilistClientId = envTrim("ANILIST_CLIENT_ID")
	}
	if raw, err := a.store.GetSetting("sync_threshold"); err == nil {
		if n, err := strconv.ParseFloat(raw, 64); err == nil {
			view.SyncThreshold = n
		}
	}
	view.DownloadRateLimit = settingInt64(a.store, "download_rate_limit")
	view.UploadRateLimit = settingInt64(a.store, "upload_rate_limit")
	return view, nil
}

func (a *App) importPath(path string) (ImportResult, error) {
	if existing, err := a.store.EpisodeByPath(path); err == nil {
		return ImportResult{Episode: toEpisodeView(existing)}, nil
	}

	parsed := media.ParseFilename(path)
	ep := storage.Episode{
		FilePath:     path,
		DisplayTitle: media.DisplayTitle(parsed),
		Status:       "COMPLETED",
	}
	if parsed.HasEpisode {
		ep.EpisodeNumber = sql.NullInt64{Int64: int64(parsed.Episode), Valid: true}
	}

	candidates := []AnimeView{}
	autoBound := false
	if parsed.Title != "" {
		found, err := anilist.New("").Search(parsed.Title)
		if err != nil {
			runtime.LogError(a.ctx, err.Error())
		} else {
			candidates = toAnimeViews(found)
			if len(found) == 1 {
				if err := a.store.UpsertAnime(toStoredAnime(found[0])); err != nil {
					return ImportResult{}, err
				}
				ep.AnilistID = sql.NullInt64{Int64: int64(found[0].ID), Valid: true}
				autoBound = true
			}
		}
	}

	id, err := a.store.InsertEpisode(ep)
	if err != nil {
		return ImportResult{}, err
	}
	saved, err := a.store.GetEpisode(id)
	if err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		Episode:    toEpisodeView(saved),
		Candidates: candidates,
		AutoBound:  autoBound,
	}, nil
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

func (a *App) maybeSync(session *playSession, percent, threshold float64) {
	if session == nil || session.anilistID == 0 || session.episodeNum == 0 {
		return
	}

	a.playMu.Lock()
	if session.synced {
		a.playMu.Unlock()
		return
	}
	a.playMu.Unlock()

	synced, err := a.store.HasSynced(session.anilistID, session.episodeNum)
	if err != nil || synced {
		return
	}

	token, err := a.tokens.Get()
	if err != nil {
		return
	}
	client := anilist.New(token)
	current, err := client.ListProgress(session.anilistID)
	if err != nil {
		runtime.EventsEmit(a.ctx, "sync:result", SyncEvent{
			EpisodeID: session.episodeID,
			OK:        false,
			Message:   err.Error(),
		})
		return
	}
	if !syncprogress.ShouldSync(percent, threshold, session.episodeNum, current, false) {
		return
	}

	a.playMu.Lock()
	if session.synced {
		a.playMu.Unlock()
		return
	}
	session.synced = true
	a.playMu.Unlock()

	if err := client.SaveProgress(session.anilistID, session.episodeNum); err != nil {
		a.playMu.Lock()
		session.synced = false
		a.playMu.Unlock()
		runtime.EventsEmit(a.ctx, "sync:result", SyncEvent{
			EpisodeID: session.episodeID,
			OK:        false,
			Message:   err.Error(),
		})
		return
	}
	_ = a.store.RecordSync(session.anilistID, session.episodeNum)
	a.playMu.Lock()
	session.synced = true
	a.playMu.Unlock()
	runtime.EventsEmit(a.ctx, "sync:result", SyncEvent{
		EpisodeID: session.episodeID,
		OK:        true,
		Message:   "AniList updated to episode " + strconv.Itoa(session.episodeNum),
	})
}

func (a *App) anilistClient() *anilist.Client {
	token, _ := a.tokens.Get()
	return anilist.New(token)
}

func toEpisodeView(e storage.Episode) EpisodeView {
	title := e.DisplayTitle
	if e.TitleEnglish != "" {
		title = e.TitleEnglish
	} else if e.TitleRomaji != "" {
		title = e.TitleRomaji
	}
	view := EpisodeView{
		ID:           e.ID,
		FilePath:     e.FilePath,
		DisplayTitle: e.DisplayTitle,
		Status:       e.Status,
		AnimeTitle:   title,
		CoverImage:   e.CoverImage,
		Bound:        e.AnilistID.Valid,
	}
	if e.AnilistID.Valid {
		view.AnilistID = int(e.AnilistID.Int64)
	}
	if e.EpisodeNumber.Valid {
		view.EpisodeNumber = int(e.EpisodeNumber.Int64)
	}
	if view.DisplayTitle == "" {
		view.DisplayTitle = filepath.Base(e.FilePath)
	}
	return view
}

func toAnimeViews(in []anilist.Anime) []AnimeView {
	out := make([]AnimeView, 0, len(in))
	for _, a := range in {
		out = append(out, AnimeView{
			ID:            a.ID,
			TitleRomaji:   a.TitleRomaji,
			TitleEnglish:  a.TitleEnglish,
			CoverImage:    a.CoverImage,
			TotalEpisodes: a.TotalEpisodes,
			Status:        a.Status,
			Synopsis:      a.Synopsis,
		})
	}
	return out
}

func toStoredAnime(a anilist.Anime) storage.Anime {
	return storage.Anime{
		AnilistID:     a.ID,
		TitleRomaji:   a.TitleRomaji,
		TitleEnglish:  a.TitleEnglish,
		CoverImage:    a.CoverImage,
		TotalEpisodes: a.TotalEpisodes,
		Status:        a.Status,
		Synopsis:      a.Synopsis,
	}
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatInt64(v int64) string {
	return strconv.FormatInt(normalizeRateLimit(v), 10)
}

func settingInt64(store *storage.Store, key string) int64 {
	raw, err := store.GetSetting(key)
	if err != nil {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return normalizeRateLimit(value)
}

func normalizeRateLimit(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func torrentRateLimits(settings SettingsView) torrentx.RateLimits {
	return torrentx.RateLimits{
		Download: normalizeRateLimit(settings.DownloadRateLimit),
		Upload:   normalizeRateLimit(settings.UploadRateLimit),
	}
}
