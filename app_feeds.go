package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sunnyegg/miru/internal/rssfeed"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	defaultRSSPollInterval = 30 * time.Minute
)

type RSSFeedView struct {
	ID         int64  `json:"id"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Enabled    bool   `json:"enabled"`
	LastPolled string `json:"lastPolled"`
	NewCount   int    `json:"newCount"`
}

type RSSFeedItemView struct {
	ID        int64  `json:"id"`
	FeedID    int64  `json:"feedId"`
	FeedTitle string `json:"feedTitle"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Magnet    string `json:"magnet"`
	Published string `json:"published"`
	IsNew     bool   `json:"isNew"`
}

type feedPoller struct {
	app      *App
	cancel   context.CancelFunc
	interval time.Duration
}

func (a *App) startFeedPoller() {
	a.feedPollerMu.Lock()
	defer a.feedPollerMu.Unlock()
	if a.feedPoller != nil {
		a.feedPoller.stop()
	}
	interval := a.rssPollInterval()
	poller := &feedPoller{
		app:      a,
		interval: interval,
	}
	poller.start()
	a.feedPoller = poller
}

func (a *App) restartFeedPoller() {
	a.startFeedPoller()
}

func (p *feedPoller) start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go p.run(ctx)
}

func (p *feedPoller) stop() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
}

func (p *feedPoller) run(ctx context.Context) {
	p.app.pollFeeds()
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.app.pollFeeds()
		}
	}
}

func (a *App) rssPollInterval() time.Duration {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultRSSPollInterval
	}
	return clampRSSPollInterval(settings.RSSPollIntervalMinutes)
}

func clampRSSPollInterval(minutes int) time.Duration {
	if minutes < 5 {
		minutes = 30
	}
	if minutes > 1440 {
		minutes = 1440
	}
	return time.Duration(minutes) * time.Minute
}

func (a *App) ListRSSFeeds() ([]RSSFeedView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	feeds, err := a.store.ListRSSFeeds()
	if err != nil {
		return nil, err
	}
	items, err := a.store.ListRSSFeedItems(true)
	if err != nil {
		return nil, err
	}
	newByFeed := map[int64]int{}
	for _, item := range items {
		newByFeed[item.FeedID]++
	}
	out := make([]RSSFeedView, 0, len(feeds))
	for _, feed := range feeds {
		view := RSSFeedView{
			ID:       feed.ID,
			URL:      feed.URL,
			Title:    feed.Title,
			Enabled:  feed.Enabled,
			NewCount: newByFeed[feed.ID],
		}
		if feed.LastPolled.Valid {
			view.LastPolled = feed.LastPolled.Time.UTC().Format(time.RFC3339)
		}
		out = append(out, view)
	}
	return out, nil
}

func (a *App) AddRSSFeed(feedURL, title string) (RSSFeedView, error) {
	if err := a.ready(); err != nil {
		return RSSFeedView{}, err
	}
	feedURL = strings.TrimSpace(feedURL)
	title = strings.TrimSpace(title)
	if feedURL == "" {
		return RSSFeedView{}, errors.New("rss feed url is required")
	}
	parsedURL, err := url.Parse(feedURL)
	if err != nil {
		return RSSFeedView{}, fmt.Errorf("parse rss feed url: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return RSSFeedView{}, errors.New("rss feed url must use http or https")
	}
	if title == "" {
		title = parsedURL.Host
	}
	id, err := a.store.InsertRSSFeed(feedURL, title)
	if err != nil {
		return RSSFeedView{}, err
	}
	go a.pollFeed(id)
	feed, err := a.store.RSSFeedByID(id)
	if err != nil {
		return RSSFeedView{}, err
	}
	return RSSFeedView{
		ID:      feed.ID,
		URL:     feed.URL,
		Title:   feed.Title,
		Enabled: feed.Enabled,
	}, nil
}

func (a *App) RemoveRSSFeed(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.DeleteRSSFeed(id)
}

func (a *App) SetRSSFeedEnabled(id int64, enabled bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.store.SetRSSFeedEnabled(id, enabled)
}

func (a *App) ListRSSFeedItems(newOnly bool) ([]RSSFeedItemView, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	items, err := a.store.ListRSSFeedItems(newOnly)
	if err != nil {
		return nil, err
	}
	out := make([]RSSFeedItemView, 0, len(items))
	for _, item := range items {
		out = append(out, feedItemView(item))
	}
	return out, nil
}

func (a *App) CountNewRSSFeedItems() (int, error) {
	if err := a.ready(); err != nil {
		return 0, err
	}
	return a.store.CountNewRSSFeedItems()
}

func (a *App) MarkRSSFeedItemsSeen(ids []int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	if err := a.store.MarkRSSFeedItemsSeen(ids); err != nil {
		return err
	}
	a.emitFeedsUpdated()
	return nil
}

func (a *App) MarkAllRSSFeedItemsSeen() error {
	if err := a.ready(); err != nil {
		return err
	}
	if err := a.store.MarkAllRSSFeedItemsSeen(); err != nil {
		return err
	}
	a.emitFeedsUpdated()
	return nil
}

func (a *App) PollRSSFeedsNow() error {
	if err := a.ready(); err != nil {
		return err
	}
	a.pollFeeds()
	return nil
}

func (a *App) pollFeeds() {
	feeds, err := a.store.ListRSSFeeds()
	if err != nil {
		a.logDebugErr("list rss feeds", err)
		return
	}
	for _, feed := range feeds {
		if !feed.Enabled {
			continue
		}
		a.pollFeed(feed.ID)
	}
}

func (a *App) pollFeed(feedID int64) {
	feed, err := a.store.RSSFeedByID(feedID)
	if err != nil {
		a.logDebugErr("load rss feed", err)
		return
	}
	if !feed.Enabled {
		return
	}
	httpClient, err := a.networkHTTPClient()
	if err != nil {
		a.logDebugErr("rss feed http client", err)
		return
	}
	items, err := rssfeed.NewWithHTTP(httpClient).Fetch(feed.URL)
	if err != nil {
		a.logDebugErr(fmt.Sprintf("poll rss feed %q", feed.Title), err)
		return
	}
	newCount := 0
	autoQueued := 0
	settings, settingsErr := a.loadSettings()
	var libraryTitles []string
	if settingsErr == nil && settings.RSSAutoDownload && settings.RSSAutoDownloadLibraryOnly {
		libraryTitles, _ = a.store.ListLibraryAnimeTitles()
	}
	for _, item := range items {
		inserted, upsertErr := a.store.UpsertRSSFeedItem(feed.ID, storage.RSSFeedItem{
			ItemKey:   item.Key,
			Title:     item.Title,
			Link:      item.Link,
			Magnet:    item.Magnet,
			Published: item.Published,
		})
		if upsertErr != nil {
			a.logDebugErr("store rss feed item", upsertErr)
			continue
		}
		if !inserted {
			continue
		}
		newCount++
		if settingsErr != nil || !settings.RSSAutoDownload {
			continue
		}
		source := rssfeed.TorrentSource(item.Magnet, item.Link)
		if !rssfeed.ShouldAutoDownload(
			settings.RSSAutoDownload,
			settings.RSSAutoDownloadLibraryOnly,
			item.Title,
			source != "",
			libraryTitles,
		) {
			continue
		}
		if err := a.startTorrent(source, nil); err != nil {
			a.logDebugErr(fmt.Sprintf("auto-queue rss item %q", item.Title), err)
			continue
		}
		autoQueued++
	}
	if err := a.store.UpdateRSSFeedLastPolled(feed.ID, time.Now().UTC()); err != nil {
		a.logDebugErr("update rss feed last polled", err)
	}
	if newCount > 0 {
		a.emitFeedsUpdated()
	}
	if autoQueued > 0 {
		a.emitRSSAutoQueued(autoQueued, settings.DownloadNotifications)
	}
}

func (a *App) emitFeedsUpdated() {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "feeds:updated", true)
}

type rssAutoQueuedEvent struct {
	Count int `json:"count"`
}

func (a *App) emitRSSAutoQueued(count int, notify bool) {
	if a.ctx == nil || count <= 0 {
		return
	}
	if notify {
		runtime.EventsEmit(a.ctx, "rss:auto_queued", rssAutoQueuedEvent{Count: count})
	}
}

func feedItemView(item storage.RSSFeedItem) RSSFeedItemView {
	magnet := item.Magnet
	if magnet == "" && strings.HasPrefix(item.Link, "magnet:") {
		magnet = item.Link
	}
	return RSSFeedItemView{
		ID:        item.ID,
		FeedID:    item.FeedID,
		FeedTitle: item.FeedTitle,
		Title:     item.Title,
		Link:      item.Link,
		Magnet:    magnet,
		Published: item.Published.UTC().Format(time.RFC3339),
		IsNew:     item.IsNew,
	}
}
