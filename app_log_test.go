package main

import (
	"errors"
	"strings"
	"testing"
)

func TestRedactLogText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
		deny []string
	}{
		{
			name: "magnet",
			in:   "add magnet:?xt=urn:btih:abc123&dn=show failed",
			want: []string{"[magnet]"},
			deny: []string{"btih", "abc123"},
		},
		{
			name: "url query",
			in:   "request https://nyaa.si/?page=rss&q=secret failed",
			want: []string{"https://nyaa.si/", "?[redacted]"},
			deny: []string{"secret", "page=rss"},
		},
		{
			name: "token",
			in:   "authorization: Bearer super-secret access_token=abc",
			want: []string{"[redacted]"},
			deny: []string{"super-secret", "abc"},
		},
		{
			name: "unix path",
			in:   "open /home/adila/.config/miru/app_data.db: permission denied",
			want: []string{"[path]"},
			deny: []string{"/home/adila"},
		},
		{
			name: "windows path",
			in:   `rename C:\Users\adila\AppData\Local\miru\miru.exe failed`,
			want: []string{"[path]"},
			deny: []string{`C:\Users\adila`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redactLogText(tc.in)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("got %q, want substring %q", got, want)
				}
			}
			for _, deny := range tc.deny {
				if strings.Contains(got, deny) {
					t.Fatalf("got %q, still contains %q", got, deny)
				}
			}
		})
	}
}

func TestRedactErrorTruncates(t *testing.T) {
	long := errors.New(strings.Repeat("x", maxDebugLogRunes+40))
	got := redactError(long)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("got %q", got)
	}
	if len([]rune(got)) != maxDebugLogRunes+1 {
		t.Fatalf("len = %d", len([]rune(got)))
	}
}
