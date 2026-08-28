package media

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nssteinbrenner/anitogo"
)

type Parsed struct {
	Title      string
	Episode    int
	HasEpisode bool
	Group      string
	Resolution string
}

func ParseFilename(path string) Parsed {
	base := filepath.Base(path)
	el := anitogo.Parse(base, anitogo.DefaultOptions)
	out := Parsed{
		Title:      strings.TrimSpace(el.AnimeTitle),
		Group:      strings.TrimSpace(el.ReleaseGroup),
		Resolution: strings.TrimSpace(el.VideoResolution),
	}
	if out.Title == "" {
		out.Title = strings.TrimSuffix(base, filepath.Ext(base))
	}
	if len(el.EpisodeNumber) == 0 {
		return out
	}
	episode, err := strconv.Atoi(el.EpisodeNumber[0])
	if err != nil {
		return out
	}
	out.Episode = episode
	out.HasEpisode = true
	return out
}

func DisplayTitle(parsed Parsed) string {
	if parsed.HasEpisode {
		return parsed.Title + " — Episode " + strconv.Itoa(parsed.Episode)
	}
	return parsed.Title
}

// EpisodeOrSingle returns a parsed episode number, or 1 when the title has only one episode.
func EpisodeOrSingle(parsed Parsed, totalEpisodes int) (int, bool) {
	if parsed.HasEpisode {
		return parsed.Episode, true
	}
	if totalEpisodes == 1 {
		return 1, true
	}
	return 0, false
}
