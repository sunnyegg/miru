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
