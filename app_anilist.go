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
		return AnilistStatus{Connected: false}, nil
	}
	client, err := a.newAnilist(token)
	if err != nil {
		return AnilistStatus{Connected: false}, nil
	}
	name, err := client.ViewerName()
	if err != nil {
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
	_ = a.store.DeleteAPICache(watchingCacheKey)
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
	return loadCachedJSON(a.store, fmt.Sprintf("airing:%d:%d", start, end), apiCacheTTL, func() ([]AiringScheduleView, error) {
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

func (a *App) ListCurrentlyWatching() ([]WatchingEntryView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	token, err := a.tokens.Get()
	if err != nil {
		return nil, errors.New("AniList not connected")
	}
	return loadCachedJSON(a.store, watchingCacheKey, apiCacheTTL, func() ([]WatchingEntryView, error) {
		client, err := a.newAnilist(token)
		if err != nil {
			return nil, err
		}
		entries, err := client.ListCurrent()
		if err != nil {
			return nil, err
		}
		out := make([]WatchingEntryView, 0, len(entries))
		for _, entry := range entries {
			out = append(out, WatchingEntryView{
				MediaID:       entry.MediaID,
				Progress:      entry.Progress,
				TitleRomaji:   entry.TitleRomaji,
				TitleEnglish:  entry.TitleEnglish,
				CoverImage:    entry.CoverImage,
				TotalEpisodes: entry.TotalEpisodes,
				MediaStatus:   entry.MediaStatus,
			})
		}
		return out, nil
	})
}

func loadCachedJSON[T any](store *storage.Store, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	if cached, ok := cachedJSON[T](store, key, ttl); ok {
		return cached, nil
	}

	result, err := fetch()
	if err != nil {
		if stale, ok := cachedJSON[T](store, key, 0); ok {
			return stale, nil
		}
		var zero T
		return zero, err
	}

	encoded, encodeErr := json.Marshal(result)
	if encodeErr == nil {
		_ = store.SetAPICache(key, string(encoded))
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
