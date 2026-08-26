package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultEndpoint = "https://graphql.anilist.co"
	AuthorizeURL    = "https://anilist.co/api/v2/oauth/authorize"
	TokenURL        = "https://anilist.co/api/v2/oauth/token"
	ListenAddr      = "127.0.0.1:58496"
	RedirectURL     = "http://127.0.0.1:58496/callback"
)

type Anime struct {
	ID            int    `json:"id"`
	TitleRomaji   string `json:"titleRomaji"`
	TitleEnglish  string `json:"titleEnglish"`
	CoverImage    string `json:"coverImage"`
	TotalEpisodes int    `json:"totalEpisodes"`
	Status        string `json:"status"`
	Synopsis      string `json:"synopsis"`
}

type AiringSchedule struct {
	ID           int    `json:"id"`
	AiringAt     int64  `json:"airingAt"`
	Episode      int    `json:"episode"`
	MediaID      int    `json:"mediaId"`
	TitleRomaji  string `json:"titleRomaji"`
	TitleEnglish string `json:"titleEnglish"`
	CoverImage   string `json:"coverImage"`
}

type CurrentEntry struct {
	MediaID       int    `json:"mediaId"`
	Progress      int    `json:"progress"`
	TitleRomaji   string `json:"titleRomaji"`
	TitleEnglish  string `json:"titleEnglish"`
	CoverImage    string `json:"coverImage"`
	TotalEpisodes int    `json:"totalEpisodes"`
	MediaStatus   string `json:"mediaStatus"`
}

type Client struct {
	HTTP     *http.Client
	Endpoint string
	Token    string
}

func New(token string) *Client {
	return &Client{
		HTTP:     &http.Client{Timeout: 20 * time.Second},
		Endpoint: DefaultEndpoint,
		Token:    token,
	}
}

func LoginURL(clientID string) (string, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return "", fmt.Errorf("AniList client ID is empty")
	}
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", RedirectURL)
	return AuthorizeURL + "?" + q.Encode(), nil
}

func ExchangeCode(httpClient *http.Client, tokenURL, clientID, clientSecret, code string) (string, error) {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	code = strings.TrimSpace(code)
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("AniList client ID or secret is empty")
	}
	if code == "" {
		return "", fmt.Errorf("empty authorization code")
	}
	if tokenURL == "" {
		tokenURL = TokenURL
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", RedirectURL)
	form.Set("code", code)
	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("anilist token http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Message     string `json:"message"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.Error != "" {
		msg := out.Error
		if out.Message != "" {
			msg = out.Error + ": " + out.Message
		}
		return "", fmt.Errorf("anilist token: %s", msg)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", fmt.Errorf("anilist token: empty access_token")
	}
	return strings.TrimSpace(out.AccessToken), nil
}

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
	userID, err := c.ViewerID()
	if err != nil {
		return nil, err
	}

	const q = `
	query ($page: Int, $userId: Int) {
	  Page(page: $page, perPage: 50) {
	    pageInfo { hasNextPage }
	    mediaList(userId: $userId, type: ANIME, status: CURRENT, sort: UPDATED_TIME_DESC) {
	      progress
	      media {
	        id
	        title { romaji english }
	        coverImage { large }
	        episodes
	        status
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

type gqlMedia struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
	} `json:"title"`
	CoverImage struct {
		Large string `json:"large"`
	} `json:"coverImage"`
	Episodes    int    `json:"episodes"`
	Status      string `json:"status"`
	Description string `json:"description"`
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

type gqlCurrentEntry struct {
	Progress int `json:"progress"`
	Media    struct {
		ID    int `json:"id"`
		Title struct {
			Romaji  string `json:"romaji"`
			English string `json:"english"`
		} `json:"title"`
		CoverImage struct {
			Large string `json:"large"`
		} `json:"coverImage"`
		Episodes int    `json:"episodes"`
		Status   string `json:"status"`
	} `json:"media"`
}

func (e gqlCurrentEntry) toCurrentEntry() CurrentEntry {
	return CurrentEntry{
		MediaID:       e.Media.ID,
		Progress:      e.Progress,
		TitleRomaji:   e.Media.Title.Romaji,
		TitleEnglish:  e.Media.Title.English,
		CoverImage:    e.Media.CoverImage.Large,
		TotalEpisodes: e.Media.Episodes,
		MediaStatus:   e.Media.Status,
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
	return Anime{
		ID:            m.ID,
		TitleRomaji:   m.Title.Romaji,
		TitleEnglish:  m.Title.English,
		CoverImage:    m.CoverImage.Large,
		TotalEpisodes: m.Episodes,
		Status:        m.Status,
		Synopsis:      m.Description,
	}
}

func (c *Client) query(query string, variables map[string]any, dest any) error {
	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("anilist http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return err
	}
	if len(envelope.Errors) > 0 {
		return fmt.Errorf("anilist: %s", envelope.Errors[0].Message)
	}
	if dest == nil || len(envelope.Data) == 0 {
		return nil
	}
	return json.Unmarshal(envelope.Data, dest)
}
