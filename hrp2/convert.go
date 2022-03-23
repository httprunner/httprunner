package hrp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/httprunner/hrp/internal/json"
)

func loadFromJSON(path string, structObj interface{}) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("load json")

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load json path failed")
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(file))
	decoder.UseNumber()
	err = decoder.Decode(structObj)
	return err
}

func loadFromYAML(path string, structObj interface{}) error {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		return err
	}
	log.Info().Str("path", path).Msg("load yaml")

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load yaml path failed")
		return err
	}

	err = yaml.Unmarshal(file, structObj)
	return err
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
	for _, step := range tc.TestSteps {
		if step.APIPath != "" {
			refAPI := APIPath(step.APIPath)
			step.APIContent = &refAPI
			apiContent, err := step.APIContent.ToAPI()
			if err != nil {
				return nil, err
			}
			step.APIContent = apiContent
			testCase.TestSteps = append(testCase.TestSteps, &StepAPIWithOptionalArgs{
				step: step,
			})
		} else if step.TestCasePath != "" {
			refTestCase := TestCasePath(step.TestCasePath)
			step.TestCaseContent = &refTestCase
			tc, err := step.TestCaseContent.ToTestCase()
			if err != nil {
				return nil, err
			}
			step.TestCaseContent = tc
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

var ErrUnsupportedFileExt = fmt.Errorf("unsupported testcase file extension")

// APIPath implements IAPI interface.
type APIPath string

func (path *APIPath) ToString() string {
	return fmt.Sprintf("%v", *path)
}

func (path *APIPath) ToAPI() (*API, error) {
	api := &API{}
	var err error

	apiPath := path.ToString()
	ext := filepath.Ext(apiPath)
	switch ext {
	case ".json":
		err = loadFromJSON(apiPath, api)
	case ".yaml", ".yml":
		err = loadFromYAML(apiPath, api)
	default:
		err = ErrUnsupportedFileExt
	}
	if err != nil {
		return nil, err
	}
	err = convertCompatValidator(api.Validators)
	return api, err
}

// TestCasePath implements ITestCase interface.
type TestCasePath string

func (path *TestCasePath) ToString() string {
	return fmt.Sprintf("%v", *path)
}

func (path *TestCasePath) ToTestCase() (*TestCase, error) {
	tc := &TCase{}
	var err error

	casePath := path.ToString()
	ext := filepath.Ext(casePath)
	switch ext {
	case ".json":
		err = loadFromJSON(casePath, tc)
	case ".yaml", ".yml":
		err = loadFromYAML(casePath, tc)
	default:
		err = ErrUnsupportedFileExt
	}
	if err != nil {
		return nil, err
	}
	err = convertCompatTestCase(tc)
	if err != nil {
		return nil, err
	}
	tc.Config.Path = path.ToString()
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
