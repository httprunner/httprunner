package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/hrp/internal/har2case"
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
		var outputFiles []string
		for _, arg := range args {
			// must choose one
			if !genYAMLFlag && !genJSONFlag {
				return errors.New("please select convert format type")
			}
			var outputPath string
			var err error

			har := har2case.NewHAR(arg)

			// specify output dir
			if outputDir != "" {
				har.SetOutputDir(outputDir)
			}
			// generate json/yaml files
			if genYAMLFlag {
				outputPath, err = har.GenYAML()
			} else {
				outputPath, err = har.GenJSON() // default
			}
			if err != nil {
				return err
			}
			outputFiles = append(outputFiles, outputPath)
		}
		log.Info().Strs("output", outputFiles).Msg("convert testcase success")
		return nil
	},
}

var (
	genJSONFlag bool
	genYAMLFlag bool
	outputDir   string
)

func init() {
	rootCmd.AddCommand(har2caseCmd)
	har2caseCmd.Flags().BoolVarP(&genJSONFlag, "to-json", "j", true, "convert to JSON format")
	har2caseCmd.Flags().BoolVarP(&genYAMLFlag, "to-yaml", "y", false, "convert to YAML format")
	har2caseCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "specify output directory, default to the same dir with har file")
}
