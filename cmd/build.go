package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/internal/sdk"
)

var CmdBuild = &cobra.Command{
	Use:   "build $path ...",
	Short: "Build plugin for testing",
	Long:  `Build python/go plugin for testing`,
	Example: `  $ hrp build plugin/debugtalk.go
  $ hrp build plugin/debugtalk.py`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		startTime := time.Now()
		defer func() {
			sdk.SendGA4Event("hrp_build", map[string]interface{}{
				"args":                 strings.Join(args, "-"),
				"success":              err == nil,
				"engagement_time_msec": time.Since(startTime).Milliseconds(),
			})
		}()
		return hrp.BuildPlugin(args[0], output)
	},
}

var output string

func init() {
	CmdBuild.Flags().StringVarP(&output, "output", "o", "", "funplugin product output path, default: cwd")
}
