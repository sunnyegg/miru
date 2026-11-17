package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/sunnyegg/miru/internal/appinfo"
	"github.com/sunnyegg/miru/internal/mpv"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// maxPlaylistEntries bounds how many episodes Miru hands to one mpv process.
const maxPlaylistEntries = 50

// playlistEntry is one episode in the mpv playlist. The session fields mirror
// it so progress, sync and Discord keep following the episode mpv shows.
type playlistEntry struct {
	episodeID  int64
	anilistID  int
	episodeNum int
	animeTitle string
	filePath   string
}

func newPlaylistEntry(episode storage.Episode) playlistEntry {
	entry := playlistEntry{
		episodeID:  episode.ID,
		animeTitle: episodeAnimeTitle(episode),
		filePath:   episode.FilePath,
	}
	if episode.AnilistID.Valid {
		entry.anilistID = int(episode.AnilistID.Int64)
	}
	if episode.EpisodeNumber.Valid {
		entry.episodeNum = int(episode.EpisodeNumber.Int64)
	}
	return entry
}

// playbackQueue is the target episode followed by the next episodes of the same
// anime that are on disk, in order and without skipping a gap. Autoplay stops
// when a file is missing so Miru never jumps an episode.
func playbackQueue(episodes []storage.Episode, target storage.Episode) []storage.Episode {
	queue := []storage.Episode{target}
	if !target.AnilistID.Valid || !target.EpisodeNumber.Valid {
		return queue
	}
	targetNumber := target.EpisodeNumber.Int64

	var candidates []storage.Episode
	for _, episode := range episodes {
		if episode.ID == target.ID || !episode.AnilistID.Valid || !episode.EpisodeNumber.Valid {
			continue
		}
		if episode.AnilistID.Int64 != target.AnilistID.Int64 || episode.EpisodeNumber.Int64 <= targetNumber {
			continue
		}
		if _, err := os.Stat(episode.FilePath); err != nil {
			continue
		}
		candidates = append(candidates, episode)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].EpisodeNumber.Int64 < candidates[j].EpisodeNumber.Int64
	})

	next := targetNumber + 1
	for _, episode := range candidates {
		if len(queue) >= maxPlaylistEntries {
			break
		}
		if episode.EpisodeNumber.Int64 != next {
			if episode.EpisodeNumber.Int64 > next {
				break
			}
			continue
		}
		queue = append(queue, episode)
		next++
	}
	return queue
}

// sameMediaPath compares the path mpv reports with the one Miru passed.
func sameMediaPath(first, second string) bool {
	if first == "" || second == "" {
		return false
	}
	return filepath.Clean(first) == filepath.Clean(second)
}

// playlistFor builds the mpv playlist for the target episode.
func (a *App) playlistFor(target storage.Episode) []playlistEntry {
	episodes, err := a.store.ListEpisodes()
	if err != nil {
		a.logDebugErr("list episodes for playlist", err)
	}
	queue := playbackQueue(episodes, target)
	entries := make([]playlistEntry, len(queue))
	for i, episode := range queue {
		entries[i] = newPlaylistEntry(episode)
	}
	return entries
}

// playlistResume is the stored position for an episode mpv is about to open.
func (a *App) playlistResume(entry playlistEntry) float64 {
	if entry.anilistID == 0 || entry.episodeNum == 0 {
		return 0
	}
	state, err := a.store.GetPlaybackState(entry.anilistID, entry.episodeNum)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			a.logDebugErr("load playback state", err)
		}
		return 0
	}
	return state.PositionSeconds
}

// followPlaylist moves the session to the episode mpv started on its own: the
// episode that just finished is saved and synced first, then the session takes
// the new episode so progress, Discord and AniList follow the player. It also
// follows a backward step through mpv's own playlist keys.
func (a *App) followPlaylist(session *playSession, finished mpv.Progress, path string, settings SettingsView) {
	a.playMu.Lock()
	index := -1
	for i := range session.queue {
		if sameMediaPath(session.queue[i].filePath, path) {
			index = i
			break
		}
	}
	current := session.queueIndex
	a.playMu.Unlock()
	if index < 0 || index == current {
		return
	}
	entry := session.queue[index]

	if finished.Duration > 0 || finished.Position > 0 {
		resume := mpv.ResumePosition(finished.Position, finished.Duration, finished.Percent, settings.SyncThreshold)
		a.saveFinalPlaybackPosition(session, resume, finished.Percent)
	}
	a.closePlaybackWriter(session, "save playback state")
	a.maybeSync(session, finished.Percent, settings.SyncThreshold)

	resume := a.playlistResume(entry)
	writer := newPlaybackStateWriter(a.store)
	if entry.anilistID > 0 && entry.episodeNum > 0 {
		if err := a.store.UpsertPlaybackState(entry.anilistID, entry.episodeNum, resume, 0); err != nil {
			a.logDebugErr("start playback state", err)
		}
	}

	a.playMu.Lock()
	session.queueIndex = index
	session.episodeID = entry.episodeID
	session.anilistID = entry.anilistID
	session.episodeNum = entry.episodeNum
	session.animeTitle = entry.animeTitle
	session.playbackWriter = writer
	session.synced = false
	session.realtimeSyncAttempted = false
	session.episodeMapped = false
	session.mapFailed = false
	session.loggedAnilist = false
	session.lastProgress = mpv.Progress{}
	a.playMu.Unlock()

	a.syncDiscordPresence(settings, entry.animeTitle, entry.episodeNum, 0)
	title := appinfo.DisplayName + " — " + playbackTitleFor(entry.animeTitle, entry.episodeNum)
	if err := a.player.SetTitle(title); err != nil {
		a.logDebugErr("set player title", err)
	}
	if resume > 0 {
		if err := a.player.Seek(resume); err != nil {
			a.logDebugErr("seek playlist entry", err)
		}
	}
	runtime.EventsEmit(a.ctx, "library:changed", true)
}
