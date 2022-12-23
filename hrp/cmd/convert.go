package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
	"github.com/httprunner/httprunner/v4/hrp/pkg/convert"
)

var convertCmd = &cobra.Command{
	Use:          "convert $path...",
	Short:        "convert multiple source format to HttpRunner JSON/YAML/gotest/pytest cases",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: false,
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		caseConverter := convert.NewConverter(outputDir, profilePath)

		var fromType convert.FromType
		if fromYAMLFlag {
			fromType = convert.FromTypeYAML
		} else if fromPostmanFlag {
			fromType = convert.FromTypePostman
		} else if fromHARFlag {
			fromType = convert.FromTypeHAR
		} else {
			fromType = convert.FromTypeJSON
			log.Info().Str("fromType", fromType.String()).Msg("set default")
		}

		var outputType convert.OutputType
		if toYAMLFlag {
			outputType = convert.OutputTypeYAML
		} else if toPyTestFlag {
			packages := []string{
				fmt.Sprintf("httprunner==%s", version.HttpRunnerMinimumVersion),
			}
			_, err := myexec.EnsurePython3Venv(venv, packages...)
			if err != nil {
				log.Error().Err(err).Msg("python3 venv is not ready")
				return err
			}

			outputType = convert.OutputTypePyTest
		} else {
			outputType = convert.OutputTypeJSON
			log.Info().Str("outputType", outputType.String()).Msg("set default")
		}

		for _, arg := range args {
			if err := caseConverter.Convert(arg, fromType, outputType); err != nil {
				log.Error().Err(err).Str("path", arg).
					Str("outputType", outputType.String()).
					Msg("convert case failed")
				return err
			}
		}

		return nil
	},
}

var (
	outputDir   string
	profilePath string

	fromJSONFlag    bool
	fromYAMLFlag    bool
	fromPostmanFlag bool
	fromHARFlag     bool

	toJSONFlag   bool
	toYAMLFlag   bool
	toPyTestFlag bool
)

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().BoolVar(&fromJSONFlag, "from-json", true, "load from json case format")
	convertCmd.Flags().BoolVar(&fromYAMLFlag, "from-yaml", false, "load from yaml case format")
	convertCmd.Flags().BoolVar(&fromHARFlag, "from-har", false, "load from HAR format")
	convertCmd.Flags().BoolVar(&fromPostmanFlag, "from-postman", false, "load from postman format")

	convertCmd.Flags().BoolVar(&toJSONFlag, "to-json", true, "convert to JSON case scripts")
	convertCmd.Flags().BoolVar(&toYAMLFlag, "to-yaml", false, "convert to YAML case scripts")
	convertCmd.Flags().BoolVar(&toPyTestFlag, "to-pytest", false, "convert to pytest scripts")

	convertCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "specify output directory")
	convertCmd.Flags().StringVarP(&profilePath, "profile", "p", "", "specify profile path to override headers and cookies")
}
