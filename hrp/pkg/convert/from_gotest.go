package convert

import (
	_ "embed"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/myexec"
)

func convert2GoTestScripts(paths ...string) error {
	log.Warn().Msg("convert to gotest scripts is not supported yet")
	os.Exit(1)

	// TODO
	var testCasePaths []hrp.ITestCase
	for _, path := range paths {
		testCasePath := hrp.TestCasePath(path)
		testCasePaths = append(testCasePaths, &testCasePath)
	}

	testCases, err := hrp.LoadTestCases(testCasePaths...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		return err
	}

	var pytestPaths []string
	for _, testCase := range testCases {
		tc := testCase.ToTCase()
		converter := TCaseConverter{
			tCase: tc,
		}
		pytestPath, err := converter.toPyTest()
		if err != nil {
			log.Error().Err(err).
				Str("originPath", tc.Config.Path).
				Msg("convert to pytest failed")
			continue
		}
		log.Info().
			Str("pytestPath", pytestPath).
			Str("originPath", tc.Config.Path).
			Msg("convert to pytest success")
		pytestPaths = append(pytestPaths, pytestPath)
	}

	// format pytest scripts with black
	return myexec.ExecPython3Command("black", pytestPaths...)
}

//go:embed testcase.tmpl
var testcaseTemplate string
