package anilist

import "fmt"

const maxPrequelWalk = 20

type relatedMedia struct {
	ID        int    `json:"id"`
	Episodes  int    `json:"episodes"`
	Format    string `json:"format"`
	Relations struct {
		Edges []relationEdge `json:"edges"`
	} `json:"relations"`
}

type relationEdge struct {
	RelationType string `json:"relationType"`
	Node         struct {
		ID       int    `json:"id"`
		Episodes int    `json:"episodes"`
		Format   string `json:"format"`
	} `json:"node"`
}

type seasonLink struct {
	ID       int
	Episodes int
}

// SeasonEpisode maps a filename episode onto one AniList season.
// Fansub groups often use absolute numbers (Re:Zero - 79) while AniList is per-season (S4E13).
func SeasonEpisode(parsed, seasonLength, prequelEpisodes int) (int, bool) {
	if parsed <= 0 {
		return parsed, false
	}
	if seasonLength > 0 && parsed <= seasonLength {
		return parsed, true
	}
	if prequelEpisodes > 0 && parsed > prequelEpisodes {
		relative := parsed - prequelEpisodes
		if relative > 0 && (seasonLength <= 0 || relative <= seasonLength) {
			return relative, true
		}
	}
	if seasonLength <= 0 {
		return parsed, true
	}
	return parsed, false
}

func isSeasonFormat(format string) bool {
	switch format {
	case "TV", "TV_SHORT", "ONA":
		return true
	default:
		return false
	}
}

func (m relatedMedia) tvPrequel() (seasonLink, bool) {
	for _, edge := range m.Relations.Edges {
		if edge.RelationType != "PREQUEL" {
			continue
		}
		if !isSeasonFormat(edge.Node.Format) {
			continue
		}
		return seasonLink{ID: edge.Node.ID, Episodes: edge.Node.Episodes}, true
	}
	return seasonLink{}, false
}

func (c *Client) MapSeasonEpisode(mediaID, parsed int) (int, error) {
	if parsed <= 0 || mediaID <= 0 {
		return parsed, nil
	}
	start, err := c.mediaWithRelations(mediaID)
	if err != nil {
		return parsed, err
	}
	if start.Episodes > 0 && parsed <= start.Episodes {
		return parsed, nil
	}
	prequelEpisodes, err := c.tvPrequelEpisodeTotal(start)
	if err != nil {
		return parsed, err
	}
	mapped, ok := SeasonEpisode(parsed, start.Episodes, prequelEpisodes)
	if !ok {
		return parsed, fmt.Errorf("episode %d does not fit this season (%d episodes)", parsed, start.Episodes)
	}
	return mapped, nil
}

func (c *Client) mediaWithRelations(id int) (relatedMedia, error) {
	const q = `
	query ($id: Int) {
	  Media(id: $id, type: ANIME) {
	    id
	    episodes
	    format
	    relations {
	      edges {
	        relationType
	        node { id episodes format }
	      }
	    }
	  }
	}`
	var out struct {
		Media relatedMedia `json:"Media"`
	}
	if err := c.query(q, map[string]any{"id": id}, &out); err != nil {
		return relatedMedia{}, err
	}
	if out.Media.ID == 0 {
		return relatedMedia{}, fmt.Errorf("anime %d not found", id)
	}
	return out.Media, nil
}

func (c *Client) tvPrequelEpisodeTotal(start relatedMedia) (int, error) {
	total := 0
	seen := map[int]struct{}{start.ID: {}}
	current := start
	for i := 0; i < maxPrequelWalk; i++ {
		prequel, ok := current.tvPrequel()
		if !ok {
			return total, nil
		}
		if _, dup := seen[prequel.ID]; dup {
			return total, nil
		}
		seen[prequel.ID] = struct{}{}
		total += prequel.Episodes
		next, err := c.mediaWithRelations(prequel.ID)
		if err != nil {
			return 0, err
		}
		current = next
	}
	return total, nil
}
