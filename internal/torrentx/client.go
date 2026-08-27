package torrentx

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/sunnyegg/miru/internal/networking"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/time/rate"
)

const rateLimiterBurst = 16 * 1024

func (m *Manager) ensureClient(dataDir string, limits RateLimits, networkConfig networking.Config) (*torrent.Client, error) {
	normalizedNetwork, err := networkConfig.Normalized()
	if err != nil {
		return nil, err
	}
	networkKey := normalizedNetwork.Mode + ":" + normalizedNetwork.Address
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil && m.dataDir == dataDir && m.networkKey == networkKey {
		m.applyRateLimitsLocked(limits)
		return m.client, nil
	}
	if m.client != nil {
		m.client.Close()
		m.client = nil
	}
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = dataDir
	cfg.ListenPort = 0
	cfg.Seed = true
	cfg.NoUpload = false
	uploadRate, downloadRate := newRateLimiters(limits)
	cfg.UploadRateLimiter = uploadRate
	cfg.DownloadRateLimiter = downloadRate
	if normalizedNetwork.Mode == networking.ModeSystem {
		cfg.HTTPProxy = http.ProxyFromEnvironment
	}
	if normalizedNetwork.Mode == networking.ModeSOCKS5 {
		cfg.HTTPDialContext = normalizedNetwork.DialContext
		cfg.TrackerDialContext = normalizedNetwork.DialContext
		cfg.DialForPeerConns = false
		cfg.DisableUTP = true
		cfg.NoDHT = true
		cfg.NoDefaultPortForwarding = true
		cfg.DisablePEX = true
		cfg.DisableWebtorrent = true
		cfg.DisableIPv6 = true
		cfg.AcceptPeerConnections = false
		cfg.TrackerListenPacket = func(network, _ string) (net.PacketConn, error) {
			return nil, errors.New("UDP trackers are disabled when using SOCKS5")
		}
	}
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if normalizedNetwork.Mode == networking.ModeSOCKS5 {
		client.AddDialer(torrent.NetworkDialer{
			Network: "tcp4",
			Dialer:  normalizedNetwork,
		})
	}
	m.client = client
	m.dataDir = dataDir
	m.networkKey = networkKey
	m.uploadRate = uploadRate
	m.downloadRate = downloadRate
	return client, nil
}

func newRateLimiters(limits RateLimits) (*rate.Limiter, *rate.Limiter) {
	return newRateLimiter(limits.Upload), newRateLimiter(limits.Download)
}

func newRateLimiter(bytesPerSecond int64) *rate.Limiter {
	if bytesPerSecond <= 0 {
		return rate.NewLimiter(rate.Inf, rateLimiterBurst)
	}
	return rate.NewLimiter(rate.Limit(bytesPerSecond), rateLimiterBurst)
}

func (m *Manager) ApplyRateLimits(limits RateLimits) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.applyRateLimitsLocked(limits)
}

func (m *Manager) applyRateLimitsLocked(limits RateLimits) {
	if m.uploadRate != nil {
		m.uploadRate.SetLimit(rateLimit(limits.Upload))
	}
	if m.downloadRate != nil {
		m.downloadRate.SetLimit(rateLimit(limits.Download))
	}
}

func rateLimit(bytesPerSecond int64) rate.Limit {
	if bytesPerSecond <= 0 {
		return rate.Inf
	}
	return rate.Limit(bytesPerSecond)
}

func addSource(client *torrent.Client, source string, httpClient *http.Client) (*torrent.Torrent, error) {
	if strings.HasPrefix(strings.ToLower(source), "magnet:") {
		return client.AddMagnet(source)
	}

	parsed, err := url.Parse(source)
	isHTTP := err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https")
	if !isHTTP {
		mi, err := metainfo.LoadFromFile(source)
		if err != nil {
			return nil, err
		}
		return client.AddTorrent(mi)
	}

	if httpClient == nil {
		return nil, errors.New("HTTP client is unavailable")
	}
	response, err := httpClient.Get(parsed.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New("torrent URL returned an HTTP error")
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	mi, err := metainfo.Load(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	return client.AddTorrent(mi)
}

func displaySource(source string) string {
	if strings.HasPrefix(strings.ToLower(source), "magnet:") {
		return "Magnet download"
	}
	return filepath.Base(source)
}

func ResolveDataPath(destDir, torrentPath string) string {
	if filepath.IsAbs(torrentPath) {
		return torrentPath
	}
	return filepath.Join(destDir, torrentPath)
}
