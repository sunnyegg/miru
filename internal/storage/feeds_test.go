package storage

import (
	"database/sql"
	"strings"
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

	items, _, err := store.ListRSSFeedItems(true, "", 0, 0)
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

func TestListRSSFeedItemsFilteredPaged(t *testing.T) {
	store := openTestStore(t)

	nyaaID, err := store.InsertRSSFeed("https://nyaa.si/?page=rss", "Nyaa")
	if err != nil {
		t.Fatal(err)
	}
	fansubID, err := store.InsertRSSFeed("https://example.com/rss", "Fansub")
	if err != nil {
		t.Fatal(err)
	}
	published := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	titles := []struct {
		feedID int64
		key    string
		title  string
		link   string
		isNew  bool
	}{
		{nyaaID, "a", "Show A Episode 01", "https://example.com/a1.torrent", true},
		{nyaaID, "b", "Show B Episode 01", "https://example.com/b1.torrent", true},
		{fansubID, "c", "Show B Episode 02", "https://example.com/b2.torrent", true},
		{nyaaID, "d", "Show C Episode 03", "https://example.com/c3.torrent", false},
	}
	for _, item := range titles {
		inserted, err := store.UpsertRSSFeedItem(item.feedID, RSSFeedItem{
			ItemKey:   item.key,
			Title:     item.title,
			Link:      item.link,
			Published: published,
		})
		if err != nil || !inserted {
			t.Fatalf("insert %s = %v, %v", item.key, inserted, err)
		}
		if !item.isNew {
			rows, err := store.db.Exec(
				`UPDATE rss_feed_items SET is_new = 0 WHERE feed_id = ? AND item_key = ?`,
				item.feedID, item.key,
			)
			if err != nil {
				t.Fatal(err)
			}
			if n, _ := rows.RowsAffected(); n != 1 {
				t.Fatalf("mark seen %s affected %d rows", item.key, n)
			}
		}
	}

	t.Run("new only", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(true, "", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 3 || len(items) != 3 {
			t.Fatalf("total = %d, items = %d", total, len(items))
		}
	})

	t.Run("query matches title", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(false, "show b", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 2 || len(items) != 2 {
			t.Fatalf("total = %d, items = %d", total, len(items))
		}
		for _, item := range items {
			if !strings.Contains(strings.ToLower(item.Title), "show b") {
				t.Fatalf("item %+v does not match", item)
			}
		}
	})

	t.Run("query matches link", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(false, "b2.torrent", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 1 || len(items) != 1 || items[0].Title != "Show B Episode 02" {
			t.Fatalf("total = %d, items = %+v", total, items)
		}
	})

	t.Run("query matches feed title", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(false, "fansub", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 1 || len(items) != 1 || items[0].FeedTitle != "Fansub" {
			t.Fatalf("total = %d, items = %+v", total, items)
		}
	})

	t.Run("query combined with new only", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(true, "show b", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 2 || len(items) != 2 {
			t.Fatalf("total = %d, items = %d", total, len(items))
		}
	})

	t.Run("paging", func(t *testing.T) {
		page1, total, err := store.ListRSSFeedItems(false, "", 2, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 4 || len(page1) != 2 {
			t.Fatalf("page1 total = %d, items = %d", total, len(page1))
		}
		page2, _, err := store.ListRSSFeedItems(false, "", 2, 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(page2) != 2 {
			t.Fatalf("page2 items = %d", len(page2))
		}
		if page1[0].ID == page2[0].ID || page1[0].ID == page2[1].ID {
			t.Fatalf("pages overlap: %+v vs %+v", page1, page2)
		}
		if page1[0].ID != page2[0].ID+2 && page1[0].ID != page2[0].ID-2 {
			t.Fatalf("pages not adjacent: %+v vs %+v", page1, page2)
		}
	})

	t.Run("paging beyond end returns empty page", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(false, "", 20, 100)
		if err != nil {
			t.Fatal(err)
		}
		if total != 4 || len(items) != 0 {
			t.Fatalf("total = %d, items = %d", total, len(items))
		}
	})

	t.Run("default limit", func(t *testing.T) {
		items, total, err := store.ListRSSFeedItems(false, "", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if total != 4 || len(items) != 4 {
			t.Fatalf("total = %d, items = %d", total, len(items))
		}
	})
}

func TestListLibraryAnimeTitles(t *testing.T) {
	store := openTestStore(t)

	titles, err := store.ListLibraryAnimeTitles()
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != 0 {
		t.Fatalf("empty library titles = %+v", titles)
	}

	if err := store.UpsertAnime(Anime{
		AnilistID:    42,
		TitleRomaji:  "Frieren",
		TitleEnglish: "Frieren: Beyond Journey's End",
	}); err != nil {
		t.Fatal(err)
	}
	episodeID, err := store.InsertEpisode(Episode{
		AnilistID: sql.NullInt64{Int64: 42, Valid: true},
		FilePath:  "/tmp/frieren.mkv",
		Status:    "COMPLETED",
	})
	if err != nil || episodeID == 0 {
		t.Fatalf("insert episode = %d, %v", episodeID, err)
	}

	titles, err = store.ListLibraryAnimeTitles()
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != 2 {
		t.Fatalf("titles = %+v", titles)
	}
}
