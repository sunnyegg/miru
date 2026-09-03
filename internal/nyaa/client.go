package nyaa

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultEndpoint = "https://nyaa.si/"
	MirrorEndpoint  = "https://nyaa.net/"
	CategoryEnglish = "1_2"
	maxFeedSize     = 4 << 20
)

// errEndpointUnreachable marks failures that mean the endpoint itself could not
// serve the search (network error, timeout, HTTP 5xx). Other failures — HTTP
// 4xx or an unparseable feed — are returned as-is and never trigger a mirror.
var errEndpointUnreachable = errors.New("nyaa endpoint unreachable")

type Result struct {
	Title     string
	Link      string
	InfoHash  string
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
	Mirror   string
}

func New() *Client {
	return &Client{
		HTTP:     &http.Client{Timeout: 15 * time.Second},
		Endpoint: DefaultEndpoint,
		Mirror:   MirrorEndpoint,
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
		return nil, errors.New("nyaa search query is empty")
	}

	base := c.Endpoint
	if base == "" {
		base = DefaultEndpoint
	}
	results, err := c.searchAt(query, base)
	if err == nil || !errors.Is(err, errEndpointUnreachable) {
		return results, err
	}
	mirror := c.Mirror
	if mirror == "" {
		mirror = MirrorEndpoint
	}
	if base == mirror {
		return results, err
	}
	return c.searchAt(query, mirror)
}

func (c *Client) searchAt(query, base string) ([]Result, error) {
	endpoint, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse Nyaa endpoint: %w", err)
	}
	values := endpoint.Query()
	values.Set("page", "rss")
	values.Set("q", query)
	values.Set("c", CategoryEnglish)
	values.Set("f", "0")
	values.Set("s", "id")
	values.Set("o", "desc")
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
		return nil, fmt.Errorf("request Nyaa RSS: %w: %w", err, errEndpointUnreachable)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("nyaa RSS http %d: %w", resp.StatusCode, errEndpointUnreachable)
		}
		return nil, fmt.Errorf("nyaa RSS http %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedSize+1))
	if err != nil {
		return nil, fmt.Errorf("read Nyaa RSS: %w", err)
	}
	if len(raw) > maxFeedSize {
		return nil, errors.New("nyaa RSS response is too large")
	}

	var feed rssFeed
	if err := xml.Unmarshal(raw, &feed); err != nil {
		return nil, fmt.Errorf("parse Nyaa RSS: %w", err)
	}
	results := make([]Result, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		result, err := item.result()
		if err != nil {
			return nil, err
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
	Title     string `xml:"title"`
	Link      string `xml:"link"`
	Published string `xml:"pubDate"`
	Fields    map[string]string
}

func (item *rssItem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	item.Fields = make(map[string]string)
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}
		switch value := token.(type) {
		case xml.StartElement:
			var text string
			if err := d.DecodeElement(&text, &value); err != nil {
				return err
			}
			switch value.Name.Local {
			case "title":
				item.Title = strings.TrimSpace(text)
			case "link":
				item.Link = strings.TrimSpace(text)
			case "pubDate":
				item.Published = strings.TrimSpace(text)
			default:
				item.Fields[value.Name.Local] = strings.TrimSpace(text)
			}
		case xml.EndElement:
			if value.Name == start.Name {
				return nil
			}
		}
	}
}

func (item rssItem) result() (Result, error) {
	published, err := time.Parse(time.RFC1123Z, item.Published)
	if err != nil {
		published, err = time.Parse(time.RFC1123, item.Published)
	}
	if err != nil {
		return Result{}, fmt.Errorf("parse Nyaa publication time %q: %w", item.Published, err)
	}
	hash := strings.ToLower(strings.TrimSpace(item.Fields["infoHash"]))
	if hash == "" {
		return Result{}, errors.New("nyaa RSS item is missing info hash")
	}
	return Result{
		Title:     item.Title,
		Link:      item.Link,
		InfoHash:  hash,
		Published: published,
		Size:      item.Fields["size"],
		Seeders:   parseNumber(item.Fields["seeders"]),
		Leechers:  parseNumber(item.Fields["leechers"]),
		Downloads: parseNumber(item.Fields["downloads"]),
		Trusted:   strings.EqualFold(item.Fields["trusted"], "yes"),
		Remake:    strings.EqualFold(item.Fields["remake"], "yes"),
	}, nil
}

func (r Result) Magnet() string {
	values := url.Values{}
	values.Set("dn", r.Title)
	return "magnet:?xt=urn:btih:" + r.InfoHash + "&" + values.Encode()
}

func parseNumber(value string) int {
	number, _ := strconv.Atoi(strings.TrimSpace(value))
	return number
}
