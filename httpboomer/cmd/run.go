package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/httprunner/httpboomer"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run path...",
	Short: "run API test",
	Long:  `run yaml/json testcase files for API test`,
	Example: `  $ httpboomer run demo.json	# run specified json testcase file
  $ httpboomer run demo.yaml	# run specified yaml testcase file
  $ httpboomer run examples/	# run testcases in specified folder`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// --silent flag
		silentFlag, err := cmd.Flags().GetBool("silent")
		if err != nil {
			return err
		}
		log.Printf("[runCmd] set --silent flag: %v", silentFlag)

		var paths []httpboomer.ITestCase
		for _, arg := range args {
			paths = append(paths, &httpboomer.TestCasePath{Path: arg})
		}
		return httpboomer.NewRunner().SetDebug(!silentFlag).Run(paths...)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("silent", "s", false, "Disable logging request & response details")
	// runCmd.Flags().BoolP("gen-html-report", "r", false, "Generate HTML report")
}
