package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAiringAvailableCachesHTTPStatus(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		available bool
	}{
		{name: "available", status: http.StatusOK, available: true},
		{name: "forbidden", status: http.StatusForbidden, available: false},
		{name: "server error", status: http.StatusInternalServerError, available: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				if test.status == http.StatusOK {
					_, _ = fmt.Fprint(w, `{"data":{"Page":{"pageInfo":{"hasNextPage":false},"airingSchedules":[]}}}`)
				}
			}))
			defer server.Close()

			a := newSettingsApp(t)
			client, err := a.newAnilist("")
			if err != nil {
				t.Fatal(err)
			}
			client.Endpoint = server.URL
			client.HTTP = server.Client()

			for range 2 {
				available, err := a.AiringAvailable()
				if err != nil {
					t.Fatal(err)
				}
				if available != test.available {
					t.Fatalf("AiringAvailable() = %t, want %t", available, test.available)
				}
			}
			if got := requests.Load(); got != 1 {
				t.Fatalf("requests = %d, want 1", got)
			}

			// The persisted cache avoids another request after memory is cleared.
			a.deleteAPIMemory(airingAvailabilityCacheKey)
			available, err := a.AiringAvailable()
			if err != nil {
				t.Fatal(err)
			}
			if available != test.available || requests.Load() != 1 {
				t.Fatalf("disk cache result = %t, requests = %d", available, requests.Load())
			}
		})
	}
}
