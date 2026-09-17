package main

import (
	"testing"
	"time"
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
