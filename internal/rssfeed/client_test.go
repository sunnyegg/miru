package rssfeed

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchParsesNyaaStyleFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:nyaa="https://nyaa.si/xmlns/nyaa">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>[Group] Show - 01</title>
      <link>https://nyaa.si/download/1.torrent</link>
      <guid>https://nyaa.si/view/1</guid>
      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>
      <nyaa:infoHash>abcdef0123456789</nyaa:infoHash>
      <nyaa:size>1.2 GiB</nyaa:size>
      <nyaa:seeders>10</nyaa:seeders>
      <nyaa:leechers>2</nyaa:leechers>
      <nyaa:downloads>50</nyaa:downloads>
      <nyaa:trusted>Yes</nyaa:trusted>
      <nyaa:remake>No</nyaa:remake>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	client := New()
	client.HTTP = server.Client()
	items, err := client.Fetch(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %+v", items)
	}
	item := items[0]
	if item.Key != "https://nyaa.si/view/1" || item.Title != "[Group] Show - 01" {
		t.Fatalf("item = %+v", item)
	}
	if !strings.HasPrefix(item.Magnet, "magnet:?xt=urn:btih:abcdef0123456789") {
		t.Fatalf("magnet = %q", item.Magnet)
	}
	if item.Seeders != 10 || item.Leechers != 2 || item.Downloads != 50 || !item.Trusted {
		t.Fatalf("counts = %+v", item)
	}
}

func TestFetchParsesMagnetInDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <item>
      <title>Fansub Release</title>
      <link>https://example.com/page</link>
      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>
      <description>Size: 500 MiB magnet:?xt=urn:btih:deadbeef</description>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	client := New()
	client.HTTP = server.Client()
	items, err := client.Fetch(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %+v", items)
	}
	item := items[0]
	if item.Magnet != "magnet:?xt=urn:btih:deadbeef" {
		t.Fatalf("magnet = %q", item.Magnet)
	}
	if item.Size != "500 MiB" {
		t.Fatalf("size = %q", item.Size)
	}
}

func TestFetchRejectsEmptyURL(t *testing.T) {
	client := New()
	if _, err := client.Fetch("  "); err == nil {
		t.Fatal("expected empty url error")
	}
}
