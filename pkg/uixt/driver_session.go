package uixt

import (
	"bytes"
	"context"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/json"
)

type Session struct {
	ID      string
	baseURL *url.URL
	client  *http.Client

	// cache uia2/wda request and response
	requests []*DriverRequests
	// cache screenshot ocr results
	screenResults []*ScreenResult
}

func (d *Session) addScreenResult(screenResult *ScreenResult) {
	d.screenResults = append(d.screenResults, screenResult)
}

func (d *Session) addRequestResult(driverResult *DriverRequests) {
	d.requests = append(d.requests, driverResult)
}

func (d *Session) Init(baseUrl string) error {
	var err error
	d.baseURL, err = url.Parse(baseUrl)
	return err
}

func (d *Session) Reset() {
	d.screenResults = make([]*ScreenResult, 0)
	d.requests = make([]*DriverRequests, 0)
}

func (d *Session) GetData(withReset bool) Attachments {
	data := Attachments{
		"screen_results": d.screenResults,
	}
	if len(d.requests) != 0 {
		data["requests"] = d.requests
	}
	if withReset {
		d.Reset()
	}
	return data
}

type Attachments map[string]interface{}

type DriverRequests struct {
	RequestMethod string    `json:"request_method"`
	RequestUrl    string    `json:"request_url"`
	RequestBody   string    `json:"request_body,omitempty"`
	RequestTime   time.Time `json:"request_time"`

	ResponseStatus   int    `json:"response_status"`
	ResponseDuration int64  `json:"response_duration(ms)"` // ms
	ResponseBody     string `json:"response_body"`

	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (wd *Session) concatURL(u *url.URL, elem ...string) string {
	var tmp *url.URL
	if u == nil {
		u = wd.baseURL
	}
	tmp, _ = url.Parse(u.String())
	tmp.Path = path.Join(append([]string{u.Path}, elem...)...)
	return tmp.String()
}

func (wd *Session) GET(pathElem ...string) (rawResp DriverRawResponse, err error) {
	return wd.Request(http.MethodGet, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Session) POST(data interface{}, pathElem ...string) (rawResp DriverRawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.Request(http.MethodPost, wd.concatURL(nil, pathElem...), bsJSON)
}

func (wd *Session) DELETE(pathElem ...string) (rawResp DriverRawResponse, err error) {
	return wd.Request(http.MethodDelete, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Session) Request(method string, rawURL string, rawBody []byte) (rawResp DriverRawResponse, err error) {
	driverResult := &DriverRequests{
		RequestMethod: method,
		RequestUrl:    rawURL,
		RequestBody:   string(rawBody),
	}

	defer func() {
		wd.addRequestResult(driverResult)

		var logger *zerolog.Event
		if err != nil {
			driverResult.Success = false
			driverResult.Error = err.Error()
			logger = log.Error().Bool("success", false).Err(err)
		} else {
			driverResult.Success = true
			logger = log.Debug().Bool("success", true)
		}

		logger = logger.Str("request_method", method).Str("request_url", rawURL).
			Str("request_body", string(rawBody))
		if !driverResult.RequestTime.IsZero() {
			logger = logger.Int64("request_time", driverResult.RequestTime.UnixMilli())
		}
		if driverResult.ResponseStatus != 0 {
			logger = logger.
				Int("response_status", driverResult.ResponseStatus).
				Int64("response_duration(ms)", driverResult.ResponseDuration).
				Str("response_body", driverResult.ResponseBody)
		}
		logger.Msg("request uixt driver")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	driverResult.RequestTime = time.Now()
	var resp *http.Response
	if resp, err = wd.client.Do(req); err != nil {
		return nil, err
	}
	defer func() {
		// https://github.com/etcd-io/etcd/blob/v3.3.25/pkg/httputil/httputil.go#L16-L22
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	rawResp, err = io.ReadAll(resp.Body)
	duration := time.Since(driverResult.RequestTime)
	driverResult.ResponseDuration = duration.Milliseconds()
	driverResult.ResponseStatus = resp.StatusCode

	if strings.HasSuffix(rawURL, "screenshot") {
		// avoid printing screenshot data
		driverResult.ResponseBody = "OMITTED"
	} else {
		driverResult.ResponseBody = string(rawResp)
	}
	if err != nil {
		return nil, err
	}

	if err = rawResp.CheckErr(); err != nil {
		if resp.StatusCode == http.StatusOK {
			return rawResp, nil
		}
		return nil, err
	}

	return
}

func (wd *Session) InitConnection(localPort int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return fmt.Errorf("create tcp connection error %v", err)
	}
	wd.client = NewHTTPClientWithConnection(conn, 30*time.Second)
	return nil
}

func NewHTTPClientWithConnection(conn net.Conn, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return conn, nil
			},
		},
		Timeout: timeout,
	}
}

type DriverRawResponse []byte

func (r DriverRawResponse) CheckErr() (err error) {
	reply := new(struct {
		Value struct {
			Err        string `json:"error"`
			Message    string `json:"message"`
			Traceback  string `json:"traceback"`  // wda
			Stacktrace string `json:"stacktrace"` // uia
		}
	})
	if err = json.Unmarshal(r, reply); err != nil {
		return err
	}
	if reply.Value.Err != "" {
		errText := reply.Value.Message
		re := regexp.MustCompile(`{.+?=(.+?)}`)
		if re.MatchString(reply.Value.Message) {
			subMatch := re.FindStringSubmatch(reply.Value.Message)
			errText = subMatch[len(subMatch)-1]
		}
		return fmt.Errorf("%s: %s", reply.Value.Err, errText)
	}
	return
}

func (r DriverRawResponse) ValueConvertToString() (s string, err error) {
	reply := new(struct{ Value string })
	if err = json.Unmarshal(r, reply); err != nil {
		return "", errors.Wrapf(err, "json.Unmarshal failed, rawResponse: %s", string(r))
	}
	s = reply.Value
	return
}

func (r DriverRawResponse) ValueConvertToBool() (b bool, err error) {
	reply := new(struct{ Value bool })
	if err = json.Unmarshal(r, reply); err != nil {
		return false, err
	}
	b = reply.Value
	return
}

func (r DriverRawResponse) ValueConvertToSessionInfo() (sessionInfo Session, err error) {
	reply := new(struct{ Value struct{ Session } })
	if err = json.Unmarshal(r, reply); err != nil {
		return Session{}, err
	}
	sessionInfo = reply.Value.Session
	return
}

func (r DriverRawResponse) ValueConvertToJsonRawMessage() (raw builtinJSON.RawMessage, err error) {
	reply := new(struct{ Value builtinJSON.RawMessage })
	if err = json.Unmarshal(r, reply); err != nil {
		return nil, err
	}
	raw = reply.Value
	return
}

func (r DriverRawResponse) ValueConvertToJsonObject() (obj map[string]interface{}, err error) {
	if err = json.Unmarshal(r, &obj); err != nil {
		return nil, err
	}
	return
}

func (r DriverRawResponse) ValueDecodeAsBase64() (raw *bytes.Buffer, err error) {
	str, err := r.ValueConvertToString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to string")
	}
	decodeString, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64 string")
	}
	raw = bytes.NewBuffer(decodeString)
	return
}
