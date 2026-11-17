package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sunnyegg/miru/internal/storage"
)

func testEpisode(id int64, anilistID, number int, path string) storage.Episode {
	return storage.Episode{
		ID:            id,
		AnilistID:     sql.NullInt64{Int64: int64(anilistID), Valid: anilistID > 0},
		EpisodeNumber: sql.NullInt64{Int64: int64(number), Valid: number > 0},
		FilePath:      path,
	}
}

func touchFiles(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func episodeNumbers(episodes []storage.Episode) []int64 {
	numbers := make([]int64, len(episodes))
	for i, episode := range episodes {
		numbers[i] = episode.EpisodeNumber.Int64
	}
	return numbers
}

func TestPlaybackQueueFollowsConsecutiveFiles(t *testing.T) {
	dir := t.TempDir()
	target := testEpisode(1, 21, 3, filepath.Join(dir, "03.mkv"))
	four := testEpisode(2, 21, 4, filepath.Join(dir, "04.mkv"))
	five := testEpisode(3, 21, 5, filepath.Join(dir, "05.mkv"))
	other := testEpisode(4, 99, 4, filepath.Join(dir, "other.mkv"))
	touchFiles(t, target.FilePath, four.FilePath, five.FilePath, other.FilePath)

	queue := playbackQueue([]storage.Episode{five, other, target, four}, target)

	if got := episodeNumbers(queue); len(got) != 3 || got[0] != 3 || got[1] != 4 || got[2] != 5 {
		t.Fatalf("queue = %v, want [3 4 5]", got)
	}
}

func TestPlaybackQueueStopsAtGap(t *testing.T) {
	dir := t.TempDir()
	target := testEpisode(1, 21, 3, filepath.Join(dir, "03.mkv"))
	five := testEpisode(2, 21, 5, filepath.Join(dir, "05.mkv"))
	six := testEpisode(3, 21, 6, filepath.Join(dir, "06.mkv"))
	touchFiles(t, target.FilePath, five.FilePath, six.FilePath)

	queue := playbackQueue([]storage.Episode{target, five, six}, target)

	if got := episodeNumbers(queue); len(got) != 1 || got[0] != 3 {
		t.Fatalf("queue = %v, want [3]", got)
	}
}

func TestPlaybackQueueSkipsMissingFileAsGap(t *testing.T) {
	dir := t.TempDir()
	target := testEpisode(1, 21, 3, filepath.Join(dir, "03.mkv"))
	four := testEpisode(2, 21, 4, filepath.Join(dir, "missing.mkv"))
	five := testEpisode(3, 21, 5, filepath.Join(dir, "05.mkv"))
	touchFiles(t, target.FilePath, five.FilePath)

	queue := playbackQueue([]storage.Episode{target, four, five}, target)

	if got := episodeNumbers(queue); len(got) != 1 || got[0] != 3 {
		t.Fatalf("queue = %v, want [3]", got)
	}
}

func TestPlaybackQueueSingleEntryWithoutAniListID(t *testing.T) {
	dir := t.TempDir()
	target := testEpisode(1, 0, 0, filepath.Join(dir, "03.mkv"))
	next := testEpisode(2, 0, 0, filepath.Join(dir, "04.mkv"))
	touchFiles(t, target.FilePath, next.FilePath)

	queue := playbackQueue([]storage.Episode{target, next}, target)

	if len(queue) != 1 || queue[0].ID != target.ID {
		t.Fatalf("queue = %+v, want only the target", queue)
	}
}

func TestPlaybackQueueCapsEntries(t *testing.T) {
	dir := t.TempDir()
	target := testEpisode(1, 21, 1, filepath.Join(dir, "001.mkv"))
	episodes := []storage.Episode{target}
	for number := 2; number <= maxPlaylistEntries+10; number++ {
		episodes = append(episodes, testEpisode(int64(number), 21, number, filepath.Join(dir, fmt.Sprintf("%03d.mkv", number))))
	}
	for _, episode := range episodes {
		touchFiles(t, episode.FilePath)
	}

	queue := playbackQueue(episodes, target)

	if len(queue) != maxPlaylistEntries {
		t.Fatalf("queue length = %d, want %d", len(queue), maxPlaylistEntries)
	}
}

func TestSameMediaPath(t *testing.T) {
	if !sameMediaPath("/media/episode.mkv", "/media/episode.mkv") {
		t.Fatal("identical paths must match")
	}
	if !sameMediaPath("/media/./episode.mkv", "/media/episode.mkv") {
		t.Fatal("cleaned paths must match")
	}
	if sameMediaPath("/media/episode.mkv", "/media/other.mkv") {
		t.Fatal("different paths must not match")
	}
	if sameMediaPath("", "/media/episode.mkv") {
		t.Fatal("empty path must not match")
	}
}
