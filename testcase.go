package hrp

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/json"
)

// ITestCase represents interface for testcases,
// includes TestCase, TestCasePath and TestCaseJSON
type ITestCase interface {
	GetTestCase() (*TestCase, error)
}

// TestCasePath implements ITestCase interface.
type TestCasePath string

// GetTestCase loads testcase path and convert to *TestCase
func (path *TestCasePath) GetTestCase() (*TestCase, error) {
	tc := &TestCaseDef{}
	casePath := string(*path)
	err := LoadFileObject(casePath, tc)
	if err != nil {
		return nil, err
	}
	if tc.Steps == nil {
		return nil, errors.Wrap(code.InvalidCaseError,
			"invalid testcase format, missing teststeps!")
	}

	if tc.Config == nil {
		tc.Config = &TConfig{Name: "please input testcase name"}
	}
	tc.Config.Path = casePath
	return tc.loadISteps()
}

// TestCaseJSON implements ITestCase interface.
type TestCaseJSON string

// GetTestCase unmarshal json string and convert to *TestCase
func (tc *TestCaseJSON) GetTestCase() (*TestCase, error) {
	var testCaseDef TestCaseDef
	err := json.Unmarshal([]byte(*tc), &testCaseDef)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal TestCaseJSON failed")
	}
	return testCaseDef.loadISteps()
}

// TestCase is a container for one testcase, which is used for testcase runner.
// TestCase implements ITestCase interface.
type TestCase struct {
	Config    IConfig `json:"config" yaml:"config"`
	TestSteps []IStep `json:"teststeps" yaml:"teststeps"`
}

func (tc *TestCase) GetTestCase() (*TestCase, error) {
	return tc, nil
}

func (tc *TestCase) Dump2JSON(targetPath string) error {
	err := builtin.Dump2JSON(tc, targetPath)
	if err != nil {
		return errors.Wrap(err, "dump testcase to json failed")
	}
	return nil
}

func (tc *TestCase) Dump2YAML(targetPath string) error {
	err := builtin.Dump2YAML(tc, targetPath)
	if err != nil {
		return errors.Wrap(err, "dump testcase to yaml failed")
	}
	return nil
}

// define struct for testcase
type TestCaseDef struct {
	Config *TConfig `json:"config" yaml:"config"`
	Steps  []*TStep `json:"teststeps" yaml:"teststeps"`
}

// loadISteps loads TestSteps([]IStep) from TSteps structs
func (tc *TestCaseDef) loadISteps() (*TestCase, error) {
	testCase := &TestCase{
		Config: tc.Config,
	}

	err := ConvertCaseCompatibility(tc)
	if err != nil {
		return nil, err
	}

	// locate project root dir by plugin path
	projectRootDir, err := GetProjectRootDirPath(tc.Config.Path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get project root dir")
	}

	// load .env file
	dotEnvPath := filepath.Join(projectRootDir, ".env")
	if builtin.IsFilePathExists(dotEnvPath) {
		envVars, err := godotenv.Read(dotEnvPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load .env file")
		}

		config := testCase.Config.Get()
		// override testcase config env with variables loaded from .env file
		// priority: .env file > testcase config env
		if config.Environs == nil {
			config.Environs = make(map[string]string)
		}
		for key, value := range envVars {
			config.Environs[key] = value
		}
	}

	for _, step := range tc.Steps {
		if step.API != nil {
			apiPath, ok := step.API.(string)
			if ok {
				path := filepath.Join(projectRootDir, apiPath)
				if !builtin.IsFilePathExists(path) {
					return nil, errors.Wrap(code.ReferencedFileNotFound,
						fmt.Sprintf("referenced api file not found: %s", path))
				}

				refAPI := APIPath(path)
				apiContent, err := refAPI.ToAPI()
				if err != nil {
					return nil, err
				}
				step.API = apiContent
			} else {
				apiMap, ok := step.API.(map[string]interface{})
				if !ok {
					return nil, errors.Wrap(code.InvalidCaseError,
						fmt.Sprintf("referenced api should be map or path(string), got %v", step.API))
				}
				api := &API{}
				err = mapstructure.Decode(apiMap, api)
				if err != nil {
					return nil, err
				}
				step.API = api
			}
			_, ok = step.API.(*API)
			if !ok {
				return nil, errors.Wrap(code.InvalidCaseError,
					fmt.Sprintf("failed to handle referenced API, got %v", step.TestCase))
			}
			testCase.TestSteps = append(testCase.TestSteps, &StepAPIWithOptionalArgs{
				StepConfig: step.StepConfig,
				API:        step.API,
			})
		} else if step.TestCase != nil {
			casePath, ok := step.TestCase.(string)
			if ok {
				path := filepath.Join(projectRootDir, casePath)
				if !builtin.IsFilePathExists(path) {
					return nil, errors.Wrap(code.ReferencedFileNotFound,
						fmt.Sprintf("referenced testcase file not found: %s", path))
				}

				refTestCase := TestCasePath(path)
				tc, err := refTestCase.GetTestCase()
				if err != nil {
					return nil, err
				}
				step.TestCase = tc
			} else {
				testCaseMap, ok := step.TestCase.(map[string]interface{})
				if !ok {
					return nil, errors.Wrap(code.InvalidCaseError,
						fmt.Sprintf("referenced testcase should be map or path(string), got %v", step.TestCase))
				}
				tCase := &TestCaseDef{}
				err = mapstructure.Decode(testCaseMap, tCase)
				if err != nil {
					return nil, err
				}
				tc, err := tCase.loadISteps()
				if err != nil {
					return nil, err
				}
				step.TestCase = tc
			}
			_, ok = step.TestCase.(*TestCase)
			if !ok {
				return nil, errors.Wrap(code.InvalidCaseError,
					fmt.Sprintf("failed to handle referenced testcase, got %v", step.TestCase))
			}
			testCase.TestSteps = append(testCase.TestSteps, &StepTestCaseWithOptionalArgs{
				StepConfig: step.StepConfig,
				TestCase:   step.TestCase,
			})
		} else if step.ThinkTime != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepThinkTime{
				StepConfig: step.StepConfig,
				ThinkTime:  step.ThinkTime,
			})
		} else if step.Request != nil {
			stepRequest := &StepRequestWithOptionalArgs{
				StepRequest: &StepRequest{
					StepConfig: step.StepConfig,
					Request:    step.Request,
				},
			}
			// init upload
			if len(step.Request.Upload) != 0 {
				initUpload(stepRequest)
			}
			testCase.TestSteps = append(testCase.TestSteps, stepRequest)
		} else if step.Transaction != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepTransaction{
				StepConfig:  step.StepConfig,
				Transaction: step.Transaction,
			})
		} else if step.Rendezvous != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepRendezvous{
				StepConfig: step.StepConfig,
				Rendezvous: step.Rendezvous,
			})
		} else if step.WebSocket != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepWebSocket{
				StepConfig: step.StepConfig,
				WebSocket:  step.WebSocket,
			})
		} else if step.IOS != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepMobile{
				StepConfig: step.StepConfig,
				IOS:        step.IOS,
			})
		} else if step.Harmony != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepMobile{
				StepConfig: step.StepConfig,
				Harmony:    step.Harmony,
			})
		} else if step.Android != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepMobile{
				StepConfig: step.StepConfig,
				Android:    step.Android,
			})
		} else if step.Shell != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepShell{
				StepConfig: step.StepConfig,
				Shell:      step.Shell,
			})
		} else {
			log.Warn().Interface("step", step).Msg("[convertTestCase] unexpected step")
		}
	}
	return testCase, nil
}
