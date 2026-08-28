package update

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	ChannelStable     = "stable"
	ChannelPrerelease = "prerelease"
	ReleasesFeed      = "https://github.com/sunnyegg/miru/releases.atom"
	userAgent         = "miru"
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

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title   string     `xml:"title"`
	Content string     `xml:"content"`
	Links   []atomLink `xml:"link"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func IsDev(version string) bool {
	trimmed := strings.TrimSpace(version)
	return trimmed == "" || trimmed == "dev"
}

func ParseChannel(channel string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(channel)) {
	case ChannelStable:
		return ChannelStable, nil
	case ChannelPrerelease:
		return ChannelPrerelease, nil
	default:
		return "", fmt.Errorf("unsupported update channel %q", channel)
	}
}

func DefaultChannel(version string) string {
	canonicalVersion := canonical(version)
	if canonicalVersion != "" && semver.Prerelease(canonicalVersion) != "" {
		return ChannelPrerelease
	}
	return ChannelStable
}

func canonical(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "v") {
		trimmed = "v" + trimmed
	}
	return semver.Canonical(trimmed)
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
	currentCanonical := canonical(current)
	latestCanonical := canonical(latest)
	if currentCanonical == "" || latestCanonical == "" {
		return false
	}
	return semver.Compare(currentCanonical, latestCanonical) < 0
}

func Check(ctx context.Context, client *http.Client, current, feedURL, channel, goos, goarch string) (Info, error) {
	info := Info{Current: current}
	if IsDev(current) {
		return info, nil
	}
	release, err := FetchLatest(ctx, client, feedURL, channel, goos, goarch)
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

func FetchLatest(ctx context.Context, client *http.Client, feedURL, channel, goos, goarch string) (Info, error) {
	if _, err := AssetName(goos, goarch, "0.0.0"); err != nil {
		return Info{}, err
	}
	parsedChannel, err := ParseChannel(channel)
	if err != nil {
		return Info{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return Info{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/atom+xml, application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return Info{}, fmt.Errorf("github releases: %s", resp.Status)
	}

	var feed atomFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return Info{}, fmt.Errorf("github releases: %w", err)
	}
	if len(feed.Entries) == 0 {
		return Info{}, fmt.Errorf("github releases: empty feed")
	}

	chosen, tag, err := pickRelease(feed.Entries, parsedChannel)
	if err != nil {
		return Info{}, err
	}
	wanted, err := AssetName(goos, goarch, tag)
	if err != nil {
		return Info{}, err
	}

	repoBase := strings.TrimSuffix(strings.TrimSuffix(feedURL, "/"), "/releases.atom")
	return Info{
		Latest:     tag,
		Notes:      strings.TrimSpace(chosen.Title),
		ReleaseURL: releaseURL(chosen, repoBase, tag),
		AssetName:  wanted,
		AssetURL:   repoBase + "/releases/download/" + tag + "/" + wanted,
	}, nil
}

func pickRelease(entries []atomEntry, channel string) (atomEntry, string, error) {
	var chosen atomEntry
	var chosenTag string
	for _, entry := range entries {
		tag, ok := tagFromEntry(entry)
		if !ok {
			continue
		}
		canonicalTag := canonical(tag)
		if canonicalTag == "" {
			continue
		}
		if channel == ChannelStable && semver.Prerelease(canonicalTag) != "" {
			continue
		}
		if chosenTag == "" || semver.Compare(canonicalTag, canonical(chosenTag)) > 0 {
			chosen = entry
			chosenTag = tag
		}
	}
	if chosenTag == "" {
		if channel == ChannelStable {
			return atomEntry{}, "", fmt.Errorf("github releases: no stable release")
		}
		return atomEntry{}, "", fmt.Errorf("github releases: no matching release")
	}
	return chosen, chosenTag, nil
}

func tagFromEntry(entry atomEntry) (string, bool) {
	for _, link := range entry.Links {
		tag, ok := tagFromPath(link.Href)
		if ok {
			return tag, true
		}
	}
	return tagFromPath(entry.Title)
}

func tagFromPath(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	const marker = "/releases/tag/"
	idx := strings.LastIndex(trimmed, marker)
	if idx >= 0 {
		trimmed = trimmed[idx+len(marker):]
	}
	trimmed = strings.Trim(trimmed, "/")
	if canonical(trimmed) == "" {
		return "", false
	}
	if !strings.HasPrefix(trimmed, "v") {
		trimmed = "v" + trimmed
	}
	return trimmed, true
}

func releaseURL(entry atomEntry, repoBase, tag string) string {
	for _, link := range entry.Links {
		if strings.Contains(link.Href, "/releases/tag/") {
			return link.Href
		}
	}
	return repoBase + "/releases/tag/" + tag
}
