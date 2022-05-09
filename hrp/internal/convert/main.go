package convert

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

func Convert2TestScripts(destType string, paths ...string) error {
	// report event
	sdk.SendEvent(sdk.EventTracking{
		Category: "ConvertTests",
		Action:   fmt.Sprintf("hrp convert --%s", destType),
	})

	if destType == "gotest" {
		return convert2GoTestScripts(paths...)
	} else {
		// default to pytest
		return convert2PyTestScripts(paths...)
	}
}

func convert2PyTestScripts(paths ...string) error {
	httprunner := fmt.Sprintf("httprunner>=%s", version.HttpRunnerMinVersion)
	python3, err := builtin.EnsurePython3Venv(httprunner)
	if err != nil {
		return err
	}

	args := append([]string{"-m", "httprunner", "make"}, paths...)
	return builtin.ExecCommand(python3, args...)
}

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
		converter := CaseConverter{
			TCase: tc,
		}
		pytestPath, err := converter.ToPyTest()
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
	python3, err := builtin.EnsurePython3Venv("black")
	if err != nil {
		return err
	}
	args := append([]string{"-m", "black"}, pytestPaths...)
	return builtin.ExecCommand(python3, args...)
}

//go:embed testcase.tmpl
var testcaseTemplate string

type CaseConverter struct {
	*hrp.TCase
}

func (c *CaseConverter) ToPyTest() (string, error) {
	script := convertConfig(c.TCase.Config)
	println(script)
	return script, nil
}

func (c *CaseConverter) ToGoTest() (string, error) {
	return "", nil
}

func convertConfig(config *hrp.TConfig) string {
	script := fmt.Sprintf("Config('%s')", config.Name)

	if config.Variables != nil {
		script += fmt.Sprintf(".variables(**{%v})", config.Variables)
	}
	if config.BaseURL != "" {
		script += fmt.Sprintf(".base_url('%s')", config.BaseURL)
	}
	if config.Export != nil {
		script += fmt.Sprintf(".export(*%v)", config.Export)
	}
	script += fmt.Sprintf(".verify(%v)", config.Verify)

	return script
}
