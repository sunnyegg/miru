package torrentx

import (
	"path/filepath"
	"testing"

	"github.com/sunnyegg/miru/internal/storage"
)

func TestJobPercent(t *testing.T) {
	if got := JobPercent(50, 100); got != 50 {
		t.Fatalf("got %v", got)
	}
	if got := JobPercent(1, 0); got != 0 {
		t.Fatalf("zero total: %v", got)
	}
}

func TestToView(t *testing.T) {
	view := ToView(storage.TorrentJob{
		ID:             7,
		Name:           "Show",
		Status:         "DOWNLOADING",
		BytesCompleted: 25,
		BytesTotal:     100,
		Source:         "magnet:?xt=urn:btih:abc",
	})
	if view.Percent != 25 || view.ID != 7 {
		t.Fatalf("%+v", view)
	}
}

func TestResolveDataPath(t *testing.T) {
	got := ResolveDataPath("/data", "Show/ep.mkv")
	if got != filepath.Join("/data", "Show/ep.mkv") && got != "/data/Show/ep.mkv" {
		t.Fatalf("got %s", got)
	}
}

func TestDisplaySource(t *testing.T) {
	if got := displaySource("magnet:?xt=urn:btih:x"); got != "Magnet download" {
		t.Fatalf("got %s", got)
	}
	if got := displaySource("/tmp/file.torrent"); got != "file.torrent" {
		t.Fatalf("got %s", got)
	}
}
