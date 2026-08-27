package tokyotosho

import (
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultEndpoint = "https://www.tokyotosho.info/rss.php"
	CategoryAnime   = "1"
	maxFeedSize     = 4 << 20
)

var (
	magnetPattern = regexp.MustCompile(`magnet:[^\s"'<>]+`)
	sizePattern   = regexp.MustCompile(`(?i)Size:\s*([^<\n]+)`)
)

type Result struct {
	Title     string
	Link      string
	Magnet    string
	Published time.Time
	Size      string
	Seeders   int
	Leechers  int
	Downloads int
	Trusted   bool
	Remake    bool
}

type Client struct {
	HTTP     *http.Client
	Endpoint string
}

func New() *Client {
	return &Client{
		HTTP:     &http.Client{Timeout: 15 * time.Second},
		Endpoint: DefaultEndpoint,
	}
}

func NewWithHTTP(httpClient *http.Client) *Client {
	client := New()
	if httpClient != nil {
		client.HTTP = httpClient
	}
	return client
}

func (c *Client) Search(query string) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("tokyo toshokan search query is empty")
	}

	base := c.Endpoint
	if base == "" {
		base = DefaultEndpoint
	}
	endpoint, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse Tokyo Toshokan endpoint: %w", err)
	}
	values := endpoint.Query()
	values.Set("terms", query)
	values.Set("type", CategoryAnime)
	values.Set("searchName", "true")
	endpoint.RawQuery = values.Encode()

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rss+xml, application/xml")
	req.Header.Set("User-Agent", "Miru/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Tokyo Toshokan RSS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tokyo toshokan RSS http %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Tokyo Toshokan RSS: %w", err)
	}
	if len(raw) > maxFeedSize {
		return nil, errors.New("tokyo toshokan RSS response is too large")
	}

	var feed rssFeed
	if err := xml.Unmarshal(raw, &feed); err != nil {
		return nil, fmt.Errorf("parse Tokyo Toshokan RSS: %w", err)
	}
	results := make([]Result, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		result, ok, err := item.result()
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate"`
}

func (item rssItem) result() (Result, bool, error) {
	title := strings.TrimSpace(item.Title)
	link := strings.TrimSpace(item.Link)
	description := html.UnescapeString(item.Description)
	magnet := ""
	if match := magnetPattern.FindString(description); match != "" {
		magnet = strings.TrimRight(match, ".,);")
	}
	if title == "" || (link == "" && magnet == "") {
		return Result{}, false, nil
	}
	if link == "" {
		link = magnet
	}

	published, err := time.Parse(time.RFC1123, item.Published)
	if err != nil {
		published, err = time.Parse(time.RFC1123Z, item.Published)
	}
	if err != nil {
		return Result{}, false, fmt.Errorf("parse Tokyo Toshokan publication time %q: %w", item.Published, err)
	}

	size := ""
	if match := sizePattern.FindStringSubmatch(description); len(match) > 1 {
		size = strings.TrimSpace(match[1])
	}

	return Result{
		Title:     title,
		Link:      link,
		Magnet:    magnet,
		Published: published,
		Size:      size,
		Trusted:   strings.Contains(strings.ToLower(description), "authorized: yes"),
	}, true, nil
}
