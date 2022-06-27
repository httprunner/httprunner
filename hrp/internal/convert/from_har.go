package convert

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

// ==================== model definition starts here ====================

/*
HTTP Archive (HAR) format
https://w3c.github.io/web-performance/specs/HAR/Overview.html
this file is copied from https://github.com/mrichman/hargo/blob/master/types.go
*/

// CaseHar is a container type for deserialization
type CaseHar struct {
	Log Log `json:"log"`
}

// Log represents the root of the exported data. This object MUST be present and its name MUST be "log".
type Log struct {
	// The object contains the following name/value pairs:

	// Required. Version number of the format.
	Version string `json:"version"`
	// Required. An object of type creator that contains the name and version
	// information of the log creator application.
	Creator Creator `json:"creator"`
	// Optional. An object of type browser that contains the name and version
	// information of the user agent.
	Browser Browser `json:"browser"`
	// Optional. An array of objects of type page, each representing one exported
	// (tracked) page. Leave out this field if the application does not support
	// grouping by pages.
	Pages []Page `json:"pages,omitempty"`
	// Required. An array of objects of type entry, each representing one
	// exported (tracked) HTTP request.
	Entries []Entry `json:"entries"`
	// Optional. A comment provided by the user or the application. Sorting
	// entries by startedDateTime (starting from the oldest) is preferred way how
	// to export data since it can make importing faster. However the reader
	// application should always make sure the array is sorted (if required for
	// the import).
	Comment string `json:"comment"`
}

// Creator contains information about the log creator application
type Creator struct {
	// Required. The name of the application that created the log.
	Name string `json:"name"`
	// Required. The version number of the application that created the log.
	Version string `json:"version"`
	// Optional. A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Browser that created the log
type Browser struct {
	// Required. The name of the browser that created the log.
	Name string `json:"name"`
	// Required. The version number of the browser that created the log.
	Version string `json:"version"`
	// Optional. A comment provided by the user or the browser.
	Comment string `json:"comment"`
}

// Page object for every exported web page and one <entry> object for every HTTP request.
// In case when an HTTP trace tool isn't able to group requests by a page,
// the <pages> object is empty and individual requests doesn't have a parent page.
type Page struct {
	/* There is one <page> object for every exported web page and one <entry>
	object for every HTTP request. In case when an HTTP trace tool isn't able to
	group requests by a page, the <pages> object is empty and individual
	requests doesn't have a parent page.
	*/

	// Date and time stamp for the beginning of the page load
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.45+01:00).
	StartedDateTime string `json:"startedDateTime"`
	// Unique identifier of a page within the . Entries use it to refer the parent page.
	ID string `json:"id"`
	// Page title.
	Title string `json:"title"`
	// Detailed timing info about page load.
	PageTiming PageTiming `json:"pageTiming"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// PageTiming describes timings for various events (states) fired during the page load.
// All times are specified in milliseconds. If a time info is not available appropriate field is set to -1.
type PageTiming struct {
	// Content of the page loaded. Number of milliseconds since page load started
	// (page.startedDateTime). Use -1 if the timing does not apply to the current
	// request.
	// Depeding on the browser, onContentLoad property represents DOMContentLoad
	// event or document.readyState == interactive.
	OnContentLoad int `json:"onContentLoad"`
	// Page is loaded (onLoad event fired). Number of milliseconds since page
	// load started (page.startedDateTime). Use -1 if the timing does not apply
	// to the current request.
	OnLoad int `json:"onLoad"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment"`
}

// Entry is a unique, optional Reference to the parent page.
// Leave out this field if the application does not support grouping by pages.
type Entry struct {
	Pageref string `json:"pageref,omitempty"`
	// Date and time stamp of the request start
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD).
	StartedDateTime string `json:"startedDateTime"`
	// Total elapsed time of the request in milliseconds. This is the sum of all
	// timings available in the timings object (i.e. not including -1 values) .
	Time float32 `json:"time"`
	// Detailed info about the request.
	Request Request `json:"request"`
	// Detailed info about the response.
	Response Response `json:"response"`
	// Info about cache usage.
	Cache Cache `json:"cache"`
	// Detailed timing info about request/response round trip.
	PageTimings PageTimings `json:"pageTimings"`
	// optional (new in 1.2) IP address of the server that was connected
	// (result of DNS resolution).
	ServerIPAddress string `json:"serverIPAddress,omitempty"`
	// optional (new in 1.2) Unique ID of the parent TCP/IP connection, can be
	// the client port number. Note that a port number doesn't have to be unique
	// identifier in cases where the port is shared for more connections. If the
	// port isn't available for the application, any other unique connection ID
	// can be used instead (e.g. connection index). Leave out this field if the
	// application doesn't support this info.
	Connection string `json:"connection,omitempty"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Request contains detailed info about performed request.
type Request struct {
	// Request method (GET, POST, ...).
	Method string `json:"method"`
	// Absolute URL of the request (fragments are not included).
	URL string `json:"url"`
	// Request HTTP Version.
	HTTPVersion string `json:"httpVersion"`
	// List of cookie objects.
	Cookies []Cookie `json:"cookies"`
	// List of header objects.
	Headers []NVP `json:"headers"`
	// List of query parameter objects.
	QueryString []NVP `json:"queryString"`
	// Posted data.
	PostData PostData `json:"postData"`
	// Total number of bytes from the start of the HTTP request message until
	// (and including) the double CRLF before the body. Set to -1 if the info
	// is not available.
	HeaderSize int `json:"headerSize"`
	// Size of the request body (POST data payload) in bytes. Set to -1 if the
	// info is not available.
	BodySize int `json:"bodySize"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment"`
}

// Response contains detailed info about the response.
type Response struct {
	// Response status.
	Status int `json:"status"`
	// Response status description.
	StatusText string `json:"statusText"`
	// Response HTTP Version.
	HTTPVersion string `json:"httpVersion"`
	// List of cookie objects.
	Cookies []Cookie `json:"cookies"`
	// List of header objects.
	Headers []NVP `json:"headers"`
	// Details about the response body.
	Content Content `json:"content"`
	// Redirection target URL from the Location response header.
	RedirectURL string `json:"redirectURL"`
	// Total number of bytes from the start of the HTTP response message until
	// (and including) the double CRLF before the body. Set to -1 if the info is
	// not available.
	// The size of received response-headers is computed only from headers that
	// are really received from the server. Additional headers appended by the
	// browser are not included in this number, but they appear in the list of
	// header objects.
	HeadersSize int `json:"headersSize"`
	// Size of the received response body in bytes. Set to zero in case of
	// responses coming from the cache (304). Set to -1 if the info is not
	// available.
	BodySize int `json:"bodySize"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Cookie contains list of all cookies (used in <request> and <response> objects).
type Cookie struct {
	// The name of the cookie.
	Name string `json:"name"`
	// The cookie value.
	Value string `json:"value"`
	// optional The path pertaining to the cookie.
	Path string `json:"path,omitempty"`
	// optional The host of the cookie.
	Domain string `json:"domain,omitempty"`
	// optional Cookie expiration time.
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.123+02:00).
	Expires string `json:"expires,omitempty"`
	// optional Set to true if the cookie is HTTP only, false otherwise.
	HTTPOnly bool `json:"httpOnly,omitempty"`
	// optional (new in 1.2) True if the cookie was transmitted over ssl, false
	// otherwise.
	Secure bool `json:"secure,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment bool `json:"comment,omitempty"`
}

// NVP is simply a name/value pair with a comment
type NVP struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

// PostData describes posted data, if any (embedded in <request> object).
type PostData struct {
	//  Mime type of posted data.
	MimeType string `json:"mimeType"`
	//  List of posted parameters (in case of URL encoded parameters).
	Params []PostParam `json:"params"`
	//  Plain text posted data
	Text string `json:"text"`
	// optional (new in 1.2) A comment provided by the user or the
	// application.
	Comment string `json:"comment,omitempty"`
}

// PostParam is a list of posted parameters, if any (embedded in <postData> object).
type PostParam struct {
	// name of a posted parameter.
	Name string `json:"name"`
	// optional value of a posted parameter or content of a posted file.
	Value string `json:"value,omitempty"`
	// optional name of a posted file.
	FileName string `json:"fileName,omitempty"`
	// optional content type of a posted file.
	ContentType string `json:"contentType,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Content describes details about response content (embedded in <response> object).
type Content struct {
	// Length of the returned content in bytes. Should be equal to
	// response.bodySize if there is no compression and bigger when the content
	// has been compressed.
	Size int `json:"size"`
	// optional Number of bytes saved. Leave out this field if the information
	// is not available.
	Compression int `json:"compression,omitempty"`
	// MIME type of the response text (value of the Content-Type response
	// header). The charset attribute of the MIME type is included (if
	// available).
	MimeType string `json:"mimeType"`
	// optional Response body sent from the server or loaded from the browser
	// cache. This field is populated with textual content only. The text field
	// is either HTTP decoded text or a encoded (e.g. "base64") representation of
	// the response body. Leave out this field if the information is not
	// available.
	Text string `json:"text,omitempty"`
	// optional (new in 1.2) Encoding used for response text field e.g
	// "base64". Leave out this field if the text field is HTTP decoded
	// (decompressed & unchunked), than trans-coded from its original character
	// set into UTF-8.
	Encoding string `json:"encoding,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Cache contains info about a request coming from browser cache.
type Cache struct {
	// optional State of a cache entry before the request. Leave out this field
	// if the information is not available.
	BeforeRequest CacheObject `json:"beforeRequest,omitempty"`
	// optional State of a cache entry after the request. Leave out this field if
	// the information is not available.
	AfterRequest CacheObject `json:"afterRequest,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// CacheObject is used by both beforeRequest and afterRequest
type CacheObject struct {
	// optional - Expiration time of the cache entry.
	Expires string `json:"expires,omitempty"`
	// The last time the cache entry was opened.
	LastAccess string `json:"lastAccess"`
	// Etag
	ETag string `json:"eTag"`
	// The number of times the cache entry has been opened.
	HitCount int `json:"hitCount"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// PageTimings describes various phases within request-response round trip.
// All times are specified in milliseconds.
type PageTimings struct {
	Blocked int `json:"blocked,omitempty"`
	// optional - Time spent in a queue waiting for a network connection. Use -1
	// if the timing does not apply to the current request.
	DNS int `json:"dns,omitempty"`
	// optional - DNS resolution time. The time required to resolve a host name.
	// Use -1 if the timing does not apply to the current request.
	Connect int `json:"connect,omitempty"`
	// optional - Time required to create TCP connection. Use -1 if the timing
	// does not apply to the current request.
	Send int `json:"send"`
	// Time required to send HTTP request to the server.
	Wait int `json:"wait"`
	// Waiting for a response from the server.
	Receive int `json:"receive"`
	// Time required to read entire response from the server (or cache).
	Ssl int `json:"ssl,omitempty"`
	// optional (new in 1.2) - Time required for SSL/TLS negotiation. If this
	// field is defined then the time is also included in the connect field (to
	// ensure backward compatibility with HAR 1.1). Use -1 if the timing does not
	// apply to the current request.
	Comment string `json:"comment,omitempty"`
	// optional (new in 1.2) - A comment provided by the user or the application.
}

// TestResult contains results for an individual HTTP request
type TestResult struct {
	URL       string    `json:"url"`
	Status    int       `json:"status"` // 200, 500, etc.
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Latency   int       `json:"latency"` // milliseconds
	Method    string    `json:"method"`
	HarFile   string    `json:"harfile"`
}

// ==================== model definition ends here ====================

func LoadHARCase(path string) (*hrp.TCase, error) {
	// load har file
	caseHAR, err := loadCaseHAR(path)
	if err != nil {
		return nil, err
	}

	// convert to TCase format
	return caseHAR.ToTCase()
}

func loadCaseHAR(path string) (*CaseHar, error) {
	caseHAR := new(CaseHar)
	err := builtin.LoadFile(path, caseHAR)
	if err != nil {
		return nil, errors.Wrap(err, "load har file failed")
	}
	if reflect.ValueOf(*caseHAR).IsZero() {
		return nil, errors.New("invalid har file")
	}
	return caseHAR, nil
}

// convert CaseHar to TCase format
func (c *CaseHar) ToTCase() (*hrp.TCase, error) {
	teststeps, err := c.prepareTestSteps()
	if err != nil {
		return nil, err
	}

	tCase := &hrp.TCase{
		Config:    c.prepareConfig(),
		TestSteps: teststeps,
	}
	err = tCase.MakeCompat()
	if err != nil {
		return nil, err
	}
	return tCase, nil
}

func (c *CaseHar) prepareConfig() *hrp.TConfig {
	return hrp.NewConfig("testcase description").
		SetVerifySSL(false)
}

func (c *CaseHar) prepareTestSteps() ([]*hrp.TStep, error) {
	var steps []*hrp.TStep
	for _, entry := range c.Log.Entries {
		step, err := c.prepareTestStep(&entry)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	return steps, nil
}

func (c *CaseHar) prepareTestStep(entry *Entry) (*hrp.TStep, error) {
	log.Info().
		Str("method", entry.Request.Method).
		Str("url", entry.Request.URL).
		Msg("convert teststep")

	step := &stepFromHAR{
		TStep: hrp.TStep{
			Request:    &hrp.Request{},
			Validators: make([]interface{}, 0),
		},
	}
	if err := step.makeRequestMethod(entry); err != nil {
		return nil, err
	}
	if err := step.makeRequestURL(entry); err != nil {
		return nil, err
	}
	if err := step.makeRequestParams(entry); err != nil {
		return nil, err
	}
	if err := step.makeRequestCookies(entry); err != nil {
		return nil, err
	}
	if err := step.makeRequestHeaders(entry); err != nil {
		return nil, err
	}
	if err := step.makeRequestBody(entry); err != nil {
		return nil, err
	}
	if err := step.makeValidate(entry); err != nil {
		return nil, err
	}
	return &step.TStep, nil
}

type stepFromHAR struct {
	hrp.TStep
}

func (s *stepFromHAR) makeRequestMethod(entry *Entry) error {
	s.Request.Method = hrp.HTTPMethod(entry.Request.Method)
	return nil
}

func (s *stepFromHAR) makeRequestURL(entry *Entry) error {
	u, err := url.Parse(entry.Request.URL)
	if err != nil {
		log.Error().Err(err).Msg("make request url failed")
		return err
	}
	s.Request.URL = fmt.Sprintf("%s://%s", u.Scheme, u.Host+u.Path)
	return nil
}

func (s *stepFromHAR) makeRequestParams(entry *Entry) error {
	s.Request.Params = make(map[string]interface{})
	for _, param := range entry.Request.QueryString {
		s.Request.Params[param.Name] = param.Value
	}
	return nil
}

func (s *stepFromHAR) makeRequestCookies(entry *Entry) error {
	// use cookies from har
	s.Request.Cookies = make(map[string]string)
	for _, cookie := range entry.Request.Cookies {
		s.Request.Cookies[cookie.Name] = cookie.Value
	}
	return nil
}

func (s *stepFromHAR) makeRequestHeaders(entry *Entry) error {
	// use headers from har
	s.Request.Headers = make(map[string]string)
	for _, header := range entry.Request.Headers {
		if strings.EqualFold(header.Name, "cookie") {
			continue
		}
		s.Request.Headers[header.Name] = header.Value
	}
	return nil
}

func (s *stepFromHAR) makeRequestBody(entry *Entry) error {
	mimeType := entry.Request.PostData.MimeType
	if mimeType == "" {
		// GET/HEAD/DELETE without body
		return nil
	}

	// POST/PUT with body
	if strings.HasPrefix(mimeType, "application/json") {
		// post json
		var body interface{}
		if entry.Request.PostData.Text == "" {
			body = nil
		} else {
			err := json.Unmarshal([]byte(entry.Request.PostData.Text), &body)
			if err != nil {
				log.Error().Err(err).Msg("make request body failed")
				return err
			}
		}
		s.Request.Body = body
	} else if strings.HasPrefix(mimeType, "application/x-www-form-urlencoded") {
		// post form
		paramsMap := make(map[string]string)
		for _, param := range entry.Request.PostData.Params {
			paramsMap[param.Name] = param.Value
		}
		s.Request.Body = paramsMap
	} else if strings.HasPrefix(mimeType, "text/plain") {
		// post raw data
		s.Request.Body = entry.Request.PostData.Text
	} else {
		// TODO
		log.Error().Msgf("makeRequestBody: Not implemented for mimeType %s", mimeType)
	}
	return nil
}

func (s *stepFromHAR) makeValidate(entry *Entry) error {
	// make validator for response status code
	s.Validators = append(s.Validators, hrp.Validator{
		Check:   "status_code",
		Assert:  "equals",
		Expect:  entry.Response.Status,
		Message: "assert response status code",
	})

	// make validators for response headers
	for _, header := range entry.Response.Headers {
		// assert Content-Type
		if strings.EqualFold(header.Name, "Content-Type") {
			s.Validators = append(s.Validators, hrp.Validator{
				Check:   "headers.\"Content-Type\"",
				Assert:  "equals",
				Expect:  header.Value,
				Message: "assert response header Content-Type",
			})
		}
	}

	// make validators for response body
	respBody := entry.Response.Content
	if respBody.Text == "" {
		// response body is empty
		return nil
	}
	if strings.HasPrefix(respBody.MimeType, "application/json") {
		var data []byte
		var err error
		// response body is json
		if respBody.Encoding == "base64" {
			// decode base64 text
			data, err = base64.StdEncoding.DecodeString(respBody.Text)
			if err != nil {
				return errors.Wrap(err, "decode base64 error")
			}
		} else if respBody.Encoding == "" {
			// no encoding
			data = []byte(respBody.Text)
		} else {
			// other encoding type
			return nil
		}
		// convert to json
		var body interface{}
		if err = json.Unmarshal(data, &body); err != nil {
			return errors.Wrap(err, "json.Unmarshal body error")
		}
		jsonBody, ok := body.(map[string]interface{})
		if !ok {
			return fmt.Errorf("response body is not json, not matched with MimeType")
		}

		// response body is json
		keys := make([]string, 0, len(jsonBody))
		for k := range jsonBody {
			keys = append(keys, k)
		}
		// sort map keys to keep validators in stable order
		sort.Strings(keys)
		for _, key := range keys {
			value := jsonBody[key]
			switch v := value.(type) {
			case map[string]interface{}:
				continue
			case []interface{}:
				continue
			default:
				s.Validators = append(s.Validators, hrp.Validator{
					Check:   fmt.Sprintf("body.%s", key),
					Assert:  "equals",
					Expect:  v,
					Message: fmt.Sprintf("assert response body %s", key),
				})
			}
		}
	}

	return nil
}
