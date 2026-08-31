package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewer(t *testing.T) {
	t.Parallel()
	cases := []struct {
		current string
		latest  string
		want    bool
	}{
		{current: "v1.0.0", latest: "v1.0.1", want: true},
		{current: "1.0.0", latest: "v1.0.1", want: true},
		{current: "v1.0.1", latest: "v1.0.0", want: false},
		{current: "v1.0.0", latest: "v1.0.0", want: false},
		{current: "v1.2.0", latest: "v2.0.0", want: true},
		{current: "dev", latest: "v1.0.0", want: false},
		{current: "v0.1.0-alpha", latest: "v0.1.0", want: true},
		{current: "v0.1.0-alpha", latest: "v0.1.0-alpha.1", want: true},
		{current: "v0.1.0", latest: "v0.1.0-alpha", want: false},
		{current: "v0.0.9", latest: "v0.1.0-alpha", want: true},
	}
	for _, tc := range cases {
		got := Newer(tc.current, tc.latest)
		if got != tc.want {
			t.Fatalf("Newer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}

func TestDefaultChannel(t *testing.T) {
	t.Parallel()
	if got := DefaultChannel("v0.1.0-alpha"); got != ChannelPrerelease {
		t.Fatalf("prerelease build: got %q", got)
	}
	if got := DefaultChannel("v0.1.0"); got != ChannelStable {
		t.Fatalf("stable build: got %q", got)
	}
	if got := DefaultChannel("dev"); got != ChannelStable {
		t.Fatalf("dev build: got %q", got)
	}
}

func TestAssetName(t *testing.T) {
	t.Parallel()
	name, err := AssetName("linux", "amd64", "v1.2.0")
	if err != nil || name != "miru-1.2.0-linux-amd64" {
		t.Fatalf("linux/amd64: got %q %v", name, err)
	}
	name, err = AssetName("windows", "amd64", "1.2.0")
	if err != nil || name != "miru-1.2.0-windows-amd64.exe" {
		t.Fatalf("windows/amd64: got %q %v", name, err)
	}
	name, err = AssetName("darwin", "arm64", "v1.2.0")
	if err != nil || name != "miru-1.2.0-mac-universal.zip" {
		t.Fatalf("darwin/arm64: got %q %v", name, err)
	}
	name, err = AssetName("linux", "amd64", "v0.1.0-alpha")
	if err != nil || name != "miru-0.1.0-alpha-linux-amd64" {
		t.Fatalf("linux alpha: got %q %v", name, err)
	}
	if _, err := AssetName("linux", "arm64", "v1.2.0"); err == nil {
		t.Fatal("expected error for linux/arm64")
	}
}

func TestCheckDevSkipsNetwork(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("dev check must not call GitHub")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "dev", server.URL, ChannelStable, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Available || info.Current != "dev" {
		t.Fatalf("got %+v", info)
	}
}

func TestCheckParsesAtomFeed(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		switch r.URL.Path {
		case "/releases.atom":
			w.Header().Set("Content-Type", "application/atom+xml")
			_, _ = w.Write([]byte(atomFeedXML))
		default:
			if strings.HasPrefix(r.URL.Path, "/releases/tags/") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			t.Errorf("path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelPrerelease, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.Latest != "v0.1.0-alpha" || info.AssetName != "miru-0.1.0-alpha-linux-amd64" {
		t.Fatalf("got %+v", info)
	}
	wantAsset := server.URL + "/releases/download/v0.1.0-alpha/miru-0.1.0-alpha-linux-amd64"
	if info.AssetURL != wantAsset {
		t.Fatalf("asset url %q, want %q", info.AssetURL, wantAsset)
	}
}

func TestCheckStableUsesLatestRedirect(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases.atom" {
			t.Error("stable channel must not fetch the atom feed")
		}
		switch r.URL.Path {
		case "/releases/latest":
			http.Redirect(w, r, "/sunnyegg/miru/releases/tag/v0.0.9", http.StatusFound)
		case "/releases/tags/v0.0.9":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelStable, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.Latest != "v0.0.9" || info.AssetName != "miru-0.0.9-linux-amd64" {
		t.Fatalf("got %+v", info)
	}
}

func TestCheckStableEmptyWhenOnlyPrerelease(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/sunnyegg/miru/releases", http.StatusFound)
	}))
	t.Cleanup(server.Close)

	_, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelStable, "linux", "amd64")
	if err == nil || err.Error() != "github releases: no stable release" {
		t.Fatalf("got %v", err)
	}
}

func TestCheckForbiddenIsShort(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	t.Cleanup(server.Close)

	_, err := Check(context.Background(), server.Client(), "v1.0.0", server.URL, ChannelStable, "linux", "amd64")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "github releases: 403 Forbidden" {
		t.Fatalf("got %q", err)
	}
}

func TestCheckStableFetchesReleaseBody(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			http.Redirect(w, r, "/sunnyegg/miru/releases/tag/v0.0.9", http.StatusFound)
		case "/releases/tags/v0.0.9":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"body":"## What\u0027s Changed\n- foo\n- bar\n"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelStable, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Notes != "## What's Changed\n- foo\n- bar\n" {
		t.Fatalf("notes: got %q", info.Notes)
	}
}

func TestCheckStableFallsBackToTagWhenBodyMissing(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			http.Redirect(w, r, "/sunnyegg/miru/releases/tag/v0.0.9", http.StatusFound)
		case "/releases/tags/v0.0.9":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"body":""}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelStable, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Notes != "v0.0.9" {
		t.Fatalf("fallback notes: got %q", info.Notes)
	}
}

func TestCheckStableFallsBackWhenBodyRequestFails(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			http.Redirect(w, r, "/sunnyegg/miru/releases/tag/v0.0.9", http.StatusFound)
		case "/releases/tags/v0.0.9":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelStable, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Notes != "v0.0.9" {
		t.Fatalf("fallback notes: got %q", info.Notes)
	}
}

func TestCheckPrereleaseFetchesReleaseBody(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases.atom":
			w.Header().Set("Content-Type", "application/atom+xml")
			_, _ = w.Write([]byte(atomFeedXML))
		case "/releases/tags/v0.1.0-alpha":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"body":"## Pre-release notes\n"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	info, err := Check(context.Background(), server.Client(), "v0.0.8", server.URL, ChannelPrerelease, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Notes != "## Pre-release notes\n" {
		t.Fatalf("notes: got %q", info.Notes)
	}
}

const atomFeedXML = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <title>v0.1.0-alpha</title>
    <link rel="alternate" href="https://github.com/sunnyegg/miru/releases/tag/v0.1.0-alpha"/>
  </entry>
  <entry>
    <title>v0.0.9</title>
    <link rel="alternate" href="https://github.com/sunnyegg/miru/releases/tag/v0.0.9"/>
  </entry>
</feed>
`
