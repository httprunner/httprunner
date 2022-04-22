package hrp

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/hrp/internal/builtin"
)

func Convert2TestScripts(destType string, paths ...string) error {
	if destType == "gotest" {
		return convert2GoTestScripts(paths...)
	} else {
		return convert2PyTestScripts(paths...)
	}
}

func convert2PyTestScripts(paths ...string) error {
	python3, err := builtin.EnsurePython3Venv("httprunner")
	if err != nil {
		return errors.Wrap(err, "ensure python venv failed")
	}

	args := append([]string{"-m", "httprunner", "make"}, paths...)
	return builtin.ExecCommand(python3, args...)
}

func convert2GoTestScripts(paths ...string) error {
	log.Warn().Msg("convert to gotest scripts is not supported yet")
	// report event
	// sdk.SendEvent(sdk.EventTracking{
	// 	Category: "Convert",
	// 	Action:   fmt.Sprintf("hrp convert to %s", destType),
	// })

	// var testCasePaths []ITestCase
	// for _, path := range paths {
	// 	testCasePath := TestCasePath(path)
	// 	testCasePaths = append(testCasePaths, &testCasePath)
	// }

	// _, err := loadTestCases(testCasePaths...)
	// if err != nil {
	// 	log.Error().Err(err).Msg("failed to load testcases")
	// 	return err
	// }
	return nil
}
