//go:build !windows

package mpv

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIpcEndpointFormat(t *testing.T) {
	first := ipcEndpoint()
	second := ipcEndpoint()

	dir := filepath.Dir(first)
	if dir != os.TempDir() {
		t.Errorf("ipcEndpoint dir = %q, want temp dir %q", dir, os.TempDir())
	}
	if ext := filepath.Ext(first); ext != ".sock" {
		t.Errorf("ipcEndpoint ext = %q, want .sock", ext)
	}
	if !strings.Contains(filepath.Base(first), "miru-mpv-") {
		t.Errorf("ipcEndpoint base = %q, want miru-mpv- prefix", filepath.Base(first))
	}
	if first == second {
		t.Errorf("ipcEndpoint returned same path twice: %q", first)
	}
}

func TestWaitSocketConnects(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.sock")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}()

	conn, err := waitSocket(path, 2*time.Second)
	if err != nil {
		t.Fatalf("waitSocket: %v", err)
	}
	defer conn.Close()
}

func TestWaitSocketTimeout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.sock")

	conn, err := waitSocket(path, 150*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Fatal("waitSocket on missing socket: want error, got nil")
	}
	if conn != nil {
		t.Fatalf("waitSocket on missing socket: want nil conn, got %v", conn)
	}
}
