package storage

import (
	"database/sql"
	"errors"
	"time"
)

type RSSFeed struct {
	ID         int64
	URL        string
	Title      string
	Enabled    bool
	LastPolled sql.NullTime
	CreatedAt  time.Time
}

type RSSFeedItem struct {
	ID        int64
	FeedID    int64
	FeedTitle string
	ItemKey   string
	Title     string
	Link      string
	Magnet    string
	Published time.Time
	IsNew     bool
	CreatedAt time.Time
}

func (s *Store) InsertRSSFeed(feedURL, title string) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO rss_feeds(url, title) VALUES(?, ?)`,
		feedURL, title,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) DeleteRSSFeed(id int64) error {
	result, err := s.db.Exec(`DELETE FROM rss_feeds WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetRSSFeedEnabled(id int64, enabled bool) error {
	enabledValue := 0
	if enabled {
		enabledValue = 1
	}
	result, err := s.db.Exec(
		`UPDATE rss_feeds SET enabled = ? WHERE id = ?`,
		enabledValue, id,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListRSSFeeds() ([]RSSFeed, error) {
	rows, err := s.db.Query(
		`SELECT id, url, title, enabled, last_polled, created_at
		 FROM rss_feeds
		 ORDER BY created_at ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feeds := []RSSFeed{}
	for rows.Next() {
		var feed RSSFeed
		var enabled int
		if err := rows.Scan(
			&feed.ID,
			&feed.URL,
			&feed.Title,
			&enabled,
			&feed.LastPolled,
			&feed.CreatedAt,
		); err != nil {
			return nil, err
		}
		feed.Enabled = enabled != 0
		feeds = append(feeds, feed)
	}
	return feeds, rows.Err()
}

func (s *Store) RSSFeedByID(id int64) (RSSFeed, error) {
	var feed RSSFeed
	var enabled int
	err := s.db.QueryRow(
		`SELECT id, url, title, enabled, last_polled, created_at
		 FROM rss_feeds WHERE id = ?`,
		id,
	).Scan(
		&feed.ID,
		&feed.URL,
		&feed.Title,
		&enabled,
		&feed.LastPolled,
		&feed.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return RSSFeed{}, ErrNotFound
	}
	if err != nil {
		return RSSFeed{}, err
	}
	feed.Enabled = enabled != 0
	return feed, nil
}

func (s *Store) UpdateRSSFeedLastPolled(id int64, polledAt time.Time) error {
	result, err := s.db.Exec(
		`UPDATE rss_feeds SET last_polled = ? WHERE id = ?`,
		polledAt.UTC(), id,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) UpsertRSSFeedItem(feedID int64, item RSSFeedItem) (bool, error) {
	result, err := s.db.Exec(
		`INSERT INTO rss_feed_items(
			feed_id, item_key, title, link, magnet, published, is_new
		) VALUES(?, ?, ?, ?, ?, ?, 1)
		ON CONFLICT(feed_id, item_key) DO NOTHING`,
		feedID,
		item.ItemKey,
		item.Title,
		item.Link,
		item.Magnet,
		item.Published.UTC(),
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *Store) ListRSSFeedItems(newOnly bool) ([]RSSFeedItem, error) {
	query := `SELECT
		items.id, items.feed_id, feeds.title, items.item_key, items.title,
		items.link, items.magnet, items.published, items.is_new, items.created_at
		FROM rss_feed_items items
		JOIN rss_feeds feeds ON feeds.id = items.feed_id`
	if newOnly {
		query += ` WHERE items.is_new = 1`
	}
	query += ` ORDER BY items.published DESC, items.id DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []RSSFeedItem{}
	for rows.Next() {
		var item RSSFeedItem
		var isNew int
		if err := rows.Scan(
			&item.ID,
			&item.FeedID,
			&item.FeedTitle,
			&item.ItemKey,
			&item.Title,
			&item.Link,
			&item.Magnet,
			&item.Published,
			&isNew,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.IsNew = isNew != 0
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) CountNewRSSFeedItems() (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM rss_feed_items WHERE is_new = 1`,
	).Scan(&count)
	return count, err
}

func (s *Store) MarkRSSFeedItemsSeen(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for index, id := range ids {
		placeholders[index] = "?"
		args[index] = id
	}
	query := `UPDATE rss_feed_items SET is_new = 0 WHERE id IN (` +
		joinPlaceholders(placeholders) + `)`
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Store) MarkAllRSSFeedItemsSeen() error {
	_, err := s.db.Exec(`UPDATE rss_feed_items SET is_new = 0 WHERE is_new = 1`)
	return err
}

func joinPlaceholders(placeholders []string) string {
	if len(placeholders) == 0 {
		return ""
	}
	joined := placeholders[0]
	for index := 1; index < len(placeholders); index++ {
		joined += "," + placeholders[index]
	}
	return joined
}
