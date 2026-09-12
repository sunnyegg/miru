package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PlaybackState struct {
	AnilistID       int
	EpisodeNumber   int
	PositionSeconds float64
	Percent         float64
	UpdatedAt       string
}

func (s *Store) GetPlaybackState(anilistID, episodeNumber int) (PlaybackState, error) {
	if anilistID <= 0 || episodeNumber <= 0 {
		return PlaybackState{}, fmt.Errorf("invalid playback key: %d/%d", anilistID, episodeNumber)
	}

	var state PlaybackState
	err := s.db.QueryRow(
		`SELECT anilist_id, episode_number, position_seconds, percent, updated_at
		 FROM episode_playback
		 WHERE anilist_id = ? AND episode_number = ?`,
		anilistID, episodeNumber,
	).Scan(
		&state.AnilistID,
		&state.EpisodeNumber,
		&state.PositionSeconds,
		&state.Percent,
		&state.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return PlaybackState{}, ErrNotFound
	}
	return state, err
}

func (s *Store) UpsertPlaybackState(anilistID, episodeNumber int, positionSeconds, percent float64) error {
	if anilistID <= 0 || episodeNumber <= 0 {
		return fmt.Errorf("invalid playback key: %d/%d", anilistID, episodeNumber)
	}
	if positionSeconds < 0 {
		positionSeconds = 0
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	_, err := s.db.Exec(
		`INSERT INTO episode_playback(
			anilist_id, episode_number, position_seconds, percent, updated_at
		) VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(anilist_id, episode_number) DO UPDATE SET
			position_seconds = excluded.position_seconds,
			percent = excluded.percent,
			updated_at = excluded.updated_at`,
		anilistID,
		episodeNumber,
		positionSeconds,
		percent,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}
