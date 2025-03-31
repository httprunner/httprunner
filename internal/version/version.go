package version

import (
	"fmt"
	"strings"

	_ "embed"
)

//go:embed VERSION
var VERSION string

func init() {
	VERSION = strings.TrimSpace(VERSION)
}

// 版本信息，在编译时通过 -ldflags 注入
var (
	GitCommit = "unknown"
	GitBranch = "unknown"
	BuildTime = "unknown"
	GitAuthor = "unknown"
)

func GetVersionInfo() string {
	commitID := GitCommit
	if len(commitID) > 8 {
		commitID = commitID[:8]
	}
	return fmt.Sprintf("%s (branch=%s, commit=%s, author=%s, build=%s)",
		VERSION, GitBranch, commitID, GitAuthor, BuildTime)
}
