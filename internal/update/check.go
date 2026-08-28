package update

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const (
	LatestPage = "https://github.com/sunnyegg/miru/releases/latest"
	userAgent  = "miru"
)

type Info struct {
	Current    string
	Latest     string
	Available  bool
	Notes      string
	ReleaseURL string
	AssetName  string
	AssetURL   string
}

func IsDev(version string) bool {
	trimmed := strings.TrimSpace(version)
	return trimmed == "" || trimmed == "dev"
}

func AssetName(goos, goarch, version string) (string, error) {
	ver := strings.TrimPrefix(strings.TrimSpace(version), "v")
	if ver == "" {
		return "", fmt.Errorf("missing version")
	}
	switch goos {
	case "linux":
		if goarch != "amd64" {
			return "", fmt.Errorf("no release build for linux/%s", goarch)
		}
		return "miru-" + ver + "-linux-amd64", nil
	case "windows":
		if goarch != "amd64" {
			return "", fmt.Errorf("no release build for windows/%s", goarch)
		}
		return "miru-" + ver + "-windows-amd64.exe", nil
	case "darwin":
		return "miru-" + ver + "-mac-universal.zip", nil
	default:
		return "", fmt.Errorf("unsupported OS %s", goos)
	}
}

func Newer(current, latest string) bool {
	currentParts, currentOK := parseVersion(current)
	latestParts, latestOK := parseVersion(latest)
	if !currentOK || !latestOK {
		return false
	}
	for i := range currentParts {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

func Check(ctx context.Context, client *http.Client, current, latestURL, goos, goarch string) (Info, error) {
	info := Info{Current: current}
	if IsDev(current) {
		return info, nil
	}
	release, err := FetchLatest(ctx, client, latestURL, goos, goarch)
	if err != nil {
		return Info{}, err
	}
	info.Latest = release.Latest
	info.Notes = release.Notes
	info.ReleaseURL = release.ReleaseURL
	info.AssetName = release.AssetName
	info.AssetURL = release.AssetURL
	info.Available = Newer(current, release.Latest)
	return info, nil
}

func FetchLatest(ctx context.Context, client *http.Client, latestURL, goos, goarch string) (Info, error) {
	if _, err := AssetName(goos, goarch, "0.0.0"); err != nil {
		return Info{}, err
	}

	pageClient := *client
	pageClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return Info{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := pageClient.Do(req)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return Info{}, fmt.Errorf("github releases: %s", resp.Status)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return Info{}, fmt.Errorf("github releases: missing Location")
	}
	resolved, err := resp.Request.URL.Parse(location)
	if err != nil {
		return Info{}, fmt.Errorf("github releases: %w", err)
	}

	const marker = "/releases/tag/"
	idx := strings.LastIndex(resolved.Path, marker)
	if idx < 0 {
		return Info{}, fmt.Errorf("github releases: unexpected Location")
	}
	tag := strings.Trim(resolved.Path[idx+len(marker):], "/")
	if tag == "" {
		return Info{}, fmt.Errorf("github releases: missing tag name")
	}
	wanted, err := AssetName(goos, goarch, tag)
	if err != nil {
		return Info{}, err
	}

	repoBase := strings.TrimSuffix(strings.TrimSuffix(latestURL, "/"), "/releases/latest")
	return Info{
		Latest:     tag,
		ReleaseURL: resolved.String(),
		AssetName:  wanted,
		AssetURL:   repoBase + "/releases/download/" + tag + "/" + wanted,
	}, nil
}

func parseVersion(raw string) ([3]int, bool) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "v")
	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var parsed [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		parsed[i] = n
	}
	return parsed, true
}
