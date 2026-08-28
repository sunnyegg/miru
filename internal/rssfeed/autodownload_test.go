package rssfeed

import "testing"

func TestShouldAutoDownload(t *testing.T) {
	libraryTitles := []string{"Frieren", "Sousou no Frieren"}

	tests := []struct {
		name      string
		enabled   bool
		libOnly   bool
		title     string
		hasMagnet bool
		want      bool
	}{
		{
			name:      "disabled",
			enabled:   false,
			libOnly:   false,
			title:     "Frieren - 01",
			hasMagnet: true,
			want:      false,
		},
		{
			name:      "no magnet",
			enabled:   true,
			libOnly:   false,
			title:     "Frieren - 01",
			hasMagnet: false,
			want:      false,
		},
		{
			name:      "all items when not library only",
			enabled:   true,
			libOnly:   false,
			title:     "Unknown Anime - 01",
			hasMagnet: true,
			want:      true,
		},
		{
			name:      "library match",
			enabled:   true,
			libOnly:   true,
			title:     "[SubsPlease] Frieren - 20 [1080p]",
			hasMagnet: true,
			want:      true,
		},
		{
			name:      "library miss",
			enabled:   true,
			libOnly:   true,
			title:     "[SubsPlease] Other Show - 01",
			hasMagnet: true,
			want:      false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := ShouldAutoDownload(
				testCase.enabled,
				testCase.libOnly,
				testCase.title,
				testCase.hasMagnet,
				libraryTitles,
			)
			if got != testCase.want {
				t.Fatalf("ShouldAutoDownload() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestTitleMatchesLibrary(t *testing.T) {
	titles := []string{"Dungeon Meshi", "Delicious in Dungeon"}
	if !TitleMatchesLibrary("[Erai-raws] Dungeon Meshi - 05", titles) {
		t.Fatal("expected library match")
	}
	if TitleMatchesLibrary("[Erai-raws] Unrelated - 01", titles) {
		t.Fatal("expected no library match")
	}
}

func TestTorrentSource(t *testing.T) {
	if got := TorrentSource("magnet:?xt=urn:btih:abc", "https://example.com"); got != "magnet:?xt=urn:btih:abc" {
		t.Fatalf("magnet first = %q", got)
	}
	if got := TorrentSource("", "magnet:?xt=urn:btih:def"); got != "magnet:?xt=urn:btih:def" {
		t.Fatalf("magnet link = %q", got)
	}
	if got := TorrentSource("", "https://example.com/torrent"); got != "" {
		t.Fatalf("http link = %q", got)
	}
}
