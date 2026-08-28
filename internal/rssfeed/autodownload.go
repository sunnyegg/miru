package rssfeed

import (
	"strings"
)

// ShouldAutoDownload decides whether a new RSS item should be queued for download.
func ShouldAutoDownload(
	autoDownloadEnabled bool,
	libraryOnly bool,
	itemTitle string,
	hasMagnet bool,
	libraryTitles []string,
) bool {
	if !autoDownloadEnabled || !hasMagnet {
		return false
	}
	if !libraryOnly {
		return true
	}
	return TitleMatchesLibrary(itemTitle, libraryTitles)
}

// TitleMatchesLibrary reports whether itemTitle contains a library anime title.
func TitleMatchesLibrary(itemTitle string, libraryTitles []string) bool {
	normalizedItem := strings.ToLower(strings.TrimSpace(itemTitle))
	if normalizedItem == "" {
		return false
	}
	for _, title := range libraryTitles {
		normalizedTitle := strings.ToLower(strings.TrimSpace(title))
		if normalizedTitle == "" {
			continue
		}
		if strings.Contains(normalizedItem, normalizedTitle) {
			return true
		}
	}
	return false
}

// TorrentSource returns the magnet link or magnet-style link for an RSS item.
func TorrentSource(magnet, link string) string {
	if strings.TrimSpace(magnet) != "" {
		return strings.TrimSpace(magnet)
	}
	link = strings.TrimSpace(link)
	if strings.HasPrefix(link, "magnet:") {
		return link
	}
	return ""
}
