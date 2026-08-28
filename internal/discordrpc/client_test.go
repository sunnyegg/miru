package discordrpc

import "testing"

func TestPlaybackState(t *testing.T) {
	tests := []struct {
		name          string
		episodeNumber int
		percent       float64
		want          string
	}{
		{name: "episode and percent", episodeNumber: 5, percent: 42.4, want: "Episode 5 · 42%"},
		{name: "no episode", episodeNumber: 0, percent: 10, want: "Watching · 10%"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := playbackState(test.episodeNumber, test.percent)
			if got != test.want {
				t.Fatalf("playbackState() = %q, want %q", got, test.want)
			}
		})
	}
}
