package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

var ListStatuses = []string{
	"CURRENT",
	"COMPLETED",
	"PLANNING",
	"PAUSED",
	"DROPPED",
	"REPEATING",
}

func ValidListStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "CURRENT", "COMPLETED", "PLANNING", "PAUSED", "DROPPED", "REPEATING":
		return true
	default:
		return false
	}
}

type Anime struct {
	ID            int    `json:"id"`
	TitleRomaji   string `json:"titleRomaji"`
	TitleEnglish  string `json:"titleEnglish"`
	CoverImage    string `json:"coverImage"`
	TotalEpisodes int    `json:"totalEpisodes"`
	Status        string `json:"status"`
	Synopsis      string `json:"synopsis"`
	ListStatus    string `json:"listStatus"`
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

type FuzzyDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type CurrentEntry struct {
	MediaID       int       `json:"mediaId"`
	ListStatus    string    `json:"listStatus"`
	Progress      int       `json:"progress"`
	ScoreRaw      int       `json:"scoreRaw"`
	Notes         string    `json:"notes"`
	Repeat        int       `json:"repeat"`
	Private       bool      `json:"private"`
	StartedAt     FuzzyDate `json:"startedAt"`
	CompletedAt   FuzzyDate `json:"completedAt"`
	TitleRomaji   string    `json:"titleRomaji"`
	TitleEnglish  string    `json:"titleEnglish"`
	CoverImage    string    `json:"coverImage"`
	TotalEpisodes     int    `json:"totalEpisodes"`
	MediaStatus       string `json:"mediaStatus"`
	NextAiringEpisode int    `json:"nextAiringEpisode"`
}

type ListEntrySave struct {
	MediaID         int
	Status          string
	Progress        int
	ScoreRaw        int
	Notes           string
	SendNotes       bool
	Repeat          int
	Private         bool
	SendPrivate     bool
	StartedAt       FuzzyDate
	SendStartedAt   bool
	CompletedAt     FuzzyDate
	SendCompletedAt bool
}

type MediaProgress struct {
	MediaID           int    `json:"mediaId"`
	Progress          int    `json:"progress"`
	TotalEpisodes     int    `json:"totalEpisodes"`
	MediaStatus       string `json:"mediaStatus"`
	NextAiringEpisode int    `json:"nextAiringEpisode"`
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

func NewWithHTTP(token string, httpClient *http.Client) *Client {
	client := New(token)
	if httpClient != nil {
		client.HTTP = httpClient
	}
	return client
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
