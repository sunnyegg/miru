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
