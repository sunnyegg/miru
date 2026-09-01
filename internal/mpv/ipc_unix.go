//go:build !windows

package mpv

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// ipcEndpoint returns the unix socket path mpv should listen on.
func ipcEndpoint() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("miru-mpv-%d.sock", time.Now().UnixNano()))
}

// removeEndpoint deletes a stale socket left behind by a crashed player.
func removeEndpoint(path string) {
	_ = os.Remove(path)
}

// dialIPC makes one connect attempt against the mpv IPC endpoint.
func dialIPC(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", addr, timeout)
}
