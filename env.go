package main

import (
	"os"
	"path/filepath"
	"strings"
)

func loadDotEnv() {
	candidates := []string{".env"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	for _, path := range candidates {
		applyDotEnvFile(path)
	}
	applyDotEnv(embeddedEnv)
}

func applyDotEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	applyDotEnv(string(data))
}

func applyDotEnv(data string) {
	for key, value := range parseDotEnv(data) {
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func parseDotEnv(data string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		out[key] = value
	}
	return out
}

func envTrim(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
