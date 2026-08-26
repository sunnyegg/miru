package torrentx

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sunnyegg/miru/internal/storage"
	"golang.org/x/time/rate"
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

func TestNewRateLimiters(t *testing.T) {
	upload, download := newRateLimiters(RateLimits{
		Upload:   2 * 1024 * 1024,
		Download: 5 * 1024 * 1024,
	})
	if upload == download {
		t.Fatal("upload and download limiters must be distinct")
	}
	if upload.Limit() != rate.Limit(2*1024*1024) {
		t.Fatalf("upload limit = %v", upload.Limit())
	}
	if download.Limit() != rate.Limit(5*1024*1024) {
		t.Fatalf("download limit = %v", download.Limit())
	}
	if upload.Burst() != rateLimiterBurst || download.Burst() != rateLimiterBurst {
		t.Fatalf("bursts = %d, %d", upload.Burst(), download.Burst())
	}
}

func TestNewRateLimitersUnlimited(t *testing.T) {
	upload, download := newRateLimiters(RateLimits{Upload: -1})
	if upload.Limit() != rate.Inf || download.Limit() != rate.Inf {
		t.Fatalf("limits = %v, %v", upload.Limit(), download.Limit())
	}
}

func TestBytesPerSecond(t *testing.T) {
	if got := bytesPerSecond(2048, 2*time.Second); got != 1024 {
		t.Fatalf("speed = %d", got)
	}
	if got := bytesPerSecond(0, time.Second); got != 0 {
		t.Fatalf("zero bytes speed = %d", got)
	}
}

func TestUploadRatio(t *testing.T) {
	if got := UploadRatio(50, 100); got != 0.5 {
		t.Fatalf("ratio = %v", got)
	}
	if got := UploadRatio(1, 0); got != 0 {
		t.Fatalf("zero total ratio = %v", got)
	}
}

func TestSeedingComplete(t *testing.T) {
	if !seedingComplete(50, 100) {
		t.Fatal("expected 0.5 ratio to complete seeding")
	}
	if seedingComplete(49, 100) {
		t.Fatal("expected less than 0.5 ratio to keep seeding")
	}
}
