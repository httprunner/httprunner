package hrp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

// ITestCase represents interface for testcases,
// includes TestCase and TestCasePath.
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

// TestCase is a container for one testcase, which is used for testcase runner.
// TestCase implements ITestCase interface.
type TestCase struct {
	Config    IConfig `json:"config" yaml:"config"`
	TestSteps []IStep `json:"teststeps" yaml:"teststeps"`
}

func (tc *TestCase) GetTestCase() (*TestCase, error) {
	return tc, nil
}

// MakeCompat converts TestCase compatible with Golang engine style
func (tc *TestCaseDef) MakeCompat() (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("[MakeCompat] convert compat testcase error: %v", p)
		}
	}()
	for _, step := range tc.Steps {
		// 1. deal with request body compatibility
		convertCompatRequestBody(step.Request)

		// 2. deal with validators compatibility
		err = convertCompatValidator(step.Validators)
		if err != nil {
			return err
		}

		// 3. deal with extract expr including hyphen
		convertExtract(step.Extract)

		// 4. deal with mobile step compatibility
		if step.Android != nil {
			convertCompatMobileStep(step.Android)
		} else if step.IOS != nil {
			convertCompatMobileStep(step.IOS)
		}
	}
	return nil
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

// loadISteps loads TestSteps([]IStep) from TSteps structs
func (tc *TestCaseDef) loadISteps() (*TestCase, error) {
	testCase := &TestCase{
		Config: tc.Config,
	}

	err := tc.MakeCompat()
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
		envVars := make(map[string]string)
		err = LoadFileObject(dotEnvPath, envVars)
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

func convertCompatRequestBody(request *Request) {
	if request != nil && request.Body == nil {
		if request.Json != nil {
			if request.Headers == nil {
				request.Headers = make(map[string]string)
			}
			request.Headers["Content-Type"] = "application/json; charset=utf-8"
			request.Body = request.Json
			request.Json = nil
		} else if request.Data != nil {
			request.Body = request.Data
			request.Data = nil
		}
	}
}

func convertCompatValidator(Validators []interface{}) (err error) {
	for i, iValidator := range Validators {
		if _, ok := iValidator.(Validator); ok {
			continue
		}

		var validatorMap map[string]interface{}
		if v, ok := iValidator.(map[string]interface{}); ok {
			validatorMap = v
		} else if v, ok := iValidator.(map[interface{}]interface{}); ok {
			// convert map[interface{}]interface{} to map[string]interface{}
			validatorMap = make(map[string]interface{})
			for key, value := range v {
				strKey := fmt.Sprintf("%v", key)
				validatorMap[strKey] = value
			}
		} else {
			return errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("unexpected validator format: %v", iValidator))
		}

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
			validator.Check = convertJmespathExpr(validator.Check)
			Validators[i] = validator
			continue
		}
		if len(validatorMap) == 1 {
			// Python engine style
			for assertMethod, iValidatorContent := range validatorMap {
				validatorContent := iValidatorContent.([]interface{})
				if len(validatorContent) > 3 {
					return errors.Wrap(code.InvalidCaseError,
						fmt.Sprintf("unexpected validator format: %v", validatorMap))
				}
				validator.Check = validatorContent[0].(string)
				validator.Assert = assertMethod
				validator.Expect = validatorContent[1]
				if len(validatorContent) == 3 {
					validator.Message = validatorContent[2].(string)
				}
			}
			validator.Check = convertJmespathExpr(validator.Check)
			Validators[i] = validator
			continue
		}
		return errors.Wrap(code.InvalidCaseError,
			fmt.Sprintf("unexpected validator format: %v", validatorMap))
	}
	return nil
}

// convertExtract deals with extract expr including hyphen
func convertExtract(extract map[string]string) {
	for key, value := range extract {
		extract[key] = convertJmespathExpr(value)
	}
}

func convertCompatMobileStep(mobileUI *MobileUI) {
	if mobileUI == nil {
		return
	}
	for i := 0; i < len(mobileUI.Actions); i++ {
		ma := mobileUI.Actions[i]
		actionOptions := uixt.NewActionOptions(ma.GetOptions()...)
		// append tap_cv params to screenshot_with_ui_types option
		if ma.Method == uixt.ACTION_TapByCV {
			uiTypes, _ := builtin.ConvertToStringSlice(ma.Params)
			ma.ActionOptions.ScreenShotWithUITypes = append(ma.ActionOptions.ScreenShotWithUITypes, uiTypes...)
			ma.ActionOptions.ScreenShotWithUpload = true
		}
		// set default max_retry_times to 10 for swipe_to_tap_texts
		if ma.Method == uixt.ACTION_SwipeToTapTexts && actionOptions.MaxRetryTimes == 0 {
			ma.ActionOptions.MaxRetryTimes = 10
		}
		// set default max_retry_times to 10 for swipe_to_tap_text
		if ma.Method == uixt.ACTION_SwipeToTapText && actionOptions.MaxRetryTimes == 0 {
			ma.ActionOptions.MaxRetryTimes = 10
		}
		if ma.Method == uixt.ACTION_Swipe {
			ma.ActionOptions.Direction = ma.Params
		}
		mobileUI.Actions[i] = ma
	}
}

// convertJmespathExpr deals with limited jmespath expression conversion
func convertJmespathExpr(checkExpr string) string {
	if strings.Contains(checkExpr, textExtractorSubRegexp) {
		return checkExpr
	}
	checkItems := strings.Split(checkExpr, ".")
	for i, checkItem := range checkItems {
		checkItem = strings.Trim(checkItem, "\"")
		lowerItem := strings.ToLower(checkItem)
		if strings.HasPrefix(lowerItem, "content-") || lowerItem == "user-agent" {
			checkItems[i] = fmt.Sprintf("\"%s\"", checkItem)
		}
	}
	return strings.Join(checkItems, ".")
}
