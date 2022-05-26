package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/convert"
)

// har2caseCmd represents the har2case command
var har2caseCmd = &cobra.Command{
	Use:   "har2case $har_path...",
	Short: "convert HAR to json/yaml testcase files",
	Long:  `convert HAR to json/yaml testcase files`,
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var flagCount int
		var har2caseOutputType convert.OutputType
		if har2caseGenJSONFlag {
			flagCount++
		}
		if har2caseGenYAMLFlag {
			flagCount++
			har2caseOutputType = convert.OutputTypeYAML
		}
		if flagCount > 1 {
			return errors.New("please specify at most one conversion flag")
		}
		convert.Run(har2caseOutputType, har2caseOutputDir, har2caseProfilePath, args)
		return nil
	},
}

var (
	har2caseGenJSONFlag bool
	har2caseGenYAMLFlag bool
	har2caseOutputDir   string
	har2caseProfilePath string
)

func init() {
	rootCmd.AddCommand(har2caseCmd)
	har2caseCmd.Flags().BoolVarP(&har2caseGenJSONFlag, "to-json", "j", false, "convert to JSON format (default)")
	har2caseCmd.Flags().BoolVarP(&har2caseGenYAMLFlag, "to-yaml", "y", false, "convert to YAML format")
	har2caseCmd.Flags().StringVarP(&har2caseOutputDir, "output-dir", "d", "", "specify output directory, default to the same dir with har file")
	har2caseCmd.Flags().StringVarP(&har2caseProfilePath, "profile", "p", "", "specify profile path to override headers and cookies")
}
