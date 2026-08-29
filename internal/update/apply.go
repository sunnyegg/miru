package update

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func CleanupOld(executable string) {
	exe := executable
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		exe = resolved
	}
	_ = os.Remove(exe + ".old")
	bundle, ok := appBundleRoot(exe)
	if !ok {
		return
	}
	_ = os.RemoveAll(bundle + ".old")
}

func Apply(ctx context.Context, client *http.Client, downloadURL, assetName, executable string) (string, error) {
	exe := executable
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		exe = resolved
	}
	if strings.HasSuffix(assetName, ".zip") {
		return applyAppBundle(ctx, client, downloadURL, assetName, exe)
	}
	return applyBinary(ctx, client, downloadURL, assetName, exe)
}

func binaryInstallTarget(assetName, executable string) string {
	return filepath.Join(filepath.Dir(executable), assetName)
}

func bundleInstallTarget(assetName, executable string) (string, error) {
	bundle, ok := appBundleRoot(executable)
	if !ok {
		return "", fmt.Errorf("cannot replace binary: run Miru from the .app bundle")
	}
	targetName := strings.TrimSuffix(assetName, ".zip") + ".app"
	return filepath.Join(filepath.Dir(bundle), targetName), nil
}

func applyBinary(ctx context.Context, client *http.Client, downloadURL, assetName, current string) (string, error) {
	target := binaryInstallTarget(assetName, current)
	newPath := target + ".new"
	if err := downloadFile(ctx, client, downloadURL, newPath); err != nil {
		_ = os.Remove(newPath)
		return "", err
	}
	mode := os.FileMode(0755)
	if info, err := os.Stat(current); err == nil {
		mode = info.Mode()
	} else if info, err := os.Stat(target); err == nil {
		mode = info.Mode()
	}
	if err := os.Chmod(newPath, mode); err != nil {
		_ = os.Remove(newPath)
		return "", fmt.Errorf("cannot replace binary: %w", err)
	}
	if err := installToTarget(current, target, newPath); err != nil {
		return "", err
	}
	return target, nil
}

func applyAppBundle(ctx context.Context, client *http.Client, downloadURL, assetName, executable string) (string, error) {
	currentBundle, ok := appBundleRoot(executable)
	if !ok {
		return "", fmt.Errorf("cannot replace binary: run Miru from the .app bundle")
	}
	targetBundle, err := bundleInstallTarget(assetName, executable)
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "miru-update-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, "update.zip")
	if err := downloadFile(ctx, client, downloadURL, zipPath); err != nil {
		return "", err
	}
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.Mkdir(extractDir, 0755); err != nil {
		return "", err
	}
	if err := extractZip(zipPath, extractDir); err != nil {
		return "", err
	}
	newBundle, err := findAppBundle(extractDir)
	if err != nil {
		return "", err
	}
	if err := installToTarget(currentBundle, targetBundle, newBundle); err != nil {
		return "", err
	}
	return filepath.Join(targetBundle, "Contents", "MacOS", filepath.Base(executable)), nil
}

func installToTarget(current, target, staged string) error {
	if filepath.Clean(current) == filepath.Clean(target) {
		return replacePath(target, staged)
	}

	stagedNew := target + ".new"
	if err := os.Rename(staged, stagedNew); err != nil {
		return fmt.Errorf("cannot replace binary: %w", err)
	}

	oldSidecar := target + ".old"
	_ = os.RemoveAll(oldSidecar)
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, oldSidecar); err != nil {
			_ = os.RemoveAll(stagedNew)
			return fmt.Errorf("cannot replace binary: %w", err)
		}
		_ = os.RemoveAll(oldSidecar)
	}
	if err := os.Rename(stagedNew, target); err != nil {
		_ = os.RemoveAll(stagedNew)
		return fmt.Errorf("cannot replace binary: %w", err)
	}
	if filepath.Clean(current) != filepath.Clean(target) {
		_ = os.RemoveAll(current)
	}
	return nil
}

func downloadFile(ctx context.Context, client *http.Client, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

func replacePath(dest, newPath string) error {
	oldPath := dest + ".old"
	_ = os.RemoveAll(oldPath)
	if _, err := os.Stat(dest); err == nil {
		if err := os.Rename(dest, oldPath); err != nil {
			return fmt.Errorf("cannot replace binary: %w", err)
		}
	}
	if err := os.Rename(newPath, dest); err != nil {
		_ = os.Rename(oldPath, dest)
		return fmt.Errorf("cannot replace binary: %w", err)
	}
	return nil
}

func appBundleRoot(executable string) (string, bool) {
	macOSDir := filepath.Dir(executable)
	if filepath.Base(macOSDir) != "MacOS" {
		return "", false
	}
	contents := filepath.Dir(macOSDir)
	if filepath.Base(contents) != "Contents" {
		return "", false
	}
	bundle := filepath.Dir(contents)
	if !strings.HasSuffix(bundle, ".app") {
		return "", false
	}
	return bundle, true
}

func extractZip(zipPath, dest string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	dest = filepath.Clean(dest)
	prefix := dest + string(os.PathSeparator)
	for _, file := range reader.File {
		target := filepath.Join(dest, file.Name)
		cleaned := filepath.Clean(target)
		if cleaned != dest && !strings.HasPrefix(cleaned, prefix) {
			return fmt.Errorf("invalid zip path %q", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleaned, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(cleaned), 0755); err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(cleaned, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			_ = src.Close()
			return err
		}
		_, copyErr := io.Copy(out, src)
		closeOutErr := out.Close()
		_ = src.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeOutErr != nil {
			return closeOutErr
		}
	}
	return nil
}

func findAppBundle(dir string) (string, error) {
	var found string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.HasSuffix(path, ".app") {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("zip does not contain a .app bundle")
	}
	return found, nil
}
