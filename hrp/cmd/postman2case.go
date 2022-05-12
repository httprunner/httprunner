package cmd

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/postman2case"
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
			if !postman2JSONFlag && !postman2YAMLFlag {
				return errors.New("please select convert format type")
			}
			var outputPath string
			var err error

			postman := postman2case.NewCollection(arg)

			// specify output dir
			if postman2Dir != "" {
				postman.SetOutputDir(postman2Dir)
			}

			// generate json/yaml files
			if genYAMLFlag {
				outputPath, err = postman.GenYAML()
			} else {
				outputPath, err = postman.GenJSON() // default
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
	postman2JSONFlag bool
	postman2YAMLFlag bool
	postman2Dir      string
)

func init() {
	rootCmd.AddCommand(postman2caseCmd)
	postman2caseCmd.Flags().BoolVarP(&postman2JSONFlag, "to-json", "j", true, "convert to JSON format")
	postman2caseCmd.Flags().BoolVarP(&postman2YAMLFlag, "to-yaml", "y", false, "convert to YAML format")
	postman2caseCmd.Flags().StringVarP(&postman2Dir, "output-dir", "d", "", "specify output directory, default to the same dir with postman collection file")
}
