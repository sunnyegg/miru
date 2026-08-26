package networking

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const (
	ModeSystem = "system"
	ModeDirect = "direct"
	ModeSOCKS5 = "socks5"
)

type Config struct {
	Mode    string
	Address string
}

func (c Config) Normalized() (Config, error) {
	mode := strings.ToLower(strings.TrimSpace(c.Mode))
	if mode == "" {
		mode = ModeSystem
	}
	switch mode {
	case ModeSystem, ModeDirect:
		return Config{Mode: mode}, nil
	case ModeSOCKS5:
		address := strings.TrimSpace(c.Address)
		if address == "" {
			return Config{}, errors.New("SOCKS5 proxy address is empty")
		}
		if _, _, err := net.SplitHostPort(address); err != nil {
			return Config{}, fmt.Errorf("invalid SOCKS5 proxy address: %w", err)
		}
		return Config{Mode: mode, Address: address}, nil
	default:
		return Config{}, fmt.Errorf("unsupported network mode %q", c.Mode)
	}
}

func (c Config) HTTPClient() (*http.Client, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	switch normalized.Mode {
	case ModeDirect:
		transport.Proxy = nil
	case ModeSOCKS5:
		transport.Proxy = nil
		transport.DialContext = normalized.DialContext
	}
	return &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
	}, nil
}

func (c Config) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return nil, err
	}
	if normalized.Mode != ModeSOCKS5 {
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}
	dialer, err := proxy.SOCKS5("tcp", normalized.Address, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
		return contextDialer.DialContext(ctx, network, address)
	}
	return dialer.Dial(network, address)
}

func (c Config) SOCKS5Enabled() bool {
	return strings.EqualFold(strings.TrimSpace(c.Mode), ModeSOCKS5)
}

func (c Config) URL() (*url.URL, error) {
	normalized, err := c.Normalized()
	if err != nil {
		return nil, err
	}
	if normalized.Mode != ModeSOCKS5 {
		return nil, errors.New("SOCKS5 proxy is not enabled")
	}
	return &url.URL{Scheme: ModeSOCKS5, Host: normalized.Address}, nil
}
