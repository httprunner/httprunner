package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/convert/postman2case"
)

// postman2caseCmd represents the postman2case command
var postman2caseCmd = &cobra.Command{
	Use:   "postman2case $postman_path...",
	Short: "convert postman collection to json/yaml testcase files",
	Long:  `convert postman collection to json/yaml testcase files`,
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var outputFiles []string
		for _, arg := range args {
			// must choose one
			if !postman2caseGenJSONFlag && !postman2caseGenYAMLFlag {
				return errors.New("please select convert format type")
			}
			var outputPath string
			var err error

			collection := postman2case.NewCollection(arg)

			// specify output dir
			if postman2caseOutputDir != "" {
				collection.SetOutputDir(postman2caseOutputDir)
			}

			// specify profile path
			if postman2caseProfilePath != "" {
				collection.SetProfile(postman2caseProfilePath)
			}

			// specify patch path
			if postman2casePatchPath != "" {
				collection.SetPatch(postman2casePatchPath)
			}

			// generate json/yaml files
			if postman2caseGenYAMLFlag {
				outputPath, err = collection.GenYAML()
			} else {
				outputPath, err = collection.GenJSON() // default
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
	postman2caseGenJSONFlag bool
	postman2caseGenYAMLFlag bool
	postman2caseOutputDir   string
	postman2caseProfilePath string
	postman2casePatchPath   string
)

func init() {
	rootCmd.AddCommand(postman2caseCmd)
	postman2caseCmd.Flags().BoolVarP(&postman2caseGenJSONFlag, "to-json", "j", true, "convert to JSON format")
	postman2caseCmd.Flags().BoolVarP(&postman2caseGenYAMLFlag, "to-yaml", "y", false, "convert to YAML format")
	postman2caseCmd.Flags().StringVarP(&postman2caseOutputDir, "output-dir", "d", "", "specify output directory, default to the same dir with postman collection file")
	postman2caseCmd.Flags().StringVarP(&postman2caseProfilePath, "profile", "p", "", "specify profile path to override original headers (except for Content-Type) and cookies")
	postman2caseCmd.Flags().StringVarP(&postman2casePatchPath, "patch", "r", "", "specify patch path to create or update headers and cookies")
}
