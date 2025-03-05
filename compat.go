package hrp

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// ConvertCaseCompatibility converts TestCase compatible with Golang engine style
func ConvertCaseCompatibility(tc *TestCaseDef) (err error) {
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
		} else if step.Harmony != nil {
			convertCompatMobileStep(step.Harmony)
		}
	}
	return nil
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
		actionOptions := option.NewActionOptions(ma.GetOptions()...)
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
