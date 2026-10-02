package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sunnyegg/miru/internal/secrets"
	"github.com/sunnyegg/miru/internal/storage"
)

func TestNewAnilistReusesClientAndHTTPTransport(t *testing.T) {
	a := newSettingsApp(t)

	first, err := a.newAnilist("token-a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := a.newAnilist("token-a")
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("expected cached anilist client for same token")
	}
	if first.HTTP != second.HTTP {
		t.Fatal("expected shared HTTP client")
	}

	empty, err := a.newAnilist("")
	if err != nil {
		t.Fatal(err)
	}
	if empty == first {
		t.Fatal("empty-token client must be distinct from authed client")
	}
	if empty.HTTP != first.HTTP {
		t.Fatal("empty-token client should reuse the same HTTP transport")
	}

	againEmpty, err := a.newAnilist("")
	if err != nil {
		t.Fatal(err)
	}
	if againEmpty != empty {
		t.Fatal("expected cached empty-token client")
	}
}

func TestNewAnilistRebuildsWhenNetworkChanges(t *testing.T) {
	a := newSettingsApp(t)

	first, err := a.newAnilist("token-a")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.SaveNetworkSettings("direct", "", ""); err != nil {
		t.Fatal(err)
	}
	second, err := a.newAnilist("token-a")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("expected new anilist client after network change")
	}
	if first.HTTP == second.HTTP {
		t.Fatal("expected new HTTP transport after network change")
	}
}

func TestLoadCachedJSONUsesInMemoryDecodedValue(t *testing.T) {
	a := newSettingsApp(t)
	key := "test:memory-cache"
	payload := []WatchingEntryView{{MediaID: 42, TitleRomaji: "Test"}}

	fetchCount := 0
	load := func() ([]WatchingEntryView, error) {
		return loadCachedJSON(a, key, time.Hour, func() ([]WatchingEntryView, error) {
			fetchCount++
			return payload, nil
		})
	}

	first, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetchCount = %d, want 1", fetchCount)
	}
	if len(first) != 1 || first[0].MediaID != 42 {
		t.Fatalf("first = %+v", first)
	}

	// Corrupt disk so a memory miss would have to refetch.
	if err := a.store.SetAPICache(key, `{not-json`); err != nil {
		t.Fatal(err)
	}

	second, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if fetchCount != 1 {
		t.Fatalf("fetchCount = %d after memory hit, want 1", fetchCount)
	}
	if len(second) != 1 || second[0].MediaID != 42 {
		t.Fatalf("second = %+v", second)
	}

	a.deleteAPIMemory(key)
	third, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if fetchCount != 2 {
		t.Fatalf("fetchCount = %d after memory invalidate, want 2", fetchCount)
	}
	if len(third) != 1 || third[0].MediaID != 42 {
		t.Fatalf("third = %+v", third)
	}
}

func TestGetAnimeSharesDetailCacheWithinAccount(t *testing.T) {
	var requests atomic.Int32
	var fail atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if fail.Load() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		status := "PLANNING"
		if r.Header.Get("Authorization") == "Bearer token-b" {
			status = "CURRENT"
		}
		if r.Header.Get("Authorization") == "" {
			status = ""
		}
		_, _ = fmt.Fprintf(w, `{"data":{"Media":{"id":21,"title":{"romaji":"Test"},"genres":["Drama"],"mediaListEntry":{"status":%q}}}}`, status)
	}))
	defer server.Close()
	a := newSettingsApp(t)
	a.tokens = &secrets.MemoryStore{}
	for _, token := range []string{"", "token-a", "token-b"} {
		client, err := a.newAnilist(token)
		if err != nil {
			t.Fatal(err)
		}
		client.Endpoint = server.URL
		client.HTTP = server.Client()
	}
	if err := a.tokens.Set("token-a"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		anime, err := a.GetAnime(21)
		if err != nil {
			t.Fatal(err)
		}
		if anime.ListStatus != "PLANNING" || strings.Join(anime.Genres, ",") != "Drama" {
			t.Fatalf("details = %+v", anime)
		}
	}
	if requests.Load() != 1 {
		t.Fatalf("memory cache requests = %d", requests.Load())
	}
	key := animeCacheKey(21, "token-a")
	if strings.Contains(key, "token-a") {
		t.Fatal("cache key exposes token")
	}
	a.deleteAPIMemory(key)
	if _, err := a.GetAnime(21); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatalf("disk cache requests = %d", requests.Load())
	}
	a.invalidateAnimeCache(21)
	if _, err := a.GetAnime(21); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 {
		t.Fatalf("invalidation requests = %d", requests.Load())
	}

	// Force expiry and remove the disk copy to exercise memory-only stale fallback.
	a.apiMemoryMu.Lock()
	entry := a.apiMemory[key]
	entry.fetchedAt = time.Now().Add(-2 * animeDetailCacheTTL)
	a.apiMemory[key] = entry
	a.apiMemoryMu.Unlock()
	if err := a.store.DeleteAPICache(key); err != nil {
		t.Fatal(err)
	}
	fail.Store(true)
	anime, err := a.GetAnime(21)
	if err != nil || anime.ListStatus != "PLANNING" {
		t.Fatalf("stale fallback = %+v, %v", anime, err)
	}
	if requests.Load() != 3 {
		t.Fatalf("expired cache requests = %d", requests.Load())
	}
	fail.Store(false)

	if err := a.tokens.Set("token-b"); err != nil {
		t.Fatal(err)
	}
	anime, err = a.GetAnime(21)
	if err != nil || anime.ListStatus != "CURRENT" {
		t.Fatalf("second account = %+v, %v", anime, err)
	}
	if requests.Load() != 4 {
		t.Fatalf("second account requests = %d", requests.Load())
	}
	a.invalidateAnimeCache(21)
	if _, ok := memoryCachedJSON[AnimeView](a, key, 0); !ok {
		t.Fatal("invalidating second account removed first account cache")
	}
	if err := a.tokens.Delete(); err != nil {
		t.Fatal(err)
	}
	anime, err = a.GetAnime(21)
	if err != nil || anime.ListStatus != "" {
		t.Fatalf("anonymous = %+v, %v", anime, err)
	}
	if requests.Load() != 5 {
		t.Fatalf("anonymous requests = %d", requests.Load())
	}
	db, err := sql.Open("sqlite", filepath.Join(a.dirs.Config, "app_data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	anonymousKey := animeCacheKey(21, "")
	if _, err := db.Exec("UPDATE api_cache SET fetched_at = 0 WHERE cache_key = ?", anonymousKey); err != nil {
		t.Fatal(err)
	}
	a.deleteAPIMemory(anonymousKey)
	fail.Store(true)
	anime, err = a.GetAnime(21)
	if err != nil || strings.Join(anime.Genres, ",") != "Drama" {
		t.Fatalf("disk stale fallback = %+v, %v", anime, err)
	}
	if requests.Load() != 6 {
		t.Fatalf("disk stale requests = %d", requests.Load())
	}
}

func TestAnimeSearchCacheIsAccountScopedAndInvalidated(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		status := "PLANNING"
		if r.Header.Get("Authorization") == "Bearer token-b" {
			status = "CURRENT"
		}
		_, _ = fmt.Fprintf(w, `{"data":{"Page":{"media":[{"id":21,"title":{"romaji":"Test"},"coverImage":{"large":"cover"},"episodes":12,"status":"FINISHED","description":"synopsis","mediaListEntry":{"status":%q}}]}}}`, status)
	}))
	defer server.Close()

	a := newSettingsApp(t)
	a.tokens = &secrets.MemoryStore{}
	for _, token := range []string{"token-a", "token-b"} {
		client, err := a.newAnilist(token)
		if err != nil {
			t.Fatal(err)
		}
		client.Endpoint = server.URL
		client.HTTP = server.Client()
	}
	if err := a.tokens.Set("token-a"); err != nil {
		t.Fatal(err)
	}
	first, err := a.SearchAnime("Test")
	if err != nil || len(first) != 1 || first[0].ListStatus != "PLANNING" {
		t.Fatalf("first search = %+v, %v", first, err)
	}
	if _, err := a.SearchAnime(" test "); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 1 {
		t.Fatalf("same normalized query requests = %d, want 1", requests.Load())
	}

	if err := a.tokens.Set("token-b"); err != nil {
		t.Fatal(err)
	}
	second, err := a.SearchAnime("test")
	if err != nil || len(second) != 1 || second[0].ListStatus != "CURRENT" {
		t.Fatalf("second account search = %+v, %v", second, err)
	}
	if requests.Load() != 2 {
		t.Fatalf("account-scoped requests = %d, want 2", requests.Load())
	}

	a.invalidateAnilistUserCaches()
	if _, err := a.SearchAnime("test"); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 3 {
		t.Fatalf("requests after invalidation = %d, want 3", requests.Load())
	}
}

func TestAnimeListCachesAreAccountScoped(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		userID := 1
		mediaID := 101
		count := 1
		if r.Header.Get("Authorization") == "Bearer token-b" {
			userID, mediaID, count = 2, 202, 2
		}
		switch {
		case strings.Contains(body.Query, "Viewer {"):
			_, _ = fmt.Fprintf(w, `{"data":{"Viewer":{"id":%d}}}`, userID)
		case strings.Contains(body.Query, "MediaListCollection"):
			entries := strings.Repeat(`{"status":"CURRENT"},`, count)
			entries = strings.TrimSuffix(entries, ",")
			_, _ = fmt.Fprintf(w, `{"data":{"MediaListCollection":{"lists":[{"entries":[%s]}]}}}`, entries)
		default:
			_, _ = fmt.Fprintf(w, `{"data":{"Page":{"pageInfo":{"hasNextPage":false},"mediaList":[{"status":"CURRENT","progress":1,"media":{"id":%d,"title":{"romaji":"Test"},"episodes":12,"status":"FINISHED"}}]}}}`, mediaID)
		}
	}))
	defer server.Close()

	a := newSettingsApp(t)
	a.tokens = &secrets.MemoryStore{}
	for _, token := range []string{"token-a", "token-b"} {
		client, err := a.newAnilist(token)
		if err != nil {
			t.Fatal(err)
		}
		client.Endpoint = server.URL
		client.HTTP = server.Client()
	}
	if err := a.tokens.Set("token-a"); err != nil {
		t.Fatal(err)
	}
	entriesA, err := a.ListAnimeList("CURRENT")
	if err != nil || len(entriesA) != 1 || entriesA[0].MediaID != 101 {
		t.Fatalf("account A list = %+v, %v", entriesA, err)
	}
	if _, err := a.ListAnimeList("CURRENT"); err != nil {
		t.Fatal(err)
	}
	if err := a.tokens.Set("token-b"); err != nil {
		t.Fatal(err)
	}
	entriesB, err := a.ListAnimeList("CURRENT")
	if err != nil || len(entriesB) != 1 || entriesB[0].MediaID != 202 {
		t.Fatalf("account B list = %+v, %v", entriesB, err)
	}
	if err := a.tokens.Set("token-a"); err != nil {
		t.Fatal(err)
	}
	countsA, err := a.ListAnimeListCounts()
	if err != nil || countsA["CURRENT"] != 1 {
		t.Fatalf("account A counts = %+v, %v", countsA, err)
	}
	if err := a.tokens.Set("token-b"); err != nil {
		t.Fatal(err)
	}
	countsB, err := a.ListAnimeListCounts()
	if err != nil || countsB["CURRENT"] != 2 {
		t.Fatalf("account B counts = %+v, %v", countsB, err)
	}
	if requests.Load() != 6 {
		t.Fatalf("requests = %d, want 6 including one cached ViewerID per account", requests.Load())
	}
}

func TestListEpisodesCachesProgressUntilUserDataInvalidation(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"Page":{"media":[{"id":42,"episodes":12,"status":"RELEASING","mediaListEntry":{"progress":6}}]}}}`)
	}))
	defer server.Close()

	a := newSettingsApp(t)
	a.tokens = &secrets.MemoryStore{}
	if err := a.tokens.Set("token-a"); err != nil {
		t.Fatal(err)
	}
	client, err := a.newAnilist("token-a")
	if err != nil {
		t.Fatal(err)
	}
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	if err := a.store.UpsertAnime(storage.Anime{AnilistID: 42, TitleRomaji: "Test", TotalEpisodes: 12}); err != nil {
		t.Fatal(err)
	}
	_, err = a.store.InsertEpisode(storage.Episode{
		AnilistID:     sql.NullInt64{Int64: 42, Valid: true},
		EpisodeNumber: sql.NullInt64{Int64: 6, Valid: true},
		FilePath:      "/tmp/test.mkv",
		Status:        "COMPLETED",
	})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		episodes, err := a.ListEpisodes()
		if err != nil || len(episodes) != 1 || episodes[0].Progress != 6 {
			t.Fatalf("ListEpisodes() = %+v, %v", episodes, err)
		}
	}
	if requests.Load() != 1 {
		t.Fatalf("cached progress requests = %d, want 1", requests.Load())
	}

	a.invalidateAnilistUserCaches()
	if _, err := a.ListEpisodes(); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 {
		t.Fatalf("requests after invalidation = %d, want 2", requests.Load())
	}
}
