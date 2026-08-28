package anilist

import (
	"fmt"
	"strings"
)

func (c *Client) ViewerName() (string, error) {
	var out struct {
		Viewer struct {
			Name string `json:"name"`
		} `json:"Viewer"`
	}
	if err := c.query(`query { Viewer { name } }`, nil, &out); err != nil {
		return "", err
	}
	if out.Viewer.Name == "" {
		return "", fmt.Errorf("invalid AniList token")
	}
	return out.Viewer.Name, nil
}

func (c *Client) ViewerID() (int, error) {
	var out struct {
		Viewer struct {
			ID int `json:"id"`
		} `json:"Viewer"`
	}
	if err := c.query(`query { Viewer { id } }`, nil, &out); err != nil {
		return 0, err
	}
	if out.Viewer.ID == 0 {
		return 0, fmt.Errorf("invalid AniList token")
	}
	return out.Viewer.ID, nil
}

func (c *Client) ListCurrent() ([]CurrentEntry, error) {
	return c.ListMediaList("CURRENT")
}

func (c *Client) ListMediaList(status string) ([]CurrentEntry, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	if !ValidListStatus(status) {
		return nil, fmt.Errorf("unsupported media list status %q", status)
	}

	userID, err := c.ViewerID()
	if err != nil {
		return nil, err
	}

	const q = `
	query ($page: Int, $userId: Int, $status: MediaListStatus) {
	  Page(page: $page, perPage: 50) {
	    pageInfo { hasNextPage }
	    mediaList(userId: $userId, type: ANIME, status: $status, sort: UPDATED_TIME_DESC) {
	      status
	      progress
	      score(format: POINT_100)
	      notes
	      repeat
	      private
	      startedAt { year month day }
	      completedAt { year month day }
	      media {
	        id
	        title { romaji english }
	        coverImage { large }
	        episodes
	        status
	        nextAiringEpisode { episode }
	      }
	    }
	  }
	}`
	var entries []CurrentEntry
	for page := 1; ; page++ {
		var out struct {
			Page struct {
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
				MediaList []gqlCurrentEntry `json:"mediaList"`
			} `json:"Page"`
		}
		if err := c.query(q, map[string]any{
			"page":   page,
			"userId": userID,
			"status": status,
		}, &out); err != nil {
			return nil, err
		}
		for _, entry := range out.Page.MediaList {
			entries = append(entries, entry.toCurrentEntry())
		}
		if !out.Page.PageInfo.HasNextPage {
			return entries, nil
		}
	}
}

func (c *Client) Search(search string) ([]Anime, error) {
	search = strings.TrimSpace(search)
	if search == "" {
		return []Anime{}, nil
	}
	const q = `
	query ($search: String) {
	  Page(perPage: 8) {
	    media(search: $search, type: ANIME) {
	      id
	      title { romaji english }
	      coverImage { large }
	      episodes
	      status
	      description(asHtml: false)
	      mediaListEntry { status }
	    }
	  }
	}`
	var out struct {
		Page struct {
			Media []gqlMedia `json:"media"`
		} `json:"Page"`
	}
	if err := c.query(q, map[string]any{"search": search}, &out); err != nil {
		return nil, err
	}
	animes := make([]Anime, 0, len(out.Page.Media))
	for _, m := range out.Page.Media {
		animes = append(animes, m.toAnime())
	}
	return animes, nil
}

func (c *Client) GetAnime(id int) (Anime, error) {
	const q = `
	query ($id: Int) {
	  Media(id: $id, type: ANIME) {
	    id
	    title { romaji english }
	    coverImage { large }
	    episodes
	    status
	    description(asHtml: false)
	  }
	}`
	var out struct {
		Media gqlMedia `json:"Media"`
	}
	if err := c.query(q, map[string]any{"id": id}, &out); err != nil {
		return Anime{}, err
	}
	if out.Media.ID == 0 {
		return Anime{}, fmt.Errorf("anime %d not found", id)
	}
	return out.Media.toAnime(), nil
}

func (c *Client) ListProgressForMedia(ids []int) (map[int]MediaProgress, error) {
	if len(ids) == 0 {
		return map[int]MediaProgress{}, nil
	}

	const chunkSize = 50
	const q = `
	query ($ids: [Int]) {
	  Page(perPage: 50) {
	    media(id_in: $ids, type: ANIME) {
	      id
	      episodes
	      status
	      nextAiringEpisode { episode }
	      mediaListEntry {
	        progress
	      }
	    }
	  }
	}`

	result := make(map[int]MediaProgress, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]

		var out struct {
			Page struct {
				Media []gqlMediaProgress `json:"media"`
			} `json:"Page"`
		}
		if err := c.query(q, map[string]any{"ids": chunk}, &out); err != nil {
			return nil, err
		}
		for _, media := range out.Page.Media {
			result[media.ID] = media.toMediaProgress()
		}
	}
	return result, nil
}

func (c *Client) ListProgress(mediaID int) (int, error) {
	const q = `
	query ($id: Int) {
	  Media(id: $id, type: ANIME) {
	    mediaListEntry { progress }
	  }
	}`
	var out struct {
		Media struct {
			MediaListEntry *struct {
				Progress int `json:"progress"`
			} `json:"mediaListEntry"`
		} `json:"Media"`
	}
	if err := c.query(q, map[string]any{"id": mediaID}, &out); err != nil {
		return 0, err
	}
	if out.Media.MediaListEntry == nil {
		return 0, nil
	}
	return out.Media.MediaListEntry.Progress, nil
}

func (c *Client) AiringSchedules(start, end int64) ([]AiringSchedule, error) {
	if start < 0 || end <= start {
		return nil, fmt.Errorf("invalid airing schedule range")
	}

	const q = `
	query ($page: Int, $start: Int, $end: Int) {
	  Page(page: $page, perPage: 50) {
	    pageInfo { hasNextPage }
	      airingSchedules(airingAt_greater: $start, airingAt_lesser: $end, sort: TIME) {
	      id
	      airingAt
	      episode
	      media {
	        id
	        title { romaji english }
	        coverImage { large }
	      }
	    }
	  }
	}`
	var schedules []AiringSchedule
	for page := 1; ; page++ {
		var out struct {
			Page struct {
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
				Schedules []gqlAiringSchedule `json:"airingSchedules"`
			} `json:"Page"`
		}
		if err := c.query(q, map[string]any{
			"page":  page,
			"start": start,
			"end":   end,
		}, &out); err != nil {
			return nil, err
		}
		for _, schedule := range out.Page.Schedules {
			schedules = append(schedules, schedule.toAiringSchedule())
		}
		if !out.Page.PageInfo.HasNextPage {
			return schedules, nil
		}
	}
}

func (c *Client) SaveProgress(mediaID, progress int) error {
	const q = `
	mutation ($mediaId: Int, $progress: Int) {
	  SaveMediaListEntry(mediaId: $mediaId, progress: $progress) {
	    id
	    progress
	  }
	}`
	var out struct {
		SaveMediaListEntry struct {
			Progress int `json:"progress"`
		} `json:"SaveMediaListEntry"`
	}
	return c.query(q, map[string]any{"mediaId": mediaID, "progress": progress}, &out)
}

func (c *Client) SaveListStatus(mediaID int, status string, progress int) error {
	save := ListEntrySave{
		MediaID: mediaID,
		Status:  status,
	}
	if progress >= 0 {
		save.Progress = progress
	} else {
		save.Progress = -1
	}
	return c.SaveListEntry(save)
}

func (c *Client) SaveListEntry(save ListEntrySave) error {
	status := strings.ToUpper(strings.TrimSpace(save.Status))
	if !validListStatus(status) {
		return fmt.Errorf("unsupported media list status %q", status)
	}
	if save.MediaID <= 0 {
		return fmt.Errorf("invalid media id %d", save.MediaID)
	}
	if save.Progress < -1 {
		return fmt.Errorf("invalid progress %d", save.Progress)
	}
	if save.ScoreRaw < -1 || save.ScoreRaw > 100 {
		return fmt.Errorf("invalid score %d", save.ScoreRaw)
	}
	if save.Repeat < -1 {
		return fmt.Errorf("invalid repeat %d", save.Repeat)
	}

	variables := map[string]any{
		"mediaId": save.MediaID,
		"status":  status,
	}
	fields := []string{"$mediaId: Int", "$status: MediaListStatus"}
	args := []string{"mediaId: $mediaId", "status: $status"}

	if save.Progress >= 0 {
		variables["progress"] = save.Progress
		fields = append(fields, "$progress: Int")
		args = append(args, "progress: $progress")
	}
	if save.ScoreRaw >= 0 {
		variables["scoreRaw"] = save.ScoreRaw
		fields = append(fields, "$scoreRaw: Int")
		args = append(args, "scoreRaw: $scoreRaw")
	}
	if save.SendNotes {
		variables["notes"] = save.Notes
		fields = append(fields, "$notes: String")
		args = append(args, "notes: $notes")
	}
	if save.Repeat >= 0 {
		variables["repeat"] = save.Repeat
		fields = append(fields, "$repeat: Int")
		args = append(args, "repeat: $repeat")
	}
	if save.SendPrivate {
		variables["private"] = save.Private
		fields = append(fields, "$private: Boolean")
		args = append(args, "private: $private")
	}
	if save.SendStartedAt {
		variables["startedAt"] = fuzzyDateVariable(save.StartedAt)
		fields = append(fields, "$startedAt: FuzzyDateInput")
		args = append(args, "startedAt: $startedAt")
	}
	if save.SendCompletedAt {
		variables["completedAt"] = fuzzyDateVariable(save.CompletedAt)
		fields = append(fields, "$completedAt: FuzzyDateInput")
		args = append(args, "completedAt: $completedAt")
	}

	query := fmt.Sprintf(`
	mutation (%s) {
	  SaveMediaListEntry(%s) {
	    id
	    status
	    progress
	    score(format: POINT_100)
	  }
	}`, strings.Join(fields, ", "), strings.Join(args, ", "))

	var out struct {
		SaveMediaListEntry struct {
			Status   string  `json:"status"`
			Progress int     `json:"progress"`
			Score    float64 `json:"score"`
		} `json:"SaveMediaListEntry"`
	}
	return c.query(query, variables, &out)
}

func validListStatus(status string) bool {
	return ValidListStatus(status)
}

func fuzzyDateVariable(date FuzzyDate) any {
	if date.Year <= 0 {
		return nil
	}
	return map[string]int{
		"year":  date.Year,
		"month": date.Month,
		"day":   date.Day,
	}
}

type gqlMedia struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
	} `json:"title"`
	CoverImage struct {
		Large string `json:"large"`
	} `json:"coverImage"`
	Episodes       int    `json:"episodes"`
	Status         string `json:"status"`
	Description    string `json:"description"`
	MediaListEntry *struct {
		Status string `json:"status"`
	} `json:"mediaListEntry"`
}

type gqlAiringSchedule struct {
	ID       int   `json:"id"`
	AiringAt int64 `json:"airingAt"`
	Episode  int   `json:"episode"`
	Media    struct {
		ID    int `json:"id"`
		Title struct {
			Romaji  string `json:"romaji"`
			English string `json:"english"`
		} `json:"title"`
		CoverImage struct {
			Large string `json:"large"`
		} `json:"coverImage"`
	} `json:"media"`
}

type gqlMediaProgress struct {
	ID                int    `json:"id"`
	Episodes          int    `json:"episodes"`
	Status            string `json:"status"`
	NextAiringEpisode *struct {
		Episode int `json:"episode"`
	} `json:"nextAiringEpisode"`
	MediaListEntry *struct {
		Progress int `json:"progress"`
	} `json:"mediaListEntry"`
}

func (m gqlMediaProgress) toMediaProgress() MediaProgress {
	progress := 0
	if m.MediaListEntry != nil {
		progress = m.MediaListEntry.Progress
	}
	nextAiringEpisode := 0
	if m.NextAiringEpisode != nil {
		nextAiringEpisode = m.NextAiringEpisode.Episode
	}
	return MediaProgress{
		MediaID:           m.ID,
		Progress:          progress,
		TotalEpisodes:     m.Episodes,
		MediaStatus:       m.Status,
		NextAiringEpisode: nextAiringEpisode,
	}
}

type gqlCurrentEntry struct {
	Status    string  `json:"status"`
	Progress  int     `json:"progress"`
	Score     float64 `json:"score"`
	Notes     string  `json:"notes"`
	Repeat    int     `json:"repeat"`
	Private   bool    `json:"private"`
	StartedAt struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"startedAt"`
	CompletedAt struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"completedAt"`
	Media struct {
		ID    int `json:"id"`
		Title struct {
			Romaji  string `json:"romaji"`
			English string `json:"english"`
		} `json:"title"`
		CoverImage struct {
			Large string `json:"large"`
		} `json:"coverImage"`
		Episodes          int    `json:"episodes"`
		Status            string `json:"status"`
		NextAiringEpisode *struct {
			Episode int `json:"episode"`
		} `json:"nextAiringEpisode"`
	} `json:"media"`
}

func (e gqlCurrentEntry) toCurrentEntry() CurrentEntry {
	nextAiringEpisode := 0
	if e.Media.NextAiringEpisode != nil {
		nextAiringEpisode = e.Media.NextAiringEpisode.Episode
	}
	return CurrentEntry{
		MediaID:           e.Media.ID,
		ListStatus:        e.Status,
		Progress:          e.Progress,
		ScoreRaw:          int(e.Score),
		Notes:             e.Notes,
		Repeat:            e.Repeat,
		Private:           e.Private,
		StartedAt:         FuzzyDate{Year: e.StartedAt.Year, Month: e.StartedAt.Month, Day: e.StartedAt.Day},
		CompletedAt:       FuzzyDate{Year: e.CompletedAt.Year, Month: e.CompletedAt.Month, Day: e.CompletedAt.Day},
		TitleRomaji:       e.Media.Title.Romaji,
		TitleEnglish:      e.Media.Title.English,
		CoverImage:        e.Media.CoverImage.Large,
		TotalEpisodes:     e.Media.Episodes,
		MediaStatus:       e.Media.Status,
		NextAiringEpisode: nextAiringEpisode,
	}
}

func (s gqlAiringSchedule) toAiringSchedule() AiringSchedule {
	return AiringSchedule{
		ID:           s.ID,
		AiringAt:     s.AiringAt,
		Episode:      s.Episode,
		MediaID:      s.Media.ID,
		TitleRomaji:  s.Media.Title.Romaji,
		TitleEnglish: s.Media.Title.English,
		CoverImage:   s.Media.CoverImage.Large,
	}
}

func (m gqlMedia) toAnime() Anime {
	listStatus := ""
	if m.MediaListEntry != nil {
		listStatus = m.MediaListEntry.Status
	}
	return Anime{
		ID:            m.ID,
		TitleRomaji:   m.Title.Romaji,
		TitleEnglish:  m.Title.English,
		CoverImage:    m.CoverImage.Large,
		TotalEpisodes: m.Episodes,
		Status:        m.Status,
		Synopsis:      m.Description,
		ListStatus:    listStatus,
	}
}
