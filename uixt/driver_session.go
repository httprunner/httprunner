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
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/uixt/option"
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
	timeout := 120 * time.Second
	session := &DriverSession{
		ctx:     context.Background(),
		ID:      "<SessionNotInit>",
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetry: 1,
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

func (s *DriverSession) GetData(withReset bool) SessionData {
	sessionData := SessionData{
		Requests:      s.History(),
		ScreenResults: s.screenResults,
	}
	if withReset {
		s.Reset()
	}
	return sessionData
}

func (s *DriverSession) SetBaseURL(baseUrl string) {
	log.Info().Str("baseUrl", baseUrl).Msg("set driver session base URL")
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

func (s *DriverSession) buildURL(urlStr string) (string, error) {
	// Handle empty URL or root path
	if urlStr == "" || urlStr == "/" {
		if s.baseUrl == "" {
			return "", fmt.Errorf("base URL is empty")
		}
		return s.baseUrl, nil
	}

	// Handle full URLs (absolute URLs)
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		u, err := url.Parse(urlStr)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}
		return u.String(), nil
	}

	// Validate base URL
	if s.baseUrl == "" {
		return "", fmt.Errorf("base URL is empty")
	}

	// Parse both base URL and relative URL upfront
	baseURL, err := url.Parse(s.baseUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	relativeURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse relative URL: %w", err)
	}

	// Special handling: when relative path starts with "/" and base URL has a non-root path,
	// we want to append to base path instead of replacing it
	if strings.HasPrefix(urlStr, "/") && baseURL.Path != "" && baseURL.Path != "/" {
		finalURL := *baseURL
		finalURL.Path = strings.TrimSuffix(baseURL.Path, "/") + relativeURL.Path
		finalURL.RawQuery = relativeURL.RawQuery
		finalURL.Fragment = relativeURL.Fragment
		return finalURL.String(), nil
	}

	// Use standard URL resolution for all other cases
	return baseURL.ResolveReference(relativeURL).String(), nil
}

func (s *DriverSession) GET(urlStr string, opts ...option.ActionOption) (rawResp DriverRawResponse, err error) {
	return s.RequestWithRetry(http.MethodGet, urlStr, nil, opts...)
}

func (s *DriverSession) POST(data interface{}, urlStr string, opts ...option.ActionOption) (rawResp DriverRawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, errors.Wrap(code.InvalidParamError, err.Error())
		}
	}
	return s.RequestWithRetry(http.MethodPost, urlStr, bsJSON, opts...)
}

func (s *DriverSession) DELETE(urlStr string, opts ...option.ActionOption) (rawResp DriverRawResponse, err error) {
	return s.RequestWithRetry(http.MethodDelete, urlStr, nil, opts...)
}

func (s *DriverSession) RequestWithRetry(method string, urlStr string, rawBody []byte, opts ...option.ActionOption) (
	rawResp DriverRawResponse, err error,
) {
	var lastError error

	maxRetry := s.maxRetry
	options := option.NewActionOptions(opts...)
	if options.MaxRetryTimes > 0 {
		maxRetry = options.MaxRetryTimes
	}

	synthesizeEventRetryAdded := false

	for attempt := 1; attempt <= maxRetry; attempt++ {
		// Execute the request
		rawResp, err = s.Request(method, urlStr, rawBody, opts...)
		if err == nil {
			if attempt > 1 {
				log.Info().Msgf("request succeeded after %d attempts", attempt)
			}
			return rawResp, nil
		}

		// only for WDA driver, check if "synthesizeEvent timeout" error and add extra retry
		if strings.Contains(err.Error(), "synthesizeEvent timeout") && !synthesizeEventRetryAdded {
			log.Warn().Err(err).Msg("synthesizeEvent timeout detected, adding one extra retry")
			maxRetry++
			synthesizeEventRetryAdded = true
		}

		// Check if it's already a DeviceOfflineError
		if errors.Is(err, code.DeviceOfflineError) {
			lastError = err
		} else {
			// Use DeviceHTTPDriverError for other errors
			lastError = errors.Wrap(code.DeviceHTTPDriverError, err.Error())
		}

		// If this was the last attempt, break
		if attempt == maxRetry {
			log.Error().Err(lastError).Msgf("request failed after %d retries, giving up", maxRetry)
			break
		} else {
			log.Warn().Err(lastError).Msgf("request failed after %d/%d retries, retrying", attempt, maxRetry)
		}

		// Wait before next attempt
		time.Sleep(3 * time.Second)

		// Try to reset the session for the next attempt
		if s.resetFn != nil {
			log.Warn().Msgf("attempting to reset driver session before attempt %d", attempt+1)
			if resetErr := s.resetFn(); resetErr != nil {
				log.Error().Err(resetErr).Msgf("failed to reset session, will retry without reset")
			} else {
				log.Info().Msg("session reset successful")
			}
		}
	}

	return nil, lastError
}

func (s *DriverSession) Request(method string, urlStr string, rawBody []byte, opts ...option.ActionOption) (
	rawResp DriverRawResponse, err error,
) {
	logid := uuid.New().String()
	timeout := s.timeout
	options := option.NewActionOptions(opts...)
	if options.Timeout > 0 {
		timeout = time.Duration(options.Timeout) * time.Second
	}

	// build final URL
	rawURL, err := s.buildURL(urlStr)
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

		logger = logger.Str("logid", logid).Str("request_method", method).Str("request_url", rawURL)
		if len(rawBody) < 1024 {
			logger = logger.Str("request_body", string(rawBody))
		}
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

	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-HTTP-Request-ID", logid)
	req.Header.Set("logid", logid)

	driverResult.RequestTime = time.Now()
	var resp *http.Response
	if resp, err = s.client.Do(req); err != nil {
		// Check for connection reset or EOF errors and classify as DeviceOfflineError
		if strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "EOF") {
			return nil, errors.Wrap(code.DeviceOfflineError, err.Error())
		}
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
	s.client.Transport = &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, fmt.Sprintf("localhost:%d", localPort))
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		TLSHandshakeTimeout: 10 * time.Second,
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
