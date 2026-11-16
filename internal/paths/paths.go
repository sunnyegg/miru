package paths

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/sunnyegg/miru/internal/appinfo"
)

type Dirs struct {
	Config string
	Data   string
	Cache  string
}

func Resolve() (Dirs, error) {
	configRoot, err := os.UserConfigDir()
	if err != nil {
		return Dirs{}, err
	}
	dataRoot, err := userDataDir()
	if err != nil {
		return Dirs{}, err
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Dirs{}, err
	}

	dirs := Dirs{
		Config: filepath.Join(configRoot, appinfo.Name),
		Data:   filepath.Join(dataRoot, appinfo.Name),
		Cache:  filepath.Join(cacheRoot, appinfo.Name),
	}
	if err := os.MkdirAll(dirs.Config, 0o700); err != nil {
		return Dirs{}, err
	}
	if err := os.MkdirAll(dirs.Data, 0o700); err != nil {
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

func (d Dirs) ShadersDir() string {
	return filepath.Join(d.Config, "shaders")
}

func (d Dirs) LogFile() string {
	return filepath.Join(d.Data, appinfo.LogName)
}

func DefaultDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Downloads", appinfo.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func userDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(home, ".local", "share"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support"), nil
	case "windows":
		if v := os.Getenv("APPDATA"); v != "" {
			return v, nil
		}
		return filepath.Join(home, "AppData", "Roaming"), nil
	default:
		return home, nil
	}
}
