package cmd

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/boomer"
	"github.com/httprunner/httprunner/v4/hrp/internal/convert"
)

var runCurlCmd = &cobra.Command{
	Use:   "curl URLs",
	Short: "run API test with go engine by curl command",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		runner := makeHRPRunner()
		if runner.Run(makeCurlTestCase(args)) != nil {
			os.Exit(1)
		}
	},
}

var boomCurlCmd = &cobra.Command{
	Use:   "curl URLs",
	Short: "run load test with boomer by curl command",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		boomer.SetUlimit(10240) // ulimit -n 10240
		if !strings.EqualFold(logLevel, "DEBUG") {
			logLevel = "WARN" // disable info logs for load testing
		}
		setLogLevel(logLevel)
	},
	Run: func(cmd *cobra.Command, args []string) {
		boomer := makeHRPBoomer()
		boomer.Run(makeCurlTestCase(args))
	},
}

var convertCurlCmd = &cobra.Command{
	Use:   "curl URLs",
	Short: "convert curl command to httprunner testcase",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		setLogLevel(logLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		curlCommand := makeCurlCommand(args)
		return convertRun(cmd, []string{curlCommand})
	},
}

var (
	cookieSlice []string
	dataSlice   []string
	formSlice   []string
	get         bool
	head        bool
	headerSlice []string
	request     string
)

func init() {
	runCmd.AddCommand(runCurlCmd)
	addCurlFlags(runCurlCmd)

	boomCmd.AddCommand(boomCurlCmd)
	addCurlFlags(boomCurlCmd)

	convertCmd.AddCommand(convertCurlCmd)
	addCurlFlags(convertCurlCmd)
}

func addCurlFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&cookieSlice, "cookie", "b", nil, "-b, --cookie in curl")
	cmd.Flags().StringSliceVarP(&dataSlice, "data", "d", nil, "-d, --data in curl")
	cmd.Flags().StringSliceVarP(&formSlice, "form", "F", nil, "-F, --form in curl")
	cmd.Flags().BoolVarP(&get, "get", "G", false, "-G, --get in curl")
	cmd.Flags().BoolVarP(&head, "head", "I", false, "-I, --head in curl")
	cmd.Flags().StringSliceVarP(&headerSlice, "header", "H", nil, "-H, --header in curl")
	cmd.Flags().StringVarP(&request, "request", "X", "", "-X, --request in curl")
}

func makeCurlTestCase(args []string) *hrp.TestCase {
	curlCommand := makeCurlCommand(args)
	tCase, err := convert.LoadSingleCurlCase(curlCommand)
	if err != nil {
		log.Error().Err(err).Msg("convert curl command failed")
		os.Exit(1)
	}
	casePath, err := os.Getwd()
	if err != nil {
		casePath = ""
		log.Error().Err(err).Msg("get working directory failed")
	}
	testCase, err := tCase.ToTestCase(casePath)
	if err != nil {
		log.Error().Err(err).Msg("convert testcase to failed")
		os.Exit(1)
	}
	return testCase
}

func makeCurlCommand(args []string) string {
	var cmdList []string
	cmdList = append(cmdList, "curl")
	for _, c := range cookieSlice {
		cmdList = append(cmdList, "--cookie", c)
	}
	for _, d := range dataSlice {
		cmdList = append(cmdList, "--data", d)
	}
	for _, f := range formSlice {
		cmdList = append(cmdList, "--form", f)
	}
	if get {
		cmdList = append(cmdList, "--get")
	}
	if head {
		cmdList = append(cmdList, "--head")
	}
	for _, h := range headerSlice {
		cmdList = append(cmdList, "--header", h)
	}
	if request != "" {
		cmdList = append(cmdList, "--request", request)
	}
	cmdList = append(cmdList, args...)
	return strings.Join(cmdList, " ")
}
