package hrp

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/imroc/req"
	"github.com/jmespath/go-jmespath"
	log "github.com/sirupsen/logrus"

	"github.com/httprunner/hrp/builtin"
)

func NewResponseObject(t *testing.T, resp *req.Resp) (*ResponseObject, error) {
	// prepare response headers
	headers := make(map[string]string)
	for k, v := range resp.Response().Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// prepare response cookies
	cookies := make(map[string]string)
	for _, cookie := range resp.Response().Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	// parse response body
	var body interface{}
	if err := json.Unmarshal(resp.Bytes(), &body); err != nil {
		// response body is not json, use raw body
		body = string(resp.Bytes())
	}

	respObjMeta := respObjMeta{
		StatusCode: resp.Response().StatusCode,
		Headers:    headers,
		Cookies:    cookies,
		Body:       body,
	}

	// convert respObjMeta to interface{}
	respObjMetaBytes, _ := json.Marshal(respObjMeta)
	var data interface{}
	if err := json.Unmarshal(respObjMetaBytes, &data); err != nil {
		log.Errorf("[NewResponseObject] convert respObjMeta to interface{} error: %v, respObjMeta: %v",
			err, string(respObjMetaBytes))
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
		log.WithField("value", extractedValue).Infof("extract value from %s", value)
		log.WithField("value", extractedValue).Infof("set variable %s", key)
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
		log.WithFields(log.Fields{
			"assertMethod": assertMethod,
			"expectValue":  expectValue,
			"checkValue":   checkValue,
			"result":       result,
		}).Infof("validate %s", checkItem)
		if !result {
			v.t.Fail()
		}
	}
	return nil
}

func (v *ResponseObject) searchJmespath(expr string) interface{} {
	checkValue, err := jmespath.Search(expr, v.respObjMeta)
	if err != nil {
		log.Errorf("[searchJmespath] jmespath.Search error: %v", err)
		return expr // jmespath not found, return the expression
	}
	return checkValue
}
