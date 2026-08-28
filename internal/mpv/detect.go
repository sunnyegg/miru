package mpv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Detect(manual string) (string, error) {
	if manual != "" {
		if isExecutable(manual) {
			return manual, nil
		}
	}

	if path, err := exec.LookPath("mpv"); err == nil && isExecutable(path) {
		return path, nil
	}

	for _, candidate := range commonPaths() {
		if isExecutable(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("mpv not found; install it or pick the binary in Settings")
}

func commonPaths() []string {
	switch runtime.GOOS {
	case "linux":
		var paths []string
		if homeDirectory, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(homeDirectory, ".local", "bin", "mpv"))
		}
		return append(paths, commonLinuxPaths...)
	case "windows":
		return commonWindowsPaths
	case "darwin":
		return commonDarwinPaths
	default:
		return nil
	}
}
