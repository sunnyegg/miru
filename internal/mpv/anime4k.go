package mpv

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const anime4KVersion = "v4.0.1"

type anime4KShader struct {
	remotePath string
	fileName   string
}

// Mode A (HQ) from Anime4K v4 — restore and upscale chain for 1080p sources.
var anime4KModeA = []anime4KShader{
	{remotePath: "glsl/Restore/Anime4K_Clamp_Highlights.glsl", fileName: "Anime4K_Clamp_Highlights.glsl"},
	{remotePath: "glsl/Restore/Anime4K_Restore_CNN_VL.glsl", fileName: "Anime4K_Restore_CNN_VL.glsl"},
	{remotePath: "glsl/Upscale/Anime4K_Upscale_CNN_x2_VL.glsl", fileName: "Anime4K_Upscale_CNN_x2_VL.glsl"},
	{remotePath: "glsl/Upscale/Anime4K_AutoDownscalePre_x2.glsl", fileName: "Anime4K_AutoDownscalePre_x2.glsl"},
	{remotePath: "glsl/Upscale/Anime4K_AutoDownscalePre_x4.glsl", fileName: "Anime4K_AutoDownscalePre_x4.glsl"},
	{remotePath: "glsl/Upscale/Anime4K_Upscale_CNN_x2_M.glsl", fileName: "Anime4K_Upscale_CNN_x2_M.glsl"},
}

func ShadersDir(configDir string) string {
	return filepath.Join(configDir, "shaders")
}

func Anime4KInstalled(configDir string) bool {
	_, err := Anime4KShaderPaths(configDir)
	return err == nil
}

func Anime4KShaderPaths(configDir string) ([]string, error) {
	shaderDir := ShadersDir(configDir)
	paths := make([]string, 0, len(anime4KModeA))
	for _, shader := range anime4KModeA {
		path := filepath.Join(shaderDir, shader.fileName)
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("Anime4K shader missing (%s): download from Settings or check network", shader.fileName)
		}
		if info.Size() == 0 {
			return nil, fmt.Errorf("Anime4K shader empty (%s): re-save playback settings to re-download", shader.fileName)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func EnsureAnime4KShaders(ctx context.Context, client *http.Client, configDir string) error {
	if client == nil {
		client = http.DefaultClient
	}
	shaderDir := ShadersDir(configDir)
	if err := os.MkdirAll(shaderDir, 0o700); err != nil {
		return fmt.Errorf("create shader folder: %w", err)
	}
	for _, shader := range anime4KModeA {
		dest := filepath.Join(shaderDir, shader.fileName)
		if info, err := os.Stat(dest); err == nil && info.Size() > 0 {
			continue
		}
		url := fmt.Sprintf("https://raw.githubusercontent.com/bloc97/Anime4K/%s/%s", anime4KVersion, shader.remotePath)
		if err := downloadShader(ctx, client, url, dest); err != nil {
			return fmt.Errorf("download %s: %w", shader.fileName, err)
		}
	}
	if !Anime4KInstalled(configDir) {
		return fmt.Errorf("Anime4K shaders incomplete after download")
	}
	return nil
}

func downloadShader(ctx context.Context, client *http.Client, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "miru")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	tempPath := dest + ".part"
	out, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return closeErr
	}
	if err := os.Rename(tempPath, dest); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}
