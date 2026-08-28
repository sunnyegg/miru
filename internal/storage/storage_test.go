package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
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

func TestTorrentJobFilesJSON(t *testing.T) {
	store := openTestStore(t)
	id, err := store.InsertTorrentJob(TorrentJob{
		Source:     "show.torrent",
		DestDir:    t.TempDir(),
		Status:     "QUEUED",
		BytesTotal: 12,
		FilesJSON:  `[{"path":"01.mkv","length":12}]`,
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if job.FilesJSON != `[{"path":"01.mkv","length":12}]` || job.BytesTotal != 12 {
		t.Fatalf("job = %+v", job)
	}
	job.FilesJSON = `[{"path":"01.mkv","length":12},{"path":"02.mkv","length":8}]`
	job.BytesTotal = 20
	if err := store.UpdateTorrentJob(job); err != nil {
		t.Fatal(err)
	}
	updated, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if updated.FilesJSON != job.FilesJSON || updated.BytesTotal != 20 {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestRecoverInterruptedDownloads(t *testing.T) {
	store := openTestStore(t)
	downloadingID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "magnet:?xt=urn:btih:abc",
		DestDir: t.TempDir(),
		Status:  "DOWNLOADING",
		Error:   "stale",
	})
	if err != nil {
		t.Fatal(err)
	}
	pausedID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "magnet:?xt=urn:btih:ghi",
		DestDir: t.TempDir(),
		Status:  "PAUSED",
	})
	if err != nil {
		t.Fatal(err)
	}
	seedingID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "magnet:?xt=urn:btih:def",
		DestDir: t.TempDir(),
		Name:    "Show",
		Status:  "SEEDING",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RecoverInterruptedDownloads(); err != nil {
		t.Fatal(err)
	}
	downloading, err := store.TorrentJobByID(downloadingID)
	if err != nil {
		t.Fatal(err)
	}
	if downloading.Status != "QUEUED" || downloading.Error != "" {
		t.Fatalf("downloading = %+v", downloading)
	}
	paused, err := store.TorrentJobByID(pausedID)
	if err != nil {
		t.Fatal(err)
	}
	if paused.Status != "FAILED" || paused.Error != "interrupted by restart" {
		t.Fatalf("paused = %+v", paused)
	}
	seeding, err := store.TorrentJobByID(seedingID)
	if err != nil {
		t.Fatal(err)
	}
	if seeding.Status != "SEEDING" {
		t.Fatalf("seeding = %+v", seeding)
	}
}

func TestNextQueuedTorrentJob(t *testing.T) {
	store := openTestStore(t)
	firstID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "first.torrent",
		DestDir: t.TempDir(),
		Status:  "QUEUED",
	})
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "second.torrent",
		DestDir: t.TempDir(),
		Status:  "QUEUED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if firstID >= secondID {
		t.Fatalf("insert order = %d, %d", firstID, secondID)
	}

	job, err := store.NextQueuedTorrentJob()
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != firstID {
		t.Fatalf("next queued = %+v, want id %d", job, firstID)
	}
}

func TestRecoverInterruptedDownloadsKeepsQueued(t *testing.T) {
	store := openTestStore(t)
	queuedID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "queued.torrent",
		DestDir: t.TempDir(),
		Status:  "QUEUED",
	})
	if err != nil {
		t.Fatal(err)
	}
	downloadingID, err := store.InsertTorrentJob(TorrentJob{
		Source:  "magnet:?xt=urn:btih:abc",
		DestDir: t.TempDir(),
		Status:  "DOWNLOADING",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RecoverInterruptedDownloads(); err != nil {
		t.Fatal(err)
	}
	queued, err := store.TorrentJobByID(queuedID)
	if err != nil {
		t.Fatal(err)
	}
	if queued.Status != "QUEUED" {
		t.Fatalf("queued = %+v", queued)
	}
	downloading, err := store.TorrentJobByID(downloadingID)
	if err != nil {
		t.Fatal(err)
	}
	if downloading.Status != "QUEUED" {
		t.Fatalf("downloading = %+v", downloading)
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

func TestDeleteTorrentJob(t *testing.T) {
	store := openTestStore(t)
	id, err := store.InsertTorrentJob(TorrentJob{
		Source:  "done.torrent",
		DestDir: t.TempDir(),
		Name:    "Show",
		Status:  "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.TorrentJobByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if job.Name != "Show" || job.Status != "COMPLETED" {
		t.Fatalf("job = %+v", job)
	}
	if err := store.DeleteTorrentJob(id); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TorrentJobByID(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("lookup after delete = %v", err)
	}
	jobs, err := store.ListTorrentJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("jobs = %+v", jobs)
	}
	if err := store.DeleteTorrentJob(id); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete = %v", err)
	}
}

func TestDeleteEpisodesByFilePrefix(t *testing.T) {
	store := openTestStore(t)
	root := filepath.Join(t.TempDir(), "Show")
	keep, err := store.InsertEpisode(Episode{
		FilePath:     filepath.Join(t.TempDir(), "Other", "ep.mkv"),
		DisplayTitle: "Keep",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	single, err := store.InsertEpisode(Episode{
		FilePath:     root,
		DisplayTitle: "Single",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	child, err := store.InsertEpisode(Episode{
		FilePath:     filepath.Join(root, "ep01.mkv"),
		DisplayTitle: "Child",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	neighbor, err := store.InsertEpisode(Episode{
		FilePath:     root + "Extra.mkv",
		DisplayTitle: "Neighbor",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteEpisodesByFilePrefix(root); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetEpisode(single); !errors.Is(err, ErrNotFound) {
		t.Fatalf("single = %v", err)
	}
	if _, err := store.GetEpisode(child); !errors.Is(err, ErrNotFound) {
		t.Fatalf("child = %v", err)
	}
	if _, err := store.GetEpisode(keep); err != nil {
		t.Fatalf("keep = %v", err)
	}
	if _, err := store.GetEpisode(neighbor); err != nil {
		t.Fatalf("neighbor = %v", err)
	}
}

func TestHealSingleEpisodeNumbers(t *testing.T) {
	store := openTestStore(t)
	if err := store.UpsertAnime(Anime{
		AnilistID:     99,
		TitleRomaji:   "Movie",
		TotalEpisodes: 1,
	}); err != nil {
		t.Fatal(err)
	}
	movieID, err := store.InsertEpisode(Episode{
		AnilistID:    sql.NullInt64{Int64: 99, Valid: true},
		FilePath:     "/tmp/movie.mkv",
		DisplayTitle: "Movie",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	episodes, err := store.ListEpisodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(episodes) != 1 {
		t.Fatalf("len = %d", len(episodes))
	}
	if !episodes[0].EpisodeNumber.Valid || episodes[0].EpisodeNumber.Int64 != 1 {
		t.Fatalf("healed episode number = %+v", episodes[0].EpisodeNumber)
	}
	got, err := store.GetEpisode(movieID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.EpisodeNumber.Valid || got.EpisodeNumber.Int64 != 1 {
		t.Fatalf("stored episode number = %+v", got.EpisodeNumber)
	}
}

func TestHealSingleEpisodeNumbersSkipsWhenSlotTaken(t *testing.T) {
	store := openTestStore(t)
	if err := store.UpsertAnime(Anime{
		AnilistID:     100,
		TitleRomaji:   "Movie",
		TotalEpisodes: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.InsertEpisode(Episode{
		AnilistID:     sql.NullInt64{Int64: 100, Valid: true},
		EpisodeNumber: sql.NullInt64{Int64: 1, Valid: true},
		FilePath:      "/tmp/first.mkv",
		DisplayTitle:  "Movie first",
		Status:        "COMPLETED",
	}); err != nil {
		t.Fatal(err)
	}
	extraID, err := store.InsertEpisode(Episode{
		AnilistID:    sql.NullInt64{Int64: 100, Valid: true},
		FilePath:     "/tmp/extra.mkv",
		DisplayTitle: "Movie extra",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.HealSingleEpisodeNumbers(); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetEpisode(extraID)
	if err != nil {
		t.Fatal(err)
	}
	if got.EpisodeNumber.Valid {
		t.Fatalf("extra file should stay unnumbered: %+v", got.EpisodeNumber)
	}
}

func TestHasEpisodeNumber(t *testing.T) {
	store := openTestStore(t)
	if err := store.UpsertAnime(Anime{AnilistID: 5, TitleRomaji: "Test"}); err != nil {
		t.Fatal(err)
	}
	id, err := store.InsertEpisode(Episode{
		FilePath:     "/tmp/slot1.mkv",
		DisplayTitle: "Test 01",
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.BindEpisode(id, 5, 1); err != nil {
		t.Fatal(err)
	}
	taken, err := store.HasEpisodeNumber(5, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !taken {
		t.Fatal("slot 1 should be taken")
	}
	taken, err = store.HasEpisodeNumber(5, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if taken {
		t.Fatal("current episode should be excluded")
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
	if err := store.SetResumePosition(id, 123.5); err != nil {
		t.Fatal(err)
	}
	ep, err = store.GetEpisode(id)
	if err != nil {
		t.Fatal(err)
	}
	if ep.ResumePosition != 123.5 {
		t.Fatalf("resume = %v", ep.ResumePosition)
	}
	_ = sql.NullInt64{}
}

func TestEpisodeByDisplayTitlePrefersBound(t *testing.T) {
	store := openTestStore(t)
	title := "Re Zero kara Hajimeru Isekai Seikatsu — Episode 80"
	if _, err := store.InsertEpisode(Episode{
		FilePath:     "/tmp/unbound.mkv",
		DisplayTitle: title,
		Status:       "COMPLETED",
	}); err != nil {
		t.Fatal(err)
	}
	boundID, err := store.InsertEpisode(Episode{
		FilePath:     "/tmp/bound.mkv",
		DisplayTitle: title,
		Status:       "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertAnime(Anime{AnilistID: 7, TitleRomaji: "Re:Zero"}); err != nil {
		t.Fatal(err)
	}
	if err := store.BindEpisode(boundID, 7, 80); err != nil {
		t.Fatal(err)
	}
	got, err := store.EpisodeByDisplayTitle(title)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != boundID || !got.AnilistID.Valid {
		t.Fatalf("got = %+v", got)
	}
}

func TestAPICacheTTL(t *testing.T) {
	store := openTestStore(t)
	if err := store.SetAPICache("watching", `[{"mediaId":1}]`); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetAPICache("watching", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got != `[{"mediaId":1}]` {
		t.Fatalf("got %q", got)
	}

	_, err = store.db.Exec(
		`UPDATE api_cache SET fetched_at = ? WHERE cache_key = ?`,
		time.Now().Add(-time.Hour).Unix(),
		"watching",
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetAPICache("watching", time.Minute)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired err = %v, want ErrNotFound", err)
	}

	stale, err := store.GetAPICache("watching", 0)
	if err != nil {
		t.Fatal(err)
	}
	if stale != `[{"mediaId":1}]` {
		t.Fatalf("stale = %q", stale)
	}

	if err := store.DeleteAPICache("watching"); err != nil {
		t.Fatal(err)
	}
	_, err = store.GetAPICache("watching", 0)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted err = %v, want ErrNotFound", err)
	}
}
