package nyaa

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSearchParsesFeedAndBuildsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("page") != "rss" ||
			query.Get("q") != "Frieren & friends" ||
			query.Get("c") != CategoryEnglish ||
			query.Get("f") != "0" ||
			query.Get("s") != "id" ||
			query.Get("o") != "desc" {
			t.Fatalf("query = %v", query)
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:nyaa="https://nyaa.si/xmlns/nyaa">
  <channel>
    <item>
      <title>[Group] Frieren &amp; Friends - 01 [1080p]</title>
      <link>https://nyaa.si/download/1.torrent</link>
      <guid>https://nyaa.si/view/1</guid>
      <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>
      <nyaa:infoHash>ABCDEF0123456789</nyaa:infoHash>
      <nyaa:size>1.2 GiB</nyaa:size>
      <nyaa:seeders>120</nyaa:seeders>
      <nyaa:leechers>4</nyaa:leechers>
      <nyaa:downloads>900</nyaa:downloads>
      <nyaa:trusted>Yes</nyaa:trusted>
      <nyaa:remake>No</nyaa:remake>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	client := New()
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	results, err := client.Search("Frieren & friends")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %+v", results)
	}
	result := results[0]
	if result.InfoHash != "abcdef0123456789" ||
		result.Seeders != 120 ||
		result.Leechers != 4 ||
		result.Downloads != 900 ||
		!result.Trusted ||
		result.Remake {
		t.Fatalf("result = %+v", result)
	}
	if !strings.HasPrefix(result.Magnet(), "magnet:?xt=urn:btih:abcdef0123456789&") {
		t.Fatalf("magnet = %s", result.Magnet())
	}
	if !strings.Contains(result.Magnet(), url.QueryEscape(result.Title)) {
		t.Fatalf("magnet title missing: %s", result.Magnet())
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	client := New()
	if _, err := client.Search("  "); err == nil {
		t.Fatal("expected empty query error")
	}
}

func TestSearchRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	client := New()
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	if _, err := client.Search("anime"); err == nil {
		t.Fatal("expected HTTP error")
	}
}
