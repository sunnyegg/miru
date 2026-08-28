package mpv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var commonLinuxPaths = []string{
	"/usr/bin/mpv",
	"/usr/local/bin/mpv",
	"/snap/bin/mpv",
}

func Detect(manual string) (string, error) {
	if manual != "" {
		if isExecutable(manual) {
			return manual, nil
		}
	}

	if path, err := exec.LookPath("mpv"); err == nil && isExecutable(path) {
		return path, nil
	}

	if runtime.GOOS == "linux" {
		if home, err := os.UserHomeDir(); err == nil {
			candidate := filepath.Join(home, ".local", "bin", "mpv")
			if isExecutable(candidate) {
				return candidate, nil
			}
		}
		for _, candidate := range commonLinuxPaths {
			if isExecutable(candidate) {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("mpv not found; install it or pick the binary in Settings")
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}
