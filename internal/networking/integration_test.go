//go:build integration

package networking

import (
	"io"
	"net/http"
	"os"
	"testing"
)

func TestSOCKS5NyaaIntegration(t *testing.T) {
	address := os.Getenv("MIRU_SOCKS5_PROXY")
	if address == "" {
		t.Skip("set MIRU_SOCKS5_PROXY to run the live SOCKS5 integration test")
	}

	client, err := (Config{
		Mode:    ModeSOCKS5,
		Address: address,
	}).HTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Get("https://nyaa.si/?page=rss&q=one+piece&c=1_2&f=0&s=id&o=desc")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Nyaa status = %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("Nyaa returned an empty response")
	}
}
