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
			_, _ = w.Write([]byte(`{"data":{"Page":{"media":[{"id":21,"title":{"romaji":"One Piece","english":"One Piece"},"coverImage":{"large":"x"},"episodes":1000,"status":"RELEASING","description":"d"}]}}}`))
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
	if err := client.SaveProgress(21, 3); err != nil {
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

func TestListCurrent(t *testing.T) {
	var pages []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		var body struct {
			Query     string `json:"query"`
			Variables struct {
				Page   int `json:"page"`
				UserID int `json:"userId"`
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
			!strings.Contains(body.Query, "CURRENT") ||
			!strings.Contains(body.Query, "UPDATED_TIME_DESC") {
			t.Fatalf("query missing current list filters: %s", body.Query)
		}
		if body.Variables.UserID != 42 {
			t.Fatalf("userId = %d", body.Variables.UserID)
		}
		pages = append(pages, body.Variables.Page)
		if body.Variables.Page == 1 {
			_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":true},"mediaList":[{"progress":3,"media":{"id":21,"title":{"romaji":"One Piece","english":"One Piece"},"coverImage":{"large":"cover"},"episodes":12,"status":"RELEASING"}}]}}}`))
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
	if entries[1].TotalEpisodes != 0 || entries[1].MediaStatus != "FINISHED" {
		t.Fatalf("ongoing mapping = %+v", entries[1])
	}
}
