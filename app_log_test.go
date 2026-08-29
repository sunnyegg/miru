package main

import (
	"errors"
	"os"
	"path/filepath"
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

func TestRotateLogFileUnderThreshold(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Miru.log")
	if err := os.WriteFile(path, []byte("small"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateLogFile(path, 5*1024*1024); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("original file should remain: %v", err)
	}
	if _, err := os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Fatalf("archive should not exist, got err=%v", err)
	}
}

func TestRotateLogFileOverThreshold(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Miru.log")
	big := strings.Repeat("x", 6*1024*1024)
	if err := os.WriteFile(path, []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateLogFile(path, 5*1024*1024); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("original should be gone, got err=%v", err)
	}
	info, err := os.Stat(path + ".1")
	if err != nil {
		t.Fatalf("archive should exist: %v", err)
	}
	if info.Size() != int64(len(big)) {
		t.Fatalf("archive size = %d, want %d", info.Size(), len(big))
	}
}

func TestRotateLogFileReplacesArchive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Miru.log")
	oldArchive := filepath.Join(dir, "Miru.log.1")
	if err := os.WriteFile(oldArchive, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	big := strings.Repeat("y", 6*1024*1024)
	if err := os.WriteFile(path, []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rotateLogFile(path, 5*1024*1024); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(oldArchive)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != int64(len(big)) {
		t.Fatalf("archive size = %d, want %d", info.Size(), len(big))
	}
}

func TestRotateLogFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Miru.log")
	if err := rotateLogFile(path, 5*1024*1024); err != nil {
		t.Fatalf("missing file should be a no-op, got %v", err)
	}
}
