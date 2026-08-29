package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sunnyegg/miru/internal/mpv"
	"github.com/sunnyegg/miru/internal/networking"
	"github.com/sunnyegg/miru/internal/nyaa"
	"github.com/sunnyegg/miru/internal/torrentx"
	"github.com/sunnyegg/miru/internal/update"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetSettings() (SettingsView, error) {
	if err := a.ready(); err != nil {
		return SettingsView{}, err
	}
	return a.loadSettings()
}

func (a *App) SaveDesktopSettings(closeToTray bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.setSettings(map[string]string{
		"close_to_tray": formatBool(closeToTray),
	})
}

func (a *App) SavePlaybackSettings(mpvPath string, anime4KEnabled, discordRpcEnabled bool) error {
	if err := a.ready(); err != nil {
		return err
	}
	if anime4KEnabled {
		client, err := a.networkHTTPClient()
		if err != nil {
			return err
		}
		if err := mpv.EnsureAnime4KShaders(a.ctx, client, a.dirs.Config); err != nil {
			return fmt.Errorf("Anime4K shaders: %w", err)
		}
	}
	if err := a.setSettings(map[string]string{
		"mpv_path":            strings.TrimSpace(mpvPath),
		"anime4k_enabled":     strconv.FormatBool(anime4KEnabled),
		"discord_rpc_enabled": strconv.FormatBool(discordRpcEnabled),
	}); err != nil {
		return err
	}
	if !discordRpcEnabled {
		a.clearDiscordPresence()
		return nil
	}
	settings, err := a.loadSettings()
	if err != nil {
		return err
	}
	a.playMu.Lock()
	session := a.play
	a.playMu.Unlock()
	if session == nil {
		return nil
	}
	a.syncDiscordPresence(settings, session.animeTitle, session.episodeNum, session.lastProgress.Percent)
	return nil
}

func (a *App) SaveDownloadSettings(
	downloadDir string,
	downloadRateLimit, uploadRateLimit int64,
	maxConcurrentDownloads int,
	seedRatio float64,
	downloadNotifications, rssAutoDownload, rssAutoDownloadLibraryOnly bool,
) error {
	if err := a.ready(); err != nil {
		return err
	}
	downloadRateLimit = normalizeRateLimit(downloadRateLimit)
	uploadRateLimit = normalizeRateLimit(uploadRateLimit)
	maxConcurrentDownloads = torrentx.ClampMaxConcurrent(maxConcurrentDownloads)
	seedRatio = torrentx.ClampSeedRatio(seedRatio)
	if err := a.setSettings(map[string]string{
		"download_dir":                   strings.TrimSpace(downloadDir),
		"download_rate_limit":            formatInt64(downloadRateLimit),
		"upload_rate_limit":              formatInt64(uploadRateLimit),
		"max_concurrent_downloads":       strconv.Itoa(maxConcurrentDownloads),
		"seed_ratio":                     strconv.FormatFloat(seedRatio, 'f', -1, 64),
		"download_notifications":         formatBool(downloadNotifications),
		"rss_auto_download":              formatBool(rssAutoDownload),
		"rss_auto_download_library_only": formatBool(rssAutoDownloadLibraryOnly),
	}); err != nil {
		return err
	}
	if a.torrents != nil {
		limits := torrentx.RateLimits{
			Download: downloadRateLimit,
			Upload:   uploadRateLimit,
		}
		settings, err := a.loadSettings()
		if err != nil {
			return err
		}
		a.torrents.SetQueueConfig(limits, networkConfig(settings))
		a.torrents.ApplyRateLimits(limits)
		a.torrents.SetMaxConcurrent(maxConcurrentDownloads)
		a.torrents.SetSeedRatio(seedRatio)
	}
	return nil
}

func (a *App) SaveNetworkSettings(networkMode, socks5Address, httpProxyURL string) error {
	if err := a.ready(); err != nil {
		return err
	}
	normalizedNetwork, err := (networking.Config{
		Mode:     networkMode,
		Address:  socks5Address,
		ProxyURL: httpProxyURL,
	}).Normalized()
	if err != nil {
		return err
	}
	if a.torrents != nil && a.torrents.Busy() {
		current, loadErr := a.loadSettings()
		if loadErr != nil {
			return loadErr
		}
		currentNetwork, currentErr := (networking.Config{
			Mode:     current.NetworkMode,
			Address:  current.Socks5Address,
			ProxyURL: current.HttpProxyURL,
		}).Normalized()
		if currentErr != nil {
			return currentErr
		}
		currentKey, currentKeyErr := currentNetwork.NetworkKey()
		newKey, newKeyErr := normalizedNetwork.NetworkKey()
		if currentKeyErr != nil || newKeyErr != nil || currentKey != newKey {
			return errors.New("stop the active download before changing networking")
		}
	}
	return a.setSettings(map[string]string{
		"network_mode":   normalizedNetwork.Mode,
		"socks5_address": normalizedNetwork.Address,
		"http_proxy_url": normalizedNetwork.ProxyURL,
	})
}

func (a *App) SaveUpdateChannel(channel string) error {
	if err := a.ready(); err != nil {
		return err
	}
	parsed, err := update.ParseChannel(channel)
	if err != nil {
		return err
	}
	return a.setSettings(map[string]string{
		"update_channel": parsed,
	})
}

func (a *App) SaveRSSPollSettings(intervalMinutes int) error {
	if err := a.ready(); err != nil {
		return err
	}
	if intervalMinutes < 5 {
		intervalMinutes = 5
	}
	if intervalMinutes > 1440 {
		intervalMinutes = 1440
	}
	if err := a.setSettings(map[string]string{
		"rss_poll_interval_minutes": strconv.Itoa(intervalMinutes),
	}); err != nil {
		return err
	}
	a.restartFeedPoller()
	return nil
}

func (a *App) SaveAnilistSettings(syncThreshold float64) error {
	if err := a.ready(); err != nil {
		return err
	}
	if syncThreshold <= 0 || syncThreshold > 100 {
		syncThreshold = 85
	}
	return a.setSettings(map[string]string{
		"sync_threshold": strconv.FormatFloat(syncThreshold, 'f', -1, 64),
	})
}

func (a *App) TestNetworkConnection(mode, socks5Address, httpProxyURL string) error {
	if err := a.ready(); err != nil {
		return err
	}
	client, err := (networking.Config{
		Mode:     mode,
		Address:  socks5Address,
		ProxyURL: httpProxyURL,
	}).HTTPClient()
	if err != nil {
		return err
	}
	response, err := client.Get(nyaa.DefaultEndpoint)
	if err != nil {
		return fmt.Errorf("network connection failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("network connection returned HTTP %d", response.StatusCode)
	}
	return nil
}

func (a *App) PickMpvPath() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select MPV binary",
	})
	return path, err
}

func (a *App) PickDownloadDir() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select download folder",
	})
}

func (a *App) DetectMpv() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	settings, err := a.loadSettings()
	if err != nil {
		return "", err
	}
	return mpv.Detect(settings.MpvPath)
}

func (a *App) loadSettings() (SettingsView, error) {
	cache := a.snapshotSettings()
	return buildSettingsView(cache, a.dirs.Config), nil
}

func (a *App) snapshotSettings() map[string]string {
	a.settingsMu.RLock()
	if a.settingsCache != nil {
		out := a.settingsCache
		a.settingsMu.RUnlock()
		a.logCache("hit", len(out))
		return out
	}
	a.settingsMu.RUnlock()

	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	if a.settingsCache != nil {
		out := a.settingsCache
		a.logCache("hit (race)", len(out))
		return out
	}
	snapshot, err := a.store.AllSettings()
	if err != nil {
		a.logCache("miss (db error)", 0)
		return map[string]string{}
	}
	a.settingsCache = snapshot
	a.logCache("miss (loaded from db)", len(snapshot))
	return snapshot
}

func (a *App) logCache(event string, count int) {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.LogDebugf(a.ctx, "settings cache: %s (keys=%d)", event, count)
}

func (a *App) logCacheInvalidate() {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.LogDebugf(a.ctx, "settings cache: invalidate")
}

func (a *App) invalidateSettingsCache() {
	a.settingsMu.Lock()
	hadCache := a.settingsCache != nil
	a.settingsCache = nil
	a.settingsMu.Unlock()
	if hadCache {
		a.logCacheInvalidate()
	}
}

func buildSettingsView(cache map[string]string, configDir string) SettingsView {
	view := SettingsView{
		SyncThreshold:         85,
		SeedRatio:             torrentx.DefaultSeedRatio,
		DownloadNotifications: true,
	}
	view.MpvPath = cache["mpv_path"]
	view.Anime4KEnabled = parseSettingBool(cache["anime4k_enabled"], false)
	view.Anime4KShadersReady = mpv.Anime4KInstalled(configDir)
	view.DownloadDir = cache["download_dir"]
	if raw, ok := cache["sync_threshold"]; ok {
		if threshold, parseErr := strconv.ParseFloat(raw, 64); parseErr == nil {
			view.SyncThreshold = threshold
		}
	}
	view.DownloadRateLimit = parseSettingInt64(cache["download_rate_limit"])
	view.UploadRateLimit = parseSettingInt64(cache["upload_rate_limit"])
	view.MaxConcurrentDownloads = torrentx.ClampMaxConcurrent(parseSettingInt(cache["max_concurrent_downloads"], 1))
	if raw, ok := cache["seed_ratio"]; ok {
		if seedRatio, parseErr := strconv.ParseFloat(raw, 64); parseErr == nil {
			view.SeedRatio = torrentx.ClampSeedRatio(seedRatio)
		}
	}
	view.NetworkMode = cache["network_mode"]
	if view.NetworkMode == "" {
		view.NetworkMode = networking.ModeSystem
	}
	view.Socks5Address = cache["socks5_address"]
	if view.Socks5Address == "" {
		view.Socks5Address = "127.0.0.1:1080"
	}
	view.HttpProxyURL = cache["http_proxy_url"]
	if view.HttpProxyURL == "" {
		view.HttpProxyURL = "http://127.0.0.1:8080"
	}
	view.DiscordRpcEnabled = parseSettingBool(cache["discord_rpc_enabled"], false)
	parsedChannel, err := update.ParseChannel(cache["update_channel"])
	if err != nil {
		view.UpdateChannel = update.DefaultChannel(version)
	} else {
		view.UpdateChannel = parsedChannel
	}
	view.RSSPollIntervalMinutes = parseSettingInt(cache["rss_poll_interval_minutes"], 30)
	if view.RSSPollIntervalMinutes < 5 {
		view.RSSPollIntervalMinutes = 5
	}
	if view.RSSPollIntervalMinutes > 1440 {
		view.RSSPollIntervalMinutes = 1440
	}
	view.DownloadNotifications = parseSettingBool(cache["download_notifications"], true)
	view.RSSAutoDownload = parseSettingBool(cache["rss_auto_download"], false)
	view.RSSAutoDownloadLibraryOnly = parseSettingBool(cache["rss_auto_download_library_only"], true)
	view.CloseToTray = parseSettingBool(cache["close_to_tray"], false)
	return view
}

func (a *App) setSettings(pairs map[string]string) error {
	for key, value := range pairs {
		if err := a.store.SetSetting(key, value); err != nil {
			return err
		}
	}
	a.invalidateSettingsCache()
	return nil
}

func formatInt64(v int64) string {
	return strconv.FormatInt(normalizeRateLimit(v), 10)
}

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func parseSettingInt64(raw string) int64 {
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return normalizeRateLimit(value)
}

func normalizeRateLimit(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func parseSettingBool(raw string, defaultValue bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return defaultValue
	case "0", "false", "no", "off":
		return false
	case "1", "true", "yes", "on":
		return true
	default:
		return defaultValue
	}
}

func parseSettingInt(raw string, defaultValue int) int {
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return value
}
