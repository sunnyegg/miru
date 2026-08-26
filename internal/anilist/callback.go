package anilist

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
)

type MuxConfig struct {
	OnToken      func(token string) error
	ExchangeCode func(code string) (string, error)
}

func NewMux(cfg MuxConfig) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		serveCallback(w, r, cfg)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		serveToken(w, r, cfg.OnToken)
	})
	return mux
}

func serveCallback(w http.ResponseWriter, r *http.Request, cfg MuxConfig) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		writeHTML(w, http.StatusBadRequest, "Login did not return a code. Try Open login again.")
		return
	}
	if cfg.ExchangeCode == nil {
		writeHTML(w, http.StatusInternalServerError, "Login server is not ready.")
		return
	}
	token, err := cfg.ExchangeCode(code)
	if err != nil {
		writeHTML(w, http.StatusBadRequest, "Could not exchange the login code: "+err.Error())
		return
	}
	if cfg.OnToken == nil {
		writeHTML(w, http.StatusInternalServerError, "Login server is not ready.")
		return
	}
	if err := cfg.OnToken(token); err != nil {
		writeHTML(w, http.StatusBadRequest, "Could not save the token: "+err.Error())
		return
	}
	writeHTML(w, http.StatusOK, "Logged in. You can close this tab.")
}

func serveToken(w http.ResponseWriter, r *http.Request, onToken func(token string) error) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	token := strings.TrimSpace(body.AccessToken)
	if token == "" {
		http.Error(w, "empty token", http.StatusBadRequest)
		return
	}
	if onToken == nil {
		http.Error(w, "not ready", http.StatusInternalServerError)
		return
	}
	if err := onToken(token); err != nil {
		http.Error(w, fmt.Sprintf("save token: %s", err.Error()), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func writeHTML(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Miru AniList login</title></head>
<body><p>%s</p></body>
</html>
`, html.EscapeString(message))
}
