package datareset

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sunnyegg/miru/internal/paths"
)

func TestMeasureCountsActiveAndStagedDataWithoutFollowingSymlinks(t *testing.T) {
	dirs := testDirs(t)
	writeTestFile(t, filepath.Join(dirs.Config, "app_data.db"), "database")
	writeTestFile(t, filepath.Join(dirs.Cache, "nested", "cache"), "cached")
	writeTestFile(t, filepath.Join(dirs.Data, "Miru.log"), "log")

	if runtime.GOOS != "windows" {
		outside := filepath.Join(t.TempDir(), "outside")
		writeTestFile(t, outside, strings.Repeat("x", 100))
		if err := os.Symlink(outside, filepath.Join(dirs.Cache, "outside-link")); err != nil {
			t.Fatal(err)
		}
	}

	usage, err := Measure(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if usage.Bytes < int64(len("databasecachedlog")) {
		t.Fatalf("measured bytes = %d", usage.Bytes)
	}
	if usage.CleanupPending {
		t.Fatal("cleanup unexpectedly pending")
	}
}

func TestSchedulePrepareAndCommit(t *testing.T) {
	dirs := testDirs(t)
	writeTestFile(t, filepath.Join(dirs.Config, "app_data.db"), "old database")
	writeTestFile(t, filepath.Join(dirs.Cache, "response"), "old cache")
	writeTestFile(t, filepath.Join(dirs.Data, "Miru.log"), "old log")

	if err := Schedule(dirs); err != nil {
		t.Fatal(err)
	}
	if err := Schedule(dirs); err == nil {
		t.Fatal("second reset schedule succeeded")
	}
	state, err := Prepare(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if !state.NeedsCommit || state.Blocked || state.CleanupPending {
		t.Fatalf("unexpected startup state: %+v", state)
	}
	resumed, err := Prepare(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if !resumed.NeedsCommit || resumed.Blocked || resumed.CleanupPending {
		t.Fatalf("unexpected resumed startup state: %+v", resumed)
	}
	for _, path := range []string{dirs.Config, dirs.Cache, dirs.Data} {
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("fresh root %q is not empty", path)
		}
	}
	if err := Commit(dirs); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(markerPath(dirs)); !os.IsNotExist(err) {
		t.Fatalf("marker still exists: %v", err)
	}
	usage, err := Measure(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if usage.Bytes != 0 || usage.CleanupPending {
		t.Fatalf("usage after reset = %+v", usage)
	}
}

func TestPrepareRollsBackStagedRoots(t *testing.T) {
	dirs := testDirs(t)
	writeTestFile(t, filepath.Join(dirs.Config, "app_data.db"), "database")
	writeTestFile(t, filepath.Join(dirs.Cache, "cache"), "cache")

	if err := Schedule(dirs); err != nil {
		t.Fatal(err)
	}
	current, err := readMarker(dirs, mustRoots(t, dirs))
	if err != nil {
		t.Fatal(err)
	}
	cacheBackup := dirs.Cache + ".reset-" + current.ResetID
	writeTestFile(t, filepath.Join(cacheBackup, "collision"), "occupied")

	state, err := Prepare(dirs)
	if err == nil {
		t.Fatal("prepare unexpectedly succeeded")
	}
	if state.Blocked {
		t.Fatalf("recoverable staging failure was blocked: %v", err)
	}
	if got, err := os.ReadFile(filepath.Join(dirs.Config, "app_data.db")); err != nil || string(got) != "database" {
		t.Fatalf("config data was not restored: %q, %v", got, err)
	}
	if _, err := os.Stat(markerPath(dirs)); !os.IsNotExist(err) {
		t.Fatalf("marker still exists after rollback: %v", err)
	}
}

func TestPrepareRecoversRenameBeforeMarkerUpdate(t *testing.T) {
	dirs := testDirs(t)
	writeTestFile(t, filepath.Join(dirs.Config, "app_data.db"), "database")
	if err := Schedule(dirs); err != nil {
		t.Fatal(err)
	}
	roots := mustRoots(t, dirs)
	current, err := readMarker(dirs, roots)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(dirs.Config, dirs.Config+".reset-"+current.ResetID); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dirs.Config, 0o700); err != nil {
		t.Fatal(err)
	}

	state, err := Prepare(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if !state.NeedsCommit {
		t.Fatalf("unexpected startup state: %+v", state)
	}
	if err := Commit(dirs); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareBlocksMalformedMarker(t *testing.T) {
	dirs := testDirs(t)
	writeTestFile(t, markerPath(dirs), "not-json")

	state, err := Prepare(dirs)
	if err == nil || !state.Blocked {
		t.Fatalf("malformed marker result = %+v, %v", state, err)
	}
}

func TestPrepareKeepsUnsafeCleanupBackup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic link setup requires additional Windows privileges")
	}
	dirs := testDirs(t)
	roots := mustRoots(t, dirs)
	current := marker{
		SchemaVersion: markerSchemaVersion,
		ResetID:       strings.Repeat("a", 32),
		Phase:         phaseCleanupPending,
		StagedRoots:   []string{"config"},
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, backupPath(roots[0], current.ResetID)); err != nil {
		t.Fatal(err)
	}
	if err := writeMarker(dirs, current); err != nil {
		t.Fatal(err)
	}

	state, err := Prepare(dirs)
	if err == nil || !state.CleanupPending || state.Blocked {
		t.Fatalf("unsafe cleanup result = %+v, %v", state, err)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("symlink target was touched: %v", err)
	}
	if err := os.Remove(backupPath(roots[0], current.ResetID)); err != nil {
		t.Fatal(err)
	}
	state, err = Prepare(dirs)
	if err != nil || state != (StartupState{}) {
		t.Fatalf("cleanup retry result = %+v, %v", state, err)
	}
	if _, err := os.Stat(markerPath(dirs)); !os.IsNotExist(err) {
		t.Fatalf("marker still exists after cleanup retry: %v", err)
	}
}

func TestDuplicateRootsAreProcessedOnce(t *testing.T) {
	dirs := testDirs(t)
	dirs.Data = dirs.Config
	writeTestFile(t, filepath.Join(dirs.Config, "shared"), "12345")

	usage, err := Measure(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if usage.Bytes != 5 {
		t.Fatalf("duplicate root measured %d bytes", usage.Bytes)
	}
	if got := len(mustRoots(t, dirs)); got != 2 {
		t.Fatalf("unique roots = %d", got)
	}
	if err := Schedule(dirs); err != nil {
		t.Fatal(err)
	}
	state, err := Prepare(dirs)
	if err != nil || !state.NeedsCommit {
		t.Fatalf("duplicate root staging result = %+v, %v", state, err)
	}
	if err := Commit(dirs); err != nil {
		t.Fatal(err)
	}
}

func TestScheduleRejectsUnsafeRoot(t *testing.T) {
	dirs := testDirs(t)
	dirs.Cache = t.TempDir()
	if err := Schedule(dirs); err == nil {
		t.Fatal("unsafe root was accepted")
	}
}

func testDirs(t *testing.T) paths.Dirs {
	t.Helper()
	base := t.TempDir()
	dirs := paths.Dirs{
		Config: filepath.Join(base, "config", "miru"),
		Cache:  filepath.Join(base, "cache", "miru"),
		Data:   filepath.Join(base, "data", "miru"),
	}
	for _, path := range []string{dirs.Config, dirs.Cache, dirs.Data} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return dirs
}

func mustRoots(t *testing.T, dirs paths.Dirs) []resetRoot {
	t.Helper()
	roots, err := rootsFor(dirs)
	if err != nil {
		t.Fatal(err)
	}
	return roots
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
