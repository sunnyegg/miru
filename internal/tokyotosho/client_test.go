package tokyotosho

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchParsesFeedAndBuildsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("terms") != "Frieren & friends" ||
			query.Get("type") != CategoryAnime ||
			query.Get("searchName") != "true" {
			t.Fatalf("query = %v", query)
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <item>
      <category>Anime</category>
      <title>[Group] Frieren &amp; Friends - 01 [1080p]</title>
      <link><![CDATA[https://example.test/download/1.torrent]]></link>
      <description><![CDATA[<a href="https://example.test/download/1.torrent">Torrent Link</a><br />
<a href="magnet:?xt=urn:btih:ABCDEF0123456789&tr=udp://tracker.example:80/announce">Magnet Link</a><br />
Size: 1.2 GiB<br />
Authorized: Yes<br />
Submitter: Group]]></description>
      <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
    </item>
    <item>
      <title></title>
      <link></link>
      <description></description>
      <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
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
	if result.Title != "[Group] Frieren & Friends - 01 [1080p]" ||
		result.Link != "https://example.test/download/1.torrent" ||
		result.Magnet != "magnet:?xt=urn:btih:ABCDEF0123456789&tr=udp://tracker.example:80/announce" ||
		result.Size != "1.2 GiB" ||
		!result.Trusted {
		t.Fatalf("result = %+v", result)
	}
	if result.Published.Year() != 2006 {
		t.Fatalf("published = %v", result.Published)
	}
}

func TestSearchUsesMagnetWhenTorrentLinkMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <item>
      <title>Magnet only</title>
      <link></link>
      <description><![CDATA[<a href="magnet:?xt=urn:btih:abcdef">Magnet Link</a><br />Size: 10MB]]></description>
      <pubDate>Mon, 02 Jan 2006 15:04:05 GMT</pubDate>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	client := New()
	client.Endpoint = server.URL
	client.HTTP = server.Client()
	results, err := client.Search("anime")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Link != "magnet:?xt=urn:btih:abcdef" {
		t.Fatalf("results = %+v", results)
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
