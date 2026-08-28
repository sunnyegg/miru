//go:build windows

package mpv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsExecutableWindows(t *testing.T) {
	dir := t.TempDir()

	executablePath := filepath.Join(dir, "mpv.exe")
	if err := os.WriteFile(executablePath, []byte("bin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isExecutable(executablePath) {
		t.Fatal("expected .exe file to pass")
	}

	extensionlessPath := filepath.Join(dir, "mpv-portable")
	if err := os.WriteFile(extensionlessPath, []byte("bin"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !isExecutable(extensionlessPath) {
		t.Fatal("expected regular file without execute bit to pass on Windows")
	}

	if isExecutable(dir) {
		t.Fatal("expected directory to fail")
	}
}

func TestDetectFindsCommonWindowsPath(t *testing.T) {
	dir := t.TempDir()
	fake := filepath.Join(dir, "mpv.exe")
	if err := os.WriteFile(fake, []byte("bin"), 0o644); err != nil {
		t.Fatal(err)
	}

	saved := commonWindowsPaths
	commonWindowsPaths = []string{fake}
	t.Cleanup(func() {
		commonWindowsPaths = saved
	})

	t.Setenv("PATH", t.TempDir())

	got, err := Detect("")
	if err != nil {
		t.Fatal(err)
	}
	if got != fake {
		t.Fatalf("got %s want %s", got, fake)
	}
}
