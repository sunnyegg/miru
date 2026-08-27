package anilist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
