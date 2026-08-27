package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
	client, clientErr := a.newAnilist(token)
	if clientErr != nil {
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
	_ = a.store.RecordSync(session.anilistID, session.episodeNum)
	_ = a.store.DeleteAPICache(watchingCacheKey)
	a.playMu.Lock()
	session.synced = true
	a.playMu.Unlock()
	runtime.EventsEmit(a.ctx, "sync:result", SyncEvent{
		EpisodeID: session.episodeID,
		OK:        true,
		Message:   "AniList updated to episode " + strconv.Itoa(session.episodeNum),
	})
}
