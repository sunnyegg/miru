package syncprogress

import "testing"

func TestShouldSync(t *testing.T) {
	cases := []struct {
		name    string
		percent float64
		episode int
		current int
		synced  bool
		want    bool
	}{
		{name: "threshold met", percent: 85, episode: 3, current: 2, want: true},
		{name: "below threshold", percent: 84.9, episode: 3, current: 2, want: false},
		{name: "already synced", percent: 90, episode: 3, current: 2, synced: true, want: false},
		{name: "would lower progress", percent: 90, episode: 3, current: 5, want: false},
		{name: "same episode", percent: 90, episode: 3, current: 3, want: false},
		{name: "missing episode", percent: 90, episode: 0, current: 0, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ShouldSync(tc.percent, 85, tc.episode, tc.current, tc.synced)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestShouldAttemptRealtimeSync(t *testing.T) {
	cases := []struct {
		name      string
		percent   float64
		threshold float64
		synced    bool
		mapFailed bool
		attempted bool
		want      bool
	}{
		{name: "crosses threshold", percent: 85, threshold: 85, want: true},
		{name: "below threshold", percent: 84.9, threshold: 85, want: false},
		{name: "already synced", percent: 90, threshold: 85, synced: true, want: false},
		{name: "map failed", percent: 90, threshold: 85, mapFailed: true, want: false},
		{name: "already attempted", percent: 90, threshold: 85, attempted: true, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ShouldAttemptRealtimeSync(tc.percent, tc.threshold, tc.synced, tc.mapFailed, tc.attempted)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
