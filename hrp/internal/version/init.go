package version

import (
	_ "embed"
)

//go:embed VERSION
var VERSION string

const HttpRunnerMinVersion = "v4.0.0-beta"
