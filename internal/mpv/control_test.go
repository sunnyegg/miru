package mpv

import (
	"bufio"
	"encoding/json"
	"net"
	"slices"
	"testing"
)

func TestSeekSendsAbsoluteExactCommand(t *testing.T) {
	player, commands := runFakeMpv(t, nil)

	if err := player.Seek(42.5); err != nil {
		t.Fatal(err)
	}

	command := <-commands
	if len(command) != 3 || command[0] != "seek" || command[1] != 42.5 || command[2] != "absolute+exact" {
		t.Fatalf("seek command = %v", command)
	}
}

func TestSetPausedSendsProperty(t *testing.T) {
	player, commands := runFakeMpv(t, nil)

	if err := player.SetPaused(true); err != nil {
		t.Fatal(err)
	}

	if command := <-commands; !slices.Equal(command, []any{"set_property", "pause", true}) {
		t.Fatalf("pause command = %v", command)
	}
}

func TestSetVolumeSendsProperty(t *testing.T) {
	player, commands := runFakeMpv(t, nil)

	if err := player.SetVolume(35); err != nil {
		t.Fatal(err)
	}

	if command := <-commands; !slices.Equal(command, []any{"set_property", "volume", float64(35)}) {
		t.Fatalf("volume command = %v", command)
	}
}

func TestSetFullscreenSendsProperty(t *testing.T) {
	player, commands := runFakeMpv(t, nil)

	if err := player.SetFullscreen(false); err != nil {
		t.Fatal(err)
	}

	if command := <-commands; !slices.Equal(command, []any{"set_property", "fullscreen", false}) {
		t.Fatalf("fullscreen command = %v", command)
	}
}

func TestPausedReadsBoolProperty(t *testing.T) {
	player, commands := runFakeMpv(t, map[string]any{"pause": true})

	paused, err := player.Paused()
	if err != nil {
		t.Fatal(err)
	}
	if !paused {
		t.Fatal("paused = false, want true")
	}
	if command := <-commands; !slices.Equal(command, []any{"get_property", "pause"}) {
		t.Fatalf("command = %v", command)
	}
}

func TestStateSnapshot(t *testing.T) {
	values := map[string]any{
		"time-pos":   12.0,
		"duration":   600.0,
		"pause":      true,
		"volume":     80.0,
		"fullscreen": true,
	}
	player, _ := runFakeMpv(t, values)

	state, err := player.State()
	if err != nil {
		t.Fatal(err)
	}
	if state.Position != 12 || state.Duration != 600 {
		t.Fatalf("state = %+v", state)
	}
	if state.Percent != 2 || !state.Paused || state.Volume != 80 || !state.Fullscreen {
		t.Fatalf("state = %+v", state)
	}
}

// runFakeMpv serves one player over an in-memory pipe and reports the commands
// it receives.
func runFakeMpv(t *testing.T, values map[string]any) (*Player, <-chan []any) {
	t.Helper()
	client, server := net.Pipe()
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
	})

	commands := make(chan []any, 16)
	go func() {
		reader := bufio.NewReader(server)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			var request struct {
				Command   []any `json:"command"`
				RequestID int   `json:"request_id"`
			}
			if err := json.Unmarshal(line, &request); err != nil {
				continue
			}
			commands <- request.Command
			reply := map[string]any{"request_id": request.RequestID, "error": "success"}
			if len(request.Command) >= 2 && request.Command[0] == "get_property" {
				name, _ := request.Command[1].(string)
				if value, ok := values[name]; ok {
					reply["data"] = value
				}
			}
			payload, err := json.Marshal(reply)
			if err != nil {
				return
			}
			if _, err := server.Write(append(payload, '\n')); err != nil {
				return
			}
		}
	}()

	return &Player{conn: client}, commands
}
