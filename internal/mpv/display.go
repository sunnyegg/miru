package mpv

import "strings"

// useXWayland reports whether mpv should be pinned to XWayland. Mutter gives
// native Wayland clients no server side decorations and mpv draws none of its
// own, so without this the player window has no title bar, resize handles or
// close button.
func useXWayland(environ []string, nativeWayland bool) bool {
	if nativeWayland {
		return false
	}
	return hasEnv(environ, "WAYLAND_DISPLAY") && hasEnv(environ, "DISPLAY")
}

func hasEnv(environ []string, key string) bool {
	prefix := key + "="
	for _, entry := range environ {
		if strings.HasPrefix(entry, prefix) && len(entry) > len(prefix) {
			return true
		}
	}
	return false
}
