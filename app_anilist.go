package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) AnilistStatus() (AnilistStatus, error) {
	if err := a.ready(); err != nil {
		return AnilistStatus{}, err
	}
	token, err := a.tokens.Get()
	if err != nil {
		a.logDebugErr("anilist status token", err)
		return AnilistStatus{Connected: false}, nil
	}
	client, err := a.newAnilist(token)
	if err != nil {
		a.logDebugErr("anilist status client", err)
		return AnilistStatus{Connected: false}, nil
	}
	name, err := client.ViewerName()
	if err != nil {
		a.logDebugErr("anilist status viewer", err)
		return AnilistStatus{Connected: false}, nil
	}
	return AnilistStatus{Connected: true, Username: name}, nil
}

func (a *App) OpenAnilistLogin() error {
	if err := a.ready(); err != nil {
		return err
	}
	loginURL, err := anilist.LoginURL(a.anilistClientID())
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
		return fmt.Errorf("login port %s is in use; try again after closing the other process: %w", anilist.ListenAddr, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	secret := envTrim("ANILIST_CLIENT_SECRET")
	if secret == "" {
		secret, _ = a.store.GetSetting("anilist_client_secret")
	}
	clientID := a.anilistClientID()
	mux := anilist.NewMux(anilist.MuxConfig{
		ExchangeCode: func(code string) (string, error) {
			httpClient, err := a.networkHTTPClient()
			if err != nil {
				return "", err
			}
			return anilist.ExchangeCode(httpClient, anilist.TokenURL, clientID, secret, code)
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
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logDebugErr("anilist login server", err)
		}
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
		if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logDebugErr("anilist login shutdown", err)
		}
	}
}

func (a *App) SaveAnilistToken(token string) error {
	if err := a.ready(); err != nil {
		return err
	}
	token = strings.TrimSpace(token)
	client, err := a.newAnilist(token)
	if err != nil {
		return err
	}
	name, err := client.ViewerName()
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
	a.invalidateAnimeListCache()
	return a.tokens.Delete()
}

func (a *App) anilistClientID() string {
	id := envTrim("ANILIST_CLIENT_ID")
	if id != "" {
		return id
	}
	id, _ = a.store.GetSetting("anilist_client_id")
	return strings.TrimSpace(id)
}

func (a *App) ListAiringSchedule(start, end int64) ([]AiringScheduleView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	if start < 0 || end <= start || end-start > 8*24*60*60 {
		return nil, errors.New("invalid airing schedule range")
	}
	return loadCachedJSON(a, fmt.Sprintf("airing:%d:%d", start, end), apiCacheTTL, func() ([]AiringScheduleView, error) {
		client, err := a.newAnilist("")
		if err != nil {
			return nil, err
		}
		schedules, err := client.AiringSchedules(start, end)
		if err != nil {
			return nil, err
		}
		out := make([]AiringScheduleView, 0, len(schedules))
		for _, schedule := range schedules {
			out = append(out, AiringScheduleView{
				ID:           int64(schedule.ID),
				AiringAt:     schedule.AiringAt,
				Episode:      schedule.Episode,
				MediaID:      schedule.MediaID,
				TitleRomaji:  schedule.TitleRomaji,
				TitleEnglish: schedule.TitleEnglish,
				CoverImage:   schedule.CoverImage,
			})
		}
		return out, nil
	})
}

func (a *App) GetAnime(mediaID int) (AnimeView, error) {
	if err := a.ready(); err != nil {
		return AnimeView{}, err
	}
	if mediaID <= 0 {
		return AnimeView{}, errors.New("invalid anime id")
	}
	return loadCachedJSON(a, animeCacheKey(mediaID), apiCacheTTL, func() (AnimeView, error) {
		token, _ := a.tokens.Get()
		client, err := a.newAnilist(token)
		if err != nil {
			return AnimeView{}, err
		}
		anime, err := client.GetAnime(mediaID)
		if err != nil {
			return AnimeView{}, err
		}
		views := toAnimeViews([]anilist.Anime{anime})
		return views[0], nil
	})
}

func (a *App) ListCurrentlyWatching() ([]WatchingEntryView, error) {
	return a.ListAnimeList("CURRENT")
}

func (a *App) ListAnimeList(status string) ([]WatchingEntryView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	if !anilist.ValidListStatus(status) {
		return nil, fmt.Errorf("unsupported list status %q", status)
	}
	token, err := a.tokens.Get()
	if err != nil {
		return nil, errors.New("AniList not connected")
	}
	cacheKey := animeListCacheKey(status)
	cacheTTL := apiCacheTTL
	if status == "CURRENT" {
		cacheTTL = currentListCacheTTL
	}
	return loadCachedJSON(a, cacheKey, cacheTTL, func() ([]WatchingEntryView, error) {
		client, err := a.newAnilist(token)
		if err != nil {
			return nil, err
		}
		entries, err := client.ListMediaList(status)
		if err != nil {
			return nil, err
		}
		return toWatchingEntryViews(entries), nil
	})
}

func (a *App) ListAnimeListCounts() (map[string]int, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	token, err := a.tokens.Get()
	if err != nil {
		return nil, errors.New("AniList not connected")
	}
	return loadCachedJSON(a, animeListCountsCacheKey, currentListCacheTTL, func() (map[string]int, error) {
		client, err := a.newAnilist(token)
		if err != nil {
			return nil, err
		}
		return client.ListMediaListCounts()
	})
}

func (a *App) SetAnimeListStatus(mediaID int, status string, totalEpisodes int) error {
	if err := a.ready(); err != nil {
		return err
	}
	if mediaID <= 0 {
		return errors.New("invalid anime id")
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	switch status {
	case "CURRENT", "COMPLETED", "PLANNING":
	default:
		return fmt.Errorf("unsupported list status %q", status)
	}
	token, err := a.tokens.Get()
	if err != nil {
		return errors.New("AniList not connected")
	}
	client, err := a.newAnilist(token)
	if err != nil {
		return err
	}
	progress := -1
	if status == "COMPLETED" && totalEpisodes > 0 {
		progress = totalEpisodes
	}
	if err := client.SaveListStatus(mediaID, status, progress); err != nil {
		return err
	}
	a.invalidateAnimeListCache()
	a.invalidateAnimeCache(mediaID)
	return nil
}

func (a *App) SaveAnimeListEntry(input AnimeListEntryInput) error {
	if err := a.ready(); err != nil {
		return err
	}
	if input.MediaID <= 0 {
		return errors.New("invalid anime id")
	}
	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if status == "" {
		return errors.New("status is required")
	}
	token, err := a.tokens.Get()
	if err != nil {
		return errors.New("AniList not connected")
	}
	client, err := a.newAnilist(token)
	if err != nil {
		return err
	}
	save := anilist.ListEntrySave{
		MediaID:         input.MediaID,
		Status:          status,
		Progress:        input.Progress,
		ScoreRaw:        input.ScoreRaw,
		Notes:           input.Notes,
		SendNotes:       true,
		Repeat:          input.Repeat,
		Private:         input.Private,
		SendPrivate:     true,
		StartedAt:       anilist.FuzzyDate{Year: input.StartedYear, Month: input.StartedMonth, Day: input.StartedDay},
		SendStartedAt:   true,
		CompletedAt:     anilist.FuzzyDate{Year: input.CompletedYear, Month: input.CompletedMonth, Day: input.CompletedDay},
		SendCompletedAt: true,
	}
	if err := client.SaveListEntry(save); err != nil {
		return err
	}
	a.invalidateAnimeListCache()
	a.invalidateAnimeCache(input.MediaID)
	return nil
}

func (a *App) invalidateAnimeCache(mediaID int) {
	_ = a.store.DeleteAPICache(animeCacheKey(mediaID))
}

func (a *App) invalidateAnimeListCache() {
	for _, status := range anilist.ListStatuses {
		_ = a.store.DeleteAPICache(animeListCacheKey(status))
	}
	_ = a.store.DeleteAPICache(watchingCacheKey)
	_ = a.store.DeleteAPICache(completedCacheKey)
	_ = a.store.DeleteAPICache(animeListCountsCacheKey)
}

func toWatchingEntryViews(entries []anilist.CurrentEntry) []WatchingEntryView {
	out := make([]WatchingEntryView, 0, len(entries))
	for _, entry := range entries {
		out = append(out, WatchingEntryView{
			MediaID:           entry.MediaID,
			ListStatus:        entry.ListStatus,
			Progress:          entry.Progress,
			ScoreRaw:          entry.ScoreRaw,
			Notes:             entry.Notes,
			Repeat:            entry.Repeat,
			Private:           entry.Private,
			StartedAt:         FuzzyDateView{Year: entry.StartedAt.Year, Month: entry.StartedAt.Month, Day: entry.StartedAt.Day},
			CompletedAt:       FuzzyDateView{Year: entry.CompletedAt.Year, Month: entry.CompletedAt.Month, Day: entry.CompletedAt.Day},
			TitleRomaji:       entry.TitleRomaji,
			TitleEnglish:      entry.TitleEnglish,
			CoverImage:        entry.CoverImage,
			BannerImage:       entry.BannerImage,
			TotalEpisodes:     entry.TotalEpisodes,
			MediaStatus:       entry.MediaStatus,
			NextAiringEpisode: entry.NextAiringEpisode,
		})
	}
	return out
}

func loadCachedJSON[T any](a *App, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	if cached, ok := cachedJSON[T](a.store, key, ttl); ok {
		return cached, nil
	}

	result, err := fetch()
	if err != nil {
		if stale, ok := cachedJSON[T](a.store, key, 0); ok {
			a.logDebugErr("api cache stale fallback", err)
			return stale, nil
		}
		var zero T
		return zero, err
	}

	encoded, encodeErr := json.Marshal(result)
	if encodeErr != nil {
		a.logDebugErr("api cache encode", encodeErr)
		return result, nil
	}
	if err := a.store.SetAPICache(key, string(encoded)); err != nil {
		a.logDebugErr("api cache write", err)
	}
	return result, nil
}

func cachedJSON[T any](store *storage.Store, key string, ttl time.Duration) (T, bool) {
	var cached T
	payload, err := store.GetAPICache(key, ttl)
	if err != nil {
		return cached, false
	}
	if json.Unmarshal([]byte(payload), &cached) != nil {
		return cached, false
	}
	return cached, true
}
