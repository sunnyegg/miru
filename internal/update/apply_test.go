package update

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyWithProgressReportsFinalBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "miru-linux-amd64")
	if err := os.WriteFile(dest, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	payload := bytes.Repeat([]byte("x"), 64*1024)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "65536")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(server.Close)

	var calls int
	var lastDownloaded, lastTotal int64
	progress := func(downloaded, total int64) {
		calls++
		lastDownloaded = downloaded
		lastTotal = total
	}

	installed, err := ApplyWithProgress(
		context.Background(),
		server.Client(),
		server.URL,
		"miru-linux-amd64",
		dest,
		progress,
	)
	if err != nil {
		t.Fatal(err)
	}
	if installed != dest {
		t.Fatalf("installed %q, want %q", installed, dest)
	}
	if calls == 0 {
		t.Fatal("progress callback was not invoked")
	}
	if lastTotal != int64(len(payload)) {
		t.Fatalf("last total = %d, want %d", lastTotal, len(payload))
	}
	if lastDownloaded != lastTotal {
		t.Fatalf("last downloaded = %d, want %d", lastDownloaded, lastTotal)
	}
	if lastDownloaded > lastTotal {
		t.Fatalf("downloaded %d exceeds total %d", lastDownloaded, lastTotal)
	}
}

func TestApplyReplacesFileInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "miru-linux-amd64")
	if err := os.WriteFile(dest, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "new-binary")
	}))
	t.Cleanup(server.Close)

	installed, err := Apply(context.Background(), server.Client(), server.URL, "miru-linux-amd64", dest)
	if err != nil {
		t.Fatal(err)
	}
	if installed != dest {
		t.Fatalf("installed %q, want %q", installed, dest)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(dest + ".old"); err != nil {
		t.Fatalf("old binary missing: %v", err)
	}
}

func TestApplyRenamesVersionedBinary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	current := filepath.Join(dir, "miru-0.1.0-linux-amd64")
	target := filepath.Join(dir, "miru-0.2.0-linux-amd64")
	if err := os.WriteFile(current, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "new-binary")
	}))
	t.Cleanup(server.Close)

	installed, err := Apply(context.Background(), server.Client(), server.URL, "miru-0.2.0-linux-amd64", current)
	if err != nil {
		t.Fatal(err)
	}
	if installed != target {
		t.Fatalf("installed %q, want %q", installed, target)
	}
	if _, err := os.Stat(current); !os.IsNotExist(err) {
		t.Fatalf("old binary still exists: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyRenamesCustomBinaryName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	current := filepath.Join(dir, "miru")
	target := filepath.Join(dir, "miru-0.2.0-linux-amd64")
	if err := os.WriteFile(current, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "new-binary")
	}))
	t.Cleanup(server.Close)

	installed, err := Apply(context.Background(), server.Client(), server.URL, "miru-0.2.0-linux-amd64", current)
	if err != nil {
		t.Fatal(err)
	}
	if installed != target {
		t.Fatalf("installed %q, want %q", installed, target)
	}
	if _, err := os.Stat(current); !os.IsNotExist(err) {
		t.Fatalf("old binary still exists: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyReplacesAppBundleInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bundle := filepath.Join(dir, "miru-mac-universal.app")
	macOSDir := filepath.Join(bundle, "Contents", "MacOS")
	if err := os.MkdirAll(macOSDir, 0755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(macOSDir, "miru")
	if err := os.WriteFile(exe, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(dir, "update.zip")
	if err := writeAppZip(zipPath, "new-binary"); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(server.Close)

	installed, err := Apply(context.Background(), server.Client(), server.URL, "miru-mac-universal.zip", exe)
	if err != nil {
		t.Fatal(err)
	}
	if installed != exe {
		t.Fatalf("installed %q, want %q", installed, exe)
	}
	got, err := os.ReadFile(filepath.Join(bundle, "Contents", "MacOS", "miru"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyRenamesAppBundle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	currentBundle := filepath.Join(dir, "miru.app")
	targetBundle := filepath.Join(dir, "miru-0.2.0-mac-universal.app")
	macOSDir := filepath.Join(currentBundle, "Contents", "MacOS")
	if err := os.MkdirAll(macOSDir, 0755); err != nil {
		t.Fatal(err)
	}
	exe := filepath.Join(macOSDir, "miru")
	if err := os.WriteFile(exe, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(dir, "update.zip")
	if err := writeAppZip(zipPath, "new-binary"); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(server.Close)

	installed, err := Apply(context.Background(), server.Client(), server.URL, "miru-0.2.0-mac-universal.zip", exe)
	if err != nil {
		t.Fatal(err)
	}
	targetExe := filepath.Join(targetBundle, "Contents", "MacOS", "miru")
	if installed != targetExe {
		t.Fatalf("installed %q, want %q", installed, targetExe)
	}
	if _, err := os.Stat(currentBundle); !os.IsNotExist(err) {
		t.Fatalf("old bundle still exists: %v", err)
	}
	got, err := os.ReadFile(targetExe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-binary" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanupOldRemovesSidecar(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	exe := filepath.Join(dir, "miru")
	if err := os.WriteFile(exe, []byte("current"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe+".old", []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	CleanupOld(exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Fatalf("expected .old gone, got %v", err)
	}
}

func writeAppZip(path, contents string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{
		Name:   "miru.app/Contents/MacOS/miru",
		Method: zip.Deflate,
	}
	header.SetMode(0755)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		_ = writer.Close()
		return err
	}
	if _, err := io.WriteString(entry, contents); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return file.Close()
}
