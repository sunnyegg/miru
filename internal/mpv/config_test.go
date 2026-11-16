package mpv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureConfigCreatesTemplate(t *testing.T) {
	dir := t.TempDir()

	path, err := EnsureConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "mpv.conf"); path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# MPV options used only by Miru playback.") {
		t.Fatalf("unexpected template: %q", data)
	}
}

func TestEnsureConfigKeepsUserEdits(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := ConfigPath(dir)
	if err := os.WriteFile(path, []byte("border=no\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureConfig(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "border=no\n" {
		t.Fatalf("config overwritten: %q", data)
	}
}

func TestEnsureConfigRewritesEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := ConfigPath(dir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureConfig(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("empty config was not rewritten")
	}
}

func TestEnsureConfigRequiresDir(t *testing.T) {
	if _, err := EnsureConfig(""); err == nil {
		t.Fatal("expected error for empty config dir")
	}
}
