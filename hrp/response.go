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

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

var fieldTags = []string{"proto", "status_code", "headers", "cookies", "body", textExtractorSubRegexp}

type httpRespObjMeta struct {
	Proto      string            `json:"proto"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Cookies    map[string]string `json:"cookies"`
	Body       interface{}       `json:"body"`
}

func newHttpResponseObject(t *testing.T, parser *Parser, resp *http.Response) (*responseObject, error) {
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

	respObjMeta := httpRespObjMeta{
		Proto:      resp.Proto,
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Cookies:    cookies,
		Body:       body,
	}

	return convertToResponseObject(t, parser, respObjMeta)
}

type wsCloseRespObject struct {
	StatusCode int    `json:"status_code"`
	Text       string `json:"body"`
}

func newWsCloseResponseObject(t *testing.T, parser *Parser, resp *wsCloseRespObject) (*responseObject, error) {
	return convertToResponseObject(t, parser, resp)
}

type wsReadRespObject struct {
	Message     interface{} `json:"body"`
	messageType int
}

func newWsReadResponseObject(t *testing.T, parser *Parser, resp *wsReadRespObject) (*responseObject, error) {
	byteMessage, ok := resp.Message.([]byte)
	if !ok {
		return nil, errors.New("websocket message type should be []byte")
	}
	var msg interface{}
	if err := json.Unmarshal(byteMessage, &msg); err != nil {
		// response body is not json, use raw body
		msg = string(byteMessage)
	}
	resp.Message = msg
	return convertToResponseObject(t, parser, resp)
}

func convertToResponseObject(t *testing.T, parser *Parser, respObjMeta interface{}) (*responseObject, error) {
	respObjMetaBytes, _ := json.Marshal(respObjMeta)
	var data interface{}
	decoder := json.NewDecoder(bytes.NewReader(respObjMetaBytes))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		log.Error().
			Str("respObjectMeta", string(respObjMetaBytes)).
			Err(err).
			Msg("[convertToResponseObject] convert respObjectMeta to interface{} failed")
		return nil, err
	}
	return &responseObject{
		t:           t,
		parser:      parser,
		respObjMeta: data,
	}, nil
}

type responseObject struct {
	t                 *testing.T
	parser            *Parser
	respObjMeta       interface{}
	validationResults []*ValidationResult
}

const textExtractorSubRegexp string = `(.*)`

func (v *responseObject) searchField(value string) interface{} {
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
		extractedValue := v.searchField(value)
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
			// parse reference variables in checkItem
			checkValue, err = v.parser.Parse(checkItem, variablesMapping)
			if err != nil {
				return err
			}
		} else {
			checkValue = checkItem
		}
		if searchCheckValue, ok := checkValue.(string); ok && checkSearchField(searchCheckValue) {
			checkValue = v.searchField(searchCheckValue)
		}

		// get assert method
		assertMethod := validator.Assert
		assertFunc, ok := builtin.Assertions[assertMethod]
		if !ok {
			return errors.New(fmt.Sprintf("unexpected assertMethod: %v", assertMethod))
		}

		// parse expected value
		expectValue, err := v.parser.Parse(validator.Expect, variablesMapping)
		if err != nil {
			return err
		}
		validResult := &ValidationResult{
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
			Str("expectValueType", builtin.InterfaceType(expectValue)).
			Interface("checkValue", checkValue).
			Str("checkValueType", builtin.InterfaceType(checkValue)).
			Bool("result", result).
			Msgf("validate %s", checkItem)
		if !result {
			v.t.Fail()
			log.Error().
				Str("checkExpr", validator.Check).
				Str("assertMethod", assertMethod).
				Interface("checkValue", checkValue).
				Str("checkValueType", builtin.InterfaceType(checkValue)).
				Interface("expectValue", expectValue).
				Str("expectValueType", builtin.InterfaceType(expectValue)).
				Msg("assert failed")
			return errors.New("step validation failed")
		}
	}
	return nil
}

func checkSearchField(expr string) bool {
	for _, t := range fieldTags {
		if strings.Contains(expr, t) {
			return true
		}
	}
	return false
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
	if len(match) > 1 {
		return match[1] // return first matched result in parentheses
	}
	log.Error().Str("expr", expr).Msg("search regexp failed")
	return expr
}
