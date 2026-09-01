//go:build windows

package mpv

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// ipcEndpoint returns the named pipe address mpv should listen on.
// mpv's --input-ipc-server uses named pipes on Windows, not unix sockets.
func ipcEndpoint() string {
	return fmt.Sprintf(`\\.\pipe\miru-mpv-%d`, time.Now().UnixNano())
}

// removeEndpoint is a no-op: named pipes disappear with the mpv process.
func removeEndpoint(string) {}

// dialIPC makes one connect attempt against the mpv IPC endpoint.
func dialIPC(addr string, timeout time.Duration) (net.Conn, error) {
	return winio.DialPipe(addr, &timeout)
}
