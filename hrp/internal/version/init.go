package version

import (
	_ "embed"
)

//go:embed VERSION
var VERSION string

// httprunner python version
const HttpRunnerMinimumVersion = "v4.3.0"
