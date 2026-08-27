package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Progress struct {
	Position float64 `json:"position"`
	Duration float64 `json:"duration"`
	Percent  float64 `json:"percent"`
}

type Player struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	conn   net.Conn
	socket string
	reqID  int
}

func Percent(position, duration float64) float64 {
	if duration <= 0 {
		return 0
	}
	p := (position / duration) * 100
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

const minResumeSeconds = 5
const endRemainingSeconds = 10

// ResumePosition is the start offset for the next play. Finished (or nearly
// finished) watches start over so AniList-synced episodes do not reopen on credits.
func ResumePosition(position, duration, percent, threshold float64) float64 {
	if percent >= threshold {
		return 0
	}
	if position < minResumeSeconds {
		return 0
	}
	if duration > 0 && duration-position <= endRemainingSeconds {
		return 0
	}
	return position
}

func (p *Player) Playing() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmd != nil && p.cmd.Process != nil
}

func (p *Player) Stop() {
	p.mu.Lock()
	cmd := p.cmd
	conn := p.conn
	socket := p.socket
	p.cmd = nil
	p.conn = nil
	p.socket = ""
	p.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	if socket != "" {
		_ = os.Remove(socket)
	}
}

func (p *Player) Play(mpvPath, mediaPath string, startSeconds float64, onProgress func(Progress), onExit func(error)) error {
	if mpvPath == "" {
		return fmt.Errorf("mpv path is empty")
	}
	if _, err := os.Stat(mediaPath); err != nil {
		return fmt.Errorf("media file: %w", err)
	}

	p.Stop()

	socket := filepath.Join(os.TempDir(), fmt.Sprintf("miru-mpv-%d.sock", time.Now().UnixNano()))
	_ = os.Remove(socket)

	args := []string{
		"--input-ipc-server=" + socket,
		"--force-window=yes",
		"--no-terminal",
		"--keep-open=yes",
	}
	if startSeconds > 0 {
		args = append(args, fmt.Sprintf("--start=%.3f", startSeconds))
	}
	args = append(args, mediaPath)

	cmd := exec.Command(mpvPath, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}

	conn, err := waitSocket(socket, 4*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		_ = os.Remove(socket)
		return fmt.Errorf("mpv ipc: %w", err)
	}

	p.mu.Lock()
	p.cmd = cmd
	p.conn = conn
	p.socket = socket
	p.mu.Unlock()

	go p.watch(cmd, socket, onProgress, onExit)
	return nil
}

func (p *Player) watch(cmd *exec.Cmd, socket string, onProgress func(Progress), onExit func(error)) {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case err := <-done:
			p.cleanup(cmd, socket)
			if onExit != nil {
				onExit(err)
			}
			return
		case <-ticker.C:
			pos, dur, err := p.properties()
			if err != nil {
				continue
			}
			if onProgress != nil {
				onProgress(Progress{
					Position: pos,
					Duration: dur,
					Percent:  Percent(pos, dur),
				})
			}
		}
	}
}

func (p *Player) cleanup(cmd *exec.Cmd, socket string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != cmd {
		_ = os.Remove(socket)
		return
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	p.cmd = nil
	p.socket = ""
	_ = os.Remove(socket)
}

func (p *Player) properties() (float64, float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn == nil {
		return 0, 0, fmt.Errorf("ipc closed")
	}
	pos, err := p.getNumber("time-pos")
	if err != nil {
		return 0, 0, err
	}
	dur, err := p.getNumber("duration")
	if err != nil {
		return 0, 0, err
	}
	return pos, dur, nil
}

func (p *Player) getNumber(name string) (float64, error) {
	p.reqID++
	req := map[string]any{
		"command":    []any{"get_property", name},
		"request_id": p.reqID,
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	_ = p.conn.SetDeadline(time.Now().Add(800 * time.Millisecond))
	if _, err := p.conn.Write(append(payload, '\n')); err != nil {
		return 0, err
	}

	reader := bufio.NewReader(p.conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return 0, err
		}
		var resp struct {
			RequestID int      `json:"request_id"`
			Error     string   `json:"error"`
			Data      *float64 `json:"data"`
		}
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		if resp.RequestID != p.reqID {
			continue
		}
		if resp.Error != "" && resp.Error != "success" {
			return 0, fmt.Errorf("mpv: %s", resp.Error)
		}
		if resp.Data == nil {
			return 0, fmt.Errorf("mpv: empty %s", name)
		}
		return *resp.Data, nil
	}
}

func waitSocket(path string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
		if err == nil {
			return conn, nil
		}
		last = err
		time.Sleep(50 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("socket timeout")
	}
	return nil, last
}
