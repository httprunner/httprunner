package hrp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/jmespath/go-jmespath"

	"github.com/httprunner/hrp/internal/builtin"
)

func NewResponseObject(t *testing.T, resp *http.Response) (*ResponseObject, error) {
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
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
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
	if err := json.Unmarshal(respObjMetaBytes, &data); err != nil {
		log.Error().
			Str("respObjMeta", string(respObjMetaBytes)).
			Err(err).
			Msg("[NewResponseObject] convert respObjMeta to interface{} failed")
		return nil, err
	}

	return &ResponseObject{
		t:           t,
		respObjMeta: data,
	}, nil
}

type respObjMeta struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Cookies    map[string]string `json:"cookies"`
	Body       interface{}       `json:"body"`
}

type ResponseObject struct {
	t                 *testing.T
	respObjMeta       interface{}
	validationResults map[string]interface{}
}

func (v *ResponseObject) Extract(extractors map[string]string) map[string]interface{} {
	if extractors == nil {
		return nil
	}

	extractMapping := make(map[string]interface{})
	for key, value := range extractors {
		extractedValue := v.searchJmespath(value)
		log.Info().Str("from", value).Interface("value", extractedValue).Msg("extract value")
		log.Info().Str("variable", key).Interface("value", extractedValue).Msg("set variable")
		extractMapping[key] = extractedValue
	}

	return extractMapping
}

func (v *ResponseObject) Validate(validators []TValidator, variablesMapping map[string]interface{}) (err error) {
	for _, validator := range validators {
		// parse check value
		checkItem := validator.Check
		var checkValue interface{}
		if strings.Contains(checkItem, "$") {
			// reference variable
			checkValue, err = parseData(checkItem, variablesMapping)
			if err != nil {
				return err
			}
		} else {
			checkValue = v.searchJmespath(checkItem)
		}

		// get assert method
		assertMethod := validator.Assert
		assertFunc := builtin.Assertions[assertMethod]

		// parse expected value
		expectValue, err := parseData(validator.Expect, variablesMapping)
		if err != nil {
			return err
		}

		// do assertion
		result := assertFunc(v.t, expectValue, checkValue)
		log.Info().
			Str("assertMethod", assertMethod).
			Interface("expectValue", expectValue).
			Interface("checkValue", checkValue).
			Bool("result", result).
			Msgf("validate %s", checkItem)
		if !result {
			v.t.Fail()
		}
	}
	return nil
}

func (v *ResponseObject) searchJmespath(expr string) interface{} {
	checkValue, err := jmespath.Search(expr, v.respObjMeta)
	if err != nil {
		log.Error().Str("expr", expr).Err(err).Msg("search jmespath failed")
		return expr // jmespath not found, return the expression
	}
	return checkValue
}
