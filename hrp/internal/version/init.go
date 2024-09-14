package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var VERSION string

func init() {
	VERSION = strings.TrimSpace(VERSION)
}
