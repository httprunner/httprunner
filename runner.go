package hrp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// run API test with default configs
func Run(t *testing.T, testcases ...ITestCase) error {
	return NewRunner().WithTestingT(t).SetDebug(true).Run(testcases...)
}

func NewRunner() *Runner {
	return &Runner{
		t:      &testing.T{},
		debug:  false, // default to turn off debug
		client: &http.Client{},
	}
}

type Runner struct {
	t      *testing.T
	debug  bool
	client *http.Client
}

func (r *Runner) WithTestingT(t *testing.T) *Runner {
	log.Info().Msg("[init] WithTestingT")
	r.t = t
	return r
}

func (r *Runner) SetDebug(debug bool) *Runner {
	log.Info().Bool("debug", debug).Msg("[init] SetDebug")
	r.debug = debug
	return r
}

func (r *Runner) SetProxyUrl(proxyUrl string) *Runner {
	log.Info().Str("proxyUrl", proxyUrl).Msg("[init] SetProxyUrl")
	// TODO
	return r
}

func (r *Runner) Run(testcases ...ITestCase) error {
	for _, iTestCase := range testcases {
		testcase, err := iTestCase.ToTestCase()
		if err != nil {
			log.Error().Err(err).Msg("[Run] convert ITestCase interface to TestCase struct failed")
			return err
		}
		if err := r.runCase(testcase); err != nil {
			log.Error().Err(err).Msg("[Run] run testcase failed")
			return err
		}
	}
	return nil
}

func (r *Runner) runCase(testcase *TestCase) error {
	config := &testcase.Config
	if err := r.parseConfig(config); err != nil {
		return err
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	extractedVariables := make(map[string]interface{})

	for _, step := range testcase.TestSteps {
		// override variables
		// step variables > extracted variables from previous steps
		stepVariables := mergeVariables(step.ToStruct().Variables, extractedVariables)
		// step variables > testcase config variables
		stepVariables = mergeVariables(stepVariables, config.Variables)

		// parse step variables
		parsedVariables, err := parseVariables(stepVariables)
		if err != nil {
			log.Error().Interface("variables", config.Variables).Err(err).Msg("parse step variables failed")
			return err
		}
		step.ToStruct().Variables = parsedVariables

		stepData, err := r.runStep(step, config)
		if err != nil {
			return err
		}
		// update extracted variables
		for k, v := range stepData.ExportVars {
			extractedVariables[k] = v
		}
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return nil
}

func (r *Runner) runStep(step IStep, config *TConfig) (stepData *StepData, err error) {
	log.Info().Str("step", step.Name()).Msg("run step start")
	if tc, ok := step.(*testcaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Info().Str("testcase", tc.step.Name).Msg("run referenced testcase")
		// TODO: override testcase config
		stepData, err = r.runStepTestCase(tc.step)
		if err != nil {
			log.Error().Err(err).Msg("run referenced testcase step failed")
			return
		}
	} else {
		// run request
		tStep := parseStep(step, config)
		stepData, err = r.runStepRequest(tStep)
		if err != nil {
			log.Error().Err(err).Msg("run request step failed")
			return
		}
	}
	log.Info().
		Str("step", step.Name()).
		Bool("success", stepData.Success).
		Interface("exportVars", stepData.ExportVars).
		Msg("run step end")
	return
}

func (r *Runner) runStepRequest(step *TStep) (stepData *StepData, err error) {
	stepData = &StepData{
		Name:           step.Name,
		Success:        false,
		ResponseLength: 0,
	}

	rawUrl := step.Request.URL
	method := step.Request.Method
	req := &http.Request{
		Method:     string(method),
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	// prepare request headers
	if len(step.Request.Headers) > 0 {
		headers, err := parseHeaders(step.Request.Headers, step.Variables)
		if err != nil {
			return nil, errors.Wrap(err, "parse headers failed")
		}
		for key, value := range headers {
			req.Header.Add(key, value)
		}
	}
	if length := req.Header.Get("Content-Length"); length != "" {
		if l, err := strconv.ParseInt(length, 10, 64); err == nil {
			req.ContentLength = l
		}
	}
	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	// prepare request params
	var queryParams url.Values
	if len(step.Request.Params) > 0 {
		params, err := parseData(step.Request.Params, step.Variables)
		if err != nil {
			return nil, errors.Wrap(err, "parse data failed")
		}
		parsedParams := params.(map[string]interface{})
		if len(parsedParams) > 0 {
			queryParams = make(url.Values)
			for k, v := range parsedParams {
				queryParams.Add(k, fmt.Sprint(v))
			}
		}
	}
	if queryParams != nil {
		// append params to url
		paramStr := queryParams.Encode()
		if strings.IndexByte(rawUrl, '?') == -1 {
			rawUrl = rawUrl + "?" + paramStr
		} else {
			rawUrl = rawUrl + "&" + paramStr
		}
	}

	// prepare request cookies
	for cookieName, cookieValue := range step.Request.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		})
	}

	// prepare request body
	if step.Request.Body != nil {
		data, err := parseData(step.Request.Body, step.Variables)
		if err != nil {
			return nil, err
		}
		var dataBytes []byte
		switch vv := data.(type) {
		case map[string]interface{}: // post json
			dataBytes, err = json.Marshal(vv)
			if err != nil {
				return nil, err
			}
			setContentType(req, "application/json; charset=UTF-8")
		case string:
			dataBytes = []byte(vv)
		case []byte:
			dataBytes = vv
		case bytes.Buffer:
			dataBytes = vv.Bytes()
		default: // unexpected body type
			return nil, errors.New("unexpected request body type")
		}
		setBodyBytes(req, dataBytes)
	}

	// prepare url
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, errors.Wrap(err, "parse url failed")
	}
	req.URL = u

	// do request action
	// req.Debug = r.debug
	// resp, err := r.client.Do(string(step.Request.Method), step.Request.URL, v...)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request failed")
	}
	defer resp.Body.Close()

	// new response object
	respObj, err := NewResponseObject(r.t, resp)
	if err != nil {
		err = errors.Wrap(err, "init ResponseObject error")
		return
	}

	// extract variables from response
	extractors := step.Extract
	extractMapping := respObj.Extract(extractors)
	stepData.ExportVars = extractMapping

	// override step variables with extracted variables
	stepVariables := mergeVariables(step.Variables, extractMapping)

	// validate response
	err = respObj.Validate(step.Validators, stepVariables)
	if err != nil {
		return
	}

	stepData.Success = true
	stepData.ResponseLength = resp.ContentLength
	return
}

func (r *Runner) runStepTestCase(step *TStep) (stepData *StepData, err error) {
	stepData = &StepData{
		Name:    step.Name,
		Success: false,
	}
	testcase := step.TestCase
	err = r.runCase(testcase)
	return
}

func (r *Runner) parseConfig(config *TConfig) error {
	// parse config variables
	parsedVariables, err := parseVariables(config.Variables)
	if err != nil {
		log.Error().Interface("variables", config.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	config.Variables = parsedVariables

	// parse config name
	parsedName, err := parseString(config.Name, config.Variables)
	if err != nil {
		return err
	}
	config.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := parseString(config.BaseURL, config.Variables)
	if err != nil {
		return err
	}
	config.BaseURL = convertString(parsedBaseURL)

	return nil
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}

func setBodyBytes(req *http.Request, data []byte) {
	req.Body = ioutil.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
}

func setContentType(req *http.Request, contentType string) {
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
}
