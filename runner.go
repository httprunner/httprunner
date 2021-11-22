package hrp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"

	"github.com/httprunner/hrp/internal/ga"
)

// run API test with default configs
func Run(testcases ...ITestCase) error {
	t := &testing.T{}
	return NewRunner(t).SetDebug(true).Run(testcases...)
}

func NewRunner(t *testing.T) *Runner {
	if t == nil {
		t = &testing.T{}
	}
	return &Runner{
		t:     t,
		debug: false, // default to turn off debug
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
		sessionVariables: make(map[string]interface{}),
	}
}

type Runner struct {
	t                *testing.T
	debug            bool
	client           *http.Client
	sessionVariables map[string]interface{}
}

func (r *Runner) SetDebug(debug bool) *Runner {
	log.Info().Bool("debug", debug).Msg("[init] SetDebug")
	r.debug = debug
	return r
}

func (r *Runner) SetProxyUrl(proxyUrl string) *Runner {
	log.Info().Str("proxyUrl", proxyUrl).Msg("[init] SetProxyUrl")
	p, err := url.Parse(proxyUrl)
	if err != nil {
		log.Error().Err(err).Str("proxyUrl", proxyUrl).Msg("[init] invalid proxyUrl")
		return r
	}
	r.client.Transport = &http.Transport{
		Proxy:           http.ProxyURL(p),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return r
}

func (r *Runner) Run(testcases ...ITestCase) error {
	event := ga.EventTracking{
		Category: "RunAPITests",
		Action:   "hrp run",
	}
	// report start event
	go ga.SendEvent(event)
	// report execution timing event
	defer ga.SendEvent(event.StartTiming("execution"))

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

	for _, step := range testcase.TestSteps {
		_, err := r.runStep(step, config)
		if err != nil {
			return err
		}
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return nil
}

func (r *Runner) runStep(step IStep, config *TConfig) (stepData *StepData, err error) {
	log.Info().Str("step", step.Name()).Msg("run step start")

	// copy step to avoid data racing
	copiedStep := &TStep{}
	if err = copier.Copy(copiedStep, step.ToStruct()); err != nil {
		log.Error().Err(err).Msg("copy step data failed")
		return
	}

	stepVariables := copiedStep.Variables
	// override variables
	// step variables > session variables (extracted variables from previous steps)
	stepVariables = mergeVariables(stepVariables, r.sessionVariables)
	// step variables > testcase config variables
	stepVariables = mergeVariables(stepVariables, config.Variables)

	// parse step variables
	parsedVariables, err := parseVariables(stepVariables)
	if err != nil {
		log.Error().Interface("variables", config.Variables).Err(err).Msg("parse step variables failed")
		return
	}
	copiedStep.Variables = parsedVariables // avoid data racing

	if _, ok := step.(*testcaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Info().Str("testcase", copiedStep.Name).Msg("run referenced testcase")
		// TODO: override testcase config
		stepData, err = r.runStepTestCase(copiedStep)
		if err != nil {
			log.Error().Err(err).Msg("run referenced testcase step failed")
			return
		}
	} else {
		// run request
		copiedStep.Request.URL = buildURL(config.BaseURL, copiedStep.Request.URL) // avoid data racing
		stepData, err = r.runStepRequest(copiedStep)
		if err != nil {
			log.Error().Err(err).Msg("run request step failed")
			return
		}
	}

	// update extracted variables
	for k, v := range stepData.ExportVars {
		r.sessionVariables[k] = v
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
		case map[string]interface{}:
			contentType := req.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				// post form data
				formData := make(url.Values)
				for k, v := range vv {
					formData.Add(k, fmt.Sprint(v))
				}
				dataBytes = []byte(formData.Encode())
			} else {
				// post json
				dataBytes, err = json.Marshal(vv)
				if err != nil {
					return nil, err
				}
				req.Header.Set("Content-Type", "application/json; charset=UTF-8")
			}
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
	req.Host = u.Host

	// log & print request
	if r.debug {
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, errors.Wrap(err, "dump request failed")
		}
		fmt.Println("-------------------- request --------------------")
		fmt.Println(string(reqDump))
	}

	// do request action
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request failed")
	}
	defer resp.Body.Close()

	// log & print response
	if r.debug {
		fmt.Println("==================== response ===================")
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, errors.Wrap(err, "dump response failed")
		}
		fmt.Println(string(respDump))
		fmt.Println("--------------------------------------------------")
	}

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
