//go:build production

package main

import (
	_ "embed"
)

//go:embed .env
var embeddedEnv string
