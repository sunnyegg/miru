//go:build production

package appinfo

// Release identity. Wails adds the production tag for `wails build`, so dev
// runs (wails dev, go run, plain go build) use the -dev identity instead and
// never share state with an installed release.
const (
	Name        = "miru"
	DisplayName = "Miru"
	LogName     = "Miru.log"
	BundleID    = "com.sunnyegg.miru"
)
