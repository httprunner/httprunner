package cmd

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/httprunner/httpboomer"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run path...",
	Short: "run API test",
	Long:  `run yaml/json testcase files`,
	Example: `  $ httpboomer run demo.json	# run specified json testcase file
  $ httpboomer run demo.yaml	# run specified yaml testcase file
  $ httpboomer run examples/	# run testcases in specified folder`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// f, _ := cmd.Flags().GetBool("gen-html-report")
		// fmt.Println(f)

		var paths []httpboomer.ITestCase
		for _, arg := range args {
			paths = append(paths, &httpboomer.TestCasePath{Path: arg})
		}
		return httpboomer.Run(&testing.T{}, paths...)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolP("gen-html-report", "r", false, "Generate HTML report")
}
