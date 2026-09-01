package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/sunnyegg/miru/internal/secrets"
	"github.com/sunnyegg/miru/internal/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const maxDebugLogRunes = 240

var (
	magnetPattern      = regexp.MustCompile(`(?i)magnet:\?[^\s]+`)
	urlPattern         = regexp.MustCompile(`(?i)(https?|file)://[^\s]+`)
	bearerPattern      = regexp.MustCompile(`(?i)\bbearer\s+[^\s,;]+`)
	accessTokenPattern = regexp.MustCompile(`(?i)\b(?:authorization|access[_-]?token|token)[=:\s]+[^\s,;]+`)
	unixPathPattern    = regexp.MustCompile(`(^|[\s"'=])(/[^/\s][^\s]*)`)
	winPathPattern     = regexp.MustCompile(`(?i)[a-z]:\\[^\s]+`)
)

func (a *App) logDebugErr(operation string, err error) {
	if err == nil || a == nil || a.ctx == nil {
		return
	}
	if skippedDebugErr(err) {
		return
	}
	runtime.LogDebugf(a.ctx, "%s: %s", operation, redactError(err))
}

func (a *App) logDebugf(format string, args ...any) {
	if a == nil || a.ctx == nil {
		return
	}
	runtime.LogDebug(a.ctx, truncateLogText(fmt.Sprintf(format, args...)))
}

func (a *App) logErr(operation string, err error) {
	if err == nil || a == nil || a.ctx == nil {
		return
	}
	if skippedDebugErr(err) {
		return
	}
	runtime.LogError(a.ctx, fmt.Sprintf("%s: %s", operation, redactError(err)))
}

func skippedDebugErr(err error) bool {
	return errors.Is(err, secrets.ErrNotFound) ||
		errors.Is(err, storage.ErrNotFound) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, http.ErrServerClosed)
}

func redactError(err error) string {
	if err == nil {
		return ""
	}
	return redactLogText(err.Error())
}

func redactLogText(text string) string {
	text = magnetPattern.ReplaceAllString(text, "[magnet]")
	text = urlPattern.ReplaceAllStringFunc(text, redactURL)
	text = bearerPattern.ReplaceAllString(text, "bearer [redacted]")
	text = accessTokenPattern.ReplaceAllString(text, "[redacted]")
	text = winPathPattern.ReplaceAllString(text, "[path]")
	text = unixPathPattern.ReplaceAllString(text, "${1}[path]")
	return truncateLogText(text)
}

func redactURL(raw string) string {
	schemeEnd := strings.Index(raw, "://")
	if schemeEnd < 0 {
		return "[url]"
	}
	rest := raw[schemeEnd+3:]
	pathStart := strings.IndexAny(rest, "/?#")
	host := rest
	if pathStart >= 0 {
		host = rest[:pathStart]
		rest = rest[pathStart:]
	} else {
		return raw[:schemeEnd+3] + host
	}
	queryStart := strings.IndexAny(rest, "?#")
	path := rest
	if queryStart >= 0 {
		path = rest[:queryStart]
		return raw[:schemeEnd+3] + host + path + "?[redacted]"
	}
	return raw[:schemeEnd+3] + host + path
}

func truncateLogText(text string) string {
	runes := []rune(text)
	if len(runes) <= maxDebugLogRunes {
		return text
	}
	return string(runes[:maxDebugLogRunes]) + "…"
}

const maxLogFileBytes = 5 * 1024 * 1024

func rotateLogFile(path string, maxBytes int64) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Size() <= maxBytes {
		return nil
	}
	archive := path + ".1"
	if err := os.Remove(archive); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(path, archive)
}
