package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/media"
	"github.com/sunnyegg/miru/internal/nyaa"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
	client, err := a.newAnilist("")
	if err != nil {
		return nil, err
	}
	results, err := client.Search(query)
	if err != nil {
		return nil, err
	}
	return toAnimeViews(results), nil
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
	var zero T
	if payload, err := store.GetAPICache(key, ttl); err == nil {
		var cached T
		if json.Unmarshal([]byte(payload), &cached) == nil {
			return cached, nil
		}
	}

	result, err := fetch()
	if err != nil {
		if payload, cacheErr := store.GetAPICache(key, 0); cacheErr == nil {
			var cached T
			if json.Unmarshal([]byte(payload), &cached) == nil {
				return cached, nil
			}
		}
		return zero, err
	}

	if encoded, encodeErr := json.Marshal(result); encodeErr == nil {
		_ = store.SetAPICache(key, string(encoded))
	}
	return result, nil
}

func (a *App) SearchNyaa(query string) ([]NyaaResultView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("Nyaa search query is empty")
	}
	if len(query) > 200 {
		return nil, errors.New("Nyaa search query is too long")
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

func (a *App) BindEpisode(episodeID int64, anilistID int) error {
	if err := a.ready(); err != nil {
		return err
	}
	ep, err := a.store.GetEpisode(episodeID)
	if err != nil {
		return err
	}
	client, err := a.anilistClient()
	if err != nil {
		return err
	}
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
		client, clientErr := a.newAnilist("")
		if clientErr != nil {
			runtime.LogError(a.ctx, clientErr.Error())
			return ImportResult{}, clientErr
		}
		found, err := client.Search(parsed.Title)
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
