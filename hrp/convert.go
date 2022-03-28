package hrp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/hrp/internal/builtin"
)

func convertCompatValidator(Validators []interface{}) (err error) {
	for i, iValidator := range Validators {
		validatorMap := iValidator.(map[string]interface{})
		validator := Validator{}
		_, checkExisted := validatorMap["check"]
		_, assertExisted := validatorMap["assert"]
		_, expectExisted := validatorMap["expect"]
		// check priority: HRP > HttpRunner
		if checkExisted && assertExisted && expectExisted {
			// HRP validator format
			validator.Check = validatorMap["check"].(string)
			validator.Assert = validatorMap["assert"].(string)
			validator.Expect = validatorMap["expect"]
			if msg, existed := validatorMap["msg"]; existed {
				validator.Message = msg.(string)
			}
			validator.Check = convertCheckExpr(validator.Check)
			Validators[i] = validator
		} else if len(validatorMap) == 1 {
			// HttpRunner validator format
			for assertMethod, iValidatorContent := range validatorMap {
				checkAndExpect := iValidatorContent.([]interface{})
				if len(checkAndExpect) != 2 {
					return fmt.Errorf("unexpected validator format: %v", validatorMap)
				}
				validator.Check = checkAndExpect[0].(string)
				validator.Assert = assertMethod
				validator.Expect = checkAndExpect[1]
			}
			validator.Check = convertCheckExpr(validator.Check)
			Validators[i] = validator
		} else {
			return fmt.Errorf("unexpected validator format: %v", validatorMap)
		}
	}
	return nil
}

func convertCompatTestCase(tc *TCase) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("convert compat testcase error: %v", p)
		}
	}()
	for _, step := range tc.TestSteps {
		// 1. deal with request body compatible with HttpRunner
		if step.Request != nil && step.Request.Body == nil {
			if step.Request.Json != nil {
				step.Request.Headers["Content-Type"] = "application/json; charset=utf-8"
				step.Request.Body = step.Request.Json
			} else if step.Request.Data != nil {
				step.Request.Body = step.Request.Data
			}
		}

		// 2. deal with validators compatible with HttpRunner
		err = convertCompatValidator(step.Validators)
		if err != nil {
			return err
		}
	}
	return nil
}

// convertCheckExpr deals with check expression including hyphen
func convertCheckExpr(checkExpr string) string {
	if strings.Contains(checkExpr, textExtractorSubRegexp) {
		return checkExpr
	}
	checkItems := strings.Split(checkExpr, ".")
	for i, checkItem := range checkItems {
		if strings.Contains(checkItem, "-") && !strings.Contains(checkItem, "\"") {
			checkItems[i] = fmt.Sprintf("\"%s\"", checkItem)
		}
	}
	return strings.Join(checkItems, ".")
}

func (tc *TCase) ToTestCase() (*TestCase, error) {
	testCase := &TestCase{
		Config: tc.Config,
	}

	// locate project root dir by plugin path
	projectRootDir, err := getProjectRootDirPath(testCase.Config.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project root dir")
	}
	log.Info().Str("dir", projectRootDir).Msg("located project root dir")

	for _, step := range tc.TestSteps {
		if step.API != nil {
			if apiContent, ok := step.API.(*API); ok {
				step.API = apiContent
				testCase.TestSteps = append(testCase.TestSteps, &StepAPIWithOptionalArgs{
					step: step,
				})
				return testCase, nil
			}

			// reference api path
			var apiFullPath string
			if apiPath, ok := step.API.(string); ok {
				apiFullPath = filepath.Join(projectRootDir, apiPath)
			} else if apiPath, ok := step.API.(APIPath); ok {
				apiFullPath = filepath.Join(projectRootDir, apiPath.GetPath())
			} else {
				return nil, errors.New("invalid api format")
			}

			if !builtin.IsFilePathExists(apiFullPath) {
				return nil, errors.New("referenced api file not found: " + apiFullPath)
			}

			refAPI := APIPath(apiFullPath)
			apiContent, err := refAPI.ToAPI()
			if err != nil {
				return nil, err
			}
			step.API = apiContent

			testCase.TestSteps = append(testCase.TestSteps, &StepAPIWithOptionalArgs{
				step: step,
			})
		} else if step.TestCase != nil {
			path := filepath.Join(projectRootDir, step.TestCase.(string))
			if !builtin.IsFilePathExists(path) {
				return nil, errors.New("referenced testcase file not found: " + path)
			}

			refTestCase := TestCasePath(path)
			tc, err := refTestCase.ToTestCase()
			if err != nil {
				return nil, err
			}
			step.TestCase = tc
			testCase.TestSteps = append(testCase.TestSteps, &StepTestCaseWithOptionalArgs{
				step: step,
			})
		} else if step.ThinkTime != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepThinkTime{
				step: step,
			})
		} else if step.Request != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepRequestWithOptionalArgs{
				step: step,
			})
		} else if step.Transaction != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepTransaction{
				step: step,
			})
		} else if step.Rendezvous != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepRendezvous{
				step: step,
			})
		} else {
			log.Warn().Interface("step", step).Msg("[convertTestCase] unexpected step")
		}
	}
	return testCase, nil
}

func getProjectRootDirPath(path string) (rootDir string, err error) {
	pluginPath, err := locatePlugin(path)
	if err == nil {
		rootDir = filepath.Dir(pluginPath)
		return
	}

	// failed to locate project root dir
	// maybe project plugin debugtalk.xx is not exist
	// use current dir instead
	return os.Getwd()
}

// APIPath implements IAPI interface.
type APIPath string

func (path *APIPath) GetPath() string {
	return fmt.Sprintf("%v", *path)
}

func (path *APIPath) ToAPI() (*API, error) {
	api := &API{}
	apiPath := path.GetPath()
	err := builtin.LoadFile(apiPath, api)
	if err != nil {
		return nil, err
	}
	err = convertCompatValidator(api.Validators)
	return api, err
}

// TestCasePath implements ITestCase interface.
type TestCasePath string

func (path *TestCasePath) GetPath() string {
	return fmt.Sprintf("%v", *path)
}

func (path *TestCasePath) ToTestCase() (*TestCase, error) {
	tc := &TCase{}
	casePath := path.GetPath()
	err := builtin.LoadFile(casePath, tc)
	if err != nil {
		return nil, err
	}

	err = convertCompatTestCase(tc)
	if err != nil {
		return nil, err
	}
	tc.Config.Path = casePath
	testcase, err := tc.ToTestCase()
	if err != nil {
		return nil, err
	}
	return testcase, nil
}

func (path *TestCasePath) ToTCase() (*TCase, error) {
	testcase, err := path.ToTestCase()
	if err != nil {
		return nil, err
	}
	return testcase.ToTCase()
}
