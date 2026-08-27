package anilist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSeasonEpisode(t *testing.T) {
	cases := []struct {
		name            string
		parsed          int
		seasonLength    int
		prequelEpisodes int
		want            int
		ok              bool
	}{
		{name: "already relative", parsed: 8, seasonLength: 13, prequelEpisodes: 12, want: 8, ok: true},
		{name: "rezero s4 absolute 79", parsed: 79, seasonLength: 19, prequelEpisodes: 66, want: 13, ok: true},
		{name: "first episode of season", parsed: 67, seasonLength: 19, prequelEpisodes: 66, want: 1, ok: true},
		{name: "no prequels overflow", parsed: 79, seasonLength: 19, prequelEpisodes: 0, want: 79, ok: false},
		{name: "unknown length uses offset", parsed: 79, seasonLength: 0, prequelEpisodes: 66, want: 13, ok: true},
		{name: "offset overshoots season", parsed: 90, seasonLength: 19, prequelEpisodes: 66, want: 90, ok: false},
		{name: "missing episode", parsed: 0, seasonLength: 19, prequelEpisodes: 66, want: 0, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := SeasonEpisode(tc.parsed, tc.seasonLength, tc.prequelEpisodes)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("got %d ok=%v want %d ok=%v", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestMapSeasonEpisodeWalksTVPrequels(t *testing.T) {
	type mediaPayload struct {
		ID        int
		Episodes  int
		Format    string
		PrequelID int
		PrequelEp int
	}
	catalog := map[int]mediaPayload{
		189046: {ID: 189046, Episodes: 19, Format: "TV", PrequelID: 163134, PrequelEp: 66},
		163134: {ID: 163134, Episodes: 66, Format: "TV"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables struct {
				ID int `json:"id"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		media, ok := catalog[body.Variables.ID]
		if !ok {
			http.Error(w, "missing media", 500)
			return
		}
		edges := []any{}
		if media.PrequelID != 0 {
			edges = append(edges, map[string]any{
				"relationType": "PREQUEL",
				"node": map[string]any{
					"id":       media.PrequelID,
					"episodes": media.PrequelEp,
					"format":   "TV",
				},
			})
		}
		payload := map[string]any{
			"data": map[string]any{
				"Media": map[string]any{
					"id":        media.ID,
					"episodes":  media.Episodes,
					"format":    media.Format,
					"relations": map[string]any{"edges": edges},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	client := New("")
	client.Endpoint = server.URL
	client.HTTP = server.Client()

	got, err := client.MapSeasonEpisode(189046, 79)
	if err != nil {
		t.Fatal(err)
	}
	if got != 13 {
		t.Fatalf("mapped = %d want 13", got)
	}

	relative, err := client.MapSeasonEpisode(189046, 8)
	if err != nil {
		t.Fatal(err)
	}
	if relative != 8 {
		t.Fatalf("relative = %d want 8", relative)
	}
}
