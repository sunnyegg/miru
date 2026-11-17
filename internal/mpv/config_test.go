package mpv

import (
	"os"
	"strings"
	"testing"
)

func TestEnsureConfigCreatesFiles(t *testing.T) {
	dir := t.TempDir()

	config, err := EnsureConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if config.DefaultsPath != DefaultsPath(dir) || config.UserPath != ConfigPath(dir) || config.NextPath != NextPath(dir) {
		t.Fatalf("config = %+v", config)
	}

	script, err := os.ReadFile(config.NextPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(script), "miru-next-activate") {
		t.Fatalf("next script missing its activation hook: %q", script)
	}

	defaults, err := os.ReadFile(config.DefaultsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"keepaspect-window=no", "script-opts=osc-vidscale=no"} {
		if !strings.Contains(string(defaults), want) {
			t.Fatalf("defaults missing %q: %q", want, defaults)
		}
	}

	user, err := os.ReadFile(config.UserPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(user), "# MPV options used only by Miru playback.") {
		t.Fatalf("unexpected user template: %q", user)
	}
}

func TestEnsureConfigKeepsUserEdits(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(ConfigPath(dir), []byte("border=no\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureConfig(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(ConfigPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "border=no\n" {
		t.Fatalf("user config overwritten: %q", data)
	}
}

func TestEnsureConfigRewritesEmptyUserFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(ConfigPath(dir), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureConfig(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(ConfigPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("empty user config was not rewritten")
	}
}

func TestEnsureConfigRestoresManagedDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(DefaultsPath(dir), []byte("keepaspect-window=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureConfig(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(DefaultsPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "keepaspect-window=no") {
		t.Fatalf("managed defaults not restored: %q", data)
	}
}

func TestEnsureConfigRequiresDir(t *testing.T) {
	if _, err := EnsureConfig(""); err == nil {
		t.Fatal("expected error for empty config dir")
	}
}
