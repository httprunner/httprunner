package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/httprunner/hrp/har2case"
)

// har2caseCmd represents the har2case command
var har2caseCmd = &cobra.Command{
	Use:   "har2case path...",
	Short: "Convert HAR to json/yaml testcase files",
	Long:  `Convert HAR to json/yaml testcase files`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var outputFiles []string
		for _, arg := range args {
			var outputPath string
			var err error
			if genYAMLFlag {
				outputPath, err = har2case.NewHAR(arg).GenYAML()
			} else {
				outputPath, err = har2case.NewHAR(arg).GenJSON()
			}
			if err != nil {
				return err
			}
			outputFiles = append(outputFiles, outputPath)
		}
		log.Printf("output: %v", outputFiles)
		return nil
	},
}

var (
	genJSONFlag bool
	genYAMLFlag bool
)

func init() {
	RootCmd.AddCommand(har2caseCmd)
	har2caseCmd.Flags().BoolVarP(&genJSONFlag, "to-json", "j", false, "convert to JSON format (default)")
	har2caseCmd.Flags().BoolVarP(&genYAMLFlag, "to-yaml", "y", false, "convert to JSON format")
}
