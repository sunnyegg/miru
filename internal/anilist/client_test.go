package anilist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginURL(t *testing.T) {
	got, err := LoginURL("123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "client_id=123") || !strings.Contains(got, "response_type=code") {
		t.Fatalf("url = %s", got)
	}
	if !strings.Contains(got, "redirect_uri=") {
		t.Fatalf("missing redirect_uri: %s", got)
	}
}

func TestSearchAndSave(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body.Query, "Page(") {
			_, _ = w.Write([]byte(`{"data":{"Page":{"media":[{"id":21,"title":{"romaji":"One Piece","english":"One Piece"},"coverImage":{"large":"x"},"episodes":1000,"status":"RELEASING","description":"d","mediaListEntry":{"status":"PLANNING"}}]}}}`))
			return
		}
		if strings.Contains(body.Query, "SaveMediaListEntry") && strings.Contains(body.Query, "status:") {
			_, _ = w.Write([]byte(`{"data":{"SaveMediaListEntry":{"id":1,"status":"CURRENT","progress":0}}}`))
			return
		}
		if strings.Contains(body.Query, "SaveMediaListEntry") {
			_, _ = w.Write([]byte(`{"data":{"SaveMediaListEntry":{"id":1,"progress":3}}}`))
			return
		}
		if strings.Contains(body.Query, "Viewer") {
			_, _ = w.Write([]byte(`{"data":{"Viewer":{"name":"adila"}}}`))
			return
		}
		http.Error(w, "unexpected", 500)
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()

	name, err := client.ViewerName()
	if err != nil || name != "adila" {
		t.Fatalf("viewer %q %v", name, err)
	}
	results, err := client.Search("one")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ID != 21 {
		t.Fatalf("search = %+v", results)
	}
	if results[0].ListStatus != "PLANNING" {
		t.Fatalf("listStatus = %q", results[0].ListStatus)
	}
	if err := client.SaveProgress(21, 3); err != nil {
		t.Fatal(err)
	}
	if err := client.SaveListStatus(21, "CURRENT", -1); err != nil {
		t.Fatal(err)
	}
}

func TestAiringSchedules(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				Page  int   `json:"page"`
				Start int64 `json:"start"`
				End   int64 `json:"end"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Variables.Start != 100 || body.Variables.End != 200 {
			t.Fatalf("range = %+v", body.Variables)
		}
		if !strings.Contains(body.Query, "airingSchedules") {
			t.Fatalf("query missing airingSchedules: %s", body.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		if body.Variables.Page == 1 {
			_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":true},"airingSchedules":[{"id":1,"airingAt":120,"episode":4,"media":{"id":21,"title":{"romaji":"One Piece","english":"One Piece"},"coverImage":{"large":"cover"}}}]}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":false},"airingSchedules":[{"id":2,"airingAt":180,"episode":5,"media":{"id":22,"title":{"romaji":"Another","english":null},"coverImage":{"large":""}}}]}}}`))
	}))
	defer server.Close()

	client := New("")
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	schedules, err := client.AiringSchedules(100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || len(schedules) != 2 {
		t.Fatalf("requests = %d schedules = %+v", requests, schedules)
	}
	if schedules[1].MediaID != 22 || schedules[1].TitleRomaji != "Another" {
		t.Fatalf("mapping = %+v", schedules[1])
	}
}

func TestAiringSchedulesRejectsInvalidRange(t *testing.T) {
	client := New("")
	if _, err := client.AiringSchedules(200, 100); err == nil {
		t.Fatal("expected invalid range error")
	}
}

func TestListProgressForMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				IDs []int `json:"ids"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(body.Query, "id_in") || !strings.Contains(body.Query, "mediaListEntry") {
			t.Fatalf("query missing batch progress fields: %s", body.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"Page":{"media":[
			{"id":21,"episodes":12,"status":"RELEASING","nextAiringEpisode":{"episode":8},"mediaListEntry":{"progress":5}},
			{"id":22,"episodes":24,"status":"FINISHED","nextAiringEpisode":null,"mediaListEntry":null}
		]}}}`))
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()

	progressByMedia, err := client.ListProgressForMedia([]int{21, 22})
	if err != nil {
		t.Fatal(err)
	}
	if len(progressByMedia) != 2 {
		t.Fatalf("progressByMedia = %+v", progressByMedia)
	}
	if progressByMedia[21].Progress != 5 || progressByMedia[21].TotalEpisodes != 12 {
		t.Fatalf("media 21 = %+v", progressByMedia[21])
	}
	if progressByMedia[21].MediaStatus != "RELEASING" || progressByMedia[21].NextAiringEpisode != 8 {
		t.Fatalf("media 21 airing = %+v", progressByMedia[21])
	}
	if progressByMedia[22].Progress != 0 || progressByMedia[22].TotalEpisodes != 24 {
		t.Fatalf("media 22 = %+v", progressByMedia[22])
	}
	if progressByMedia[22].MediaStatus != "FINISHED" || progressByMedia[22].NextAiringEpisode != 0 {
		t.Fatalf("media 22 status = %+v", progressByMedia[22])
	}
}

func TestListCurrent(t *testing.T) {
	var pages []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				Page   int    `json:"page"`
				UserID int    `json:"userId"`
				Status string `json:"status"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body.Query, "Viewer") {
			_, _ = w.Write([]byte(`{"data":{"Viewer":{"id":42}}}`))
			return
		}
		if !strings.Contains(body.Query, "mediaList") ||
			!strings.Contains(body.Query, "UPDATED_TIME_DESC") ||
			!strings.Contains(body.Query, "nextAiringEpisode") {
			t.Fatalf("query missing media list fields: %s", body.Query)
		}
		if body.Variables.UserID != 42 {
			t.Fatalf("userId = %d", body.Variables.UserID)
		}
		if body.Variables.Status != "CURRENT" {
			t.Fatalf("status = %q", body.Variables.Status)
		}
		pages = append(pages, body.Variables.Page)
		if body.Variables.Page == 1 {
			_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":true},"mediaList":[{"status":"CURRENT","progress":3,"score":85,"notes":"great","repeat":1,"private":false,"startedAt":{"year":2024,"month":1,"day":2},"completedAt":{"year":0,"month":0,"day":0},"media":{"id":21,"title":{"romaji":"One Piece","english":"One Piece"},"coverImage":{"large":"cover"},"episodes":12,"status":"RELEASING","nextAiringEpisode":{"episode":8}}}]}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":false},"mediaList":[{"progress":7,"media":{"id":22,"title":{"romaji":"Another","english":null},"coverImage":{"large":""},"episodes":null,"status":"FINISHED"}}]}}}`))
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	entries, err := client.ListCurrent()
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[0] != 1 || pages[1] != 2 {
		t.Fatalf("pages = %v", pages)
	}
	if len(entries) != 2 || entries[0].Progress != 3 || entries[1].MediaID != 22 {
		t.Fatalf("entries = %+v", entries)
	}
	if entries[0].ListStatus != "CURRENT" || entries[0].ScoreRaw != 85 || entries[0].Notes != "great" {
		t.Fatalf("entry fields = %+v", entries[0])
	}
	if entries[0].StartedAt.Year != 2024 || entries[0].StartedAt.Month != 1 || entries[0].StartedAt.Day != 2 {
		t.Fatalf("startedAt = %+v", entries[0].StartedAt)
	}
	if entries[0].NextAiringEpisode != 8 {
		t.Fatalf("nextAiringEpisode = %d", entries[0].NextAiringEpisode)
	}
	if entries[1].TotalEpisodes != 0 || entries[1].MediaStatus != "FINISHED" {
		t.Fatalf("ongoing mapping = %+v", entries[1])
	}
}

func TestListMediaListCompleted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				Status string `json:"status"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body.Query, "Viewer") {
			_, _ = w.Write([]byte(`{"data":{"Viewer":{"id":42}}}`))
			return
		}
		if body.Variables.Status != "COMPLETED" {
			t.Fatalf("status = %q", body.Variables.Status)
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":false},"mediaList":[{"progress":12,"media":{"id":21,"title":{"romaji":"Done Show","english":null},"coverImage":{"large":""},"episodes":12,"status":"FINISHED"}}]}}}`))
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	entries, err := client.ListMediaList("COMPLETED")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].MediaID != 21 || entries[0].Progress != 12 {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestListMediaListRejectsUnknownStatus(t *testing.T) {
	client := New("tok")
	if _, err := client.ListMediaList("INVALID"); err == nil {
		t.Fatal("expected error for unsupported status")
	}
}

func TestListMediaListPlanning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				Status string `json:"status"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(body.Query, "Viewer") {
			_, _ = w.Write([]byte(`{"data":{"Viewer":{"id":42}}}`))
			return
		}
		if body.Variables.Status != "PLANNING" {
			t.Fatalf("status = %q", body.Variables.Status)
		}
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":false},"mediaList":[{"status":"PLANNING","progress":0,"score":0,"notes":"","repeat":0,"private":false,"startedAt":{"year":0,"month":0,"day":0},"completedAt":{"year":0,"month":0,"day":0},"media":{"id":21,"title":{"romaji":"Plan Show","english":null},"coverImage":{"large":""},"episodes":12,"status":"NOT_YET_RELEASED"}}]}}}`))
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	entries, err := client.ListMediaList("PLANNING")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].ListStatus != "PLANNING" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestSaveListEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if !strings.Contains(body.Query, "SaveMediaListEntry") {
			t.Fatalf("unexpected query: %s", body.Query)
		}
		if body.Variables["mediaId"] != float64(21) {
			t.Fatalf("mediaId = %v", body.Variables["mediaId"])
		}
		if body.Variables["status"] != "PLANNING" {
			t.Fatalf("status = %v", body.Variables["status"])
		}
		if body.Variables["scoreRaw"] != float64(90) {
			t.Fatalf("scoreRaw = %v", body.Variables["scoreRaw"])
		}
		startedAt, ok := body.Variables["startedAt"].(map[string]any)
		if !ok || startedAt["year"] != float64(2020) {
			t.Fatalf("startedAt = %v", body.Variables["startedAt"])
		}
		completedAt := body.Variables["completedAt"]
		if completedAt != nil {
			t.Fatalf("completedAt = %v", completedAt)
		}
		_, _ = w.Write([]byte(`{"data":{"SaveMediaListEntry":{"id":1,"status":"PLANNING","progress":4,"score":90}}}`))
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	err := client.SaveListEntry(ListEntrySave{
		MediaID:         21,
		Status:          "PLANNING",
		Progress:        4,
		ScoreRaw:        90,
		Notes:           "note",
		SendNotes:       true,
		Repeat:          0,
		SendPrivate:     true,
		Private:         true,
		StartedAt:       FuzzyDate{Year: 2020, Month: 3, Day: 15},
		SendStartedAt:   true,
		CompletedAt:     FuzzyDate{},
		SendCompletedAt: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSaveListEntryRejectsInvalidStatus(t *testing.T) {
	client := New("tok")
	err := client.SaveListEntry(ListEntrySave{MediaID: 1, Status: "INVALID"})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}
