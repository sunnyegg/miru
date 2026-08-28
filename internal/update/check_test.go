package update

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	}
	for _, tc := range cases {
		got := Newer(tc.current, tc.latest)
		if got != tc.want {
			t.Fatalf("Newer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
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

	info, err := Check(context.Background(), server.Client(), "dev", server.URL, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if info.Available || info.Current != "dev" {
		t.Fatalf("got %+v", info)
	}
}

func TestCheckParsesLatestRedirect(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		if r.URL.Path != "/releases/latest" {
			t.Errorf("path %s", r.URL.Path)
		}
		http.Redirect(w, r, "/sunnyegg/miru/releases/tag/v1.2.0", http.StatusFound)
	}))
	t.Cleanup(server.Close)

	latestURL := server.URL + "/releases/latest"
	info, err := Check(context.Background(), server.Client(), "v1.0.0", latestURL, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Available || info.Latest != "v1.2.0" || info.AssetName != "miru-1.2.0-linux-amd64" {
		t.Fatalf("got %+v", info)
	}
	wantAsset := server.URL + "/releases/download/v1.2.0/miru-1.2.0-linux-amd64"
	if info.AssetURL != wantAsset {
		t.Fatalf("asset url %q, want %q", info.AssetURL, wantAsset)
	}
}

func TestCheckForbiddenIsShort(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	t.Cleanup(server.Close)

	_, err := Check(context.Background(), server.Client(), "v1.0.0", server.URL+"/releases/latest", "linux", "amd64")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "github releases: 403 Forbidden" {
		t.Fatalf("got %q", err)
	}
}
