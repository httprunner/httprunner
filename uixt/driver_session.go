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
		ID:      "<SessionNotInit>",
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetry: 5,
	}
	session.Reset()
	return session
}

type DriverSession struct {
	ctx      context.Context
	ID       string
	baseUrl  string
	client   *http.Client
	timeout  time.Duration
	maxRetry int

	// used to reset driver session when request failed
	resetFn func() error

	// cache driver request and response
	requests []*DriverRequests

	// cache screenshot results
	screenResults []*ScreenResult
}

func (s *DriverSession) Reset() {
	s.requests = make([]*DriverRequests, 0)
	s.screenResults = make([]*ScreenResult, 0)
}

func (s *DriverSession) SetBaseURL(baseUrl string) {
	s.baseUrl = baseUrl
}

func (s *DriverSession) RegisterResetHandler(fn func() error) {
	s.resetFn = fn
}

func (s *DriverSession) addRequestResult(driverResult *DriverRequests) {
	s.requests = append(s.requests, driverResult)
}

func (s *DriverSession) History() []*DriverRequests {
	return s.requests
}

func (s *DriverSession) concatURL(urlStr string) (string, error) {
	if urlStr == "" || urlStr == "/" {
		if s.baseUrl == "" {
			return "", fmt.Errorf("base URL is empty")
		}
		return s.baseUrl, nil
	}

	// replace with session ID
	if s.ID != "" && !strings.Contains(urlStr, s.ID) {
		sessionPattern := regexp.MustCompile(`/session/([^/]+)/`)
		if matches := sessionPattern.FindStringSubmatch(urlStr); len(matches) != 0 {
			urlStr = strings.Replace(urlStr, matches[1], s.ID, 1)
		}
	}

	// 处理完整 URL
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		u, err := url.Parse(urlStr)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		return u.String(), nil
	}

	// 处理相对路径
	if s.baseUrl == "" {
		return "", fmt.Errorf("base URL is empty")
	}
	u, err := url.Parse(s.baseUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	// 处理路径和查询参数
	parts := strings.SplitN(urlStr, "?", 2)
	u.Path = path.Join(u.Path, parts[0])
	if len(parts) > 1 {
		query, err := url.ParseQuery(parts[1])
		if err != nil {
			return "", fmt.Errorf("failed to parse query params: %w", err)
		}
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

func (s *DriverSession) GET(urlStr string) (rawResp DriverRawResponse, err error) {
	return s.RequestWithRetry(http.MethodGet, urlStr, nil)
}

func (s *DriverSession) POST(data interface{}, urlStr string) (rawResp DriverRawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return s.RequestWithRetry(http.MethodPost, urlStr, bsJSON)
}

func (s *DriverSession) DELETE(urlStr string) (rawResp DriverRawResponse, err error) {
	return s.RequestWithRetry(http.MethodDelete, urlStr, nil)
}

func (s *DriverSession) RequestWithRetry(method string, urlStr string, rawBody []byte) (
	rawResp DriverRawResponse, err error) {
	for count := 1; count <= s.maxRetry; count++ {
		rawResp, err = s.Request(method, urlStr, rawBody)
		if err == nil {
			return
		}
		time.Sleep(3 * time.Second)

		if s.resetFn != nil {
			log.Warn().Msg("reset driver session")
			if err2 := s.resetFn(); err2 != nil {
				log.Error().Err(err2).Msgf(
					"failed to reset session, try count %v", count)
			} else {
				log.Info().Msgf(
					"reset session success, try count %v", count)
			}
		}
	}
	return
}

func (s *DriverSession) Request(method string, urlStr string, rawBody []byte) (
	rawResp DriverRawResponse, err error) {

	// concat url with base url
	rawURL, err := s.concatURL(urlStr)
	if err != nil {
		return nil, err
	}

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

func (s *DriverSession) SetupPortForward(localPort int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return fmt.Errorf("create tcp connection error %v", err)
	}
	s.client.Transport = &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return conn, nil
		},
	}
	return nil
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
