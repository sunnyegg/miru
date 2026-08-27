package torrentx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sunnyegg/miru/internal/networking"
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

func openManager(t *testing.T) (*Manager, *storage.Store) {
	t.Helper()
	store, err := storage.Open(filepath.Join(t.TempDir(), "app_data.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return NewManager(store), store
}

func insertJob(t *testing.T, store *storage.Store, destDir, name string) int64 {
	t.Helper()
	id, err := store.InsertTorrentJob(storage.TorrentJob{
		Source:  "magnet:?xt=urn:btih:abc",
		DestDir: destDir,
		Name:    name,
		Status:  "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestRemoveDeletesJobAndFiles(t *testing.T) {
	manager, store := openManager(t)
	destDir := t.TempDir()
	showDir := filepath.Join(destDir, "Show")
	if err := os.MkdirAll(showDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(showDir, "ep.mkv")
	if err := os.WriteFile(payload, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	id := insertJob(t, store, destDir, "Show")

	if err := manager.Remove(id, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(payload); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("payload still present: %v", err)
	}
	jobs, err := store.ListTorrentJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("jobs = %+v", jobs)
	}
}

func TestRemoveListOnlyKeepsFiles(t *testing.T) {
	manager, store := openManager(t)
	destDir := t.TempDir()
	showDir := filepath.Join(destDir, "Show")
	if err := os.MkdirAll(showDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(showDir, "ep.mkv")
	if err := os.WriteFile(payload, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	id := insertJob(t, store, destDir, "Show")

	if err := manager.Remove(id, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(payload); err != nil {
		t.Fatalf("payload missing: %v", err)
	}
	jobs, err := store.ListTorrentJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("jobs = %+v", jobs)
	}
}

func TestRemoveSkipsPlaceholderAndEmptyName(t *testing.T) {
	manager, store := openManager(t)
	destDir := t.TempDir()
	placeholderDir := filepath.Join(destDir, "Magnet download")
	if err := os.MkdirAll(placeholderDir, 0o755); err != nil {
		t.Fatal(err)
	}
	placeholderFile := filepath.Join(placeholderDir, "x")
	if err := os.WriteFile(placeholderFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(destDir, "marker")
	if err := os.WriteFile(marker, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	placeholderID := insertJob(t, store, destDir, "Magnet download")
	if err := manager.Remove(placeholderID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(placeholderFile); err != nil {
		t.Fatalf("placeholder files removed: %v", err)
	}

	emptyID := insertJob(t, store, destDir, "")
	if err := manager.Remove(emptyID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("dest dir wiped: %v", err)
	}
}

func TestRemoveSkipsPathEscape(t *testing.T) {
	manager, store := openManager(t)
	parent := t.TempDir()
	destDir := filepath.Join(parent, "downloads")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(parent, "secret")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	id := insertJob(t, store, destDir, "..")
	if err := manager.Remove(id, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("escaped delete: %v", err)
	}
	if _, err := os.Stat(destDir); err != nil {
		t.Fatalf("dest dir missing: %v", err)
	}
}

func TestRemoveMissingFilesStillDeletesJob(t *testing.T) {
	manager, store := openManager(t)
	id := insertJob(t, store, t.TempDir(), "Gone")
	if err := manager.Remove(id, true); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TorrentJobByID(id); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("job = %v", err)
	}
}

func TestCloseKeepsSeedingJob(t *testing.T) {
	manager, store := openManager(t)
	id, err := store.InsertTorrentJob(storage.TorrentJob{
		Source:        "magnet:?xt=urn:btih:abc",
		DestDir:       t.TempDir(),
		Name:          "Show",
		Status:        "SEEDING",
		BytesUploaded: 10,
		BytesTotal:    100,
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	manager.mu.Lock()
	manager.job = job
	manager.mu.Unlock()
	manager.Close()

	got, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "SEEDING" || got.BytesUploaded != 10 {
		t.Fatalf("job = %+v", got)
	}
}

func TestCloseCancelsDownloadingJob(t *testing.T) {
	manager, store := openManager(t)
	id, err := store.InsertTorrentJob(storage.TorrentJob{
		Source:  "magnet:?xt=urn:btih:abc",
		DestDir: t.TempDir(),
		Status:  "DOWNLOADING",
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	manager.mu.Lock()
	manager.job = job
	manager.mu.Unlock()
	manager.Close()

	got, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "CANCELLED" {
		t.Fatalf("job = %+v", got)
	}
}

func TestFinishStoredSeedingJob(t *testing.T) {
	manager, store := openManager(t)
	destDir := t.TempDir()
	showDir := filepath.Join(destDir, "Show")
	if err := os.MkdirAll(showDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(showDir, "ep.mkv")
	if err := os.WriteFile(payload, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, err := store.InsertTorrentJob(storage.TorrentJob{
		Source:  "magnet:?xt=urn:btih:abc",
		DestDir: destDir,
		Name:    "Show",
		Status:  "SEEDING",
	})
	if err != nil {
		t.Fatal(err)
	}

	var ingested []string
	manager.SetCallbacks(nil, func(files []string) {
		ingested = files
	})
	if err := manager.Finish(id); err != nil {
		t.Fatal(err)
	}
	job, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != "COMPLETED" {
		t.Fatalf("job = %+v", job)
	}
	if len(ingested) != 1 || ingested[0] != payload {
		t.Fatalf("ingested = %v", ingested)
	}
}

func TestResumeSeedingRejectsOtherStatus(t *testing.T) {
	manager, store := openManager(t)
	id := insertJob(t, store, t.TempDir(), "Show")
	err := manager.ResumeSeeding(id, RateLimits{}, networking.Config{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVideoFilesFromDisk(t *testing.T) {
	destDir := t.TempDir()
	showDir := filepath.Join(destDir, "Show")
	if err := os.MkdirAll(showDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(showDir, "ep.mkv")
	if err := os.WriteFile(payload, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := videoFilesFromDisk(destDir, "Show")
	if len(got) != 1 || got[0] != payload {
		t.Fatalf("got %v", got)
	}
	if files := videoFilesFromDisk(destDir, "Magnet download"); len(files) != 0 {
		t.Fatalf("placeholder = %v", files)
	}
	if files := videoFilesFromDisk(destDir, ".."); len(files) != 0 {
		t.Fatalf("escape = %v", files)
	}
}
