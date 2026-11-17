package mpv

import (
	"bufio"
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"
)

func TestPropertiesFetchesDurationOnce(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	counts := &propertyCounts{}
	go servePropertyReplies(server, func(property string) float64 {
		counts.add(property)
		if property == "duration" {
			return 1440
		}
		return 12
	})

	player := &Player{conn: client}
	if state, err := player.media(); err != nil {
		t.Fatal(err)
	} else if state.Duration != 1440 {
		t.Fatalf("duration = %v, want 1440", state.Duration)
	}
	if state, err := player.media(); err != nil {
		t.Fatal(err)
	} else if state.Duration != 1440 {
		t.Fatalf("cached duration = %v, want 1440", state.Duration)
	}

	if got := counts.get("time-pos"); got != 2 {
		t.Fatalf("time-pos fetches = %d, want 2", got)
	}
	if got := counts.get("duration"); got != 1 {
		t.Fatalf("duration fetches = %d, want 1", got)
	}
}

func TestPropertiesRetriesZeroDuration(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	counts := &propertyCounts{}
	go servePropertyReplies(server, func(property string) float64 {
		n := counts.add(property)
		if property == "duration" && n == 1 {
			return 0
		}
		if property == "duration" {
			return 1440
		}
		return 5
	})

	player := &Player{conn: client}
	if state, err := player.media(); err != nil {
		t.Fatal(err)
	} else if state.Duration != 0 {
		t.Fatalf("first duration = %v, want 0", state.Duration)
	}
	if state, err := player.media(); err != nil {
		t.Fatal(err)
	} else if state.Duration != 1440 {
		t.Fatalf("second duration = %v, want 1440", state.Duration)
	}

	if got := counts.get("duration"); got != 2 {
		t.Fatalf("duration fetches = %d, want 2", got)
	}
}

func TestMediaRefetchesDurationAfterPlaylistAdvance(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	counts := &propertyCounts{}
	go servePropertyReplies(server, func(property string) float64 {
		n := counts.add(property)
		switch property {
		case "playlist-pos":
			if n == 1 {
				return 0
			}
			return 1
		case "duration":
			if n == 1 {
				return 1440
			}
			return 480
		}
		return 12
	})

	player := &Player{conn: client}
	if state, err := player.media(); err != nil {
		t.Fatal(err)
	} else if state.Duration != 1440 || state.PlaylistPos != 0 {
		t.Fatalf("first entry = %+v", state)
	}
	if state, err := player.media(); err != nil {
		t.Fatal(err)
	} else if state.Duration != 480 || state.PlaylistPos != 1 {
		t.Fatalf("next entry must refetch its duration, got %+v", state)
	}
	if got := counts.get("duration"); got != 2 {
		t.Fatalf("duration fetches = %d, want 2", got)
	}
}

func TestStopUnblocksStalledProperties(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()

	received := make(chan struct{})
	go func() {
		reader := bufio.NewReader(server)
		_, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		close(received)
		_, _ = reader.ReadBytes('\n')
	}()

	player := &Player{conn: client}
	errCh := make(chan error, 1)
	go func() {
		_, err := player.media()
		errCh <- err
	}()

	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("properties did not send an IPC request")
	}

	started := time.Now()
	player.Stop()
	elapsed := time.Since(started)
	if elapsed > 400*time.Millisecond {
		t.Fatalf("Stop took %v; stalled IPC should not hold the player mutex", elapsed)
	}

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("properties: want error after Stop, got nil")
		}
	case <-time.After(time.Second):
		t.Fatal("properties still blocked after Stop")
	}
}

type propertyCounts struct {
	mu     sync.Mutex
	values map[string]int
}

func (c *propertyCounts) add(property string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values == nil {
		c.values = map[string]int{}
	}
	c.values[property]++
	return c.values[property]
}

func (c *propertyCounts) get(property string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.values[property]
}

func servePropertyReplies(conn net.Conn, valueFor func(property string) float64) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}
		var req struct {
			Command   []any `json:"command"`
			RequestID int   `json:"request_id"`
		}
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		if len(req.Command) < 2 {
			continue
		}
		property, _ := req.Command[1].(string)
		value := valueFor(property)
		resp, err := json.Marshal(map[string]any{
			"request_id": req.RequestID,
			"error":      "success",
			"data":       value,
		})
		if err != nil {
			return
		}
		if _, err := conn.Write(append(resp, '\n')); err != nil {
			return
		}
	}
}
