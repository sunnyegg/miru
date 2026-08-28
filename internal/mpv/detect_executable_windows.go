//go:build windows

package mpv

import (
	"os"
	"path/filepath"
	"strings"
)

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".exe", ".bat", ".cmd":
		return true
	}

	return info.Mode().IsRegular()
}
