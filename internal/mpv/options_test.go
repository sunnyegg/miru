package mpv

import "testing"

func TestOptionsArgsWindowFlags(t *testing.T) {
	opts := Options{
		MediaPath:    "/media/episode.mkv",
		Title:        "Miru — Frieren — Episode 3",
		StartSeconds: 12.5,
		GLSLShaders:  []string{"/shaders/restore.glsl"},
	}
	config := Config{
		DefaultsPath: "/config/mpv-defaults.conf",
		UserPath:     "/config/mpv.conf",
	}

	args := opts.args("/tmp/mpv.sock", config, false)

	if args[0] != "--input-ipc-server=/tmp/mpv.sock" {
		t.Fatalf("args[0] = %q", args[0])
	}
	if args[1] != "--include=/config/mpv-defaults.conf" || args[2] != "--include=/config/mpv.conf" {
		t.Fatalf("defaults must be included before the user config, got %v", args[:3])
	}
	for _, want := range []string{
		"--force-window=yes",
		"--keep-open=yes",
		"--title=Miru — Frieren — Episode 3",
		"--start=12.500",
		"--glsl-shader=/shaders/restore.glsl",
	} {
		if !containsArg(args, want) {
			t.Fatalf("missing %q in %v", want, args)
		}
	}
	if last := args[len(args)-1]; last != opts.MediaPath {
		t.Fatalf("media path must be last, got %q", last)
	}
}

func TestOptionsArgsPinsXWayland(t *testing.T) {
	opts := Options{MediaPath: "/media/episode.mkv"}

	args := opts.args("/tmp/mpv.sock", Config{}, true)

	if !containsArg(args, "--gpu-context=x11egl") {
		t.Fatalf("missing x11 context in %v", args)
	}
}

func TestOptionsArgsWithoutOptionalFlags(t *testing.T) {
	opts := Options{MediaPath: "/media/episode.mkv"}

	args := opts.args("/tmp/mpv.sock", Config{}, false)

	if containsArg(args, "--include=") || containsArg(args, "--title=") {
		t.Fatalf("unexpected optional flags in %v", args)
	}
	if len(args) != 5 {
		t.Fatalf("args = %v, want ipc + 3 base flags + media", args)
	}
}

func TestOptionsArgsOnlyDefaultConfig(t *testing.T) {
	opts := Options{MediaPath: "/media/episode.mkv"}

	args := opts.args("/tmp/mpv.sock", Config{DefaultsPath: "/config/mpv-defaults.conf"}, false)

	if !containsArg(args, "--include=/config/mpv-defaults.conf") {
		t.Fatalf("missing defaults include in %v", args)
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
