package main

import (
	"errors"
	"fmt"

	"github.com/sunnyegg/miru/internal/mpv"
)

func playbackTitle(session *playSession) string {
	return playbackTitleFor(session.animeTitle, session.episodeNum)
}

func playbackTitleFor(animeTitle string, episodeNum int) string {
	if episodeNum > 0 {
		return fmt.Sprintf("%s — Episode %d", animeTitle, episodeNum)
	}
	return animeTitle
}

// playbackEvent builds the event sent to the UI from an mpv state snapshot.
func (a *App) playbackEvent(state mpv.Progress) PlaybackEvent {
	a.playMu.Lock()
	session := a.play
	if session == nil {
		a.playMu.Unlock()
		return PlaybackEvent{}
	}
	event := PlaybackEvent{
		EpisodeID:  session.episodeID,
		Title:      playbackTitle(session),
		Position:   state.Position,
		Duration:   state.Duration,
		Percent:    state.Percent,
		Paused:     state.Paused,
		Volume:     state.Volume,
		Fullscreen: state.Fullscreen,
	}
	a.playMu.Unlock()
	return event
}

func (a *App) playbackSession() (*playSession, error) {
	a.playMu.Lock()
	session := a.play
	a.playMu.Unlock()
	if session == nil || !a.player.Playing() {
		return nil, errors.New("nothing is playing")
	}
	return session, nil
}

// playbackState reads the current mpv state for a control response.
func (a *App) playbackState() (PlaybackEvent, error) {
	if _, err := a.playbackSession(); err != nil {
		return PlaybackEvent{}, err
	}
	state, err := a.player.State()
	if err != nil {
		return PlaybackEvent{}, err
	}
	return a.playbackEvent(state), nil
}

// TogglePlaybackPause flips mpv's pause state and returns the new state.
func (a *App) TogglePlaybackPause() (PlaybackEvent, error) {
	if _, err := a.playbackSession(); err != nil {
		return PlaybackEvent{}, err
	}
	paused, err := a.player.Paused()
	if err != nil {
		return PlaybackEvent{}, err
	}
	if err := a.player.SetPaused(!paused); err != nil {
		return PlaybackEvent{}, err
	}
	return a.playbackState()
}

// SeekPlayback jumps to an absolute position in seconds.
func (a *App) SeekPlayback(position float64) (PlaybackEvent, error) {
	if _, err := a.playbackSession(); err != nil {
		return PlaybackEvent{}, err
	}
	if err := a.player.Seek(position); err != nil {
		return PlaybackEvent{}, err
	}
	return a.playbackState()
}

// SetPlaybackVolume sets mpv volume on a 0–100 scale.
func (a *App) SetPlaybackVolume(volume int) (PlaybackEvent, error) {
	if _, err := a.playbackSession(); err != nil {
		return PlaybackEvent{}, err
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	if err := a.player.SetVolume(volume); err != nil {
		return PlaybackEvent{}, err
	}
	return a.playbackState()
}

// TogglePlaybackFullscreen flips the mpv window fullscreen state.
func (a *App) TogglePlaybackFullscreen() (PlaybackEvent, error) {
	if _, err := a.playbackSession(); err != nil {
		return PlaybackEvent{}, err
	}
	fullscreen, err := a.player.Fullscreen()
	if err != nil {
		return PlaybackEvent{}, err
	}
	if err := a.player.SetFullscreen(!fullscreen); err != nil {
		return PlaybackEvent{}, err
	}
	return a.playbackState()
}

// StopPlayback closes the mpv window. The usual exit path saves the final
// position and clears the now-playing state.
func (a *App) StopPlayback() error {
	if _, err := a.playbackSession(); err != nil {
		return err
	}
	a.player.Stop()
	return nil
}
