package mpv

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the pair of mpv config files Miru passes to the player. The
// defaults file is Miru-managed and included first, so the user file can
// override any option in it.
type Config struct {
	DefaultsPath string
	UserPath     string
}

const defaultsTemplate = `# MPV defaults managed by Miru. Miru rewrites this file on launch.
# Override any option in mpv.conf, which is included after this file.

# Resize the window freely instead of locking it to the video aspect ratio.
keepaspect-window=no

# Keep the OSC controls at a readable size instead of shrinking them with the
# window. Tune with script-opts=osc-scalewindowed=<factor> in mpv.conf, and pin
# OSD text too with osd-scale-by-window=no.
script-opts=osc-vidscale=no
`

const userTemplate = `# MPV options used only by Miru playback.
#
# Included after Miru's own defaults, so options here win. Miru's window
# flags (title, GPU context, IPC socket) still win over both files. Miru
# writes this file only when it is missing or empty.
#
# Window examples:
# autofit=70%x70%        initial window size
# geometry=50%:50%       initial window position
# border=no              undecorated window (drag the video to move it)
# ontop=yes              keep the player above other windows
# keepaspect-window=yes  lock the window to the video aspect ratio
# osd-scale-by-window=yes  let OSD text shrink with the window
# script-opts=osc-vidscale=auto  let the OSC scale with the window
# osd-font-size=40       OSD text size in pixels (fixed while scaling is off)
#
# Image examples:
# hwdec=auto-safe
# deband=yes
# sub-font-size=44
`

// DefaultsPath is the Miru-managed mpv config, included before the user file.
func DefaultsPath(configDir string) string {
	return filepath.Join(configDir, "mpv-defaults.conf")
}

// ConfigPath is the user-editable mpv config Miru includes for its playback.
func ConfigPath(configDir string) string {
	return filepath.Join(configDir, "mpv.conf")
}

// EnsureConfig prepares both mpv config files and returns their paths. The
// defaults file follows Miru's template, the user file only gets the template
// when it is missing or empty, so user edits survive.
func EnsureConfig(configDir string) (Config, error) {
	if configDir == "" {
		return Config{}, errors.New("mpv config dir is empty")
	}
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return Config{}, fmt.Errorf("create mpv config folder: %w", err)
	}
	config := Config{
		DefaultsPath: DefaultsPath(configDir),
		UserPath:     ConfigPath(configDir),
	}
	if err := writeConfig(config.DefaultsPath, defaultsTemplate, true); err != nil {
		return Config{}, err
	}
	if err := writeConfig(config.UserPath, userTemplate, false); err != nil {
		return Config{}, err
	}
	return config, nil
}

func writeConfig(path, content string, managed bool) error {
	current, err := os.ReadFile(path)
	switch {
	case err == nil && managed && string(current) == content:
		return nil
	case err == nil && !managed && len(current) > 0:
		return nil
	case err != nil && !errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("read %s: %w", filepath.Base(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	return nil
}
