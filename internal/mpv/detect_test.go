package mpv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPercent(t *testing.T) {
	if got := Percent(85, 100); got != 85 {
		t.Fatalf("got %v", got)
	}
	if got := Percent(10, 0); got != 0 {
		t.Fatalf("zero duration: %v", got)
	}
	if got := Percent(200, 100); got != 100 {
		t.Fatalf("clamped: %v", got)
	}
}

func TestDetectFindsPATH(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "mpv")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	got, err := Detect("")
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("got %s want %s", got, fake)
	}
}

func TestDetectManualPath(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "custom-mpv")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Detect(fake)
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("got %s", got)
	}
}

func TestDetectMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	saved := commonLinuxPaths
	commonLinuxPaths = nil
	t.Cleanup(func() {
		commonLinuxPaths = saved
	})

	if _, err := Detect(""); err == nil {
		t.Fatal("expected error")
	}
}
