package media

import "testing"

func TestParseFilename(t *testing.T) {
	got := ParseFilename("[Subs] Boku no Hero Academia - 73 [1080p].mkv")
	if got.Title == "" {
		t.Fatalf("empty title: %+v", got)
	}
	if !got.HasEpisode || got.Episode != 73 {
		t.Fatalf("episode: %+v", got)
	}
}

func TestParseFallbackTitle(t *testing.T) {
	got := ParseFilename("/tmp/plainfile.mkv")
	if got.Title != "plainfile" {
		t.Fatalf("title = %q", got.Title)
	}
}

func TestDisplayTitle(t *testing.T) {
	got := DisplayTitle(Parsed{Title: "Test", Episode: 2, HasEpisode: true})
	if got != "Test — Episode 2" {
		t.Fatalf("got %q", got)
	}
}

func TestParseFilenameMovieWithoutEpisode(t *testing.T) {
	got := ParseFilename("[ASW] Overlord Movie - Sei Oukoku-hen [1080p HEVC][886AE26D].mkv")
	if got.HasEpisode {
		t.Fatalf("movie filename should not parse an episode: %+v", got)
	}
	if got.Title == "" {
		t.Fatalf("empty title: %+v", got)
	}
}

func TestEpisodeOrSingle(t *testing.T) {
	parsed := Parsed{Title: "Movie", HasEpisode: false}
	if number, ok := EpisodeOrSingle(parsed, 1); !ok || number != 1 {
		t.Fatalf("single episode movie: got (%d, %v)", number, ok)
	}
	if number, ok := EpisodeOrSingle(parsed, 12); ok || number != 0 {
		t.Fatalf("multi episode without parse: got (%d, %v)", number, ok)
	}
	parsed = Parsed{Title: "Show", Episode: 3, HasEpisode: true}
	if number, ok := EpisodeOrSingle(parsed, 12); !ok || number != 3 {
		t.Fatalf("parsed episode: got (%d, %v)", number, ok)
	}
}
