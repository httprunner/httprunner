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

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/httpstat"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type HTTPMethod string

const (
	HTTP_GET     HTTPMethod = "GET"
	HTTP_HEAD    HTTPMethod = "HEAD"
	HTTP_POST    HTTPMethod = "POST"
	HTTP_PUT     HTTPMethod = "PUT"
	HTTP_DELETE  HTTPMethod = "DELETE"
	HTTP_OPTIONS HTTPMethod = "OPTIONS"
	HTTP_PATCH   HTTPMethod = "PATCH"
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

func initUpload(step *StepRequestWithOptionalArgs) {
	if step.Request.Headers == nil {
		step.Request.Headers = make(map[string]string)
	}
	step.Request.Headers["Content-Type"] = "${multipart_content_type($m_encoder)}"
	step.Request.Body = "$m_encoder"
}

func prepareUpload(parser *Parser, stepRequest *StepRequest, stepVariables map[string]interface{}) (err error) {
	if len(stepRequest.Request.Upload) == 0 {
		return
	}
	uploadMap, err := parser.Parse(stepRequest.Request.Upload, stepVariables)
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

func runStepRequest(r *SessionRunner, step IStep) (stepResult *StepResult, err error) {
	stepRequest := step.(*StepRequestWithOptionalArgs)
	start := time.Now()
	stepResult = &StepResult{
		Name:        stepRequest.StepName,
		StepType:    StepTypeRequest,
		Success:     false,
		ContentSize: 0,
		StartTime:   start.Unix(),
	}

	defer func() {
		stepResult.Elapsed = time.Since(start).Milliseconds()
		// update testcase summary
		if err != nil {
			stepResult.Attachments = err.Error()
		}
	}()

	err = prepareUpload(r.caseRunner.parser, stepRequest.StepRequest, stepRequest.Variables)
	if err != nil {
		return
	}

	sessionData := &SessionData{
		ReqResps: &ReqResps{},
	}
	parser := r.caseRunner.parser
	config := r.caseRunner.Config.Get()

	rb := newRequestBuilder(parser, config, stepRequest.Request)
	rb.req.Method = strings.ToUpper(string(stepRequest.Request.Method))

	err = rb.prepareUrlParams(stepRequest.Variables)
	if err != nil {
		return
	}

	err = rb.prepareHeaders(stepRequest.Variables)
	if err != nil {
		return
	}

	err = rb.prepareBody(stepRequest.Variables)
	if err != nil {
		return
	}

	// add request object to step variables, could be used in setup hooks
	stepRequest.Variables["hrp_step_name"] = step.Name
	stepRequest.Variables["hrp_step_request"] = rb.requestMap
	stepRequest.Variables["request"] = rb.requestMap // setup hooks compatible with v3

	// deal with setup hooks
	for _, setupHook := range stepRequest.SetupHooks {
		_, err := parser.Parse(setupHook, stepRequest.Variables)
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
	if stepRequest.Request.HTTP2 {
		client = r.caseRunner.hrpRunner.http2Client
	} else {
		client = r.caseRunner.hrpRunner.httpClient
	}

	// set step timeout
	if stepRequest.Request.Timeout != 0 {
		client.Timeout = time.Duration(stepRequest.Request.Timeout*1000) * time.Millisecond
	}

	// do request action
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

	if r.caseRunner.hrpRunner.httpStatOn {
		// resp.Body has been ReadAll
		httpStat.Finish()
		stepResult.HttpStat = httpStat.Durations()
		httpStat.Print()
	}

	// add response object to step variables, could be used in teardown hooks
	stepRequest.Variables["hrp_step_response"] = respObj.respObjMeta
	stepRequest.Variables["response"] = respObj.respObjMeta

	// deal with teardown hooks
	for _, teardownHook := range stepRequest.TeardownHooks {
		_, err := parser.Parse(teardownHook, stepRequest.Variables)
		if err != nil {
			return stepResult, errors.Wrap(err, "run teardown hooks failed")
		}
	}

	sessionData.ReqResps.Request = rb.requestMap
	sessionData.ReqResps.Response = builtin.FormatResponse(respObj.respObjMeta)

	// extract variables from response
	extractors := stepRequest.StepRequest.Extract
	extractMapping := respObj.Extract(extractors, stepRequest.Variables)
	stepResult.ExportVars = extractMapping

	// override step variables with extracted variables
	stepRequest.Variables = mergeVariables(stepRequest.Variables, extractMapping)

	// validate response
	err = respObj.Validate(stepRequest.Validators, stepRequest.Variables)
	sessionData.Validators = respObj.validationResults
	if err == nil {
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
		StepConfig: StepConfig{
			StepName:  name,
			Variables: make(map[string]interface{}),
		},
	}
}

type StepRequest struct {
	StepConfig
	Request *Request `json:"request,omitempty" yaml:"request,omitempty"`
}

// WithVariables sets variables for current teststep.
func (s *StepRequest) WithVariables(variables map[string]interface{}) *StepRequest {
	s.Variables = variables
	return s
}

// SetupHook adds a setup hook for current teststep.
func (s *StepRequest) SetupHook(hook string) *StepRequest {
	s.SetupHooks = append(s.SetupHooks, hook)
	return s
}

// HTTP2 enables HTTP/2 protocol
func (s *StepRequest) HTTP2() *StepRequest {
	s.Request = &Request{
		HTTP2: true,
	}
	return s
}

// Loop specify running times for the current step
func (s *StepRequest) Loop(times int) *StepRequest {
	s.Loops = times
	return s
}

// GET makes a HTTP GET request.
func (s *StepRequest) GET(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_GET
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_GET,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// HEAD makes a HTTP HEAD request.
func (s *StepRequest) HEAD(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_HEAD
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_HEAD,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// POST makes a HTTP POST request.
func (s *StepRequest) POST(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_POST
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_POST,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// PUT makes a HTTP PUT request.
func (s *StepRequest) PUT(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_PUT
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_PUT,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// DELETE makes a HTTP DELETE request.
func (s *StepRequest) DELETE(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_DELETE
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_DELETE,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// OPTIONS makes a HTTP OPTIONS request.
func (s *StepRequest) OPTIONS(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_OPTIONS
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_OPTIONS,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// PATCH makes a HTTP PATCH request.
func (s *StepRequest) PATCH(url string) *StepRequestWithOptionalArgs {
	if s.Request != nil {
		s.Request.Method = HTTP_PATCH
		s.Request.URL = url
	} else {
		s.Request = &Request{
			Method: HTTP_PATCH,
			URL:    url,
		}
	}
	return &StepRequestWithOptionalArgs{
		StepRequest: s,
	}
}

// CallRefCase calls a referenced testcase.
func (s *StepRequest) CallRefCase(tc ITestCase) *StepTestCaseWithOptionalArgs {
	testCase, err := tc.GetTestCase()
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcase")
		os.Exit(code.GetErrorCode(err))
	}
	return &StepTestCaseWithOptionalArgs{
		StepConfig: s.StepConfig,
		TestCase:   testCase,
	}
}

// CallRefAPI calls a referenced api.
func (s *StepRequest) CallRefAPI(api IAPI) *StepAPIWithOptionalArgs {
	api, err := api.ToAPI()
	if err != nil {
		log.Error().Err(err).Msg("failed to load api")
		os.Exit(code.GetErrorCode(err))
	}
	return &StepAPIWithOptionalArgs{
		StepConfig: s.StepConfig,
		API:        api,
	}
}

// StartTransaction starts a transaction.
func (s *StepRequest) StartTransaction(name string) *StepTransaction {
	return &StepTransaction{
		StepConfig: s.StepConfig,
		Transaction: &Transaction{
			Name: name,
			Type: TransactionStart,
		},
	}
}

// EndTransaction ends a transaction.
func (s *StepRequest) EndTransaction(name string) *StepTransaction {
	return &StepTransaction{
		StepConfig: s.StepConfig,
		Transaction: &Transaction{
			Name: name,
			Type: TransactionEnd,
		},
	}
}

// SetThinkTime sets think time.
func (s *StepRequest) SetThinkTime(time float64) *StepThinkTime {
	return &StepThinkTime{
		StepConfig: s.StepConfig,
		ThinkTime: &ThinkTime{
			Time: time,
		},
	}
}

// SetRendezvous creates a new rendezvous
func (s *StepRequest) SetRendezvous(name string) *StepRendezvous {
	return &StepRendezvous{
		Rendezvous: &Rendezvous{
			Name: name,
		},
	}
}

// WebSocket creates a new websocket action
func (s *StepRequest) WebSocket() *StepWebSocket {
	return &StepWebSocket{
		StepConfig: s.StepConfig,
		WebSocket:  &WebSocketAction{},
	}
}

// MobileUI creates a new mobile step session
func (s *StepRequest) MobileUI() *StepMobile {
	return &StepMobile{
		StepConfig: s.StepConfig,
		Mobile:     &MobileUI{},
	}
}

// Android creates a new android step session
func (s *StepRequest) Android(opts ...option.AndroidDeviceOption) *StepMobile {
	androidOptions := option.NewAndroidDeviceOptions(opts...)
	return &StepMobile{
		StepConfig: s.StepConfig,
		Android: &MobileUI{
			Serial: androidOptions.SerialNumber,
		},
	}
}

// IOS creates a new ios step session
func (s *StepRequest) IOS(opts ...option.IOSDeviceOption) *StepMobile {
	iosOptions := option.NewIOSDeviceOptions(opts...)
	return &StepMobile{
		StepConfig: s.StepConfig,
		IOS: &MobileUI{
			Serial: iosOptions.UDID,
		},
	}
}

// Harmony creates a new harmony step session
func (s *StepRequest) Harmony(opts ...option.HarmonyDeviceOption) *StepMobile {
	harmonyOptions := option.NewHarmonyDeviceOptions(opts...)
	return &StepMobile{
		StepConfig: s.StepConfig,
		Harmony: &MobileUI{
			Serial: harmonyOptions.ConnectKey,
		},
	}
}

// Shell creates a new shell step session
func (s *StepRequest) Shell(content string) *StepShell {
	return &StepShell{
		StepConfig: s.StepConfig,
		Shell: &Shell{
			String:         content,
			ExpectExitCode: 0,
		},
	}
}

// Function creates a new function step session
func (s *StepRequest) Function(fn func()) *StepFunction {
	return &StepFunction{
		StepConfig: s.StepConfig,
		Fn:         fn,
	}
}

// StepRequestWithOptionalArgs implements IStep interface.
type StepRequestWithOptionalArgs struct {
	*StepRequest
}

// SetVerify sets whether to verify SSL for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetVerify(verify bool) *StepRequestWithOptionalArgs {
	log.Info().Bool("verify", verify).Msg("set step request verify")
	s.Request.Verify = verify
	return s
}

// SetTimeout sets timeout for current HTTP request.
func (s *StepRequestWithOptionalArgs) SetTimeout(timeout time.Duration) *StepRequestWithOptionalArgs {
	log.Info().Float64("timeout(seconds)", timeout.Seconds()).Msg("set step request timeout")
	s.Request.Timeout = timeout.Seconds()
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
	s.Request.AllowRedirects = allowRedirects
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
	s.Request.Params = params
	return s
}

// WithHeaders sets HTTP request headers for current step.
func (s *StepRequestWithOptionalArgs) WithHeaders(headers map[string]string) *StepRequestWithOptionalArgs {
	s.Request.Headers = headers
	return s
}

// WithCookies sets HTTP request cookies for current step.
func (s *StepRequestWithOptionalArgs) WithCookies(cookies map[string]string) *StepRequestWithOptionalArgs {
	s.Request.Cookies = cookies
	return s
}

// WithBody sets HTTP request body for current step.
func (s *StepRequestWithOptionalArgs) WithBody(body interface{}) *StepRequestWithOptionalArgs {
	s.Request.Body = body
	return s
}

// WithUpload sets HTTP request body for uploading file(s).
func (s *StepRequestWithOptionalArgs) WithUpload(upload map[string]interface{}) *StepRequestWithOptionalArgs {
	// init upload
	initUpload(s)
	s.Request.Upload = upload
	return s
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepRequestWithOptionalArgs) TeardownHook(hook string) *StepRequestWithOptionalArgs {
	s.TeardownHooks = append(s.TeardownHooks, hook)
	return s
}

// Validate switches to step validation.
func (s *StepRequestWithOptionalArgs) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		StepRequestWithOptionalArgs: s,
	}
}

// Extract switches to step extraction.
func (s *StepRequestWithOptionalArgs) Extract() *StepRequestExtraction {
	s.StepConfig.Extract = make(map[string]string)
	return &StepRequestExtraction{
		StepRequestWithOptionalArgs: s,
	}
}

func (s *StepRequestWithOptionalArgs) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	return fmt.Sprintf("%v %s", s.Request.Method, s.Request.URL)
}

func (s *StepRequestWithOptionalArgs) Type() StepType {
	return StepType(fmt.Sprintf("request-%v", s.Request.Method))
}

func (s *StepRequestWithOptionalArgs) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepRequestWithOptionalArgs) Run(r *SessionRunner) (*StepResult, error) {
	return runStepRequest(r, s)
}

// StepRequestExtraction implements IStep interface.
type StepRequestExtraction struct {
	*StepRequestWithOptionalArgs
}

// WithJmesPath sets the JMESPath expression to extract from the response.
func (s *StepRequestExtraction) WithJmesPath(jmesPath string, varName string) *StepRequestExtraction {
	s.StepConfig.Extract[varName] = jmesPath
	return s
}

// Validate switches to step validation.
func (s *StepRequestExtraction) Validate() *StepRequestValidation {
	return &StepRequestValidation{
		StepRequestWithOptionalArgs: s.StepRequestWithOptionalArgs,
	}
}

func (s *StepRequestExtraction) Name() string {
	return s.StepName
}

func (s *StepRequestExtraction) Type() StepType {
	stepType := StepType(fmt.Sprintf("request-%v", s.Request.Method))
	return stepType + stepTypeSuffixExtraction
}

func (s *StepRequestExtraction) Struct() *StepConfig {
	return &s.StepConfig
}

func (s *StepRequestExtraction) Run(r *SessionRunner) (*StepResult, error) {
	if s.StepRequestWithOptionalArgs != nil {
		return runStepRequest(r, s.StepRequestWithOptionalArgs)
	}
	return nil, errors.New("unexpected protocol type")
}

// StepRequestValidation implements IStep interface.
type StepRequestValidation struct {
	*StepRequestWithOptionalArgs
}

func (s *StepRequestValidation) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	return fmt.Sprintf("%s %s", s.Request.Method, s.Request.URL)
}

func (s *StepRequestValidation) Type() StepType {
	stepType := StepType(fmt.Sprintf("request-%v", s.Request.Method))
	return stepType + stepTypeSuffixValidation
}

func (s *StepRequestValidation) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepRequestValidation) Run(r *SessionRunner) (*StepResult, error) {
	if s.StepRequestWithOptionalArgs != nil {
		return runStepRequest(r, s.StepRequestWithOptionalArgs)
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
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertGreater(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "greater_than",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLess(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "less_than",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertGreaterOrEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "greater_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLessOrEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "less_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertNotEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "not_equal",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertContains(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contains",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertTypeMatch(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "type_match",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertRegexp(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "regex_match",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertStartsWith(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "startswith",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertEndsWith(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "endswith",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertContainedBy(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "contained_by",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthLessThan(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_less_than",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertStringEqual(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "string_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertEqualFold(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "equal_fold",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthLessOrEquals(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_less_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthGreaterThan(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_greater_than",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepRequestValidation) AssertLengthGreaterOrEquals(jmesPath string, expected interface{}, msg string) *StepRequestValidation {
	v := Validator{
		Check:   jmesPath,
		Assert:  "length_greater_or_equals",
		Expect:  expected,
		Message: msg,
	}
	s.Validators = append(s.Validators, v)
	return s
}

// Validator represents validator for one HTTP response.
type Validator struct {
	Check   string      `json:"check" yaml:"check"` // get value with jmespath
	Assert  string      `json:"assert" yaml:"assert"`
	Expect  interface{} `json:"expect" yaml:"expect"`
	Message string      `json:"msg,omitempty" yaml:"msg,omitempty"` // optional
}
