package storage

import (
	"testing"
	"time"
)

func TestRSSFeedLifecycle(t *testing.T) {
	store := openTestStore(t)

	feedID, err := store.InsertRSSFeed("https://nyaa.si/?page=rss", "Nyaa")
	if err != nil {
		t.Fatal(err)
	}
	feed, err := store.RSSFeedByID(feedID)
	if err != nil {
		t.Fatal(err)
	}
	if feed.URL != "https://nyaa.si/?page=rss" || !feed.Enabled {
		t.Fatalf("feed = %+v", feed)
	}

	polledAt := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	if err := store.UpdateRSSFeedLastPolled(feedID, polledAt); err != nil {
		t.Fatal(err)
	}
	if err := store.SetRSSFeedEnabled(feedID, false); err != nil {
		t.Fatal(err)
	}

	feeds, err := store.ListRSSFeeds()
	if err != nil {
		t.Fatal(err)
	}
	if len(feeds) != 1 || feeds[0].Enabled || !feeds[0].LastPolled.Valid {
		t.Fatalf("feeds = %+v", feeds)
	}

	inserted, err := store.UpsertRSSFeedItem(feedID, RSSFeedItem{
		ItemKey:   "item-1",
		Title:     "Episode 01",
		Link:      "https://example.com/1.torrent",
		Magnet:    "magnet:?xt=urn:btih:abc",
		Published: polledAt,
	})
	if err != nil || !inserted {
		t.Fatalf("insert = %v, %v", inserted, err)
	}
	inserted, err = store.UpsertRSSFeedItem(feedID, RSSFeedItem{
		ItemKey:   "item-1",
		Title:     "Episode 01",
		Link:      "https://example.com/1.torrent",
		Magnet:    "magnet:?xt=urn:btih:abc",
		Published: polledAt,
	})
	if err != nil || inserted {
		t.Fatalf("duplicate insert = %v, %v", inserted, err)
	}

	count, err := store.CountNewRSSFeedItems()
	if err != nil || count != 1 {
		t.Fatalf("count = %d, err = %v", count, err)
	}

	items, err := store.ListRSSFeedItems(true)
	if err != nil || len(items) != 1 || items[0].FeedTitle != "Nyaa" {
		t.Fatalf("items = %+v, err = %v", items, err)
	}

	if err := store.MarkRSSFeedItemsSeen([]int64{items[0].ID}); err != nil {
		t.Fatal(err)
	}
	count, err = store.CountNewRSSFeedItems()
	if err != nil || count != 0 {
		t.Fatalf("count after mark = %d, err = %v", count, err)
	}

	if err := store.DeleteRSSFeed(feedID); err != nil {
		t.Fatal(err)
	}
	feeds, err = store.ListRSSFeeds()
	if err != nil || len(feeds) != 0 {
		t.Fatalf("feeds after delete = %+v, err = %v", feeds, err)
	}
}
