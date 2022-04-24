package cmd

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/hrp/internal/convert"
)

var convertCmd = &cobra.Command{
	Use:   "convert $path...",
	Short: "convert JSON/YAML testcases to pytest/gotest scripts",
	Args:  cobra.ExactValidArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !pytestFlag && !gotestFlag {
			return errors.New("please specify convertion type")
		}

		var err error
		if gotestFlag {
			err = convert.Convert2TestScripts("gotest", args...)
		} else {
			err = convert.Convert2TestScripts("pytest", args...)
		}
		if err != nil {
			log.Error().Err(err).Msg("convert test scripts failed")
			os.Exit(1)
		}
		return nil
	},
}

var (
	pytestFlag bool
	gotestFlag bool
)

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().BoolVar(&pytestFlag, "pytest", true, "convert to pytest scripts")
	convertCmd.Flags().BoolVar(&gotestFlag, "gotest", false, "convert to gotest scripts (TODO)")
}
