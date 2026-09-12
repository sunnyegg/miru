package main

import (
	"path/filepath"
	"testing"

	"github.com/sunnyegg/miru/internal/storage"
)

func TestPlaybackStateWriterFlushesLatestPosition(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "app_data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	writer := newPlaybackStateWriter(store)
	defer writer.Close()

	writer.Submit(playbackStateUpdate{
		anilistID:     21,
		episodeNumber: 4,
		position:      120,
		percent:       18,
	})
	writer.Submit(playbackStateUpdate{
		anilistID:     21,
		episodeNumber: 4,
		position:      180,
		percent:       42.5,
	})
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}

	state, err := store.GetPlaybackState(21, 4)
	if err != nil {
		t.Fatal(err)
	}
	if state.PositionSeconds != 180 {
		t.Fatalf("position = %v", state.PositionSeconds)
	}
	if state.Percent != 42.5 {
		t.Fatalf("percent = %v", state.Percent)
	}
}
