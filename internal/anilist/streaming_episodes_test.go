package anilist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEpisodeNumberFromStreamingTitle(t *testing.T) {
	tests := []struct {
		title   string
		number  int
		matched bool
	}{
		{title: "Episode 12", number: 12, matched: true},
		{title: "ep 3", number: 3, matched: true},
		{title: "Ep. 7 - The Battle", number: 7, matched: true},
		{title: "5. First Day", number: 5, matched: true},
		{title: "10: Something", number: 10, matched: true},
		{title: "2 - Title", number: 2, matched: true},
		{title: "15", number: 15, matched: true},
		{title: "Special Episode", number: 0, matched: false},
		{title: "", number: 0, matched: false},
	}

	for _, test := range tests {
		number, matched := EpisodeNumberFromStreamingTitle(test.title)
		if matched != test.matched || number != test.number {
			t.Fatalf("title %q => (%d, %v), want (%d, %v)", test.title, number, matched, test.number, test.matched)
		}
	}
}

func TestStreamingEpisodeThumbnails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body.Query, "streamingEpisodes") {
			_, _ = w.Write([]byte(`{"data":{"Media":{"streamingEpisodes":[
				{"title":"Episode 1","thumbnail":"https://example.com/1.jpg"},
				{"title":"Episode 1","thumbnail":"https://example.com/dup.jpg"},
				{"title":"Episode 2","thumbnail":""},
				{"title":"5. Bonus","thumbnail":"https://example.com/5.jpg"},
				{"title":"No number","thumbnail":"https://example.com/x.jpg"}
			]}}}`))
			return
		}
		http.Error(w, "unexpected", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()

	thumbnails, err := client.StreamingEpisodeThumbnails(21)
	if err != nil {
		t.Fatal(err)
	}
	if len(thumbnails) != 2 {
		t.Fatalf("len = %d, want 2", len(thumbnails))
	}
	if thumbnails[0].EpisodeNumber != 1 || thumbnails[0].Thumbnail != "https://example.com/1.jpg" {
		t.Fatalf("first = %+v", thumbnails[0])
	}
	if thumbnails[1].EpisodeNumber != 5 || thumbnails[1].Thumbnail != "https://example.com/5.jpg" {
		t.Fatalf("second = %+v", thumbnails[1])
	}
}
