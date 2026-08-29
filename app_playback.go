package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sunnyegg/miru/internal/anilist"
	"github.com/sunnyegg/miru/internal/mpv"
	syncprogress "github.com/sunnyegg/miru/internal/syncprogress"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

	session := &playSession{
		episodeID:  episodeID,
		animeTitle: episodeAnimeTitle(ep),
	}
	if ep.AnilistID.Valid {
		session.anilistID = int(ep.AnilistID.Int64)
	}
	if ep.EpisodeNumber.Valid {
		session.episodeNum = int(ep.EpisodeNumber.Int64)
	}

	a.playMu.Lock()
	a.play = session
	a.playMu.Unlock()

	a.syncDiscordPresence(settings, session.animeTitle, session.episodeNum, 0)

	var glslShaders []string
	if settings.Anime4KEnabled {
		glslShaders, err = mpv.Anime4KShaderPaths(a.dirs.Config)
		if err != nil {
			return err
		}
		runtime.LogInfo(a.ctx, fmt.Sprintf("Anime4K enabled: attaching %d shader(s) to mpv", len(glslShaders)))
	} else {
		runtime.LogDebugf(a.ctx, "Anime4K disabled: no shaders attached")
	}

	return a.player.Play(mpvPath, ep.FilePath, ep.ResumePosition, glslShaders, func(p mpv.Progress) {
		a.playMu.Lock()
		session.lastProgress = p
		needsMap := !session.episodeMapped && !session.mapFailed
		shouldSync := syncprogress.ShouldAttemptRealtimeSync(
			p.Percent,
			settings.SyncThreshold,
			session.synced,
			session.mapFailed,
			session.realtimeSyncAttempted,
		)
		if shouldSync {
			session.realtimeSyncAttempted = true
		}
		a.playMu.Unlock()

		runtime.EventsEmit(a.ctx, "mpv:progress", PlaybackEvent{
			EpisodeID: episodeID,
			Percent:   p.Percent,
		})
		a.updateDiscordFromProgress(session, p, settings)
		if needsMap {
			client, err := a.playbackAnilist()
			if err != nil {
				a.logPlaybackAnilist(session, err)
				return
			}
			_ = a.ensureSeasonEpisode(session, client)
		}
		if shouldSync {
			a.maybeSync(session, p.Percent, settings.SyncThreshold)
		}
	}, func(exitErr error) {
		a.clearDiscordPresence()
		a.onMpvClosed(session, settings.SyncThreshold, exitErr)
	})
}

func (a *App) onMpvClosed(session *playSession, threshold float64, exitErr error) {
	a.playMu.Lock()
	progress := session.lastProgress
	a.playMu.Unlock()

	if progress.Duration > 0 || progress.Position > 0 {
		resume := mpv.ResumePosition(progress.Position, progress.Duration, progress.Percent, threshold)
		if err := a.store.SetResumePosition(session.episodeID, resume); err != nil {
			a.logDebugErr("save resume position", err)
		}
	}
	a.maybeSync(session, progress.Percent, threshold)

	msg := ""
	if exitErr != nil && !strings.Contains(exitErr.Error(), "signal: killed") {
		msg = exitErr.Error()
	}
	runtime.EventsEmit(a.ctx, "mpv:ended", SyncEvent{EpisodeID: session.episodeID, OK: true, Message: msg})
}

func (a *App) maybeSync(session *playSession, percent, threshold float64) {
	if session == nil || session.anilistID == 0 || session.episodeNum == 0 {
		return
	}

	a.playMu.Lock()
	if session.synced || session.mapFailed {
		a.playMu.Unlock()
		return
	}
	needsMap := !session.episodeMapped
	a.playMu.Unlock()

	if needsMap {
		client, err := a.playbackAnilist()
		if err != nil {
			a.logPlaybackAnilist(session, err)
			return
		}
		if err := a.ensureSeasonEpisode(session, client); err != nil {
			return
		}
	}

	if percent < threshold {
		return
	}

	synced, err := a.store.HasSynced(session.anilistID, session.episodeNum)
	if err != nil {
		a.logDebugErr("has synced", err)
		return
	}
	if synced {
		return
	}

	client, err := a.playbackAnilist()
	if err != nil {
		a.logPlaybackAnilist(session, err)
		return
	}
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
	if err := a.store.RecordSync(session.anilistID, session.episodeNum); err != nil {
		a.logDebugErr("record sync", err)
	}
	a.invalidateAnimeListCache()
	a.playMu.Lock()
	session.synced = true
	a.playMu.Unlock()
	runtime.EventsEmit(a.ctx, "sync:result", SyncEvent{
		EpisodeID: session.episodeID,
		OK:        true,
		Message:   "AniList updated to episode " + strconv.Itoa(session.episodeNum),
	})
}

func (a *App) playbackAnilist() (*anilist.Client, error) {
	token, err := a.tokens.Get()
	if err != nil {
		return nil, err
	}
	return a.newAnilist(token)
}

func (a *App) ensureSeasonEpisode(session *playSession, client *anilist.Client) error {
	a.playMu.Lock()
	if session.episodeMapped {
		a.playMu.Unlock()
		return nil
	}
	parsed := session.episodeNum
	anilistID := session.anilistID
	episodeID := session.episodeID
	a.playMu.Unlock()

	mapped, err := client.MapSeasonEpisode(anilistID, parsed)
	if err != nil {
		a.playMu.Lock()
		session.mapFailed = true
		a.playMu.Unlock()
		runtime.EventsEmit(a.ctx, "sync:result", SyncEvent{
			EpisodeID: episodeID,
			OK:        false,
			Message:   err.Error(),
		})
		return err
	}

	a.playMu.Lock()
	session.episodeNum = mapped
	session.episodeMapped = true
	a.playMu.Unlock()

	if mapped != parsed {
		if err := a.store.BindEpisode(episodeID, anilistID, mapped); err != nil {
			a.logDebugErr("bind mapped episode", err)
		}
		runtime.EventsEmit(a.ctx, "library:changed", true)
		runtime.LogInfo(a.ctx, fmt.Sprintf("mapped episode %d to season episode %d", parsed, mapped))
	}
	return nil
}

func (a *App) logPlaybackAnilist(session *playSession, err error) {
	if session == nil {
		a.logDebugErr("playback anilist client", err)
		return
	}
	a.playMu.Lock()
	already := session.loggedAnilist
	session.loggedAnilist = true
	a.playMu.Unlock()
	if already {
		return
	}
	a.logDebugErr("playback anilist client", err)
}
