package rssfeed

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const maxFeedSize = 4 << 20

var (
	magnetPattern = regexp.MustCompile(`magnet:[^\s"'<>]+`)
	sizePattern   = regexp.MustCompile(`(?i)Size:\s*([^<\n]+?)(?:\s+magnet:|$)`)
)

type Item struct {
	Key       string
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
	HTTP *http.Client
}

func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func NewWithHTTP(httpClient *http.Client) *Client {
	client := New()
	if httpClient != nil {
		client.HTTP = httpClient
	}
	return client
}

func (c *Client) Fetch(feedURL string) ([]Item, error) {
	feedURL = strings.TrimSpace(feedURL)
	if feedURL == "" {
		return nil, errors.New("rss feed url is empty")
	}
	parsedURL, err := url.Parse(feedURL)
	if err != nil {
		return nil, fmt.Errorf("parse rss feed url: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, errors.New("rss feed url must use http or https")
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	request, err := http.NewRequest(http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")
	request.Header.Set("User-Agent", "Miru/1.0")
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request rss feed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("rss feed http %d", response.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(response.Body, maxFeedSize+1))
	if err != nil {
		return nil, fmt.Errorf("read rss feed: %w", err)
	}
	if len(raw) > maxFeedSize {
		return nil, errors.New("rss feed response is too large")
	}

	var feed rssFeed
	if err := xml.Unmarshal(raw, &feed); err != nil {
		return nil, fmt.Errorf("parse rss feed: %w", err)
	}

	items := make([]Item, 0, len(feed.Channel.Items))
	for _, rawItem := range feed.Channel.Items {
		item, ok, err := rawItem.item()
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

type rssFeed struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string
	Link        string
	GUID        string
	Description string
	Published   string
	Fields      map[string]string
}

func (item *rssItem) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	item.Fields = make(map[string]string)
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch value := token.(type) {
		case xml.StartElement:
			var text string
			if err := decoder.DecodeElement(&text, &value); err != nil {
				return err
			}
			text = strings.TrimSpace(text)
			switch value.Name.Local {
			case "title":
				item.Title = text
			case "link":
				item.Link = text
			case "guid":
				item.GUID = text
			case "description":
				item.Description = text
			case "pubDate":
				item.Published = text
			default:
				item.Fields[value.Name.Local] = text
			}
		case xml.EndElement:
			if value.Name == start.Name {
				return nil
			}
		}
	}
}

func (item rssItem) item() (Item, bool, error) {
	title := strings.TrimSpace(item.Title)
	link := strings.TrimSpace(item.Link)
	description := html.UnescapeString(item.Description)

	magnet := ""
	if hash := strings.ToLower(strings.TrimSpace(item.Fields["infoHash"])); hash != "" {
		values := url.Values{}
		values.Set("dn", title)
		magnet = "magnet:?xt=urn:btih:" + hash + "&" + values.Encode()
	}
	if magnet == "" {
		if match := magnetPattern.FindString(description); match != "" {
			magnet = strings.TrimRight(match, ".,);")
		}
	}
	if title == "" || (link == "" && magnet == "") {
		return Item{}, false, nil
	}
	if link == "" {
		link = magnet
	}

	published, err := time.Parse(time.RFC1123Z, item.Published)
	if err != nil {
		published, err = time.Parse(time.RFC1123, item.Published)
	}
	if err != nil {
		return Item{}, false, fmt.Errorf("parse rss publication time %q: %w", item.Published, err)
	}

	size := strings.TrimSpace(item.Fields["size"])
	if size == "" {
		if match := sizePattern.FindStringSubmatch(description); len(match) > 1 {
			size = strings.TrimSpace(match[1])
		}
	}

	key := strings.TrimSpace(item.GUID)
	if key == "" {
		key = link
	}
	if key == "" {
		key = magnet
	}
	if key == "" {
		digest := sha256.Sum256([]byte(title + "|" + published.UTC().Format(time.RFC3339Nano)))
		key = hex.EncodeToString(digest[:])
	}

	return Item{
		Key:       key,
		Title:     title,
		Link:      link,
		Magnet:    magnet,
		Published: published,
		Size:      size,
		Seeders:   parseNumber(item.Fields["seeders"]),
		Leechers:  parseNumber(item.Fields["leechers"]),
		Downloads: parseNumber(item.Fields["downloads"]),
		Trusted: strings.EqualFold(item.Fields["trusted"], "yes") ||
			strings.Contains(strings.ToLower(description), "authorized: yes"),
		Remake: strings.EqualFold(item.Fields["remake"], "yes"),
	}, true, nil
}

func parseNumber(value string) int {
	number, _ := strconv.Atoi(strings.TrimSpace(value))
	return number
}
