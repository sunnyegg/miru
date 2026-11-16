//go:build !production

package appinfo

// Development identity, used by wails dev, go run and plain go build. It keeps
// config, data, cache, logs, the single-instance id and the window title apart
// from an installed release.
const (
	Name        = "miru-dev"
	DisplayName = "Miru Dev"
	LogName     = "Miru-dev.log"
	BundleID    = "com.sunnyegg.miru-dev"
)
