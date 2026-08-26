package main

import "testing"

func TestParseDotEnv(t *testing.T) {
	got := parseDotEnv("# comment\nANILIST_CLIENT_ID=49496\nANILIST_CLIENT_SECRET=\"abc def\"\n\n=skip\n")
	if got["ANILIST_CLIENT_ID"] != "49496" {
		t.Fatalf("id = %q", got["ANILIST_CLIENT_ID"])
	}
	if got["ANILIST_CLIENT_SECRET"] != "abc def" {
		t.Fatalf("secret = %q", got["ANILIST_CLIENT_SECRET"])
	}
	if _, ok := got[""]; ok {
		t.Fatal("empty key should be skipped")
	}
}
