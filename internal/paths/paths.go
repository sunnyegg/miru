package paths

import (
	"os"
	"path/filepath"
)

const appName = "miru"

type Dirs struct {
	Config string
	Cache  string
}

func Resolve() (Dirs, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return Dirs{}, err
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Dirs{}, err
	}

	dirs := Dirs{
		Config: filepath.Join(configRoot, appName),
		Cache:  filepath.Join(cacheRoot, appName),
	}
	if err := os.MkdirAll(dirs.Config, 0o700); err != nil {
		return Dirs{}, err
	}
	if err := os.MkdirAll(dirs.Cache, 0o700); err != nil {
		return Dirs{}, err
	}
	return dirs, nil
}

func (d Dirs) DatabaseFile() string {
	return filepath.Join(d.Config, "app_data.db")
}

func (d Dirs) TokenFile() string {
	return filepath.Join(d.Config, "anilist.token")
}

func DefaultDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Downloads", appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
