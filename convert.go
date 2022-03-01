package hrp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func loadFromJSON(path string) (*TCase, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		return nil, err
	}
	log.Info().Str("path", path).Msg("load json testcase")

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load json path failed")
		return nil, err
	}

	tc := &TCase{}
	decoder := json.NewDecoder(bytes.NewReader(file))
	decoder.UseNumber()
	err = decoder.Decode(tc)
	if err != nil {
		return tc, err
	}
	err = convertCompatTestCase(tc)
	return tc, err
}

func loadFromYAML(path string) (*TCase, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Error().Str("path", path).Err(err).Msg("convert absolute path failed")
		return nil, err
	}
	log.Info().Str("path", path).Msg("load yaml testcase")

	file, err := os.ReadFile(path)
	if err != nil {
		log.Error().Err(err).Msg("load yaml path failed")
		return nil, err
	}

	tc := &TCase{}
	err = yaml.Unmarshal(file, tc)
	if err != nil {
		return tc, nil
	}
	err = convertCompatTestCase(tc)
	return tc, err
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
		for i, iValidator := range step.Validators {
			validatorMap := iValidator.(map[string]interface{})
			if len(validatorMap) == 0 {
				// pass invalid or empty validator
				continue
			}
			// check priority: HRP > HttpRunner
			validator := Validator{}
			if len(validatorMap) == 4 {
				// HRP validator format
				validator.Check = validatorMap["check"].(string)
				validator.Assert = validatorMap["assert"].(string)
				validator.Expect = validatorMap["expect"]
				if msg, exist := validatorMap["msg"]; exist {
					validator.Message = msg.(string)
				}
				convertCompatHeader(&validator)
				step.Validators[i] = validator
			} else if len(validatorMap) == 1 {
				// HttpRunner validator format
				for assertMethod, iValidatorContent := range validatorMap {
					checkAndExpect := iValidatorContent.([]interface{})
					validator.Check = checkAndExpect[0].(string)
					validator.Assert = assertMethod
					validator.Expect = checkAndExpect[1]
				}
				convertCompatHeader(&validator)
				step.Validators[i] = validator
			} else {
				log.Error().Msgf("[convert compat testcase] unexpected validator format: %v", validatorMap)
			}
		}
	}
	return err
}

// convertCompatHeader deals with headers format in HttpRunner
// e.g. headers.Content-Type => headers.\"Content-Type\"
func convertCompatHeader(validator *Validator) {
	if strings.Contains(validator.Check, "headers.") &&
		!strings.Contains(validator.Check, "\"") &&
		strings.Contains(validator.Check, "-") {
		replacedHeader := fmt.Sprintf("headers.\"%s\"", validator.Check[len("headers."):])
		validator.Check = replacedHeader
	}
}

func (tc *TCase) ToTestCase() (*TestCase, error) {
	testCase := &TestCase{
		Config: tc.Config,
	}
	for _, step := range tc.TestSteps {
		if step.Request != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepRequestWithOptionalArgs{
				step: step,
			})
		} else if step.TestCase != nil {
			testCase.TestSteps = append(testCase.TestSteps, &StepTestCaseWithOptionalArgs{
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

// TestCasePath implements ITestCase interface.
type TestCasePath struct {
	Path string
}

func (path *TestCasePath) ToTestCase() (*TestCase, error) {
	var tc *TCase
	var err error

	casePath := path.Path
	ext := filepath.Ext(casePath)
	switch ext {
	case ".json":
		tc, err = loadFromJSON(casePath)
	case ".yaml", ".yml":
		tc, err = loadFromYAML(casePath)
	default:
		err = ErrUnsupportedFileExt
	}
	if err != nil {
		return nil, err
	}
	tc.Config.Path = path.Path
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
