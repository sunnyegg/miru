//go:build !windows

package mpv

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsExecutableUnix(t *testing.T) {
	dir := t.TempDir()

	executablePath := filepath.Join(dir, "mpv")
	if err := os.WriteFile(executablePath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !isExecutable(executablePath) {
		t.Fatal("expected executable file to pass")
	}

	nonExecutablePath := filepath.Join(dir, "mpv-noexec")
	if err := os.WriteFile(nonExecutablePath, []byte("bin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if isExecutable(nonExecutablePath) {
		t.Fatal("expected non-executable file to fail")
	}

	if isExecutable(dir) {
		t.Fatal("expected directory to fail")
	}
}

func TestDetectFindsCommonLinuxPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only")
	}

	dir := t.TempDir()
	fake := filepath.Join(dir, "mpv")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	saved := commonLinuxPaths
	commonLinuxPaths = []string{fake}
	t.Cleanup(func() {
		commonLinuxPaths = saved
	})

	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	got, err := Detect("")
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("got %s want %s", got, fake)
	}
}

func TestDetectFindsCommonDarwinPath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only")
	}

	dir := t.TempDir()
	fake := filepath.Join(dir, "mpv")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	saved := commonDarwinPaths
	commonDarwinPaths = []string{fake}
	t.Cleanup(func() {
		commonDarwinPaths = saved
	})

	t.Setenv("PATH", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	got, err := Detect("")
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("got %s want %s", got, fake)
	}
}
