package discordrpc

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hugolgst/rich-go/client"
)

type Client struct {
	mu        sync.Mutex
	appID     string
	connected bool
	startedAt time.Time
}

func New() *Client {
	return &Client{}
}

func (presence *Client) Connect(appID string) error {
	presence.mu.Lock()
	defer presence.mu.Unlock()

	appID = strings.TrimSpace(appID)
	if appID == "" {
		return fmt.Errorf("discord app id is empty")
	}
	if presence.connected && presence.appID == appID {
		return nil
	}
	if presence.connected {
		client.Logout()
		presence.connected = false
	}
	if err := client.Login(appID); err != nil {
		return err
	}
	presence.appID = appID
	presence.connected = true
	return nil
}

func (presence *Client) Disconnect() {
	presence.mu.Lock()
	defer presence.mu.Unlock()
	if !presence.connected {
		return
	}
	client.Logout()
	presence.connected = false
	presence.appID = ""
}

func (presence *Client) SetWatching(animeTitle string, episodeNumber int, percent float64) error {
	presence.mu.Lock()
	defer presence.mu.Unlock()
	if !presence.connected {
		return fmt.Errorf("discord rpc is not connected")
	}

	animeTitle = strings.TrimSpace(animeTitle)
	if animeTitle == "" {
		animeTitle = "Miru"
	}
	if presence.startedAt.IsZero() {
		presence.startedAt = time.Now()
	}
	startedAt := presence.startedAt

	return client.SetActivity(client.Activity{
		Details: animeTitle,
		State:   playbackState(episodeNumber, percent),
		Timestamps: &client.Timestamps{
			Start: &startedAt,
		},
	})
}

func (presence *Client) Clear() {
	presence.mu.Lock()
	defer presence.mu.Unlock()
	if !presence.connected {
		return
	}
	client.Logout()
	presence.connected = false
	presence.appID = ""
	presence.startedAt = time.Time{}
}

func playbackState(episodeNumber int, percent float64) string {
	if episodeNumber > 0 {
		return fmt.Sprintf("Episode %d · %.0f%%", episodeNumber, percent)
	}
	return fmt.Sprintf("Watching · %.0f%%", percent)
}
