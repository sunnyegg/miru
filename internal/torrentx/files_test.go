package torrentx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/storage"
)

func TestContentsFromInfoMultiFile(t *testing.T) {
	info := metainfo.Info{
		Name: "Show",
		Files: []metainfo.FileInfo{
			{Path: []string{"01.mkv"}, Length: 100},
			{Path: []string{"readme.txt"}, Length: 20},
		},
	}
	contents := contentsFromInfo(&info)
	if contents.Name != "Show" || contents.BytesTotal != 120 || len(contents.Files) != 2 {
		t.Fatalf("contents = %+v", contents)
	}
	if contents.Files[0].Path != "01.mkv" || !contents.Files[0].IsVideo || !contents.Files[0].Selected {
		t.Fatalf("video file = %+v", contents.Files[0])
	}
	if contents.Files[1].Path != "readme.txt" || contents.Files[1].IsVideo {
		t.Fatalf("text file = %+v", contents.Files[1])
	}
}

func TestContentsFromInfoSingleFile(t *testing.T) {
	info := metainfo.Info{Name: "episode.mkv", Length: 50}
	contents := contentsFromInfo(&info)
	if contents.Name != "episode.mkv" || len(contents.Files) != 1 {
		t.Fatalf("contents = %+v", contents)
	}
	if contents.Files[0].Path != "episode.mkv" || !contents.Files[0].IsVideo {
		t.Fatalf("file = %+v", contents.Files[0])
	}
}

func TestEncodeDecodeFiles(t *testing.T) {
	if got := decodeFiles(""); got != nil {
		t.Fatalf("empty = %v", got)
	}
	raw := encodeFiles([]FileView{
		{Path: "01.mkv", Length: 10},
		{Path: "note.txt", Length: 2},
	})
	got := decodeFiles(raw)
	if len(got) != 2 || got[0].Path != "01.mkv" || !got[0].IsVideo || !got[0].Selected {
		t.Fatalf("got = %+v", got)
	}
	if got[1].IsVideo {
		t.Fatalf("txt marked video: %+v", got[1])
	}
}

func TestSelectedBytesTotal(t *testing.T) {
	if got := selectedBytesTotal([]FileView{{Length: 10}, {Length: 5}}); got != 15 {
		t.Fatalf("total = %d", got)
	}
}

func TestSelectedPathSetEmptyMeansAll(t *testing.T) {
	if selectedPathSet(nil) != nil {
		t.Fatal("nil should mean all files")
	}
	if selectedPathSet([]FileView{}) != nil {
		t.Fatal("empty should mean all files")
	}
	wanted := selectedPathSet([]FileView{{Path: "01.mkv"}, {Path: "  "}})
	if _, ok := wanted["01.mkv"]; !ok || len(wanted) != 1 {
		t.Fatalf("wanted = %v", wanted)
	}
}

func TestToViewUsesStoredFiles(t *testing.T) {
	view := ToView(storage.TorrentJob{
		ID:        3,
		Name:      "Show",
		Status:    "QUEUED",
		FilesJSON: encodeFiles([]FileView{{Path: "01.mkv", Length: 80}}),
	})
	if view.BytesTotal != 80 || len(view.Files) != 1 || view.Files[0].Path != "01.mkv" {
		t.Fatalf("view = %+v", view)
	}
}

func TestInspectLocalTorrent(t *testing.T) {
	manager, _ := openManager(t)
	torrentPath := writeTestTorrent(t)
	contents, err := manager.Inspect(torrentPath, "", RateLimits{}, networking.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if contents.Name != "Show" {
		t.Fatalf("name = %q", contents.Name)
	}
	if len(contents.Files) != 2 {
		t.Fatalf("files = %+v", contents.Files)
	}
	paths := map[string]FileView{}
	for _, file := range contents.Files {
		paths[file.Path] = file
	}
	if !paths["01.mkv"].IsVideo || paths["01.mkv"].Length != 3 {
		t.Fatalf("video = %+v", paths["01.mkv"])
	}
	if paths["readme.txt"].IsVideo || paths["readme.txt"].Length != 4 {
		t.Fatalf("text = %+v", paths["readme.txt"])
	}
}

func TestStartStoresSelectedFiles(t *testing.T) {
	manager, store := openManager(t)
	t.Cleanup(manager.Close)
	destDir := t.TempDir()
	err := manager.Start(
		"file.torrent",
		destDir,
		[]FileView{{Path: "01.mkv", Length: 100}, {Path: "02.mkv", Length: 50}},
		RateLimits{},
		networking.Config{},
	)
	if err != nil {
		t.Fatal(err)
	}
	jobs, err := store.ListTorrentJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %+v", jobs)
	}
	job := jobs[0]
	if job.BytesTotal != 150 {
		t.Fatalf("bytes total = %d", job.BytesTotal)
	}
	files := decodeFiles(job.FilesJSON)
	if len(files) != 2 || files[0].Path != "01.mkv" || files[1].Path != "02.mkv" {
		t.Fatalf("files = %+v", files)
	}
}

func TestUserStartedTracksSessionStart(t *testing.T) {
	manager, store := openManager(t)
	t.Cleanup(manager.Close)
	destDir := t.TempDir()
	if err := manager.Start("file.torrent", destDir, nil, RateLimits{}, networking.Config{}); err != nil {
		t.Fatal(err)
	}
	jobs, err := store.ListTorrentJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("jobs = %+v", jobs)
	}
	if !manager.UserStarted(jobs[0].ID) {
		t.Fatal("expected user-started download to be tracked")
	}

	recoveredID, err := store.InsertTorrentJob(storage.TorrentJob{
		Source:  "magnet:?xt=urn:btih:recovered",
		DestDir: destDir,
		Name:    "Recovered",
		Status:  "QUEUED",
	})
	if err != nil {
		t.Fatal(err)
	}
	if manager.UserStarted(recoveredID) {
		t.Fatal("expected recovered queued job not to be user-started")
	}
}

func writeTestTorrent(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	showDir := filepath.Join(root, "Show")
	if err := os.MkdirAll(showDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(showDir, "01.mkv"), []byte("vid"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(showDir, "readme.txt"), []byte("note"), 0o644); err != nil {
		t.Fatal(err)
	}

	info := metainfo.Info{PieceLength: 256 * 1024}
	if err := info.BuildFromFilePath(showDir); err != nil {
		t.Fatal(err)
	}
	var meta metainfo.MetaInfo
	meta.SetDefaults()
	raw, err := bencode.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	meta.InfoBytes = raw

	torrentPath := filepath.Join(root, "show.torrent")
	file, err := os.Create(torrentPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := meta.Write(file); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return torrentPath
}
