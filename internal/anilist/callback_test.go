package anilist

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackExchangesCode(t *testing.T) {
	var got string
	srv := httptest.NewServer(NewMux(MuxConfig{
		ExchangeCode: func(code string) (string, error) {
			if code != "abc" {
				t.Fatalf("code = %q", code)
			}
			return "tok-1", nil
		},
		OnToken: func(token string) error {
			got = token
			return nil
		},
	}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/callback?code=abc")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	if got != "tok-1" {
		t.Fatalf("token = %q", got)
	}
}

func TestCallbackMissingCode(t *testing.T) {
	srv := httptest.NewServer(NewMux(MuxConfig{
		ExchangeCode: func(code string) (string, error) {
			t.Fatal("should not exchange")
			return "", nil
		},
	}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/callback")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestTokenPost(t *testing.T) {
	var got string
	srv := httptest.NewServer(NewMux(MuxConfig{
		OnToken: func(token string) error {
			got = token
			return nil
		},
	}))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/token", "application/json", strings.NewReader(`{"access_token":"abc"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	if got != "abc" {
		t.Fatalf("token = %q", got)
	}
}

func TestTokenPostEmpty(t *testing.T) {
	srv := httptest.NewServer(NewMux(MuxConfig{
		OnToken: func(token string) error {
			t.Fatal("should not save")
			return nil
		},
	}))
	defer srv.Close()

	res, err := http.Post(srv.URL+"/token", "application/json", strings.NewReader(`{"access_token":"  "}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestTokenPostSaveError(t *testing.T) {
	srv := httptest.NewServer(NewMux(MuxConfig{
		OnToken: func(token string) error {
			return errors.New("invalid token")
		},
	}))
	defer srv.Close()

	payload, _ := json.Marshal(map[string]string{"access_token": "bad"})
	res, err := http.Post(srv.URL+"/token", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "code" {
			t.Fatalf("form = %v", r.Form)
		}
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Fatalf("content-type = %s", ct)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"jwt-here"}`))
	}))
	defer server.Close()

	token, err := ExchangeCode(server.Client(), server.URL, "id", "secret", "code")
	if err != nil {
		t.Fatal(err)
	}
	if token != "jwt-here" {
		t.Fatalf("token = %q", token)
	}
}
