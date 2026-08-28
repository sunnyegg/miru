package networking

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestConfigNormalizesModes(t *testing.T) {
	got, err := (Config{Mode: " SOCKS5 ", Address: "127.0.0.1:1080"}).Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != ModeSOCKS5 || got.Address != "127.0.0.1:1080" {
		t.Fatalf("config = %+v", got)
	}
}

func TestConfigRejectsInvalidSOCKS5(t *testing.T) {
	for _, config := range []Config{
		{Mode: ModeSOCKS5},
		{Mode: ModeSOCKS5, Address: "localhost"},
		{Mode: "unknown"},
	} {
		if _, err := config.Normalized(); err == nil {
			t.Fatalf("expected error for %+v", config)
		}
	}
}

func TestConfigNormalizesHTTPProxy(t *testing.T) {
	got, err := (Config{Mode: " HTTP_PROXY ", ProxyURL: " http://127.0.0.1:8080 "}).
		Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != ModeHTTPProxy || got.ProxyURL != "http://127.0.0.1:8080" {
		t.Fatalf("config = %+v", got)
	}
}

func TestConfigRejectsInvalidHTTPProxy(t *testing.T) {
	for _, config := range []Config{
		{Mode: ModeHTTPProxy},
		{Mode: ModeHTTPProxy, ProxyURL: "127.0.0.1:8080"},
		{Mode: ModeHTTPProxy, ProxyURL: "ftp://127.0.0.1:8080"},
		{Mode: ModeHTTPProxy, ProxyURL: "http://"},
	} {
		if _, err := config.Normalized(); err == nil {
			t.Fatalf("expected error for %+v", config)
		}
	}
}

func TestHTTPClientHTTPProxyMode(t *testing.T) {
	client, err := (Config{
		Mode:     ModeHTTPProxy,
		ProxyURL: "http://127.0.0.1:8080",
	}).HTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", client.Transport)
	}
	if transport.Proxy == nil {
		t.Fatal("http_proxy mode must set transport.Proxy")
	}
	request, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	proxyURL, err := transport.Proxy(request)
	if err != nil {
		t.Fatal(err)
	}
	if proxyURL.String() != "http://127.0.0.1:8080" {
		t.Fatalf("proxy URL = %q", proxyURL)
	}
}

func TestHTTPProxyClientRoutesRequestThroughProxy(t *testing.T) {
	var proxyHits atomic.Int32
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer origin.Close()

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyHits.Add(1)
		response, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
	}))
	defer proxyServer.Close()

	client, err := (Config{
		Mode:     ModeHTTPProxy,
		ProxyURL: "http://" + proxyServer.Listener.Addr().String(),
	}).HTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Get(origin.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if got := proxyHits.Load(); got != 1 {
		t.Fatalf("proxy hits = %d", got)
	}
}

func TestHTTPClientModes(t *testing.T) {
	for _, mode := range []string{ModeSystem, ModeDirect} {
		client, err := (Config{Mode: mode}).HTTPClient()
		if err != nil {
			t.Fatal(err)
		}
		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("transport type = %T", client.Transport)
		}
		if mode == ModeDirect && transport.Proxy != nil {
			t.Fatal("direct mode must not use a proxy")
		}
		if mode == ModeSystem && transport.Proxy == nil {
			t.Fatal("system mode must use environment proxy resolution")
		}
	}
}

func TestSOCKS5HTTPClientRoutesRequestThroughProxy(t *testing.T) {
	var proxyConnections atomic.Int32
	proxy := newTestSOCKS5Proxy(t, &proxyConnections)
	defer proxy.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := (Config{
		Mode:    ModeSOCKS5,
		Address: proxy.Addr().String(),
	}).HTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if got := proxyConnections.Load(); got != 1 {
		t.Fatalf("proxy connections = %d", got)
	}
}

func newTestSOCKS5Proxy(t *testing.T, connections *atomic.Int32) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			connections.Add(1)
			go serveTestSOCKS5Connection(connection)
		}
	}()
	return listener
}

func serveTestSOCKS5Connection(connection net.Conn) {
	defer connection.Close()
	reader := bufio.NewReader(connection)
	if _, err := io.ReadFull(reader, make([]byte, 2)); err != nil {
		return
	}
	methods := make([]byte, 1)
	if _, err := io.ReadFull(reader, methods); err != nil {
		return
	}
	if _, err := connection.Write([]byte{5, 0}); err != nil {
		return
	}

	requestHeader := make([]byte, 4)
	if _, err := io.ReadFull(reader, requestHeader); err != nil {
		return
	}
	address, err := readTestSOCKS5Address(reader, requestHeader[3])
	if err != nil {
		return
	}
	target, err := net.Dial("tcp", address)
	if err != nil {
		_, _ = connection.Write([]byte{5, 5, 0, 1, 0, 0, 0, 0, 0, 0})
		return
	}
	defer target.Close()
	if _, err := connection.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}
	go io.Copy(target, reader)
	_, _ = io.Copy(connection, target)
}

func readTestSOCKS5Address(reader *bufio.Reader, addressType byte) (string, error) {
	var host string
	switch addressType {
	case 1:
		raw := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(reader, raw); err != nil {
			return "", err
		}
		host = net.IP(raw).String()
	case 3:
		length, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		raw := make([]byte, length)
		if _, err := io.ReadFull(reader, raw); err != nil {
			return "", err
		}
		host = string(raw)
	case 4:
		raw := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(reader, raw); err != nil {
			return "", err
		}
		host = net.IP(raw).String()
	default:
		return "", io.ErrUnexpectedEOF
	}

	port := make([]byte, 2)
	if _, err := io.ReadFull(reader, port); err != nil {
		return "", err
	}
	return net.JoinHostPort(host, strconv.Itoa(int(binary.BigEndian.Uint16(port)))), nil
}
