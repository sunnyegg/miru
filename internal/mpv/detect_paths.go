package mpv

import (
	"os"
	"path/filepath"
)

var commonLinuxPaths = []string{
	"/usr/bin/mpv",
	"/usr/local/bin/mpv",
	"/snap/bin/mpv",
}

var commonDarwinPaths = []string{
	"/opt/homebrew/bin/mpv",
	"/usr/local/bin/mpv",
	"/Applications/mpv.app/Contents/MacOS/mpv",
}

var commonWindowsPaths = buildWindowsCommonPaths()

func buildWindowsCommonPaths() []string {
	var paths []string

	for _, programFilesRoot := range []string{
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
	} {
		if programFilesRoot == "" {
			continue
		}
		paths = append(paths, filepath.Join(programFilesRoot, "mpv", "mpv.exe"))
	}

	if programDataRoot := os.Getenv("ProgramData"); programDataRoot != "" {
		paths = append(paths, filepath.Join(programDataRoot, "chocolatey", "bin", "mpv.exe"))
	}

	if homeDirectory, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(homeDirectory, "scoop", "apps", "mpv", "current", "mpv.exe"),
			filepath.Join(homeDirectory, "scoop", "shims", "mpv.exe"),
		)
	}

	return paths
}
