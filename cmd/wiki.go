package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/internal/wiki"
)

var CmdWiki = &cobra.Command{
	Use:     "wiki",
	Aliases: []string{"info", "docs", "doc"},
	Short:   "visit https://httprunner.com",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_wiki", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()
		return wiki.OpenWiki()
	},
}
