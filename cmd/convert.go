package cmd

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/funplugin/myexec"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/convert"
	"github.com/httprunner/httprunner/v5/internal/builtin"
)

var CmdConvert = &cobra.Command{
	Use:          "convert $path...",
	Short:        "Convert multiple source format to HttpRunner JSON/YAML/gotest/pytest cases",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		caseConverter := convert.NewConverter(outputDir, profilePath)

		var fromType convert.FromType
		if fromYAMLFlag {
			fromType = convert.FromTypeYAML
		} else if fromPostmanFlag {
			fromType = convert.FromTypePostman
		} else if fromHARFlag {
			fromType = convert.FromTypeHAR
		} else if fromCurlFlag {
			fromType = convert.FromTypeCurl
		} else {
			fromType = convert.FromTypeJSON
			log.Info().Str("fromType", fromType.String()).Msg("set default")
		}

		var outputType convert.OutputType
		if toYAMLFlag {
			outputType = convert.OutputTypeYAML
		} else if toPyTestFlag {
			packages := []string{"httprunner"}
			_, err := myexec.EnsurePython3Venv(venv, packages...)
			if err != nil {
				log.Error().Err(err).Msg("python3 venv is not ready")
				return errors.Wrap(code.InvalidPython3Venv, err.Error())
			}

			outputType = convert.OutputTypePyTest
		} else {
			outputType = convert.OutputTypeJSON
			log.Info().Str("outputType", outputType.String()).Msg("set default")
		}

		var files []string
		for _, arg := range args {
			if builtin.IsFolderPathExists(arg) {
				fs, err := os.ReadDir(arg)
				if err != nil {
					log.Error().Err(err).Str("path", arg).Msg("read dir failed")
					continue
				}
				for _, f := range fs {
					files = append(files, filepath.Join(arg, f.Name()))
				}
			} else {
				files = append(files, arg)
			}
		}

		for _, file := range files {
			extName := filepath.Ext(file)
			if !builtin.Contains(fromType.Extensions(), extName) {
				log.Warn().Str("path", file).
					Strs("expectExtensions", fromType.Extensions()).
					Msg("skip file")
				continue
			}

			if err := caseConverter.Convert(file, fromType, outputType); err != nil {
				log.Error().Err(err).Str("path", file).
					Str("outputType", outputType.String()).
					Msg("convert case failed")
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
	fromCurlFlag    bool

	toJSONFlag   bool
	toYAMLFlag   bool
	toPyTestFlag bool
)

func init() {
	CmdConvert.Flags().BoolVar(&fromJSONFlag, "from-json", true, "load from json case format")
	CmdConvert.Flags().BoolVar(&fromYAMLFlag, "from-yaml", false, "load from yaml case format")
	CmdConvert.Flags().BoolVar(&fromHARFlag, "from-har", false, "load from HAR format")
	CmdConvert.Flags().BoolVar(&fromPostmanFlag, "from-postman", false, "load from postman format")
	CmdConvert.Flags().BoolVar(&fromCurlFlag, "from-curl", false, "load from curl format")

	CmdConvert.Flags().BoolVar(&toJSONFlag, "to-json", true, "convert to JSON case scripts")
	CmdConvert.Flags().BoolVar(&toYAMLFlag, "to-yaml", false, "convert to YAML case scripts")
	CmdConvert.Flags().BoolVar(&toPyTestFlag, "to-pytest", false, "convert to pytest scripts")

	CmdConvert.Flags().StringVarP(&outputDir, "output-dir", "d", "", "specify output directory")
	CmdConvert.Flags().StringVarP(&profilePath, "profile", "p", "", "specify profile path to override headers and cookies")
}
