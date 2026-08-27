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
	"github.com/sunnyegg/miru/internal/storage"
	"github.com/sunnyegg/miru/internal/torrentx"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetSettings() (SettingsView, error) {
	if err := a.ready(); err != nil {
		return SettingsView{}, err
	}
	return a.loadSettings()
}

func (a *App) SaveSettings(view SettingsView) error {
	if err := a.ready(); err != nil {
		return err
	}
	threshold := view.SyncThreshold
	if threshold <= 0 || threshold > 100 {
		threshold = 85
	}
	view.DownloadRateLimit = normalizeRateLimit(view.DownloadRateLimit)
	view.UploadRateLimit = normalizeRateLimit(view.UploadRateLimit)
	networkConfig := networking.Config{
		Mode:    view.NetworkMode,
		Address: view.Socks5Address,
	}
	normalizedNetwork, err := networkConfig.Normalized()
	if err != nil {
		return err
	}
	if a.torrents != nil && a.torrents.Busy() {
		current, _ := a.loadSettings()
		if normalizeNetworkMode(current.NetworkMode) != normalizedNetwork.Mode ||
			strings.TrimSpace(current.Socks5Address) != normalizedNetwork.Address {
			return errors.New("stop the active download before changing networking")
		}
	}
	view.NetworkMode = normalizedNetwork.Mode
	view.Socks5Address = normalizedNetwork.Address
	pairs := map[string]string{
		"mpv_path":            strings.TrimSpace(view.MpvPath),
		"download_dir":        strings.TrimSpace(view.DownloadDir),
		"sync_threshold":      formatFloat(threshold),
		"anilist_client_id":   strings.TrimSpace(view.AnilistClientId),
		"download_rate_limit": formatInt64(view.DownloadRateLimit),
		"upload_rate_limit":   formatInt64(view.UploadRateLimit),
		"network_mode":        normalizeNetworkMode(view.NetworkMode),
		"socks5_address":      strings.TrimSpace(view.Socks5Address),
	}
	for key, value := range pairs {
		if err := a.store.SetSetting(key, value); err != nil {
			return err
		}
	}
	if a.torrents != nil {
		a.torrents.ApplyRateLimits(torrentx.RateLimits{
			Download: normalizeRateLimit(view.DownloadRateLimit),
			Upload:   normalizeRateLimit(view.UploadRateLimit),
		})
	}
	return nil
}

func (a *App) TestNetworkConnection(mode, socks5Address string) error {
	if err := a.ready(); err != nil {
		return err
	}
	client, err := (networking.Config{
		Mode:    mode,
		Address: socks5Address,
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
	view := SettingsView{SyncThreshold: 85}
	view.MpvPath, _ = a.store.GetSetting("mpv_path")
	view.DownloadDir, _ = a.store.GetSetting("download_dir")
	view.AnilistClientId, _ = a.store.GetSetting("anilist_client_id")
	if strings.TrimSpace(view.AnilistClientId) == "" {
		view.AnilistClientId = envTrim("ANILIST_CLIENT_ID")
	}
	if raw, err := a.store.GetSetting("sync_threshold"); err == nil {
		if n, err := strconv.ParseFloat(raw, 64); err == nil {
			view.SyncThreshold = n
		}
	}
	view.DownloadRateLimit = settingInt64(a.store, "download_rate_limit")
	view.UploadRateLimit = settingInt64(a.store, "upload_rate_limit")
	view.NetworkMode, _ = a.store.GetSetting("network_mode")
	if view.NetworkMode == "" {
		view.NetworkMode = networking.ModeSystem
	}
	view.Socks5Address, _ = a.store.GetSetting("socks5_address")
	if view.Socks5Address == "" {
		view.Socks5Address = "127.0.0.1:1080"
	}
	return view, nil
}

func normalizeNetworkMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return networking.ModeSystem
	}
	return mode
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatInt64(v int64) string {
	return strconv.FormatInt(normalizeRateLimit(v), 10)
}

func settingInt64(store *storage.Store, key string) int64 {
	raw, err := store.GetSetting(key)
	if err != nil {
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
