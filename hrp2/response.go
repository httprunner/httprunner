package hrp

import (
	"bytes"
	builtinJSON "encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/json"
)

func newResponseObject(t *testing.T, parser *parser, resp *http.Response) (*responseObject, error) {
	// prepare response headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// prepare response cookies
	cookies := make(map[string]string)
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	// read response body
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// parse response body
	var body interface{}
	if err := json.Unmarshal(respBodyBytes, &body); err != nil {
		// response body is not json, use raw body
		body = string(respBodyBytes)
	}

	respObjMeta := respObjMeta{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Cookies:    cookies,
		Body:       body,
	}

	// convert respObjMeta to interface{}
	respObjMetaBytes, _ := json.Marshal(respObjMeta)
	var data interface{}
	decoder := json.NewDecoder(bytes.NewReader(respObjMetaBytes))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		log.Error().
			Str("respObjMeta", string(respObjMetaBytes)).
			Err(err).
			Msg("[NewResponseObject] convert respObjMeta to interface{} failed")
		return nil, err
	}

	return &responseObject{
		t:           t,
		parser:      parser,
		respObjMeta: data,
	}, nil
}

type respObjMeta struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Cookies    map[string]string `json:"cookies"`
	Body       interface{}       `json:"body"`
}

type responseObject struct {
	t                 *testing.T
	parser            *parser
	respObjMeta       interface{}
	validationResults []*validationResult
}

const textExtractorSubRegexp string = `(.*)`

func (v *responseObject) extractField(value string) interface{} {
	var result interface{}
	if strings.Contains(value, textExtractorSubRegexp) {
		result = v.searchRegexp(value)
	} else {
		result = v.searchJmespath(value)
	}
	return result
}

func (v *responseObject) Extract(extractors map[string]string) map[string]interface{} {
	if extractors == nil {
		return nil
	}

	extractMapping := make(map[string]interface{})
	for key, value := range extractors {
		extractedValue := v.extractField(value)
		log.Info().Str("from", value).Interface("value", extractedValue).Msg("extract value")
		log.Info().Str("variable", key).Interface("value", extractedValue).Msg("set variable")
		extractMapping[key] = extractedValue
	}

	return extractMapping
}

func (v *responseObject) Validate(iValidators []interface{}, variablesMapping map[string]interface{}) (err error) {
	for _, iValidator := range iValidators {
		validator, ok := iValidator.(Validator)
		if !ok {
			return errors.New("validator type error")
		}
		// parse check value
		checkItem := validator.Check
		var checkValue interface{}
		if strings.Contains(checkItem, "$") {
			// reference variable
			checkValue, err = v.parser.parseData(checkItem, variablesMapping)
			if err != nil {
				return err
			}
		} else {
			// regExp or jmesPath
			checkValue = v.extractField(checkItem)
		}

		// get assert method
		assertMethod := validator.Assert
		assertFunc, ok := builtin.Assertions[assertMethod]
		if !ok {
			return errors.New(fmt.Sprintf("unexpected assertMethod: %v", assertMethod))
		}

		// parse expected value
		expectValue, err := v.parser.parseData(validator.Expect, variablesMapping)
		if err != nil {
			return err
		}
		validResult := &validationResult{
			Validator: Validator{
				Check:   validator.Check,
				Expect:  expectValue,
				Assert:  assertMethod,
				Message: validator.Message,
			},
			CheckValue:  checkValue,
			CheckResult: "fail",
		}

		// do assertion
		result := assertFunc(v.t, checkValue, expectValue)
		if result {
			validResult.CheckResult = "pass"
		}
		v.validationResults = append(v.validationResults, validResult)
		log.Info().
			Str("checkExpr", validator.Check).
			Str("assertMethod", assertMethod).
			Interface("expectValue", expectValue).
			Interface("checkValue", checkValue).
			Bool("result", result).
			Msgf("validate %s", checkItem)
		if !result {
			v.t.Fail()
			return errors.New(fmt.Sprintf(
				"do assertion failed, checkExpr: %v, assertMethod: %v, checkValue: %v, expectValue: %v",
				validator.Check,
				assertMethod,
				checkValue,
				expectValue,
			))
		}
	}
	return nil
}

func (v *responseObject) searchJmespath(expr string) interface{} {
	checkValue, err := jmespath.Search(expr, v.respObjMeta)
	if err != nil {
		log.Error().Str("expr", expr).Err(err).Msg("search jmespath failed")
		return expr // jmespath not found, return the expression
	}
	if number, ok := checkValue.(builtinJSON.Number); ok {
		checkNumber, err := parseJSONNumber(number)
		if err != nil {
			log.Error().Interface("json number", number).Err(err).Msg("convert json number failed")
		}
		return checkNumber
	}
	return checkValue
}

func (v *responseObject) searchRegexp(expr string) interface{} {
	respMap, ok := v.respObjMeta.(map[string]interface{})
	if !ok {
		log.Error().Interface("resp", v.respObjMeta).Msg("convert respObjMeta to map failed")
		return expr
	}
	bodyStr, ok := respMap["body"].(string)
	if !ok {
		log.Error().Interface("resp", respMap).Msg("convert body to string failed")
		return expr
	}
	regexpCompile, err := regexp.Compile(expr)
	if err != nil {
		log.Error().Str("expr", expr).Err(err).Msg("compile expr failed")
		return expr
	}
	match := regexpCompile.FindStringSubmatch(bodyStr)
	if match != nil || len(match) > 1 {
		return match[1] //return first matched result in parentheses
	}
	log.Error().Str("expr", expr).Msg("search regexp failed")
	return expr
}
