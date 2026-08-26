package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "app_data.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestMigrateAndSettings(t *testing.T) {
	store := openTestStore(t)
	if err := store.SetSetting("mpv_path", "/usr/bin/mpv"); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetSetting("mpv_path")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/usr/bin/mpv" {
		t.Fatalf("got %q", got)
	}
}

func TestSyncEventsUnique(t *testing.T) {
	store := openTestStore(t)
	if err := store.RecordSync(21, 3); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordSync(21, 3); err != nil {
		t.Fatal(err)
	}
	ok, err := store.HasSynced(21, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected sync event")
	}
	var n int
	if err := store.db.QueryRow(`SELECT COUNT(1) FROM sync_events`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("count = %d, want 1", n)
	}
}

func TestFailInterruptedDownloads(t *testing.T) {
	store := openTestStore(t)
	id, err := store.InsertTorrentJob(TorrentJob{
		Source:  "magnet:?xt=urn:btih:abc",
		DestDir: t.TempDir(),
		Status:  "DOWNLOADING",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.FailInterruptedDownloads(); err != nil {
		t.Fatal(err)
	}
	job, err := store.LatestTorrentJob()
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != id || job.Status != "FAILED" {
		t.Fatalf("job = %+v", job)
	}
}

func TestListTorrentJobs(t *testing.T) {
	store := openTestStore(t)
	first, err := store.InsertTorrentJob(TorrentJob{
		Source:  "first.torrent",
		DestDir: t.TempDir(),
		Status:  "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.InsertTorrentJob(TorrentJob{
		Source:        "second.torrent",
		DestDir:       t.TempDir(),
		Status:        "SEEDING",
		BytesUploaded: 10,
	})
	if err != nil {
		t.Fatal(err)
	}

	jobs, err := store.ListTorrentJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].ID != second || jobs[1].ID != first {
		t.Fatalf("jobs = %+v", jobs)
	}
	if jobs[0].Status != "SEEDING" || jobs[0].BytesUploaded != 10 {
		t.Fatalf("latest job = %+v", jobs[0])
	}
}

func TestEpisodeBind(t *testing.T) {
	store := openTestStore(t)
	if err := store.UpsertAnime(Anime{AnilistID: 1, TitleRomaji: "Test"}); err != nil {
		t.Fatal(err)
	}
	id, err := store.InsertEpisode(Episode{
		FilePath:     "/tmp/ep.mkv",
		DisplayTitle: "Test 01",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.BindEpisode(id, 1, 1); err != nil {
		t.Fatal(err)
	}
	ep, err := store.GetEpisode(id)
	if err != nil {
		t.Fatal(err)
	}
	if !ep.AnilistID.Valid || ep.AnilistID.Int64 != 1 {
		t.Fatalf("anilist = %+v", ep.AnilistID)
	}
	if ep.TitleRomaji != "Test" {
		t.Fatalf("title = %s", ep.TitleRomaji)
	}
	_ = sql.NullInt64{}
}
