package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCreatesDirs(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	dirs, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dirs.Config, "miru") {
		t.Fatalf("config dir = %s, want suffix miru", dirs.Config)
	}
	if _, err := os.Stat(dirs.Config); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dirs.DatabaseFile()); os.IsNotExist(err) {
		// file is not created yet; only the parent dir
	}
	if filepath.Base(dirs.DatabaseFile()) != "app_data.db" {
		t.Fatalf("db file = %s", dirs.DatabaseFile())
	}
}

func TestDefaultDownloadDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir, err := DefaultDownloadDir()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
}
