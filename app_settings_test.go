package main

import (
	"path/filepath"
	"testing"

	"github.com/sunnyegg/miru/internal/paths"
	"github.com/sunnyegg/miru/internal/storage"
)

func newSettingsApp(t *testing.T) *App {
	t.Helper()
	return newSettingsStore(t, t.TempDir())
}

func newSettingsBenchApp(b *testing.B) *App {
	b.Helper()
	return newSettingsStore(b, b.TempDir())
}

func newSettingsStore(tb testing.TB, configDir string) *App {
	tb.Helper()
	dirs := paths.Dirs{Config: configDir, Cache: tb.TempDir()}
	store, err := storage.Open(filepath.Join(dirs.Config, "app_data.db"))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { _ = store.Close() })
	return &App{store: store, dirs: dirs}
}

func TestBuildSettingsViewAppliesDefaults(t *testing.T) {
	view := buildSettingsView(map[string]string{}, t.TempDir())
	if view.SyncThreshold != 85 {
		t.Fatalf("SyncThreshold = %v, want 85", view.SyncThreshold)
	}
	if !view.DownloadNotifications {
		t.Fatal("DownloadNotifications default = false, want true")
	}
	if view.NetworkMode == "" {
		t.Fatal("NetworkMode default empty")
	}
	if view.Socks5Address == "" {
		t.Fatal("Socks5Address default empty")
	}
	if view.HttpProxyURL == "" {
		t.Fatal("HttpProxyURL default empty")
	}
	if view.UpdateChannel == "" {
		t.Fatal("UpdateChannel default empty")
	}
	if view.RSSPollIntervalMinutes != 30 {
		t.Fatalf("RSSPollIntervalMinutes = %d, want 30", view.RSSPollIntervalMinutes)
	}
	if view.MaxConcurrentDownloads != 1 {
		t.Fatalf("MaxConcurrentDownloads = %d, want 1", view.MaxConcurrentDownloads)
	}
}

func TestBuildSettingsViewOverridesFromCache(t *testing.T) {
	cache := map[string]string{
		"mpv_path":                       "/usr/local/bin/mpv",
		"download_dir":                   "/data/miru",
		"sync_threshold":                 "92",
		"network_mode":                   "socks5",
		"socks5_address":                 "10.0.0.1:1080",
		"http_proxy_url":                 "http://10.0.0.2:8080",
		"anime4k_enabled":                "true",
		"discord_rpc_enabled":            "1",
		"update_channel":                 "beta",
		"rss_poll_interval_minutes":      "60",
		"download_notifications":         "false",
		"rss_auto_download":              "yes",
		"rss_auto_download_library_only": "no",
		"close_to_tray":                  "on",
		"max_concurrent_downloads":       "4",
		"seed_ratio":                     "1.5",
		"download_rate_limit":            "1024",
		"upload_rate_limit":              "512",
	}
	view := buildSettingsView(cache, t.TempDir())
	if view.MpvPath != "/usr/local/bin/mpv" {
		t.Fatalf("MpvPath = %q", view.MpvPath)
	}
	if view.SyncThreshold != 92 {
		t.Fatalf("SyncThreshold = %v", view.SyncThreshold)
	}
	if !view.Anime4KEnabled || !view.DiscordRpcEnabled || !view.RSSAutoDownload || !view.CloseToTray {
		t.Fatalf("bool overrides wrong: %+v", view)
	}
	if view.DownloadNotifications || view.RSSAutoDownloadLibraryOnly {
		t.Fatalf("bool false overrides wrong: %+v", view)
	}
	if view.NetworkMode != "socks5" || view.Socks5Address != "10.0.0.1:1080" || view.HttpProxyURL != "http://10.0.0.2:8080" {
		t.Fatalf("network overrides wrong: %+v", view)
	}
	if view.RSSPollIntervalMinutes != 60 {
		t.Fatalf("RSSPollIntervalMinutes = %d", view.RSSPollIntervalMinutes)
	}
	if view.MaxConcurrentDownloads != 4 {
		t.Fatalf("MaxConcurrentDownloads = %d", view.MaxConcurrentDownloads)
	}
	if view.DownloadRateLimit != 1024 || view.UploadRateLimit != 512 {
		t.Fatalf("rate limits: down=%d up=%d", view.DownloadRateLimit, view.UploadRateLimit)
	}
}

func TestLoadSettingsCachesAcrossCalls(t *testing.T) {
	a := newSettingsApp(t)
	if err := a.setSettings(map[string]string{"mpv_path": "/usr/bin/mpv"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		view, err := a.loadSettings()
		if err != nil {
			t.Fatal(err)
		}
		if view.MpvPath != "/usr/bin/mpv" {
			t.Fatalf("iter %d: MpvPath = %q", i, view.MpvPath)
		}
	}
	a.settingsMu.RLock()
	cached := a.settingsCache != nil
	a.settingsMu.RUnlock()
	if !cached {
		t.Fatal("expected settings cache to be populated after loadSettings")
	}
}

func TestSetSettingsInvalidatesCache(t *testing.T) {
	a := newSettingsApp(t)
	view, err := a.loadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if view.MpvPath != "" {
		t.Fatalf("initial MpvPath = %q, want empty", view.MpvPath)
	}
	if err := a.store.SetSetting("mpv_path", "/tmp/direct-write"); err != nil {
		t.Fatal(err)
	}
	view, err = a.loadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if view.MpvPath != "" {
		t.Fatalf("after direct write, cache returned %q (stale)", view.MpvPath)
	}
	if err := a.setSettings(map[string]string{"mpv_path": "/tmp/via-setter"}); err != nil {
		t.Fatal(err)
	}
	view, err = a.loadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if view.MpvPath != "/tmp/via-setter" {
		t.Fatalf("after setSettings, MpvPath = %q, want /tmp/via-setter", view.MpvPath)
	}
}

func benchCache() map[string]string {
	return map[string]string{
		"mpv_path":                       "/usr/local/bin/mpv",
		"download_dir":                   "/data/miru",
		"sync_threshold":                 "92",
		"network_mode":                   "socks5",
		"socks5_address":                 "10.0.0.1:1080",
		"http_proxy_url":                 "http://10.0.0.2:8080",
		"anime4k_enabled":                "true",
		"discord_rpc_enabled":            "1",
		"update_channel":                 "beta",
		"rss_poll_interval_minutes":      "60",
		"download_notifications":         "false",
		"rss_auto_download":              "yes",
		"rss_auto_download_library_only": "no",
		"close_to_tray":                  "on",
		"max_concurrent_downloads":       "4",
		"seed_ratio":                     "1.5",
		"download_rate_limit":            "1024",
		"upload_rate_limit":              "512",
	}
}

func BenchmarkBuildSettingsView(b *testing.B) {
	cache := benchCache()
	dir := b.TempDir()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildSettingsView(cache, dir)
	}
}

func BenchmarkLoadSettingsWarmCache(b *testing.B) {
	a := newSettingsBenchApp(b)
	if err := a.setSettings(map[string]string{
		"mpv_path":                 "/usr/local/bin/mpv",
		"download_dir":             "/data/miru",
		"sync_threshold":           "92",
		"max_concurrent_downloads": "4",
		"seed_ratio":               "1.5",
	}); err != nil {
		b.Fatal(err)
	}
	if _, err := a.loadSettings(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.loadSettings(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadSettingsColdEachCall(b *testing.B) {
	a := newSettingsBenchApp(b)
	if err := a.setSettings(map[string]string{
		"mpv_path":                 "/usr/local/bin/mpv",
		"download_dir":             "/data/miru",
		"sync_threshold":           "92",
		"max_concurrent_downloads": "4",
		"seed_ratio":               "1.5",
	}); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.invalidateSettingsCache()
		if _, err := a.loadSettings(); err != nil {
			b.Fatal(err)
		}
	}
}
