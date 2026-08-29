package anilist

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	streamingEpisodePrefixPattern  = regexp.MustCompile(`(?i)(?:episode|ep\.?)\s*(\d+)`)
	streamingEpisodeLeadingPattern = regexp.MustCompile(`^(\d+)\s*[.:\-–]\s*`)
)

func EpisodeNumberFromStreamingTitle(title string) (int, bool) {
	title = strings.TrimSpace(title)
	if title == "" {
		return 0, false
	}

	if match := streamingEpisodePrefixPattern.FindStringSubmatch(title); len(match) == 2 {
		number, err := strconv.Atoi(match[1])
		if err == nil && number > 0 {
			return number, true
		}
	}

	if match := streamingEpisodeLeadingPattern.FindStringSubmatch(title); len(match) == 2 {
		number, err := strconv.Atoi(match[1])
		if err == nil && number > 0 {
			return number, true
		}
	}

	number, err := strconv.Atoi(title)
	if err == nil && number > 0 {
		return number, true
	}

	return 0, false
}

type StreamingEpisodeThumbnail struct {
	EpisodeNumber int    `json:"episodeNumber"`
	Thumbnail     string `json:"thumbnail"`
}

func (c *Client) StreamingEpisodeThumbnails(mediaID int) ([]StreamingEpisodeThumbnail, error) {
	if mediaID <= 0 {
		return nil, nil
	}

	const query = `
	query ($id: Int) {
	  Media(id: $id, type: ANIME) {
	    streamingEpisodes {
	      title
	      thumbnail
	    }
	  }
	}`

	var response struct {
		Media struct {
			StreamingEpisodes []struct {
				Title     string `json:"title"`
				Thumbnail string `json:"thumbnail"`
			} `json:"streamingEpisodes"`
		} `json:"Media"`
	}

	if err := c.query(query, map[string]any{"id": mediaID}, &response); err != nil {
		return nil, err
	}

	thumbnails := make([]StreamingEpisodeThumbnail, 0, len(response.Media.StreamingEpisodes))
	seenEpisodeNumbers := make(map[int]struct{})

	for _, episode := range response.Media.StreamingEpisodes {
		thumbnail := strings.TrimSpace(episode.Thumbnail)
		if thumbnail == "" {
			continue
		}
		episodeNumber, ok := EpisodeNumberFromStreamingTitle(episode.Title)
		if !ok {
			continue
		}
		if _, exists := seenEpisodeNumbers[episodeNumber]; exists {
			continue
		}
		seenEpisodeNumbers[episodeNumber] = struct{}{}
		thumbnails = append(thumbnails, StreamingEpisodeThumbnail{
			EpisodeNumber: episodeNumber,
			Thumbnail:     thumbnail,
		})
	}

	return thumbnails, nil
}
