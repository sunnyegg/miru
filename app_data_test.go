package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sunnyegg/miru/internal/datareset"
	"github.com/sunnyegg/miru/internal/paths"
)

func TestGetDataSizeIncludesMiruRoots(t *testing.T) {
	dirs := newDataResetDirs(t)
	if err := os.WriteFile(filepath.Join(dirs.Cache, "cached"), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := &App{dirs: dirs}

	view, err := a.GetDataSize()
	if err != nil {
		t.Fatal(err)
	}
	if view.Bytes != 5 || view.CleanupPending || view.ResetError != "" {
		t.Fatalf("data size view = %+v", view)
	}
}

func TestDeleteAllDataSchedulesResetAndForcesQuit(t *testing.T) {
	dirs := newDataResetDirs(t)
	a := &App{dirs: dirs}

	if err := a.DeleteAllData(); err != nil {
		t.Fatal(err)
	}
	if !a.forceQuit.Load() {
		t.Fatal("force quit was not enabled")
	}
	if err := a.DeleteAllData(); err == nil {
		t.Fatal("second reset request succeeded")
	}
	state, err := datareset.Prepare(dirs)
	if err != nil {
		t.Fatal(err)
	}
	if !state.NeedsCommit {
		t.Fatalf("reset was not staged: %+v", state)
	}
	if err := datareset.Commit(dirs); err != nil {
		t.Fatal(err)
	}
}

func TestFinishDataResetStartupCleansBackups(t *testing.T) {
	dirs := newDataResetDirs(t)
	if err := os.WriteFile(filepath.Join(dirs.Config, "old"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := datareset.Schedule(dirs); err != nil {
		t.Fatal(err)
	}
	state, err := datareset.Prepare(dirs)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{}
	a.configureDataResetStartup(dirs, state, nil)
	a.finishDataResetStartup()

	view, err := a.GetDataSize()
	if err != nil {
		t.Fatal(err)
	}
	if view.Bytes != 0 || view.CleanupPending || view.ResetError != "" {
		t.Fatalf("data reset did not finish: %+v", view)
	}
}

func newDataResetDirs(t *testing.T) paths.Dirs {
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
