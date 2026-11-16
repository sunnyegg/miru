package mpv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configTemplate = `# MPV options used only by Miru playback.
#
# Miru passes this file with --include=, and its own window flags come after
# it, so options here win over your global mpv.conf. Remove a line to fall
# back to mpv defaults. Delete this file to have Miru write it again.
#
# Window examples:
# autofit=70%x70%      initial window size
# geometry=50%:50%     initial window position
# border=no            undecorated window (drag the video to move it)
# ontop=yes            keep the player above other windows
# keepaspect-window=no allow a freely resizable window
#
# Image examples:
# hwdec=auto-safe
# deband=yes
# sub-font-size=44
`

// ConfigPath is the mpv config file Miru includes for its own playback.
func ConfigPath(configDir string) string {
	return filepath.Join(configDir, "mpv.conf")
}

// EnsureConfig writes the default Miru mpv config once and returns its path.
// A non-empty existing file is kept, so user edits survive.
func EnsureConfig(configDir string) (string, error) {
	if configDir == "" {
		return "", errors.New("mpv config dir is empty")
	}
	path := ConfigPath(configDir)
	info, err := os.Stat(path)
	if err == nil && info.Size() > 0 {
		return path, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read mpv config: %w", err)
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return "", fmt.Errorf("create mpv config folder: %w", err)
	}
	if err := os.WriteFile(path, []byte(configTemplate), 0o600); err != nil {
		return "", fmt.Errorf("write mpv config: %w", err)
	}
	return path, nil
}
