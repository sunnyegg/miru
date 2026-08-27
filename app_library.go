package main

import (
	"database/sql"
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
	episodeNum := int(ep.EpisodeNumber.Int64)
	if !ep.EpisodeNumber.Valid {
		parsed := media.ParseFilename(ep.FilePath)
		if parsed.HasEpisode {
			episodeNum = parsed.Episode
		}
	}
	mapped, mapErr := client.MapSeasonEpisode(anilistID, episodeNum)
	if mapErr == nil {
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
		runtime.LogError(a.ctx, err.Error())
		return nil, false, err
	}
	found, err := client.Search(parsed.Title)
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
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
	if !parsed.HasEpisode {
		return candidates, true, nil
	}
	mapped, mapErr := client.MapSeasonEpisode(found[0].ID, parsed.Episode)
	if mapErr == nil {
		ep.EpisodeNumber = sql.NullInt64{Int64: int64(mapped), Valid: true}
	}
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
