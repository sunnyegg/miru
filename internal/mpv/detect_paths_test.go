package mpv

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestBuildWindowsCommonPaths(t *testing.T) {
	homeDirectory := t.TempDir()
	t.Setenv("ProgramFiles", `C:\Program Files`)
	t.Setenv("ProgramFiles(x86)", `C:\Program Files (x86)`)
	t.Setenv("ProgramData", `C:\ProgramData`)
	t.Setenv("HOME", homeDirectory)

	paths := buildWindowsCommonPaths()

	want := []string{
		filepath.Join(`C:\Program Files`, "mpv", "mpv.exe"),
		filepath.Join(`C:\Program Files (x86)`, "mpv", "mpv.exe"),
		filepath.Join(`C:\ProgramData`, "chocolatey", "bin", "mpv.exe"),
		filepath.Join(homeDirectory, "scoop", "apps", "mpv", "current", "mpv.exe"),
		filepath.Join(homeDirectory, "scoop", "shims", "mpv.exe"),
	}

	for _, expectedPath := range want {
		if !slices.Contains(paths, expectedPath) {
			t.Fatalf("missing path %q in %v", expectedPath, paths)
		}
	}
}

func TestBuildWindowsCommonPathsSkipsEmptyEnv(t *testing.T) {
	t.Setenv("ProgramFiles", "")
	t.Setenv("ProgramFiles(x86)", "")
	t.Setenv("ProgramData", "")

	homeDirectory := t.TempDir()
	t.Setenv("HOME", homeDirectory)

	paths := buildWindowsCommonPaths()

	for _, candidate := range paths {
		if candidate == "" {
			t.Fatal("expected no empty paths")
		}
	}

	wantScoop := filepath.Join(homeDirectory, "scoop", "shims", "mpv.exe")
	if !slices.Contains(paths, wantScoop) {
		t.Fatalf("missing scoop path %q in %v", wantScoop, paths)
	}
}
