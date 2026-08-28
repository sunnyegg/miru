package update

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyReplacesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "miru")
	if err := os.WriteFile(dest, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "new-binary")
	}))
	t.Cleanup(server.Close)

	if err := Apply(context.Background(), server.Client(), server.URL, "miru-linux-amd64", dest); err != nil {
		t.Fatal(err)
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

func TestApplyReplacesAppBundle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bundle := filepath.Join(dir, "Miru.app")
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

	if err := Apply(context.Background(), server.Client(), server.URL, "miru-mac-universal.zip", exe); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(bundle, "Contents", "MacOS", "miru"))
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
		Name:   "miru-mac-universal.app/Contents/MacOS/miru",
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
