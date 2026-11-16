package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Progress struct {
	Position   float64 `json:"position"`
	Duration   float64 `json:"duration"`
	Percent    float64 `json:"percent"`
	Paused     bool    `json:"paused"`
	Volume     int     `json:"volume"`
	Fullscreen bool    `json:"fullscreen"`
}

// Options describe one mpv playback session.
type Options struct {
	MPVPath      string
	MediaPath    string
	Title        string
	StartSeconds float64
	GLSLShaders  []string
	ConfigDir    string
	// NativeWayland keeps mpv on Wayland instead of XWayland. XWayland is the
	// default because GNOME gives native Wayland clients no title bar, resize
	// handles or close button.
	NativeWayland bool
}

type Player struct {
	mu       sync.Mutex
	cmdMu    sync.Mutex
	cmd      *exec.Cmd
	conn     net.Conn
	ipc      string
	reqID    int
	duration float64
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
	ipc := p.ipc
	p.cmd = nil
	p.conn = nil
	p.ipc = ""
	p.duration = 0
	p.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	if ipc != "" {
		removeEndpoint(ipc)
	}
}

// args builds the mpv command line. The Miru config is included first so the
// window options below always win over a user's own mpv.conf.
func (o Options) args(ipc string, config Config, xwayland bool) []string {
	args := []string{"--input-ipc-server=" + ipc}
	for _, path := range []string{config.DefaultsPath, config.UserPath} {
		if path != "" {
			args = append(args, "--include="+path)
		}
	}
	args = append(args,
		"--force-window=yes",
		"--no-terminal",
		"--keep-open=yes",
	)
	if xwayland {
		// Unsetting WAYLAND_DISPLAY is not enough: mpv still falls back to the
		// default wayland-0 socket. The context has to be pinned to X11.
		args = append(args, "--gpu-context=x11egl")
	}
	if o.Title != "" {
		args = append(args, "--title="+o.Title)
	}
	if o.StartSeconds > 0 {
		args = append(args, fmt.Sprintf("--start=%.3f", o.StartSeconds))
	}
	for _, shaderPath := range o.GLSLShaders {
		args = append(args, "--glsl-shader="+shaderPath)
	}
	return append(args, o.MediaPath)
}

func (p *Player) Play(opts Options, onProgress func(Progress), onExit func(error)) error {
	if opts.MPVPath == "" {
		return fmt.Errorf("mpv path is empty")
	}
	if _, err := os.Stat(opts.MediaPath); err != nil {
		return fmt.Errorf("media file: %w", err)
	}

	p.Stop()

	ipc := ipcEndpoint()
	removeEndpoint(ipc)

	xwayland := useXWayland(os.Environ(), opts.NativeWayland)
	config, err := EnsureConfig(opts.ConfigDir)
	if err != nil {
		log.Printf("mpv: %v", err)
		config = Config{}
	}
	if len(opts.GLSLShaders) > 0 {
		log.Printf("mpv: launching with %d Anime4K shader(s)", len(opts.GLSLShaders))
	}

	cmd := exec.Command(opts.MPVPath, opts.args(ipc, config, xwayland)...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}

	conn, err := waitSocket(ipc, 4*time.Second)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		removeEndpoint(ipc)
		return fmt.Errorf("mpv ipc: %w", err)
	}

	p.mu.Lock()
	p.cmd = cmd
	p.conn = conn
	p.ipc = ipc
	p.mu.Unlock()

	go p.watch(cmd, ipc, onProgress, onExit)
	return nil
}

func (p *Player) watch(cmd *exec.Cmd, ipc string, onProgress func(Progress), onExit func(error)) {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case err := <-done:
			p.cleanup(cmd, ipc)
			if onExit != nil {
				onExit(err)
			}
			return
		case <-ticker.C:
			state, err := p.State()
			if err != nil {
				continue
			}
			if onProgress != nil {
				onProgress(state)
			}
		}
	}
}

func (p *Player) cleanup(cmd *exec.Cmd, ipc string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd != cmd {
		removeEndpoint(ipc)
		return
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	p.cmd = nil
	p.ipc = ""
	p.duration = 0
	removeEndpoint(ipc)
}

func (p *Player) properties() (float64, float64, error) {
	p.mu.Lock()
	conn := p.conn
	cachedDuration := p.duration
	p.mu.Unlock()
	if conn == nil {
		return 0, 0, fmt.Errorf("ipc closed")
	}

	position, err := p.getNumber("time-pos")
	if err != nil {
		return 0, 0, err
	}
	if cachedDuration > 0 {
		return position, cachedDuration, nil
	}

	duration, err := p.getNumber("duration")
	if err != nil {
		return 0, 0, err
	}
	if duration <= 0 {
		return position, duration, nil
	}

	p.mu.Lock()
	if p.conn == conn {
		p.duration = duration
	}
	p.mu.Unlock()
	return position, duration, nil
}

// State is the current playback snapshot used by the progress ticker and the
// in-app controls.
func (p *Player) State() (Progress, error) {
	position, duration, err := p.properties()
	if err != nil {
		return Progress{}, err
	}
	paused, err := p.getBool("pause")
	if err != nil {
		return Progress{}, err
	}
	volume, err := p.getNumber("volume")
	if err != nil {
		return Progress{}, err
	}
	fullscreen, err := p.getBool("fullscreen")
	if err != nil {
		return Progress{}, err
	}
	return Progress{
		Position:   position,
		Duration:   duration,
		Percent:    Percent(position, duration),
		Paused:     paused,
		Volume:     int(volume),
		Fullscreen: fullscreen,
	}, nil
}

// Paused reports mpv's pause state, which also changes from mpv's own keys.
func (p *Player) Paused() (bool, error) {
	return p.getBool("pause")
}

func (p *Player) SetPaused(paused bool) error {
	return p.setProperty("pause", paused)
}

func (p *Player) Seek(position float64) error {
	if position < 0 {
		position = 0
	}
	_, err := p.request("seek", position, "absolute+exact")
	return err
}

func (p *Player) SetVolume(volume int) error {
	return p.setProperty("volume", volume)
}

func (p *Player) Fullscreen() (bool, error) {
	return p.getBool("fullscreen")
}

func (p *Player) SetFullscreen(on bool) error {
	return p.setProperty("fullscreen", on)
}

func (p *Player) setProperty(name string, value any) error {
	_, err := p.request("set_property", name, value)
	return err
}

// request sends one IPC command and waits for its reply. Requests are
// serialized so two callers never read each other's response.
func (p *Player) request(command ...any) (json.RawMessage, error) {
	p.cmdMu.Lock()
	defer p.cmdMu.Unlock()

	p.mu.Lock()
	conn := p.conn
	p.reqID++
	requestID := p.reqID
	p.mu.Unlock()
	if conn == nil {
		return nil, fmt.Errorf("ipc closed")
	}

	payload, err := json.Marshal(map[string]any{
		"command":    command,
		"request_id": requestID,
	})
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(800 * time.Millisecond))
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		var resp struct {
			RequestID int             `json:"request_id"`
			Error     string          `json:"error"`
			Data      json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		if resp.RequestID != requestID {
			continue
		}
		if resp.Error != "" && resp.Error != "success" {
			return nil, fmt.Errorf("mpv: %s", resp.Error)
		}
		return resp.Data, nil
	}
}

func (p *Player) getNumber(name string) (float64, error) {
	data, err := p.request("get_property", name)
	if err != nil {
		return 0, err
	}
	var value float64
	if err := json.Unmarshal(data, &value); err != nil {
		return 0, fmt.Errorf("mpv: %s: %w", name, err)
	}
	return value, nil
}

func (p *Player) getBool(name string) (bool, error) {
	data, err := p.request("get_property", name)
	if err != nil {
		return false, err
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return false, fmt.Errorf("mpv: %s: %w", name, err)
	}
	return value, nil
}

func waitSocket(addr string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		conn, err := dialIPC(addr, 200*time.Millisecond)
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
