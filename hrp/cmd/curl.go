package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/boomer"
	"github.com/httprunner/httprunner/v4/hrp/pkg/convert"
)

var runCurlCmd = &cobra.Command{
	Use:                "curl URLs",
	Short:              "run API test with curl command",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := makeHRPRunner()
		return runner.Run(makeCurlTestCase(args))
	},
}

var boomCurlCmd = &cobra.Command{
	Use:                "curl URLs",
	Short:              "run load test with curl command",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		boomer.SetUlimit(10240)
		if !strings.EqualFold(logLevel, "DEBUG") {
			logLevel = "WARN" // disable info logs for load testing
		}
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		boomer, err := makeHRPBoomer()
		if err != nil {
			return err
		}
		boomer.Run(makeCurlTestCase(args))
		return nil
	},
}

var convertCurlCmd = &cobra.Command{
	Use:                "curl URLs",
	Short:              "convert curl command to httprunner testcase",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		curlCommand := makeCurlCommand(args)
		return convertRun(cmd, []string{curlCommand})
	},
}

func init() {
	runCmd.AddCommand(runCurlCmd)
	boomCmd.AddCommand(boomCurlCmd)
	convertCmd.AddCommand(convertCurlCmd)
}

func makeCurlTestCase(args []string) *hrp.TestCase {
	curlCommand := makeCurlCommand(args)
	tCase, err := convert.LoadSingleCurlCase(curlCommand)
	if err != nil {
		log.Error().Err(err).Msg("convert curl command failed")
		os.Exit(1)
	}
	casePath, _ := os.Getwd()
	testCase, err := tCase.ToTestCase(casePath)
	if err != nil {
		log.Error().Err(err).Msg("convert testcase to failed")
		os.Exit(1)
	}
	return testCase
}

func makeCurlCommand(args []string) string {
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") {
			args[i] = fmt.Sprintf("\"%s\"", args[i])
		}
	}
	var curlCmd []string
	curlCmd = append(curlCmd, "curl")
	curlCmd = append(curlCmd, args...)
	return strings.Join(curlCmd, " ")
}
