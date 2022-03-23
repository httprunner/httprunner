package hrp

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/builtin"
	"github.com/httprunner/hrp/internal/ga"
	"github.com/httprunner/hrp/internal/json"
)

const (
	summaryPath string = "reports/summary-%v.json"
	reportPath  string = "reports/report-%v.html"
)

// Run starts to run API test with default configs.
func Run(testcases ...ITestCase) error {
	t := &testing.T{}
	return NewRunner(t).SetRequestsLogOn().Run(testcases...)
}

// NewRunner constructs a new runner instance.
func NewRunner(t *testing.T) *HRPRunner {
	if t == nil {
		t = &testing.T{}
	}
	return &HRPRunner{
		t:             t,
		failfast:      true, // default to failfast
		genHTMLReport: false,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 30 * time.Second,
		},
	}
}

type HRPRunner struct {
	t             *testing.T
	failfast      bool
	requestsLogOn bool
	pluginLogOn   bool
	saveTests     bool
	genHTMLReport bool
	client        *http.Client
}

// SetClientTransport configures transport of http client for high concurrency load testing
func (r *HRPRunner) SetClientTransport(maxConns int, disableKeepAlive bool, disableCompression bool) *HRPRunner {
	log.Info().Int("maxConns", maxConns).Msg("[init] SetClientTransport")
	r.client.Transport = &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DialContext:         (&net.Dialer{}).DialContext,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: maxConns,
		DisableKeepAlives:   disableKeepAlive,
		DisableCompression:  disableCompression,
	}
	return r
}

// SetFailfast configures whether to stop running when one step fails.
func (r *HRPRunner) SetFailfast(failfast bool) *HRPRunner {
	log.Info().Bool("failfast", failfast).Msg("[init] SetFailfast")
	r.failfast = failfast
	return r
}

// SetRequestsLogOn turns on request & response details logging.
func (r *HRPRunner) SetRequestsLogOn() *HRPRunner {
	log.Info().Msg("[init] SetRequestsLogOn")
	r.requestsLogOn = true
	return r
}

// SetPluginLogOn turns on plugin logging.
func (r *HRPRunner) SetPluginLogOn() *HRPRunner {
	log.Info().Msg("[init] SetPluginLogOn")
	r.pluginLogOn = true
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

// SetSaveTests configures whether to save summary of tests.
func (r *HRPRunner) SetSaveTests(saveTests bool) *HRPRunner {
	log.Info().Bool("saveTests", saveTests).Msg("[init] SetSaveTests")
	r.saveTests = saveTests
	return r
}

// GenHTMLReport configures whether to gen html report of api tests.
func (r *HRPRunner) GenHTMLReport() *HRPRunner {
	log.Info().Bool("genHTMLReport", true).Msg("[init] SetgenHTMLReport")
	r.genHTMLReport = true
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
	// record execution data to summary
	s := newOutSummary()
	for _, iTestCase := range testcases {
		testcase, err := iTestCase.ToTestCase()
		if err != nil {
			log.Error().Err(err).Msg("[Run] convert ITestCase interface to TestCase struct failed")
			return err
		}
		cfg := testcase.Config
		// parse config parameters
		err = initParameterIterator(cfg, "runner")
		if err != nil {
			log.Error().Interface("parameters", cfg.Parameters).Err(err).Msg("parse config parameters failed")
			return err
		}
		// 在runner模式下，指定整体策略，cfg.ParametersSetting.Iterators仅包含一个CartesianProduct的迭代器
		for it := cfg.ParametersSetting.Iterators[0]; it.HasNext(); {
			// iterate through all parameter iterators and update case variables
			for _, it := range cfg.ParametersSetting.Iterators {
				if it.HasNext() {
					cfg.Variables = mergeVariables(it.Next(), cfg.Variables)
				}
			}
			caseRunnerObj := r.newCaseRunner(testcase)
			if err = caseRunnerObj.run(); err != nil {
				log.Error().Err(err).Msg("[Run] run testcase failed")
				return err
			}
			caseSummary := caseRunnerObj.getSummary()
			s.appendCaseSummary(caseSummary)
		}
	}
	s.Time.Duration = time.Since(s.Time.StartAt).Seconds()
	// save summary
	if r.saveTests {
		dir, _ := filepath.Split(summaryPath)
		err := builtin.EnsureFolderExists(dir)
		if err != nil {
			return err
		}
		err = builtin.Dump2JSON(s, fmt.Sprintf(summaryPath, s.Time.StartAt.Unix()))
		if err != nil {
			return err
		}
	}
	// generate HTML report
	if r.genHTMLReport {
		err := s.genHTMLReport()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *HRPRunner) newCaseRunner(testcase *TestCase) *caseRunner {
	caseRunner := &caseRunner{
		TestCase:  testcase,
		hrpRunner: r,
		parser:    newParser(),
		summary:   newSummary(),
	}
	caseRunner.reset()
	return caseRunner
}

// caseRunner is used to run testcase and its steps.
// each testcase has its own caseRunner instance and share session variables.
type caseRunner struct {
	*TestCase
	hrpRunner        *HRPRunner
	parser           *parser
	sessionVariables map[string]interface{}
	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[transactionType]time.Time
	startTime    time.Time        // record start time of the testcase
	summary      *testCaseSummary // record test case summary
}

// reset clears runner session variables.
func (r *caseRunner) reset() *caseRunner {
	log.Info().Msg("[init] Reset session variables")
	r.sessionVariables = make(map[string]interface{})
	r.transactions = make(map[string]map[transactionType]time.Time)
	r.startTime = time.Now()
	r.summary.Name = r.Config.Name
	return r
}

func (r *caseRunner) run() error {
	config := r.TestCase.Config
	// init plugin
	var err error
	if r.parser.plugin, err = initPlugin(config.Path, r.hrpRunner.pluginLogOn); err != nil {
		return err
	}
	defer func() {
		if r.parser.plugin != nil {
			r.parser.plugin.Quit()
		}
	}()
	if err := r.parseConfig(config); err != nil {
		return err
	}
	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	r.startTime = time.Now()
	for index := range r.TestCase.TestSteps {
		stepDataObj, err := r.runStep(index, config)
		if stepDataObj == nil {
			stepDataObj = &stepData{
				Name:    r.TestCase.TestSteps[index].Name(),
				Success: false,
			}
		}
		if stepDataObj.StepType == stepTypeTestCase {
			// merge test case if the step is test case
			summary, ok := stepDataObj.Data.(*testCaseSummary)
			if ok {
				r.summary.Records = append(r.summary.Records, summary.Records...)
				r.summary.Stat.Total += summary.Stat.Total
				r.summary.Stat.Successes += summary.Stat.Successes
				r.summary.Stat.Failures += summary.Stat.Failures
			}
		} else if stepDataObj.StepType == stepTypeRequest {
			// only record that the test step is the request step
			r.summary.Records = append(r.summary.Records, stepDataObj)
			r.summary.Stat.Total += 1
			if stepDataObj.Success {
				r.summary.Stat.Successes += 1
			} else {
				r.summary.Stat.Failures += 1
			}
		}
		r.summary.Success = r.summary.Success && stepDataObj.Success
		if err != nil {
			stepDataObj.Attachment = err.Error()
			if r.hrpRunner.failfast {
				return errors.Wrap(err, "abort running due to failfast setting")
			}
		}
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return nil
}

func (r *caseRunner) runStep(index int, caseConfig *TConfig) (stepResult *stepData, err error) {
	step := r.TestCase.TestSteps[index]

	// step type priority order: transaction > rendezvous > thinktime > testcase > request
	if stepTran, ok := step.(*StepTransaction); ok {
		// transaction step
		return r.runStepTransaction(stepTran.step.Transaction)
	} else if stepRend, ok := step.(*StepRendezvous); ok {
		// rendezvous step
		return r.runStepRendezvous(stepRend.step.Rendezvous)
	} else if stepThink, ok := step.(*StepThinkTime); ok {
		// think time step
		return r.runStepThinkTime(stepThink.step, caseConfig.ThinkTime)
	}

	log.Info().Str("step", step.Name()).Msg("run step start")

	// copy step and config to avoid data racing
	copiedStep := &TStep{}
	if err = copier.Copy(copiedStep, step.ToStruct()); err != nil {
		log.Error().Err(err).Msg("copy step data failed")
		return nil, err
	}

	stepVariables := copiedStep.Variables
	// override variables
	// step variables > session variables (extracted variables from previous steps)
	stepVariables = mergeVariables(stepVariables, r.sessionVariables)
	// step variables > testcase config variables
	stepVariables = mergeVariables(stepVariables, caseConfig.Variables)

	// parse step variables
	parsedVariables, err := r.parser.parseVariables(stepVariables)
	if err != nil {
		log.Error().Interface("variables", caseConfig.Variables).Err(err).Msg("parse step variables failed")
		return nil, err
	}
	copiedStep.Variables = parsedVariables // avoid data racing

	// step type priority order: testcase > request
	if _, ok := step.(*StepTestCaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Info().Str("testcase", copiedStep.Name).Msg("run referenced testcase")
		stepResult, err = r.runStepTestCase(copiedStep)
		if err != nil {
			log.Error().Err(err).Msg("run referenced testcase step failed")
		}
	} else {
		if _, ok := step.(*StepAPIWithOptionalArgs); ok {
			// run referenced API
			log.Info().Str("api", copiedStep.Name).Msg("run referenced api")
			api, _ := copiedStep.APIContent.ToAPI()
			extendWithAPI(copiedStep, api)
		}
		// override headers
		if caseConfig.Headers != nil {
			copiedStep.Request.Headers = mergeMap(copiedStep.Request.Headers, caseConfig.Headers)
		}
		// parse step request url
		var requestUrl interface{}
		requestUrl, err = r.parser.parseString(copiedStep.Request.URL, copiedStep.Variables)
		if err != nil {
			log.Error().Err(err).Msg("parse request url failed")
			requestUrl = copiedStep.Variables
		}
		copiedStep.Request.URL = buildURL(caseConfig.BaseURL, convertString(requestUrl)) // avoid data racing
		// run request
		stepResult, err = r.runStepRequest(copiedStep)
		if err != nil {
			log.Error().Err(err).Msg("run request step failed")
		}
	}

	// update extracted variables
	for k, v := range stepResult.ExportVars {
		r.sessionVariables[k] = v
	}

	log.Info().
		Str("step", step.Name()).
		Bool("success", stepResult.Success).
		Interface("exportVars", stepResult.ExportVars).
		Msg("run step end")
	return stepResult, err
}

func (r *caseRunner) runStepThinkTime(step *TStep, ttc *ThinkTimeConfig) (stepResult *stepData, err error) {
	thinkTime := step.ThinkTime
	log.Info().
		Str("name", step.Name).
		Float64("time", thinkTime.Time).
		Msg("think time")
	stepResult = &stepData{
		Name:     step.Name,
		StepType: stepTypeThinkTime,
		Success:  true,
	}
	if ttc == nil {
		ttc = &ThinkTimeConfig{thinkTimeDefault, nil, 0}
	}
	var tt time.Duration
	switch ttc.Strategy {
	case thinkTimeDefault:
		tt = time.Duration(thinkTime.Time*1000) * time.Millisecond
	case thinkTimeRandomPercentage:
		m, ok := ttc.Setting.(map[string]float64) // e.g. {"min_percentage": 0.5, "max_percentage": 1.5}
		if !ok {
			tt = time.Duration(thinkTime.Time*1000) * time.Millisecond
			break
		}
		res := builtin.GetRandomNumber(int(thinkTime.Time*m["min_percentage"]*1000), int(thinkTime.Time*m["max_percentage"]*1000))
		tt = time.Duration(res) * time.Millisecond
	case thinkTimeMultiply:
		value, ok := ttc.Setting.(float64) // e.g. 0.5
		if !ok || value <= 0 {
			value = thinkTimeDefaultMultiply
		}
		tt = time.Duration(thinkTime.Time*value*1000) * time.Millisecond
	case thinkTimeIgnore:
		// nothing to do
	}
	// no more than limit
	if ttc.Limit > 0 {
		limit := time.Duration(ttc.Limit*1000) * time.Millisecond
		if limit < tt {
			tt = limit
		}
	}
	time.Sleep(tt)
	return stepResult, nil
}

func (r *caseRunner) runStepTransaction(transaction *Transaction) (stepResult *stepData, err error) {
	log.Info().
		Str("name", transaction.Name).
		Str("type", string(transaction.Type)).
		Msg("transaction")

	stepResult = &stepData{
		Name:        transaction.Name,
		StepType:    stepTypeTransaction,
		Success:     true,
		Elapsed:     0,
		ContentSize: 0, // TODO: record transaction total response length
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
		stepResult.Elapsed = duration.Milliseconds()
		log.Info().Str("name", transaction.Name).Dur("elapsed", duration).Msg("transaction")
	}

	return stepResult, nil
}

func (r *caseRunner) runStepRendezvous(rendezvous *Rendezvous) (stepResult *stepData, err error) {
	log.Info().
		Str("name", rendezvous.Name).
		Float32("percent", rendezvous.Percent).
		Int64("number", rendezvous.Number).
		Int64("timeout", rendezvous.Timeout).
		Msg("rendezvous")
	stepResult = &stepData{
		Name:     rendezvous.Name,
		StepType: stepTypeRendezvous,
		Success:  true,
	}

	// pass current rendezvous if already released, activate rendezvous sequentially after spawn done
	if rendezvous.isReleased() || !r.isPreRendezvousAllReleased(rendezvous) || !rendezvous.isSpawnDone() {
		return stepResult, nil
	}

	// activate the rendezvous only once during each cycle
	rendezvous.once.Do(func() {
		close(rendezvous.activateChan)
	})

	// check current cnt using double check lock before updating to avoid negative WaitGroup counter
	if atomic.LoadInt64(&rendezvous.cnt) < rendezvous.Number {
		rendezvous.lock.Lock()
		if atomic.LoadInt64(&rendezvous.cnt) < rendezvous.Number {
			atomic.AddInt64(&rendezvous.cnt, 1)
			rendezvous.wg.Done()
			rendezvous.timerResetChan <- struct{}{}
		}
		rendezvous.lock.Unlock()
	}

	// block until current rendezvous released
	<-rendezvous.releaseChan
	return stepResult, nil
}

func (r *caseRunner) isPreRendezvousAllReleased(rendezvous *Rendezvous) bool {
	tCase, _ := r.ToTCase()
	for _, step := range tCase.TestSteps {
		preRendezvous := step.Rendezvous
		if preRendezvous == nil {
			continue
		}
		// meet current rendezvous, all previous rendezvous released, return true
		if preRendezvous == rendezvous {
			return true
		}
		if !preRendezvous.isReleased() {
			return false
		}
	}
	return true
}

func (r *Rendezvous) reset() {
	r.cnt = 0
	r.releasedFlag = 0
	r.wg.Add(int(r.Number))
	// timerResetChan channel will not be closed, thus init only once
	if r.timerResetChan == nil {
		r.timerResetChan = make(chan struct{})
	}
	r.activateChan = make(chan struct{})
	r.releaseChan = make(chan struct{})
	r.once = new(sync.Once)
}

func (r *Rendezvous) isSpawnDone() bool {
	return atomic.LoadUint32(&r.spawnDoneFlag) == 1
}

func (r *Rendezvous) setSpawnDone() {
	atomic.StoreUint32(&r.spawnDoneFlag, 1)
}

func (r *Rendezvous) isReleased() bool {
	return atomic.LoadUint32(&r.releasedFlag) == 1
}

func (r *Rendezvous) setReleased() {
	atomic.StoreUint32(&r.releasedFlag, 1)
}

func initRendezvous(testcase *TestCase, total int64) []*Rendezvous {
	tCase, _ := testcase.ToTCase()
	var rendezvousList []*Rendezvous
	for _, step := range tCase.TestSteps {
		if step.Rendezvous == nil {
			continue
		}
		rendezvous := step.Rendezvous

		// either number or percent should be correctly put, otherwise set to default (total)
		if rendezvous.Number == 0 && rendezvous.Percent > 0 && rendezvous.Percent <= defaultRendezvousPercent {
			rendezvous.Number = int64(rendezvous.Percent * float32(total))
		} else if rendezvous.Number > 0 && rendezvous.Number <= total && rendezvous.Percent == 0 {
			rendezvous.Percent = float32(rendezvous.Number) / float32(total)
		} else {
			log.Warn().
				Str("name", rendezvous.Name).
				Int64("default number", total).
				Float32("default percent", defaultRendezvousPercent).
				Msg("rendezvous parameter not defined or error, set to default value")
			rendezvous.Number = total
			rendezvous.Percent = defaultRendezvousPercent
		}

		if rendezvous.Timeout <= 0 {
			rendezvous.Timeout = defaultRendezvousTimeout
		}

		rendezvous.reset()
		rendezvousList = append(rendezvousList, rendezvous)
	}
	return rendezvousList
}

func waitRendezvous(rendezvousList []*Rendezvous) {
	if rendezvousList != nil {
		lastRendezvous := rendezvousList[len(rendezvousList)-1]
		for _, rendezvous := range rendezvousList {
			go waitSingleRendezvous(rendezvous, rendezvousList, lastRendezvous)
		}
	}
}

func waitSingleRendezvous(rendezvous *Rendezvous, rendezvousList []*Rendezvous, lastRendezvous *Rendezvous) {
	for {
		// cycle start: block current checking until current rendezvous activated
		<-rendezvous.activateChan
		stop := make(chan struct{})
		timeout := time.Duration(rendezvous.Timeout) * time.Millisecond
		timer := time.NewTimer(timeout)
		go func() {
			defer close(stop)
			rendezvous.wg.Wait()
		}()
		for !rendezvous.isReleased() {
			select {
			case <-rendezvous.timerResetChan:
				timer.Reset(timeout)
			case <-stop:
				rendezvous.setReleased()
				close(rendezvous.releaseChan)
				log.Info().
					Str("name", rendezvous.Name).
					Float32("percent", rendezvous.Percent).
					Int64("number", rendezvous.Number).
					Int64("timeout(ms)", rendezvous.Timeout).
					Int64("cnt", rendezvous.cnt).
					Str("reason", "rendezvous release condition satisfied").
					Msg("rendezvous released")
			case <-timer.C:
				rendezvous.setReleased()
				close(rendezvous.releaseChan)
				log.Info().
					Str("name", rendezvous.Name).
					Float32("percent", rendezvous.Percent).
					Int64("number", rendezvous.Number).
					Int64("timeout(ms)", rendezvous.Timeout).
					Int64("cnt", rendezvous.cnt).
					Str("reason", "time's up").
					Msg("rendezvous released")
			}
		}
		// cycle end: reset all previous rendezvous after last rendezvous released
		// otherwise, block current checker until the last rendezvous end
		if rendezvous == lastRendezvous {
			for _, r := range rendezvousList {
				r.reset()
			}
		} else {
			<-lastRendezvous.releaseChan
		}
	}
}

func (r *caseRunner) runStepRequest(step *TStep) (stepResult *stepData, err error) {
	stepResult = &stepData{
		Name:        step.Name,
		StepType:    stepTypeRequest,
		Success:     false,
		ContentSize: 0,
	}
	sessionData := newSessionData()

	// convert request struct to map
	jsonRequest, _ := json.Marshal(&step.Request)
	var requestMap map[string]interface{}
	_ = json.Unmarshal(jsonRequest, &requestMap)

	rawUrl := step.Request.URL
	method := step.Request.Method
	req := &http.Request{
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	// prepare request headers
	if len(step.Request.Headers) > 0 {
		headers, err := r.parser.parseHeaders(step.Request.Headers, step.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "parse headers failed")
		}
		for key, value := range headers {
			// omit pseudo header names for HTTP/1, e.g. :authority, :method, :path, :scheme
			if strings.HasPrefix(key, ":") {
				continue
			}
			req.Header.Add(key, value)

			// prepare content length
			if strings.EqualFold(key, "Content-Length") && value != "" {
				if l, err := strconv.ParseInt(value, 10, 64); err == nil {
					req.ContentLength = l
				}
			}
		}
	}

	// prepare request params
	var queryParams url.Values
	if len(step.Request.Params) > 0 {
		params, err := r.parser.parseData(step.Request.Params, step.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "parse request params failed")
		}
		parsedParams := params.(map[string]interface{})
		requestMap["params"] = parsedParams
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
		value, err := r.parser.parseData(cookieValue, step.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "parse cookie value failed")
		}
		req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: fmt.Sprintf("%v", value),
		})
	}

	// prepare request body
	if step.Request.Body != nil {
		data, err := r.parser.parseData(step.Request.Body, step.Variables)
		if err != nil {
			return stepResult, err
		}
		// check request body format if Content-Type specified as application/json
		if strings.HasPrefix(req.Header.Get("Content-Type"), "application/json") {
			switch data.(type) {
			case bool, float64, string, map[string]interface{}, []interface{}, nil:
				break
			default:
				return stepResult, errors.Errorf("request body type inconsistent with Content-Type: %v", req.Header.Get("Content-Type"))
			}
		}
		requestMap["body"] = data
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
					return stepResult, err
				}
				if contentType == "" {
					req.Header.Set("Content-Type", "application/json; charset=utf-8")
				}
			}
		case []interface{}:
			contentType := req.Header.Get("Content-Type")
			// post json
			dataBytes, err = json.Marshal(vv)
			if err != nil {
				return stepResult, err
			}
			if contentType == "" {
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
			}
		case string:
			dataBytes = []byte(vv)
		case []byte:
			dataBytes = vv
		case bytes.Buffer:
			dataBytes = vv.Bytes()
		default: // unexpected body type
			return stepResult, errors.New("unexpected request body type")
		}
		setBodyBytes(req, dataBytes)
	}
	// update header
	headers := make(map[string]string)
	for key, value := range req.Header {
		headers[key] = value[0]
	}
	requestMap["headers"] = headers

	// prepare url
	u, err := url.Parse(rawUrl)
	if err != nil {
		return stepResult, errors.Wrap(err, "parse url failed")
	}
	req.URL = u
	req.Host = u.Host

	// add request object to step variables, could be used in setup hooks
	step.Variables["hrp_step_name"] = step.Name
	step.Variables["hrp_step_request"] = requestMap

	// deal with setup hooks
	for _, setupHook := range step.SetupHooks {
		_, err = r.parser.parseData(setupHook, step.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run setup hooks failed")
		}
	}

	// log & print request
	if err := r.printRequest(req); err != nil {
		return stepResult, err
	}

	// do request action
	start := time.Now()
	resp, err := r.hrpRunner.client.Do(req)
	stepResult.Elapsed = time.Since(start).Milliseconds()
	if err != nil {
		return stepResult, errors.Wrap(err, "do request failed")
	}
	defer resp.Body.Close()

	// decode response body in br/gzip/deflate formats
	err = decodeResponseBody(resp)
	if err != nil {
		return stepResult, errors.Wrap(err, "decode response body failed")
	}

	// log & print response
	if err := r.printResponse(resp); err != nil {
		return stepResult, err
	}

	// new response object
	respObj, err := newResponseObject(r.hrpRunner.t, r.parser, resp)
	if err != nil {
		err = errors.Wrap(err, "init ResponseObject error")
		return
	}

	// add response object to step variables, could be used in teardown hooks
	step.Variables["hrp_step_response"] = respObj.respObjMeta

	// deal with teardown hooks
	for _, teardownHook := range step.TeardownHooks {
		_, err = r.parser.parseData(teardownHook, step.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run teardown hooks failed")
		}
	}

	sessionData.ReqResps.Request = requestMap
	sessionData.ReqResps.Response = builtin.FormatResponse(respObj.respObjMeta)

	// extract variables from response
	extractors := step.Extract
	extractMapping := respObj.Extract(extractors)
	stepResult.ExportVars = extractMapping

	// override step variables with extracted variables
	stepVariables := mergeVariables(step.Variables, extractMapping)

	// validate response
	err = respObj.Validate(step.Validators, stepVariables)
	sessionData.Validators = respObj.validationResults
	if err == nil {
		sessionData.Success = true
		stepResult.Success = true
	}
	stepResult.ContentSize = resp.ContentLength
	stepResult.Data = sessionData

	return stepResult, err
}

func (r *caseRunner) printRequest(req *http.Request) error {
	if !r.hrpRunner.requestsLogOn {
		return nil
	}
	reqContentType := req.Header.Get("Content-Type")
	printBody := shouldPrintBody(reqContentType)
	reqDump, err := httputil.DumpRequest(req, printBody)
	if err != nil {
		return errors.Wrap(err, "dump request failed")
	}
	fmt.Println("-------------------- request --------------------")
	reqContent := string(reqDump)
	if req.Body != nil && !printBody {
		reqContent += fmt.Sprintf("(request body omitted for Content-Type: %v)", reqContentType)
	}
	fmt.Println(reqContent)
	return nil
}

func (r *caseRunner) printResponse(resp *http.Response) error {
	if !r.hrpRunner.requestsLogOn {
		return nil
	}
	fmt.Println("==================== response ===================")
	respContentType := resp.Header.Get("Content-Type")
	printBody := shouldPrintBody(respContentType)
	respDump, err := httputil.DumpResponse(resp, printBody)
	if err != nil {
		return errors.Wrap(err, "dump response failed")
	}
	respContent := string(respDump)
	if !printBody {
		respContent += fmt.Sprintf("(response body omitted for Content-Type: %v)", respContentType)
	}
	fmt.Println(respContent)
	fmt.Println("--------------------------------------------------")
	return nil
}

// shouldPrintBody return true if the Content-Type is printable
// including text/*, application/json, application/xml, application/www-form-urlencoded
func shouldPrintBody(contentType string) bool {
	if strings.HasPrefix(contentType, "text/") {
		return true
	}
	if strings.HasPrefix(contentType, "application/json") {
		return true
	}
	if strings.HasPrefix(contentType, "application/xml") {
		return true
	}
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return true
	}
	return false
}

func decodeResponseBody(resp *http.Response) (err error) {
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		resp.Body = io.NopCloser(brotli.NewReader(resp.Body))
	case "gzip":
		resp.Body, err = gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		resp.ContentLength = -1 // set to unknown to avoid Content-Length mismatched
	case "deflate":
		resp.Body, err = zlib.NewReader(resp.Body)
		if err != nil {
			return err
		}
		resp.ContentLength = -1 // set to unknown to avoid Content-Length mismatched
	}
	return nil
}

func (r *caseRunner) runStepTestCase(step *TStep) (stepResult *stepData, err error) {
	stepResult = &stepData{
		Name:     step.Name,
		StepType: stepTypeTestCase,
		Success:  false,
	}
	testcase := step.TestCaseContent

	// copy testcase to avoid data racing
	copiedTestCase := &TestCase{}
	if err = copier.Copy(copiedTestCase, testcase); err != nil {
		log.Error().Err(err).Msg("copy testcase failed")
		return stepResult, err
	}
	// override testcase config
	extendWithTestCase(step, copiedTestCase)

	start := time.Now()
	caseRunnerObj := r.hrpRunner.newCaseRunner(copiedTestCase)
	err = caseRunnerObj.run()
	stepResult.Elapsed = time.Since(start).Milliseconds()
	if err != nil {
		return stepResult, err
	}
	stepResult.Data = caseRunnerObj.getSummary()
	// export testcase export variables
	stepResult.ExportVars = caseRunnerObj.summary.InOut.ExportVars
	stepResult.Success = true
	return stepResult, nil
}

func (r *caseRunner) parseConfig(cfg *TConfig) error {
	// parse config variables
	parsedVariables, err := r.parser.parseVariables(cfg.Variables)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	cfg.Variables = parsedVariables

	// parse config name
	parsedName, err := r.parser.parseString(cfg.Name, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := r.parser.parseString(cfg.BaseURL, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.BaseURL = convertString(parsedBaseURL)

	// ensure correction of think time config
	cfg.ThinkTime.checkThinkTime()

	return nil
}

func newSummary() *testCaseSummary {
	return &testCaseSummary{
		Success: true,
		Stat:    &testStepStat{},
		Time:    &testCaseTime{},
		InOut:   &testCaseInOut{},
	}
}

func (r *caseRunner) getSummary() *testCaseSummary {
	caseSummary := r.summary
	caseSummary.Time.StartAt = r.startTime
	caseSummary.Time.Duration = time.Since(r.startTime).Seconds()
	exportVars := make(map[string]interface{})
	for _, value := range r.Config.Export {
		exportVars[value] = r.sessionVariables[value]
	}
	caseSummary.InOut.ExportVars = exportVars
	caseSummary.InOut.ConfigVars = r.Config.Variables
	return caseSummary
}

func setBodyBytes(req *http.Request, data []byte) {
	req.Body = io.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
}

//go:embed internal/report/template.html
var reportTemplate string

func (s *Summary) genHTMLReport() error {
	dir, _ := filepath.Split(reportPath)
	err := builtin.EnsureFolderExists(dir)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(fmt.Sprintf(reportPath, s.Time.StartAt.Unix()), os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		log.Error().Err(err).Msg("open file failed")
		return err
	}
	writer := bufio.NewWriter(file)
	tmpl := template.Must(template.New("report").Parse(reportTemplate))
	err = tmpl.Execute(writer, s)
	if err != nil {
		log.Error().Err(err).Msg("execute applies a parsed template to the specified data object failed")
		return err
	}
	err = writer.Flush()
	return err
}
