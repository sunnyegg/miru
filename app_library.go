package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/media"
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
	a.applyAnilistProgress(out)
	return out, nil
}

func (a *App) ListStreamingEpisodeThumbnails(mediaID int) ([]StreamingEpisodeThumbnailView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	if mediaID <= 0 {
		return nil, errors.New("invalid anime id")
	}
	return loadCachedJSON(a, fmt.Sprintf("streaming:%d", mediaID), apiCacheTTL, func() ([]StreamingEpisodeThumbnailView, error) {
		token, _ := a.tokens.Get()
		client, err := a.newAnilist(token)
		if err != nil {
			return nil, err
		}
		thumbnails, err := client.StreamingEpisodeThumbnails(mediaID)
		if err != nil {
			return nil, err
		}
		out := make([]StreamingEpisodeThumbnailView, 0, len(thumbnails))
		for _, thumbnail := range thumbnails {
			out = append(out, StreamingEpisodeThumbnailView{
				EpisodeNumber: thumbnail.EpisodeNumber,
				Thumbnail:     thumbnail.Thumbnail,
			})
		}
		return out, nil
	})
}

func (a *App) applyAnilistProgress(episodes []EpisodeView) {
	mediaIDs := make(map[int]struct{})
	for _, episode := range episodes {
		if episode.AnilistID > 0 {
			mediaIDs[episode.AnilistID] = struct{}{}
		}
	}
	if len(mediaIDs) == 0 {
		return
	}
	ids := make([]int, 0, len(mediaIDs))
	for mediaID := range mediaIDs {
		ids = append(ids, mediaID)
	}
	token, _ := a.tokens.Get()
	client, err := a.newAnilist(token)
	if err != nil {
		a.logDebugErr("anilist progress client", err)
		return
	}
	progressByMedia, err := client.ListProgressForMedia(ids)
	if err != nil {
		a.logDebugErr("anilist progress fetch", err)
		return
	}
	for index := range episodes {
		if episodes[index].AnilistID <= 0 {
			continue
		}
		mediaProgress, ok := progressByMedia[episodes[index].AnilistID]
		if !ok {
			continue
		}
		episodes[index].Progress = mediaProgress.Progress
		if mediaProgress.TotalEpisodes > 0 {
			episodes[index].TotalEpisodes = mediaProgress.TotalEpisodes
		}
		if mediaProgress.MediaStatus != "" {
			episodes[index].MediaStatus = mediaProgress.MediaStatus
		}
		episodes[index].NextAiringEpisode = mediaProgress.NextAiringEpisode
	}
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
	token, _ := a.tokens.Get()
	client, err := a.newAnilist(token)
	if err != nil {
		return nil, err
	}
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
	token, _ := a.tokens.Get()
	client, err := a.newAnilist(token)
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
	episodeNum := 0
	hasStoredNumber := ep.EpisodeNumber.Valid && ep.EpisodeNumber.Int64 > 0
	if hasStoredNumber {
		episodeNum = int(ep.EpisodeNumber.Int64)
	}
	parsed := media.ParseFilename(ep.FilePath)
	number, parsedOK := media.EpisodeOrSingle(parsed, anime.TotalEpisodes)
	if !hasStoredNumber && parsedOK {
		taken, takenErr := a.store.HasEpisodeNumber(anilistID, number, episodeID)
		if takenErr != nil {
			return takenErr
		}
		if !taken {
			episodeNum = number
		}
	}
	mapped, mapErr := client.MapSeasonEpisode(anilistID, episodeNum)
	if mapErr != nil {
		a.logDebugErr("bind episode season map", mapErr)
	} else {
		episodeNum = mapped
	}
	return a.store.BindEpisode(episodeID, anilistID, episodeNum)
}

func (a *App) importPath(path string) (ImportResult, error) {
	path = filepath.Clean(path)
	if absolute, err := filepath.Abs(path); err == nil {
		path = absolute
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
	}
	if existing, err := a.store.EpisodeByPath(path); err == nil {
		return ImportResult{Episode: toEpisodeView(existing)}, nil
	}

	parsed := media.ParseFilename(path)
	displayTitle := media.DisplayTitle(parsed)
	if displayTitle != "" {
		if existing, err := a.store.EpisodeByDisplayTitle(displayTitle); err == nil {
			return ImportResult{Episode: toEpisodeView(existing)}, nil
		}
	}

	ep := storage.Episode{
		FilePath:     path,
		DisplayTitle: displayTitle,
		Status:       "COMPLETED",
	}
	if parsed.HasEpisode {
		ep.EpisodeNumber = sql.NullInt64{Int64: int64(parsed.Episode), Valid: true}
	}

	candidates, autoBound, err := a.resolveImportMatch(parsed, &ep)
	if err != nil {
		return ImportResult{}, err
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

func (a *App) resolveImportMatch(parsed media.Parsed, ep *storage.Episode) ([]AnimeView, bool, error) {
	if parsed.Title == "" {
		return nil, false, nil
	}
	client, err := a.newAnilist("")
	if err != nil {
		return nil, false, err
	}
	found, err := client.Search(parsed.Title)
	if err != nil {
		a.logDebugErr("import anilist search", err)
		return nil, false, nil
	}
	candidates := toAnimeViews(found)
	if len(found) != 1 {
		return candidates, false, nil
	}
	if err := a.store.UpsertAnime(toStoredAnime(found[0])); err != nil {
		return nil, false, err
	}
	ep.AnilistID = sql.NullInt64{Int64: int64(found[0].ID), Valid: true}
	number, hasNumber := media.EpisodeOrSingle(parsed, found[0].TotalEpisodes)
	if !hasNumber {
		return candidates, true, nil
	}
	taken, err := a.store.HasEpisodeNumber(found[0].ID, number, 0)
	if err != nil {
		return nil, false, err
	}
	if taken {
		return candidates, true, nil
	}
	mapped, mapErr := client.MapSeasonEpisode(found[0].ID, number)
	if mapErr != nil {
		a.logDebugErr("import season map", mapErr)
	} else {
		number = mapped
	}
	ep.EpisodeNumber = sql.NullInt64{Int64: int64(number), Valid: true}
	return candidates, true, nil
}

func toEpisodeView(e storage.Episode) EpisodeView {
	title := e.DisplayTitle
	if e.TitleEnglish != "" {
		title = e.TitleEnglish
	} else if e.TitleRomaji != "" {
		title = e.TitleRomaji
	}
	view := EpisodeView{
		ID:            e.ID,
		FilePath:      e.FilePath,
		DisplayTitle:  e.DisplayTitle,
		Status:        e.Status,
		AnimeTitle:    title,
		CoverImage:    e.CoverImage,
		Bound:         e.AnilistID.Valid,
		TotalEpisodes: e.TotalEpisodes,
		MediaStatus:   e.MediaStatus,
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
			ListStatus:    a.ListStatus,
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
