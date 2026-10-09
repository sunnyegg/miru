package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const shareImageMaxBytes = 2 << 20 // 2 MiB

var shareImageAllowedHosts = map[string]struct{}{
	"s4.anilist.co": {},
	"img.anili.st":  {},
}

func shareImageURLAllowed(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return errors.New("image URL must be https")
	}
	host := strings.ToLower(parsed.Hostname())
	if _, ok := shareImageAllowedHosts[host]; !ok {
		return fmt.Errorf("image host not allowed: %s", host)
	}
	return nil
}

func shareBannerFilename(defaultFilename string) string {
	base := filepath.Base(strings.TrimSpace(defaultFilename))
	if base == "" || base == "." || base == ".." {
		return "miru-share.png"
	}
	base = strings.ReplaceAll(base, string(os.PathSeparator), "_")
	if !strings.HasSuffix(strings.ToLower(base), ".png") {
		base += ".png"
	}
	return base
}

// FetchShareImage downloads an AniList CDN cover and returns it as base64.
func (a *App) FetchShareImage(imageURL string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	if err := shareImageURLAllowed(imageURL); err != nil {
		return "", err
	}

	client, err := a.networkHTTPClient()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(a.ctx, 20*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("image fetch returned HTTP %d", response.StatusCode)
	}

	limited := io.LimitReader(response.Body, shareImageMaxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if len(data) > shareImageMaxBytes {
		return "", errors.New("image too large")
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// SaveShareBanner prompts for a path and writes a base64-encoded PNG banner.
// An empty path with a nil error means the user cancelled the dialog.
func (a *App) SaveShareBanner(pngBase64 string, defaultFilename string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	if strings.TrimSpace(pngBase64) == "" {
		return "", errors.New("banner image is empty")
	}

	png, err := base64.StdEncoding.DecodeString(pngBase64)
	if err != nil {
		return "", fmt.Errorf("invalid banner image: %w", err)
	}
	if len(png) == 0 {
		return "", errors.New("banner image is empty")
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save share banner",
		DefaultFilename: shareBannerFilename(defaultFilename),
		Filters: []runtime.FileFilter{
			{DisplayName: "PNG image (*.png)", Pattern: "*.png"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if !strings.HasSuffix(strings.ToLower(path), ".png") {
		path += ".png"
	}
	if err := os.WriteFile(path, png, 0o644); err != nil {
		return "", err
	}
	return path, nil
}
