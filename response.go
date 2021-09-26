package httpboomer

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/imroc/req"
	"github.com/jmespath/go-jmespath"
	"github.com/stretchr/testify/assert"
)

var assertFunctionsMap = map[string]func(t assert.TestingT, expected interface{}, actual interface{}, msgAndArgs ...interface{}) bool{
	"equals":            assert.EqualValues,
	"equal":             assert.EqualValues, // alias for equals
	"greater_than":      assert.Greater,
	"less_than":         assert.Less,
	"greater_or_equals": assert.GreaterOrEqual,
	"less_or_equals":    assert.LessOrEqual,
	"not_equal":         assert.NotEqual,
	"contains":          assert.Contains,
	"regex_match":       assert.Regexp,
}

func NewResponseObject(t *testing.T, resp *req.Resp) *ResponseObject {
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
		log.Fatalf("[NewResponseObject] json.Unmarshal response body err: %v, body: %v",
			err, string(resp.Bytes()))
		return nil
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
		log.Fatalf("[NewResponseObject] json.Unmarshal respObjMeta err: %v, respObjMetaBytes: %v",
			err, string(respObjMetaBytes))
		return nil
	}

	return &ResponseObject{
		t:           t,
		respObjMeta: data,
	}
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

func (v *ResponseObject) Validate(validators []TValidator) error {
	for _, validator := range validators {
		// parse check value
		checkItem := validator.Check
		checkValue := v.searchJmespath(checkItem)
		// get assert method
		assertMethod := validator.Assert
		assertFunc := assertFunctionsMap[assertMethod]
		// parse expected value
		expectValue := validator.Expect
		// do assertion
		result := assertFunc(v.t, expectValue, checkValue)
		log.Printf("assert %s %s %v => %v", checkItem, assertMethod, expectValue, result)
		if !result {
			v.t.Fail()
		}
	}
	return nil
}

func (v *ResponseObject) searchJmespath(expr string) interface{} {
	checkValue, err := jmespath.Search(expr, v.respObjMeta)
	if err != nil {
		log.Printf("[searchJmespath] jmespath.Search error: %v", err)
		return nil
	}
	return checkValue
}
