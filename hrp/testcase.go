package hrp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/httprunner/httprunner/hrp/internal/builtin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ITestCase represents interface for testcases,
// includes TestCase and TestCasePath.
type ITestCase interface {
	GetPath() string
	ToTestCase() (*TestCase, error)
}

// TestCase is a container for one testcase, which is used for testcase runner.
// TestCase implements ITestCase interface.
type TestCase struct {
	Config    *TConfig
	TestSteps []IStep
}

func (tc *TestCase) GetPath() string {
	return tc.Config.Path
}

func (tc *TestCase) ToTestCase() (*TestCase, error) {
	return tc, nil
}

func (tc *TestCase) ToTCase() *TCase {
	tCase := &TCase{
		Config: tc.Config,
	}
	for _, step := range tc.TestSteps {
		tCase.TestSteps = append(tCase.TestSteps, step.ToStruct())
	}
	return tCase
}

// TestCasePath implements ITestCase interface.
type TestCasePath string

func (path *TestCasePath) GetPath() string {
	return fmt.Sprintf("%v", *path)
}

// ToTestCase loads testcase path and convert to *TestCase
func (path *TestCasePath) ToTestCase() (*TestCase, error) {
	tc := &TCase{}
	casePath := path.GetPath()
	err := builtin.LoadFile(casePath, tc)
	if err != nil {
		return nil, err
	}

	err = tc.makeCompat()
	if err != nil {
		return nil, err
	}
	tc.Config.Path = casePath

	testCase := &TestCase{
		Config: tc.Config,
	}

	// locate project root dir by plugin path
	projectRootDir, err := getProjectRootDirPath(casePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project root dir")
	}

	for _, step := range tc.TestSteps {
		if step.API != nil {
			apiPath, ok := step.API.(string)
			if !ok {
				return nil, fmt.Errorf("referenced api path should be string, got %v", step.API)
			}
			path := filepath.Join(projectRootDir, apiPath)
			if !builtin.IsFilePathExists(path) {
				return nil, errors.New("referenced api file not found: " + path)
			}

			refAPI := APIPath(path)
			apiContent, err := refAPI.ToAPI()
			if err != nil {
				return nil, err
			}
			step.API = apiContent

			testCase.TestSteps = append(testCase.TestSteps, &StepAPIWithOptionalArgs{
				step: step,
			})
		} else if step.TestCase != nil {
			casePath, ok := step.TestCase.(string)
			if !ok {
				return nil, fmt.Errorf("referenced testcase path should be string, got %v", step.TestCase)
			}
			path := filepath.Join(projectRootDir, casePath)
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

// TCase represents testcase data structure.
// Each testcase includes one public config and several sequential teststeps.
type TCase struct {
	Config    *TConfig `json:"config" yaml:"config"`
	TestSteps []*TStep `json:"teststeps" yaml:"teststeps"`
}

// makeCompat converts TCase to compatible testcase
func (tc *TCase) makeCompat() error {
	var err error
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

func loadTestCases(iTestCases ...ITestCase) ([]*TestCase, error) {
	testCases := make([]*TestCase, 0)

	for _, iTestCase := range iTestCases {
		if _, ok := iTestCase.(*TestCase); ok {
			testcase, err := iTestCase.ToTestCase()
			if err != nil {
				log.Error().Err(err).Msg("failed to convert ITestCase interface to TestCase struct")
				return nil, err
			}
			testCases = append(testCases, testcase)
			continue
		}

		// iTestCase should be a TestCasePath, file path or folder path
		tcPath, ok := iTestCase.(*TestCasePath)
		if !ok {
			return nil, errors.New("invalid iTestCase type")
		}

		casePath := tcPath.GetPath()
		err := fs.WalkDir(os.DirFS(casePath), ".", func(path string, dir fs.DirEntry, e error) error {
			if dir == nil {
				// casePath is a file other than a dir
				path = casePath
			} else if dir.IsDir() && path != "." && strings.HasPrefix(path, ".") {
				// skip hidden folders
				return fs.SkipDir
			} else {
				// casePath is a dir
				path = filepath.Join(casePath, path)
			}

			// ignore non-testcase files
			ext := filepath.Ext(path)
			if ext != ".yml" && ext != ".yaml" && ext != ".json" {
				return nil
			}

			// filtered testcases
			testCasePath := TestCasePath(path)
			tc, err := testCasePath.ToTestCase()
			if err != nil {
				log.Error().Err(err).Str("path", path).Msg("load testcase failed")
				return errors.Wrap(err, "load testcase failed")
			}
			testCases = append(testCases, tc)
			return nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "read dir failed")
		}
	}

	log.Info().Int("count", len(testCases)).Msg("load testcases successfully")
	return testCases, nil
}
