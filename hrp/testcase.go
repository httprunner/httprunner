package hrp

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
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
		tCase.TestSteps = append(tCase.TestSteps, step.Struct())
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
	if tc.Config == nil {
		return nil, errors.New("incorrect testcase file format, expected config in file")
	}

	err = tc.MakeCompat()
	if err != nil {
		return nil, err
	}
	tc.Config.Path = casePath

	testCase := &TestCase{
		Config: tc.Config,
	}

	// locate project root dir by plugin path
	projectRootDir, err := GetProjectRootDirPath(casePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project root dir")
	}

	// load .env file
	dotEnvPath := filepath.Join(projectRootDir, ".env")
	if builtin.IsFilePathExists(dotEnvPath) {
		envVars := make(map[string]string)
		err = builtin.LoadFile(dotEnvPath, envVars)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load .env file")
		}

		// override testcase config env with variables loaded from .env file
		// priority: .env file > testcase config env
		if testCase.Config.Environs == nil {
			testCase.Config.Environs = make(map[string]string)
		}
		for key, value := range envVars {
			testCase.Config.Environs[key] = value
		}
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
		} else if step.WebSocket != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepWebSocket{
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

// MakeCompat converts TCase compatible with Golang engine style
func (tc *TCase) MakeCompat() (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("[MakeCompat] convert compat testcase error: %v", p)
		}
	}()
	for _, step := range tc.TestSteps {
		// 1. deal with request body compatibility
		if step.Request != nil && step.Request.Body == nil {
			if step.Request.Json != nil {
				step.Request.Headers["Content-Type"] = "application/json; charset=utf-8"
				step.Request.Body = step.Request.Json
				step.Request.Json = nil
			} else if step.Request.Data != nil {
				step.Request.Body = step.Request.Data
				step.Request.Data = nil
			}
		}

		// 2. deal with validators compatibility
		err = convertCompatValidator(step.Validators)
		if err != nil {
			return err
		}

		// 3. deal with extract expr including hyphen
		convertExtract(step.Extract)
	}
	return nil
}

func convertCompatValidator(Validators []interface{}) (err error) {
	for i, iValidator := range Validators {
		if _, ok := iValidator.(Validator); ok {
			continue
		}
		validatorMap := iValidator.(map[string]interface{})
		validator := Validator{}
		iCheck, checkExisted := validatorMap["check"]
		iAssert, assertExisted := validatorMap["assert"]
		iExpect, expectExisted := validatorMap["expect"]
		// validator check priority: Golang > Python engine style
		if checkExisted && assertExisted && expectExisted {
			// Golang engine style
			validator.Check = iCheck.(string)
			validator.Assert = iAssert.(string)
			validator.Expect = iExpect
			if iMsg, msgExisted := validatorMap["msg"]; msgExisted {
				validator.Message = iMsg.(string)
			}
			validator.Check = convertCheckExpr(validator.Check)
			Validators[i] = validator
			continue
		}
		if len(validatorMap) == 1 {
			// Python engine style
			for assertMethod, iValidatorContent := range validatorMap {
				validatorContent := iValidatorContent.([]interface{})
				if len(validatorContent) > 3 {
					return fmt.Errorf("unexpected validator format: %v", validatorMap)
				}
				validator.Check = validatorContent[0].(string)
				validator.Assert = assertMethod
				validator.Expect = validatorContent[1]
				if len(validatorContent) == 3 {
					validator.Message = validatorContent[2].(string)
				}
			}
			validator.Check = convertCheckExpr(validator.Check)
			Validators[i] = validator
			continue
		}
		return fmt.Errorf("unexpected validator format: %v", validatorMap)
	}
	return nil
}

// convertExtract deals with extract expr including hyphen
func convertExtract(extract map[string]string) {
	for key, value := range extract {
		extract[key] = convertCheckExpr(value)
	}
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

func LoadTestCases(iTestCases ...ITestCase) ([]*TestCase, error) {
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
				log.Warn().Err(err).Str("path", path).Msg("load testcase failed")
				return nil
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
