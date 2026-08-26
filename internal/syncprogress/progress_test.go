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
