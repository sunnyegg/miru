package mpv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAnime4KShaderPathsMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := Anime4KShaderPaths(dir)
	if err == nil {
		t.Fatal("expected error for missing shaders")
	}
}

func TestAnime4KInstalled(t *testing.T) {
	dir := t.TempDir()
	if Anime4KInstalled(dir) {
		t.Fatal("expected shaders to be missing")
	}
	writeTestShaders(t, dir)
	if !Anime4KInstalled(dir) {
		t.Fatal("expected shaders to be installed")
	}
}

func TestEnsureAnime4KShadersDownloads(t *testing.T) {
	dir := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("#version 330\nvoid main(){}\n"))
	}))
	defer server.Close()

	original := anime4KModeA
	t.Cleanup(func() { anime4KModeA = original })
	anime4KModeA = []anime4KShader{
		{remotePath: "test.glsl", fileName: "Anime4K_Clamp_Highlights.glsl"},
	}

	client := server.Client()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})
	client.Transport = transport

	if err := EnsureAnime4KShaders(context.Background(), client, dir); err != nil {
		t.Fatalf("EnsureAnime4KShaders: %v", err)
	}
	paths, err := Anime4KShaderPaths(dir)
	if err != nil {
		t.Fatalf("Anime4KShaderPaths: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths = %d, want 1", len(paths))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func writeTestShaders(t *testing.T, configDir string) {
	t.Helper()
	shaderDir := ShadersDir(configDir)
	if err := os.MkdirAll(shaderDir, 0o700); err != nil {
		t.Fatalf("mkdir shaders: %v", err)
	}
	for _, shader := range anime4KModeA {
		path := filepath.Join(shaderDir, shader.fileName)
		if err := os.WriteFile(path, []byte("#version 330\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", shader.fileName, err)
		}
	}
}
