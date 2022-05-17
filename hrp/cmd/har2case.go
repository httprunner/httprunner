package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/convert/har2case"
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
			if !har2caseGenYAMLFlag && !har2caseGenJSONFlag {
				return errors.New("please select convert format type")
			}
			var outputPath string
			var err error

			har := har2case.NewHAR(arg)

			// specify output dir
			if har2caseOutputDir != "" {
				har.SetOutputDir(har2caseOutputDir)
			}

			// specify profile
			if har2caseProfilePath != "" {
				har.SetProfile(har2caseProfilePath)
			}

			// specify profile
			if har2casePatchPath != "" {
				har.SetPatch(har2casePatchPath)
			}

			// generate json/yaml files
			if har2caseGenYAMLFlag {
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
	har2caseGenJSONFlag bool
	har2caseGenYAMLFlag bool
	har2caseOutputDir   string
	har2caseProfilePath string
	har2casePatchPath   string
)

func init() {
	rootCmd.AddCommand(har2caseCmd)
	har2caseCmd.Flags().BoolVarP(&har2caseGenJSONFlag, "to-json", "j", true, "convert to JSON format")
	har2caseCmd.Flags().BoolVarP(&har2caseGenYAMLFlag, "to-yaml", "y", false, "convert to YAML format")
	har2caseCmd.Flags().StringVarP(&har2caseOutputDir, "output-dir", "d", "", "specify output directory, default to the same dir with har file")
	har2caseCmd.Flags().StringVarP(&har2caseProfilePath, "profile", "p", "", "specify profile path to override headers and cookies")
	har2caseCmd.Flags().StringVarP(&har2casePatchPath, "patch", "r", "", "specify the path of the file used to replace headers and cookies")
}
