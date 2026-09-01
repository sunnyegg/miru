//go:build windows

package mpv

import (
	"strings"
	"testing"
)

func TestIpcEndpointFormat(t *testing.T) {
	first := ipcEndpoint()
	second := ipcEndpoint()

	if !strings.HasPrefix(first, `\\.\pipe\miru-mpv-`) {
		t.Errorf("ipcEndpoint = %q, want \\.\\pipe\\miru-mpv- prefix", first)
	}
	if first == second {
		t.Errorf("ipcEndpoint returned same pipe name twice: %q", first)
	}
}
