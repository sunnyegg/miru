package anilist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeAdvancedSearch(t *testing.T) {
	got, err := NormalizeAdvancedSearch(AdvancedSearch{
		Search:   "  frieren ",
		Genres:   []string{"fantasy", "ADVENTURE"},
		Formats:  []string{"tv"},
		Statuses: []string{"finished"},
		Season:   " fall ",
		Year:     2023,
		MinScore: 70,
		Page:     0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Search != "frieren" {
		t.Fatalf("search = %q", got.Search)
	}
	if strings.Join(got.Genres, ",") != "Fantasy,Adventure" {
		t.Fatalf("genres = %v", got.Genres)
	}
	if strings.Join(got.Formats, ",") != "TV" || strings.Join(got.Statuses, ",") != "FINISHED" {
		t.Fatalf("formats = %v statuses = %v", got.Formats, got.Statuses)
	}
	if got.Season != "FALL" {
		t.Fatalf("season = %q", got.Season)
	}
	if got.Sort != "POPULARITY_DESC" {
		t.Fatalf("sort = %q", got.Sort)
	}
	if got.Page != 1 {
		t.Fatalf("page = %d", got.Page)
	}
}

func TestNormalizeAdvancedSearchRejectsBadInput(t *testing.T) {
	tests := []struct {
		name  string
		input AdvancedSearch
		want  string
	}{
		{"unknown genre", AdvancedSearch{Genres: []string{"Space Opera"}}, "unknown genre"},
		{"unknown format", AdvancedSearch{Formats: []string{"SHORT"}}, "unknown format"},
		{"unknown status", AdvancedSearch{Statuses: []string{"AIRING"}}, "unknown status"},
		{"unknown sort", AdvancedSearch{Sort: "OLDEST"}, "unknown sort"},
		{"unknown season", AdvancedSearch{Season: "MONSOON", Year: 2020}, "unknown season"},
		{"season without year", AdvancedSearch{Season: "FALL"}, "together with a season"},
		{"year too old", AdvancedSearch{Year: 1900}, "year must be between"},
		{"score too high", AdvancedSearch{MinScore: 101}, "minimum score"},
		{"long title", AdvancedSearch{Search: strings.Repeat("a", 201)}, "too long"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NormalizeAdvancedSearch(test.input); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("err = %v, want contains %q", err, test.want)
			}
		})
	}
}

func TestAdvancedSearchIsEmpty(t *testing.T) {
	if !(AdvancedSearch{Sort: "SCORE_DESC"}).IsEmpty() {
		t.Fatal("sort alone must not count as a filter")
	}
	if (AdvancedSearch{ExcludeGenres: []string{"Horror"}}).IsEmpty() {
		t.Fatal("excluded genre must count as a filter")
	}
}

func TestSearchAdvancedQuery(t *testing.T) {
	var captured struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":true},"media":[{"id":154587,"title":{"romaji":"Sousou no Frieren","english":"Frieren"},"coverImage":{"large":"cover"},"episodes":28,"status":"FINISHED","season":"FALL","seasonYear":2023,"genres":["Adventure","Fantasy"],"description":"d","mediaListEntry":{"status":"COMPLETED"}}]}}}`))
	}))
	defer server.Close()

	client := New("tok")
	client.Endpoint = server.URL
	client.HTTP = server.Client()

	page, err := client.SearchAdvanced(AdvancedSearch{
		Search:        "frieren",
		Genres:        []string{"Fantasy"},
		ExcludeGenres: []string{"Ecchi"},
		Year:          2023,
		Season:        "FALL",
		Formats:       []string{"TV"},
		Statuses:      []string{"FINISHED"},
		MinScore:      70,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != 154587 || !page.HasNextPage {
		t.Fatalf("page = %+v", page)
	}
	if page.Items[0].ListStatus != "COMPLETED" || page.Items[0].SeasonYear != 2023 {
		t.Fatalf("item = %+v", page.Items[0])
	}
	for _, arg := range []string{"genre_in: $genres", "genre_not_in: $excludeGenres", "seasonYear: $seasonYear", "format_in: $formats", "status_in: $statuses", "averageScore_greater: $minScore", "sort: $sort"} {
		if !strings.Contains(captured.Query, arg) {
			t.Fatalf("query missing %q:\n%s", arg, captured.Query)
		}
	}
	if captured.Variables["search"] != "frieren" {
		t.Fatalf("search variable = %v", captured.Variables["search"])
	}
	if captured.Variables["minScore"] != float64(69) {
		t.Fatalf("minScore variable = %v, want 69 so 70 stays included", captured.Variables["minScore"])
	}
	if got, _ := captured.Variables["sort"].([]any); len(got) != 1 || got[0] != "POPULARITY_DESC" {
		t.Fatalf("sort variable = %v", captured.Variables["sort"])
	}
	if _, ok := captured.Variables["page"]; !ok {
		t.Fatal("page variable missing")
	}
}

func TestSearchAdvancedOmitsUnsetFilters(t *testing.T) {
	var variables map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		variables = body.Variables
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"Page":{"pageInfo":{"hasNextPage":false},"media":[]}}}`))
	}))
	defer server.Close()

	client := New("")
	client.Endpoint = server.URL
	client.HTTP = server.Client()

	if _, err := client.SearchAdvanced(AdvancedSearch{Genres: []string{"Mecha"}}); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"search", "excludeGenres", "season", "seasonYear", "formats", "statuses", "minScore"} {
		if _, ok := variables[key]; ok {
			t.Fatalf("%q must be omitted when unset: %v", key, variables)
		}
	}
}

func TestSearchAdvancedNeedsFilter(t *testing.T) {
	if _, err := New("").SearchAdvanced(AdvancedSearch{Sort: "SCORE_DESC"}); err == nil || !strings.Contains(err.Error(), "at least one filter") {
		t.Fatalf("err = %v", err)
	}
}
