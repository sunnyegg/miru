package main

import "testing"

func TestShareImageURLAllowed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{
			name:   "anilist cdn",
			rawURL: "https://s4.anilist.co/file/anilistcdn/media/anime/cover/large/bx1.jpg",
		},
		{
			name:   "img anilist",
			rawURL: "https://img.anili.st/media/123",
		},
		{
			name:    "http rejected",
			rawURL:  "http://s4.anilist.co/file/x.jpg",
			wantErr: true,
		},
		{
			name:    "other host rejected",
			rawURL:  "https://example.com/cover.jpg",
			wantErr: true,
		},
		{
			name:    "anilist api host rejected",
			rawURL:  "https://graphql.anilist.co/",
			wantErr: true,
		},
		{
			name:    "empty rejected",
			rawURL:  "",
			wantErr: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := shareImageURLAllowed(testCase.rawURL)
			if testCase.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !testCase.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestShareBannerFilename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{input: "miru-watching.png", want: "miru-watching.png"},
		{input: "miru-watching", want: "miru-watching.png"},
		{input: "", want: "miru-share.png"},
		{input: "../escape", want: "escape.png"},
		{input: "/tmp/miru-completed.png", want: "miru-completed.png"},
	}

	for _, testCase := range cases {
		got := shareBannerFilename(testCase.input)
		if got != testCase.want {
			t.Fatalf("shareBannerFilename(%q) = %q, want %q", testCase.input, got, testCase.want)
		}
	}
}
