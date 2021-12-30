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
	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/ga"
)

// Run starts to run API test with default configs.
func Run(testcases ...ITestCase) error {
	t := &testing.T{}
	return NewRunner(t).SetDebug(true).Run(testcases...)
}

// NewRunner constructs a new runner instance.
func NewRunner(t *testing.T) *HRPRunner {
	if t == nil {
		t = &testing.T{}
	}
	return &HRPRunner{
		t:        t,
		failfast: true,  // default to failfast
		debug:    false, // default to turn off debug
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
	}
}

type HRPRunner struct {
	t        *testing.T
	failfast bool
	debug    bool
	client   *http.Client
}

// SetFailfast configures whether to stop running when one step fails.
func (r *HRPRunner) SetFailfast(failfast bool) *HRPRunner {
	log.Info().Bool("failfast", failfast).Msg("[init] SetFailfast")
	r.failfast = failfast
	return r
}

// SetDebug configures whether to log HTTP request and response content.
func (r *HRPRunner) SetDebug(debug bool) *HRPRunner {
	log.Info().Bool("debug", debug).Msg("[init] SetDebug")
	r.debug = debug
	return r
}

// SetProxyUrl configures the proxy URL, which is usually used to capture HTTP packets for debugging.
func (r *HRPRunner) SetProxyUrl(proxyUrl string) *HRPRunner {
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

// Run starts to execute one or multiple testcases.
func (r *HRPRunner) Run(testcases ...ITestCase) error {
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
		if err := r.newCaseRunner(testcase).run(); err != nil {
			log.Error().Err(err).Msg("[Run] run testcase failed")
			return err
		}
	}
	return nil
}

func (r *HRPRunner) newCaseRunner(testcase *TestCase) *caseRunner {
	caseRunner := &caseRunner{
		TestCase:  testcase,
		hrpRunner: r,
	}
	caseRunner.reset()
	return caseRunner
}

// caseRunner is used to run testcase and its steps.
// each testcase has its own caseRunner instance and share session variables.
type caseRunner struct {
	*TestCase
	hrpRunner        *HRPRunner
	sessionVariables map[string]interface{}
	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[transactionType]time.Time
	startTime    time.Time // record start time of the testcase
}

// reset clears runner session variables.
func (r *caseRunner) reset() *caseRunner {
	log.Info().Msg("[init] Reset session variables")
	r.sessionVariables = make(map[string]interface{})
	r.transactions = make(map[string]map[transactionType]time.Time)
	r.startTime = time.Now()
	return r
}

func (r *caseRunner) run() error {
	config := r.TestCase.Config
	if err := r.parseConfig(config); err != nil {
		return err
	}

	log.Info().Str("testcase", config.Name()).Msg("run testcase start")
	r.startTime = time.Now()
	for index := range r.TestCase.TestSteps {
		_, err := r.runStep(index)
		if err != nil {
			if r.hrpRunner.failfast {
				return errors.Wrap(err, "abort running due to failfast setting")
			}
			log.Warn().Err(err).Msg("run step failed, continue next step")
		}
	}

	log.Info().Str("testcase", config.Name()).Msg("run testcase end")
	return nil
}

func (r *caseRunner) runStep(index int) (stepResult *stepData, err error) {
	config := r.TestCase.Config
	step := r.TestCase.TestSteps[index]

	// step type priority order: transaction > rendezvous > testcase > request
	if stepTran, ok := step.(*StepTransaction); ok {
		// transaction step
		return r.runStepTransaction(stepTran.step.Transaction)
	} else if stepRend, ok := step.(*StepRendezvous); ok {
		// rendezvous step
		return r.runStepRendezvous(stepRend.step.Rendezvous)
	}

	log.Info().Str("step", step.Name()).Msg("run step start")

	// copy step and config to avoid data racing
	copiedStep := &TStep{}
	if err = copier.Copy(copiedStep, step.ToStruct()); err != nil {
		log.Error().Err(err).Msg("copy step data failed")
		return nil, err
	}
	copiedConfig := &TConfig{}
	if err = copier.Copy(copiedConfig, config.ToStruct()); err != nil {
		log.Error().Err(err).Msg("copy config data failed")
		return nil, err
	}

	stepVariables := copiedStep.Variables
	// override variables
	// step variables > session variables (extracted variables from previous steps)
	stepVariables = mergeVariables(stepVariables, r.sessionVariables)
	// step variables > testcase config variables
	stepVariables = mergeVariables(stepVariables, copiedConfig.Variables)

	// parse step variables
	parsedVariables, err := parseVariables(stepVariables)
	if err != nil {
		log.Error().Interface("variables", copiedConfig.Variables).Err(err).Msg("parse step variables failed")
		return nil, err
	}
	copiedStep.Variables = parsedVariables // avoid data racing

	// step type priority order: testcase > request
	if _, ok := step.(*StepTestCaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Info().Str("testcase", copiedStep.Name).Msg("run referenced testcase")
		// TODO: override testcase config
		stepResult, err = r.runStepTestCase(copiedStep)
		if err != nil {
			log.Error().Err(err).Msg("run referenced testcase step failed")
			return
		}
	} else {
		// run request
		copiedStep.Request.URL = buildURL(copiedConfig.BaseURL, copiedStep.Request.URL) // avoid data racing
		stepResult, err = r.runStepRequest(copiedStep)
		if err != nil {
			log.Error().Err(err).Msg("run request step failed")
			return
		}
	}

	// update extracted variables
	for k, v := range stepResult.exportVars {
		r.sessionVariables[k] = v
	}

	log.Info().
		Str("step", step.Name()).
		Bool("success", stepResult.success).
		Interface("exportVars", stepResult.exportVars).
		Msg("run step end")
	return stepResult, nil
}

func (r *caseRunner) runStepTransaction(transaction *Transaction) (stepResult *stepData, err error) {
	log.Info().
		Str("name", transaction.Name).
		Str("type", string(transaction.Type)).
		Msg("transaction")

	stepResult = &stepData{
		name:        transaction.Name,
		stepType:    stepTypeTransaction,
		success:     true,
		elapsed:     0,
		contentSize: 0, // TODO: record transaction total response length
	}

	// create transaction if not exists
	if _, ok := r.transactions[transaction.Name]; !ok {
		r.transactions[transaction.Name] = make(map[transactionType]time.Time)
	}

	// record transaction start time, override if already exists
	if transaction.Type == transactionStart {
		r.transactions[transaction.Name][transactionStart] = time.Now()
	}
	// record transaction end time, override if already exists
	if transaction.Type == transactionEnd {
		r.transactions[transaction.Name][transactionEnd] = time.Now()

		// if transaction start time not exists, use testcase start time instead
		if _, ok := r.transactions[transaction.Name][transactionStart]; !ok {
			r.transactions[transaction.Name][transactionStart] = r.startTime
		}

		// calculate transaction duration
		duration := r.transactions[transaction.Name][transactionEnd].Sub(
			r.transactions[transaction.Name][transactionStart])
		stepResult.elapsed = duration.Milliseconds()
		log.Info().Str("name", transaction.Name).Dur("elapsed", duration).Msg("transaction")
	}

	return stepResult, nil
}

func (r *caseRunner) runStepRendezvous(rend *Rendezvous) (stepResult *stepData, err error) {
	log.Info().
		Str("name", rend.Name).
		Float32("percent", rend.Percent).
		Int64("number", rend.Number).
		Int64("timeout", rend.Timeout).
		Msg("rendezvous")
	stepResult = &stepData{
		name:     rend.Name,
		stepType: stepTypeRendezvous,
		success:  true,
	}
	return stepResult, nil
}

func (r *caseRunner) runStepRequest(step *TStep) (stepResult *stepData, err error) {
	stepResult = &stepData{
		name:        step.Name,
		stepType:    stepTypeRequest,
		success:     false,
		contentSize: 0,
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
	if r.hrpRunner.debug {
		reqDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, errors.Wrap(err, "dump request failed")
		}
		fmt.Println("-------------------- request --------------------")
		fmt.Println(string(reqDump))
	}

	// do request action
	start := time.Now()
	resp, err := r.hrpRunner.client.Do(req)
	stepResult.elapsed = time.Since(start).Milliseconds()
	if err != nil {
		return nil, errors.Wrap(err, "do request failed")
	}
	defer resp.Body.Close()

	// log & print response
	if r.hrpRunner.debug {
		fmt.Println("==================== response ===================")
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, errors.Wrap(err, "dump response failed")
		}
		fmt.Println(string(respDump))
		fmt.Println("--------------------------------------------------")
	}

	// new response object
	respObj, err := newResponseObject(r.hrpRunner.t, resp)
	if err != nil {
		err = errors.Wrap(err, "init ResponseObject error")
		return
	}

	// extract variables from response
	extractors := step.Extract
	extractMapping := respObj.Extract(extractors)
	stepResult.exportVars = extractMapping

	// override step variables with extracted variables
	stepVariables := mergeVariables(step.Variables, extractMapping)

	// validate response
	err = respObj.Validate(step.Validators, stepVariables)
	if err != nil {
		return
	}

	stepResult.success = true
	stepResult.contentSize = resp.ContentLength
	return stepResult, nil
}

func (r *caseRunner) runStepTestCase(step *TStep) (stepResult *stepData, err error) {
	stepResult = &stepData{
		name:     step.Name,
		stepType: stepTypeTestCase,
		success:  false,
	}
	testcase := step.TestCase
	start := time.Now()
	err = r.hrpRunner.newCaseRunner(testcase).run()
	stepResult.elapsed = time.Since(start).Milliseconds()
	if err != nil {
		return stepResult, err
	}
	stepResult.success = true
	return stepResult, nil
}

func (r *caseRunner) parseConfig(config IConfig) error {
	cfg := config.ToStruct()
	// parse config variables
	parsedVariables, err := parseVariables(cfg.Variables)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	cfg.Variables = parsedVariables

	// parse config name
	parsedName, err := parseString(cfg.Name, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := parseString(cfg.BaseURL, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.BaseURL = convertString(parsedBaseURL)

	return nil
}

func (r *caseRunner) getSummary() *testCaseSummary {
	return &testCaseSummary{}
}

func setBodyBytes(req *http.Request, data []byte) {
	req.Body = ioutil.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
}
