package cmd

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
	"github.com/httprunner/httprunner/v4/hrp/pkg/convert"
)

var convertCmd = &cobra.Command{
	Use:   "convert $path...",
	Short: "convert to JSON/YAML/gotest/pytest testcases",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: convertRun,
}

var (
	toJSONFlag   bool
	toYAMLFlag   bool
	toGoTestFlag bool
	toPyTestFlag bool
	outputDir    string
	profilePath  string

	outputType convert.OutputType
)

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().BoolVar(&toPyTestFlag, "to-pytest", false, "convert to pytest scripts")
	convertCmd.Flags().BoolVar(&toGoTestFlag, "to-gotest", false, "convert to gotest scripts (TODO)")
	convertCmd.Flags().BoolVar(&toJSONFlag, "to-json", false, "convert to JSON scripts (default)")
	convertCmd.Flags().BoolVar(&toYAMLFlag, "to-yaml", false, "convert to YAML scripts")
	convertCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "specify output directory, default to the same dir with har file")
	convertCmd.Flags().StringVarP(&profilePath, "profile", "p", "", "specify profile path to override headers and cookies")
}

func convertRun(cmd *cobra.Command, args []string) error {
	var flagCount int
	if toJSONFlag {
		flagCount++
	}
	if toYAMLFlag {
		flagCount++
		outputType = convert.OutputTypeYAML
	}
	if toGoTestFlag {
		flagCount++
		outputType = convert.OutputTypeGoTest
	}
	if toPyTestFlag {
		flagCount++
		outputType = convert.OutputTypePyTest

		packages := []string{
			fmt.Sprintf("httprunner==%s", version.HttpRunnerMinimumVersion),
		}
		_, err := myexec.EnsurePython3Venv(venv, packages...)
		if err != nil {
			log.Error().Err(err).Msg("python3 venv is not ready")
			return err
		}
	}
	if flagCount > 1 {
		return errors.New("please specify at most one conversion flag")
	}
	convert.Run(outputType, outputDir, profilePath, args)
	return nil
}
