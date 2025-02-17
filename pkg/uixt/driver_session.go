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
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

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

func NewDriverSession() *DriverSession {
	timeout := 30 * time.Second
	session := &DriverSession{
		ctx:     context.Background(),
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		requests: make([]*DriverRequests, 0),
		maxRetry: 5,
	}
	return session
}

type DriverSession struct {
	ctx      context.Context
	ID       string
	baseUrl  string
	client   *http.Client
	timeout  time.Duration
	maxRetry int

	// cache driver request and response
	requests []*DriverRequests
}

func (s *DriverSession) Init(capabilities option.Capabilities) (err error) {
	data := make(map[string]interface{})
	if len(capabilities) == 0 {
		data["capabilities"] = make(map[string]interface{})
	} else {
		data["capabilities"] = map[string]interface{}{"alwaysMatch": capabilities}
	}
	var rawResp DriverRawResponse
	if rawResp, err = s.POST(data, "/session"); err != nil {
		return err
	}
	reply := new(struct{ Value struct{ SessionId string } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return err
	}
	s.ID = reply.Value.SessionId

	// WDA
	// sessionInfo, err := rawResp.ValueConvertToSessionInfo()
	// if err != nil {
	// 	return err
	// }
	// // update session ID
	// wd.Session.ID = sessionInfo.ID

	return nil
}

func (s *DriverSession) Reset() {
	s.requests = make([]*DriverRequests, 0)
}

func (s *DriverSession) SetBaseURL(baseUrl string) {
	s.baseUrl = baseUrl
}

func (s *DriverSession) addRequestResult(driverResult *DriverRequests) {
	s.requests = append(s.requests, driverResult)
}

func (s *DriverSession) History() []*DriverRequests {
	return s.requests
}

func (s *DriverSession) concatURL(elem ...string) (string, error) {
	if s.baseUrl == "" {
		return "", fmt.Errorf("base URL is empty")
	}

	u, err := url.Parse(s.baseUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	// 分离路径和查询参数
	lastElem := elem[len(elem)-1]
	parts := strings.SplitN(lastElem, "?", 2)
	elem[len(elem)-1] = parts[0]

	// 合并基础路径
	u.Path = path.Join(append([]string{u.Path}, elem...)...)

	// 如果有查询参数，添加到 URL
	if len(parts) > 1 {
		u.RawQuery = parts[1]
	}

	return u.String(), nil
}

func (s *DriverSession) GET(pathElem ...string) (rawResp DriverRawResponse, err error) {
	urlStr, err := s.concatURL(pathElem...)
	if err != nil {
		return nil, err
	}
	return s.Request(http.MethodGet, urlStr, nil)
}

func (s *DriverSession) POST(data interface{}, pathElem ...string) (rawResp DriverRawResponse, err error) {
	urlStr, err := s.concatURL(pathElem...)
	if err != nil {
		return nil, err
	}
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return s.Request(http.MethodPost, urlStr, bsJSON)
}

func (s *DriverSession) DELETE(pathElem ...string) (rawResp DriverRawResponse, err error) {
	urlStr, err := s.concatURL(pathElem...)
	if err != nil {
		return nil, err
	}
	return s.Request(http.MethodDelete, urlStr, nil)
}

func (s *DriverSession) Request(method string, rawURL string, rawBody []byte) (
	rawResp DriverRawResponse, err error) {
	driverResult := &DriverRequests{
		RequestMethod: method,
		RequestUrl:    rawURL,
		RequestBody:   string(rawBody),
	}

	defer func() {
		s.addRequestResult(driverResult)

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

	ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	driverResult.RequestTime = time.Now()
	var resp *http.Response
	if resp, err = s.client.Do(req); err != nil {
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

// TODO: FIXME
func (s *DriverSession) RequestWithRetry(method string, rawURL string, rawBody []byte) (
	rawResp DriverRawResponse, err error) {
	for count := 1; count <= s.maxRetry; count++ {
		rawResp, err = s.Request(method, rawURL, rawBody)
		if err == nil {
			return
		}
		time.Sleep(3 * time.Second)
		oldSessionID := s.ID
		if err2 := s.Init(nil); err2 != nil {
			log.Error().Err(err2).Msgf(
				"failed to reset session, try count %v", count)
			continue
		}
		log.Debug().Str("new session", s.ID).Str("old session", oldSessionID).Msgf(
			"reset session successfully, try count %v", count)
		if oldSessionID != "" {
			rawURL = strings.Replace(rawURL, oldSessionID, s.ID, 1)
		}
	}
	return
}

func (s *DriverSession) InitConnection(localPort int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return fmt.Errorf("create tcp connection error %v", err)
	}
	s.client = NewHTTPClientWithConnection(conn, s.timeout)
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

func (r DriverRawResponse) ValueConvertToSessionInfo() (sessionInfo DriverSession, err error) {
	reply := new(struct{ Value struct{ DriverSession } })
	if err = json.Unmarshal(r, reply); err != nil {
		return DriverSession{}, err
	}
	sessionInfo = reply.Value.DriverSession
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
