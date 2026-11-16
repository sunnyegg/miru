package mpv

import "testing"

func TestUseXWayland(t *testing.T) {
	wayland := []string{"WAYLAND_DISPLAY=wayland-0", "DISPLAY=:0"}

	tests := []struct {
		name    string
		environ []string
		native  bool
		want    bool
	}{
		{name: "wayland session", environ: wayland, want: true},
		{name: "native wayland setting", environ: wayland, native: true, want: false},
		{name: "x11 session", environ: []string{"DISPLAY=:0"}, want: false},
		{name: "wayland without xwayland", environ: []string{"WAYLAND_DISPLAY=wayland-0"}, want: false},
		{name: "empty wayland display", environ: []string{"WAYLAND_DISPLAY=", "DISPLAY=:0"}, want: false},
	}
	for _, test := range tests {
		if got := useXWayland(test.environ, test.native); got != test.want {
			t.Errorf("%s: useXWayland = %v, want %v", test.name, got, test.want)
		}
	}
}
