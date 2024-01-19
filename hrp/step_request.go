package hrp

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"github.com/httprunner/httprunner/v4/hrp/pkg/httpstat"
)

type HTTPMethod string

const (
	httpGET     HTTPMethod = "GET"
	httpHEAD    HTTPMethod = "HEAD"
	httpPOST    HTTPMethod = "POST"
	httpPUT     HTTPMethod = "PUT"
	httpDELETE  HTTPMethod = "DELETE"
	httpOPTIONS HTTPMethod = "OPTIONS"
	httpPATCH   HTTPMethod = "PATCH"
)

// Request represents HTTP request data structure.
// This is used for teststep.
type Request struct {
	Method         HTTPMethod             `json:"method" yaml:"method"` // required
	URL            string                 `json:"url" yaml:"url"`       // required
	HTTP2          bool                   `json:"http2,omitempty" yaml:"http2,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	Headers        map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	Cookies        map[string]string      `json:"cookies,omitempty" yaml:"cookies,omitempty"`
	Body           interface{}            `json:"body,omitempty" yaml:"body,omitempty"`
	Json           interface{}            `json:"json,omitempty" yaml:"json,omitempty"`
	Data           interface{}            `json:"data,omitempty" yaml:"data,omitempty"`
	Timeout        float64                `json:"timeout,omitempty" yaml:"timeout,omitempty"` // timeout in seconds
	AllowRedirects bool                   `json:"allow_redirects,omitempty" yaml:"allow_redirects,omitempty"`
	Verify         bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
	Upload         map[string]interface{} `json:"upload,omitempty" yaml:"upload,omitempty"`
}

func newRequestBuilder(parser *Parser, config *TConfig, stepRequest *Request) *requestBuilder {
	// convert request struct to map
	jsonRequest, _ := json.Marshal(stepRequest)
	var requestMap map[string]interface{}
	_ = json.Unmarshal(jsonRequest, &requestMap)

	request := &http.Request{
		Header: make(http.Header),
	}
	if stepRequest.HTTP2 {
		request.ProtoMajor = 2
		request.ProtoMinor = 0
	} else {
		request.ProtoMajor = 1
		request.ProtoMinor = 1
	}

	return &requestBuilder{
		stepRequest: stepRequest,
		req:         request,
		config:      config,
		parser:      parser,
		requestMap:  requestMap,
	}
}

type requestBuilder struct {
	stepRequest *Request
	req         *http.Request
	parser      *Parser
	config      *TConfig
	requestMap  map[string]interface{}
}

func (r *requestBuilder) prepareHeaders(stepVariables map[string]interface{}) error {
	// prepare request headers
	stepHeaders := r.stepRequest.Headers
	if r.config.Headers != nil {
		// override headers
		stepHeaders = mergeMap(stepHeaders, r.config.Headers)
	}

	if len(stepHeaders) > 0 {
		headers, err := r.parser.ParseHeaders(stepHeaders, stepVariables)
		if err != nil {
			return errors.Wrap(err, "parse headers failed")
		}
		for key, value := range headers {
			// omit pseudo header names for HTTP/1, e.g. :authority, :method, :path, :scheme
			if strings.HasPrefix(key, ":") {
				continue
			}
			r.req.Header.Add(key, value)

			// prepare content length
			if strings.EqualFold(key, "Content-Length") && value != "" {
				if l, err := strconv.ParseInt(value, 10, 64); err == nil {
					r.req.ContentLength = l
				}
			}
		}
	}

	// prepare request cookies
	for cookieName, cookieValue := range r.stepRequest.Cookies {
		value, err := r.parser.Parse(cookieValue, stepVariables)
		if err != nil {
			return errors.Wrap(err, "parse cookie value failed")
		}
		r.req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: convertString(value),
		})
	}

	// update header
	headers := make(map[string]string)
	for key, value := range r.req.Header {
		headers[key] = value[0]
	}
	r.requestMap["headers"] = headers
	return nil
}

func (r *requestBuilder) prepareUrlParams(stepVariables map[string]interface{}) error {
	// parse step request url
	requestUrl, err := r.parser.ParseString(r.stepRequest.URL, stepVariables)
	if err != nil {
		log.Error().Err(err).Msg("parse request url failed")
		return err
	}
	var baseURL string
	if stepVariables["base_url"] != nil {
		baseURL, _ = stepVariables["base_url"].(string)
	}

	// prepare request params
	var queryParams url.Values
	if len(r.stepRequest.Params) > 0 {
		params, err := r.parser.Parse(r.stepRequest.Params, stepVariables)
		if err != nil {
			return errors.Wrap(err, "parse request params failed")
		}
		parsedParams := params.(map[string]interface{})
		if len(parsedParams) > 0 {
			queryParams = make(url.Values)
			for k, v := range parsedParams {
				queryParams.Add(k, convertString(v))
			}
		}

		// request params has been appended to url, thus delete it here
		delete(r.requestMap, "params")
	}

	// prepare url
	preparedURL := buildURL(baseURL, convertString(requestUrl), queryParams)
	r.req.URL = preparedURL
	r.req.Host = preparedURL.Host

	// update url
	r.requestMap["url"] = preparedURL.String()
	return nil
}

func (r *requestBuilder) prepareBody(stepVariables map[string]interface{}) error {
	// prepare request body
	if r.stepRequest.Body == nil {
		return nil
	}

	data, err := r.parser.Parse(r.stepRequest.Body, stepVariables)
	if err != nil {
		return err
	}
	// check request body format if Content-Type specified as application/json
	if strings.HasPrefix(r.req.Header.Get("Content-Type"), "application/json") {
		switch data.(type) {
		case bool, float64, string, map[string]interface{}, []interface{}, nil:
			break
		default:
			return errors.Errorf("request body type inconsistent with Content-Type: %v",
				r.req.Header.Get("Content-Type"))
		}
	}
	r.requestMap["body"] = data
	var dataBytes []byte
	switch vv := data.(type) {
	case map[string]interface{}:
		contentType := r.req.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			// post form data
			formData := make(url.Values)
			for k, v := range vv {
				formData.Add(k, convertString(v))
			}
			dataBytes = []byte(formData.Encode())
		} else {
			// post json
			dataBytes, err = json.Marshal(vv)
			if err != nil {
				return err
			}
			if contentType == "" {
				r.req.Header.Set("Content-Type", "application/json; charset=utf-8")
			}
		}
	case []interface{}:
		contentType := r.req.Header.Get("Content-Type")
		// post json
		dataBytes, err = json.Marshal(vv)
		if err != nil {
			return err
		}
		if contentType == "" {
			r.req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}
	case string:
		dataBytes = []byte(vv)
	case []byte:
		dataBytes = vv
	case bytes.Buffer:
		dataBytes = vv.Bytes()
	case *builtin.TFormDataWriter:
		dataBytes = vv.Payload.Bytes()
	default: // unexpected body type
		return errors.New("unexpected request body type")
	}

	r.req.Body = io.NopCloser(bytes.NewReader(dataBytes))
	r.req.ContentLength = int64(len(dataBytes))

	return nil
}

func initUpload(step *TStep) {
	if step.Request.Headers == nil {
		step.Request.Headers = make(map[string]string)
	}
	step.Request.Headers["Content-Type"] = "${multipart_content_type($m_encoder)}"
	step.Request.Body = "$m_encoder"
}

func prepareUpload(parser *Parser, step *TStep, stepVariables map[string]interface{}) (err error) {
	if len(step.Request.Upload) == 0 {
		return
	}
	uploadMap, err := parser.Parse(step.Request.Upload, stepVariables)
	if err != nil {
		return
	}
	stepVariables["m_upload"] = uploadMap
	mEncoder, err := parser.Parse("${multipart_encoder($m_upload)}", stepVariables)
	if err != nil {
		return
	}
	stepVariables["m_encoder"] = mEncoder
	return
}

func runStepRequest(r *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeRequest,
		Success:     false,
		ContentSize: 0,
	}

	// merge step variables with session variables
	stepVariables, err := r.ParseStepVariables(step.Variables)
	if err != nil {
		err = errors.Wrap(err, "parse step variables failed")
		return
	}

	defer func() {
		// update testcase summary
		if err != nil {
			stepResult.Attachments = err.Error()
		}
	}()

	err = prepareUpload(r.caseRunner.parser, step, stepVariables)
	if err != nil {
		return
	}

	sessionData := newSessionData()
	parser := r.caseRunner.parser
	config := r.caseRunner.parsedConfig

	rb := newRequestBuilder(parser, config, step.Request)
	rb.req.Method = strings.ToUpper(string(step.Request.Method))

	err = rb.prepareUrlParams(stepVariables)
	if err != nil {
		return
	}

	err = rb.prepareHeaders(stepVariables)
	if err != nil {
		return
	}

	err = rb.prepareBody(stepVariables)
	if err != nil {
		return
	}

	// add request object to step variables, could be used in setup hooks
	stepVariables["hrp_step_name"] = step.Name
	stepVariables["hrp_step_request"] = rb.requestMap
	stepVariables["request"] = rb.requestMap // setup hooks compatible with v3

	// deal with setup hooks
	for _, setupHook := range step.SetupHooks {
		_, err := parser.Parse(setupHook, stepVariables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run setup hooks failed")
		}
	}

	// log & print request
	if r.caseRunner.hrpRunner.requestsLogOn {
		if err := printRequest(rb.req); err != nil {
			return stepResult, err
		}
	}

	// stat HTTP request
	var httpStat httpstat.Stat
	if r.caseRunner.hrpRunner.httpStatOn {
		ctx := httpstat.WithHTTPStat(rb.req, &httpStat)
		rb.req = rb.req.WithContext(ctx)
	}

	// select HTTP client
	var client *http.Client
	if step.Request.HTTP2 {
		client = r.caseRunner.hrpRunner.http2Client
	} else {
		client = r.caseRunner.hrpRunner.httpClient
	}

	// set step timeout
	if step.Request.Timeout != 0 {
		client.Timeout = time.Duration(step.Request.Timeout*1000) * time.Millisecond
	}

	// do request action
	start := time.Now()
	resp, err := client.Do(rb.req)
	if err != nil {
		return stepResult, errors.Wrap(err, "do request failed")
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// decode response body in br/gzip/deflate formats
	err = decodeResponseBody(resp)
	if err != nil {
		return stepResult, errors.Wrap(err, "decode response body failed")
	}
	defer resp.Body.Close()

	// log & print response
	if r.caseRunner.hrpRunner.requestsLogOn {
		if err := printResponse(resp); err != nil {
			return stepResult, err
		}
	}

	// new response object
	respObj, err := newHttpResponseObject(r.caseRunner.hrpRunner.t, parser, resp)
	if err != nil {
		err = errors.Wrap(err, "init ResponseObject error")
		return
	}

	stepResult.Elapsed = time.Since(start).Milliseconds()
	if r.caseRunner.hrpRunner.httpStatOn {
		// resp.Body has been ReadAll
		httpStat.Finish()
		stepResult.HttpStat = httpStat.Durations()
		httpStat.Print()
	}

	// add response object to step variables, could be used in teardown hooks
	stepVariables["hrp_step_response"] = respObj.respObjMeta
	stepVariables["response"] = respObj.respObjMeta

	// deal with teardown hooks
	for _, teardownHook := range step.TeardownHooks {
		_, err := parser.Parse(teardownHook, stepVariables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run teardown hooks failed")
		}
	}

	sessionData.ReqResps.Request = rb.requestMap
	sessionData.ReqResps.Response = builtin.FormatResponse(respObj.respObjMeta)

	// extract variables from response
	extractors := step.Extract
	extractMapping := respObj.Extract(extractors, stepVariables)
	stepResult.ExportVars = extractMapping

	// override step variables with extracted variables
	stepVariables = mergeVariables(stepVariables, extractMapping)

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

func printRequest(req *http.Request) error {
	reqContentType := req.Header.Get("Content-Type")
	printBody := shouldPrintBody(reqContentType)
	reqDump, err := httputil.DumpRequest(req, printBody)
	if err != nil {
		return errors.Wrap(err, "dump request failed")
	}
	fmt.Println("-------------------- request --------------------")
	reqContent := string(reqDump)
	if reqContentType != "" && !printBody {
		reqContent += fmt.Sprintf("(request body omitted for Content-Type: %v)", reqContentType)
	}
	fmt.Println(reqContent)
	return nil
}

func printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(color.Output, format, a...)
}

func printResponse(resp *http.Response) error {
	fmt.Println("==================== response ====================")
	connectedVia := "plaintext"
	if resp.TLS != nil {
		switch resp.TLS.Version {
		case tls.VersionTLS12:
			connectedVia = "TLSv1.2"
		case tls.VersionTLS13:
			connectedVia = "TLSv1.3"
		}
	}
	printf("%s %s\n", color.CyanString("Connected via"), color.BlueString("%s", connectedVia))
	respContentType := resp.Header.Get("Content-Type")
	printBody := shouldPrintBody(respContentType)
	respDump, err := httputil.DumpResponse(resp, printBody)
	if err != nil {
		return errors.Wrap(err, "dump response failed")
	}
	respContent := string(respDump)
	if respContentType != "" && !printBody {
		respContent += fmt.Sprintf("(response body omitted for Content-Type: %v)", respContentType)
	}
	fmt.Println(respContent)
	fmt.Println("--------------------------------------------------")
	return nil
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

// NewStep returns a new constructed teststep with specified step name.
func NewStep(name string) *StepRequest {
	return &StepRequest{
		step: &TStep{
			Name:      name,
			Variables: make(map[string]interface{}),
		},
	}
}

type StepRequest struct {
	step *TStep
}

// WithVariables sets variables for current teststep.
func (s *StepRequest) WithVariables(variables map[string]interface{}) *StepRequest {
	s.step.Variables = variables
	return s
}

// SetupHook adds a setup hook for current teststep.
func (s *StepRequest) SetupHook(hook string) *StepRequest {
	s.step.SetupHooks = append(s.step.SetupHooks, hook)
	return s
}

// HTTP2 enables HTTP/2 protocol
func (s *StepRequest) HTTP2() *StepRequest {
	s.step.Request = &Request{
		HTTP2: true,
	}
	return s
}

// Loop specify running times for the current step
func (s *StepRequest) Loop(times int) *StepRequest {
	s.step.Loops = times
	return s
}

// GET makes a HTTP GET request.
func (s *StepRequest) GET(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpGET
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpGET,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// HEAD makes a HTTP HEAD request.
func (s *StepRequest) HEAD(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpHEAD
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpHEAD,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// POST makes a HTTP POST request.
func (s *StepRequest) POST(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpPOST
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpPOST,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// PUT makes a HTTP PUT request.
func (s *StepRequest) PUT(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpPUT
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpPUT,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// DELETE makes a HTTP DELETE request.
func (s *StepRequest) DELETE(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpDELETE
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpDELETE,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// OPTIONS makes a HTTP OPTIONS request.
func (s *StepRequest) OPTIONS(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpOPTIONS
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpOPTIONS,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// PATCH makes a HTTP PATCH request.
func (s *StepRequest) PATCH(url string) *StepRequestWithOptionalArgs {
	if s.step.Request != nil {
		s.step.Request.Method = httpPATCH
		s.step.Request.URL = url
	} else {
		s.step.Request = &Request{
			Method: httpPATCH,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		step: s.step,
	}
}

// CallRefCase calls a referenced testcase.
func (s *StepRequest) CallRefCase(tc ITestCase) *StepTestCaseWithOptionalArgs {
	var err error
	s.step.TestCase, err = tc.ToTestCase()
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcase")
		os.Exit(code.GetErrorCode(err))
	}
	return &StepTestCaseWithOptionalArgs{
		step: s.step,
	}
}

// CallRefAPI calls a referenced api.
func (s *StepRequest) CallRefAPI(api IAPI) *StepAPIWithOptionalArgs {
	var err error
	s.step.API, err = api.ToAPI()
	if err != nil {
		log.Error().Err(err).Msg("failed to load api")
		os.Exit(code.GetErrorCode(err))
	}
	return &StepAPIWithOptionalArgs{
		step: s.step,
	}
}

// StartTransaction starts a transaction.
func (s *StepRequest) StartTransaction(name string) *StepTransaction {
	s.step.Transaction = &Transaction{
		Name: name,
		Type: transactionStart,
	}
	return &StepTransaction{
		step: s.step,
	}
}

// EndTransaction ends a transaction.
func (s *StepRequest) EndTransaction(name string) *StepTransaction {
	s.step.Transaction = &Transaction{
		Name: name,
		Type: transactionEnd,
	}
	return &StepTransaction{
		step: s.step,
	}
}

// SetThinkTime sets think time.
func (s *StepRequest) SetThinkTime(time float64) *StepThinkTime {
	s.step.ThinkTime = &ThinkTime{
		Time: time,
	}
	return &StepThinkTime{
		step: s.step,
	}
}

// SetRendezvous creates a new rendezvous
func (s *StepRequest) SetRendezvous(name string) *StepRendezvous {
	s.step.Rendezvous = &Rendezvous{
		Name: name,
	}
	return &StepRendezvous{
		step: s.step,
	}
}

// WebSocket creates a new websocket action
func (s *StepRequest) WebSocket() *StepWebSocket {
	s.step.WebSocket = &WebSocketAction{}
	return &StepWebSocket{
		step: s.step,
	}
}

// Android creates a new android action
func (s *StepRequest) Android() *StepMobile {
	s.step.Android = &MobileStep{}
	return &StepMobile{
		step: s.step,
	}
}

// IOS creates a new ios action
func (s *StepRequest) IOS() *StepMobile {
	s.step.IOS = &MobileStep{}
	return &StepMobile{
		step: s.step,
	}
}

// Shell creates a new shell action
func (s *StepRequest) Shell(content string) *StepShell {
	s.step.Shell = &Shell{
		String:         content,
		ExpectExitCode: 0,
	}

	return &StepShell{
		step: s.step,
	}
}

// StepRequestWithOptionalArgs implements IStep interface.
type StepRequestWithOptionalArgs struct {
	step *TStep
}

// SetVerify sets whether to verify SSL for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetVerify(verify bool) *StepRequestWithOptionalArgs {
	log.Info().Bool("verify", verify).Msg("set step request verify")
	s.step.Request.Verify = verify
	return s
}

// SetTimeout sets timeout for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetTimeout(timeout time.Duration) *StepRequestWithOptionalArgs {
	log.Info().Float64("timeout(seconds)", timeout.Seconds()).Msg("set step request timeout")
	s.step.Request.Timeout = timeout.Seconds()
	return s
}

// SetProxies sets proxies for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetProxies(proxies map[string]string) *StepRequestWithOptionalArgs {
	log.Info().Interface("proxies", proxies).Msg("set step request proxies")
	// TODO
	return s
}

// SetAllowRedirects sets whether to allow redirects for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetAllowRedirects(allowRedirects bool) *StepRequestWithOptionalArgs {
	log.Info().Bool("allowRedirects", allowRedirects).Msg("set step request allowRedirects")
	s.step.Request.AllowRedirects = allowRedirects
	return s
}

// SetAuth sets auth for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetAuth(auth map[string]string) *StepRequestWithOptionalArgs {
	log.Info().Interface("auth", auth).Msg("set step request auth")
	// TODO
	return s
}

// WithParams sets HTTP request params for current step.
func (s *StepRequestWithOptionalArgs) WithParams(params map[string]interface{}) *StepRequestWithOptionalArgs {
	s.step.Request.Params = params
	return s
}

// WithHeaders sets HTTP request headers for current step.
func (s *StepRequestWithOptionalArgs) WithHeaders(headers map[string]string) *StepRequestWithOptionalArgs {
	s.step.Request.Headers = headers
	return s
}

// WithCookies sets HTTP request cookies for current step.
func (s *StepRequestWithOptionalArgs) WithCookies(cookies map[string]string) *StepRequestWithOptionalArgs {
	s.step.Request.Cookies = cookies
	return s
}

// WithBody sets HTTP request body for current step.
func (s *StepRequestWithOptionalArgs) WithBody(body interface{}) *StepRequestWithOptionalArgs {
	s.step.Request.Body = body
	return s
}

// WithUpload sets HTTP request body for uploading file(s).
func (s *StepRequestWithOptionalArgs) WithUpload(upload map[string]interface{}) *StepRequestWithOptionalArgs {
	// init upload
	initUpload(s.step)
	s.step.Request.Upload = upload
	return s
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepRequestWithOptionalArgs) TeardownHook(hook string) *StepRequestWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Validate switches to step validation.
func (s *StepRequestWithOptionalArgs) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		step: s.step,
	}
}

// Extract switches to step extraction.
func (s *StepRequestWithOptionalArgs) Extract() *StepRequestExtraction {
	s.step.Extract = make(map[string]string)
	return &StepRequestExtraction{
		step: s.step,
	}
}

func (s *StepRequestWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%v %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *StepRequestWithOptionalArgs) Type() StepType {
	return StepType(fmt.Sprintf("request-%v", s.step.Request.Method))
}

func (s *StepRequestWithOptionalArgs) Struct() *TStep {
	return s.step
}

func (s *StepRequestWithOptionalArgs) Run(r *SessionRunner) (*StepResult, error) {
	return runStepRequest(r, s.step)
}

// StepRequestExtraction implements IStep interface.
type StepRequestExtraction struct {
	step *TStep
}

// WithJmesPath sets the JMESPath expression to extract from the response.
func (s *StepRequestExtraction) WithJmesPath(jmesPath string, varName string) *StepRequestExtraction {
	s.step.Extract[varName] = jmesPath
	return s
}

// Validate switches to step validation.
func (s *StepRequestExtraction) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		step: s.step,
	}
}

func (s *StepRequestExtraction) Name() string {
	return s.step.Name
}

func (s *StepRequestExtraction) Type() StepType {
	var stepType StepType
	if s.step.WebSocket != nil {
		stepType = StepType(fmt.Sprintf("websocket-%v", s.step.WebSocket.Type))
	} else {
		stepType = StepType(fmt.Sprintf("request-%v", s.step.Request.Method))
	}
	return stepType + stepTypeSuffixExtraction
}

func (s *StepRequestExtraction) Struct() *TStep {
	return s.step
}

func (s *StepRequestExtraction) Run(r *SessionRunner) (*StepResult, error) {
	if s.step.Request != nil {
		return runStepRequest(r, s.step)
	}
	if s.step.WebSocket != nil {
		return runStepWebSocket(r, s.step)
	}
	return nil, errors.New("unexpected protocol type")
}

// StepRequestValidation implements IStep interface.
type StepRequestValidation struct {
	step *TStep
}

func (s *StepRequestValidation) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	return fmt.Sprintf("%s %s", s.step.Request.Method, s.step.Request.URL)
}

func (s *StepRequestValidation) Type() StepType {
	var stepType StepType
	if s.step.WebSocket != nil {
		stepType = StepType(fmt.Sprintf("websocket-%v", s.step.WebSocket.Type))
	} else {
		stepType = StepType(fmt.Sprintf("request-%v", s.step.Request.Method))
	}
	return stepType + stepTypeSuffixValidation
}

func (s *StepRequestValidation) Struct() *TStep {
	return s.step
}

func (s *StepRequestValidation) Run(r *SessionRunner) (*StepResult, error) {
	if s.step.Request != nil {
		return runStepRequest(r, s.step)
	}
	if s.step.WebSocket != nil {
		return runStepWebSocket(r, s.step)
	}
	return nil, errors.New("unexpected protocol type")
}

func (s *StepRequestValidation) AssertEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertGreater(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "greater_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLess(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "less_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertGreaterOrEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "greater_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLessOrEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "less_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertNotEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "not_equal",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertContains(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contains",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertTypeMatch(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "type_match",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertRegexp(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "regex_match",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertStartsWith(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "startswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertEndsWith(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "endswith",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertContainedBy(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contained_by",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthLessThan(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_less_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertStringEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "string_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertEqualFold(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equal_fold",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthLessOrEquals(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_less_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthGreaterThan(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_greater_than",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthGreaterOrEquals(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_greater_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.step.Validators = append(s.step.Validators, v)
	return s
}

// Validator represents validator for one HTTP response.
type Validator struct {
	Check   string      `json:"check" yaml:"check"` // get value with jmespath
	Assert  string      `json:"assert" yaml:"assert"`
	Expect  interface{} `json:"expect" yaml:"expect"`
	Message string      `json:"msg,omitempty" yaml:"msg,omitempty"` // optional
}
