package uixt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Driver struct {
	urlPrefix *url.URL
	sessionId string
	client    *http.Client
	scale     float64
}

func (wd *Driver) concatURL(u *url.URL, elem ...string) string {
	var tmp *url.URL
	if u == nil {
		u = wd.urlPrefix
	}
	tmp, _ = url.Parse(u.String())
	tmp.Path = path.Join(append([]string{u.Path}, elem...)...)
	return tmp.String()
}

func (wd *Driver) httpGET(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.httpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Driver) httpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.httpRequest(http.MethodPost, wd.concatURL(nil, pathElem...), bsJSON)
}

func (wd *Driver) httpDELETE(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.httpRequest(http.MethodDelete, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Driver) httpRequest(method string, rawURL string, rawBody []byte) (rawResp rawResponse, err error) {
	log.Debug().Str("method", method).Str("url", rawURL).Str("body", string(rawBody)).Msg("request driver agent")

	var req *http.Request
	if req, err = http.NewRequest(method, rawURL, bytes.NewBuffer(rawBody)); err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	var resp *http.Response
	if resp, err = wd.client.Do(req); err != nil {
		return nil, err
	}
	defer func() {
		// https://github.com/etcd-io/etcd/blob/v3.3.25/pkg/httputil/httputil.go#L16-L22
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	rawResp, err = ioutil.ReadAll(resp.Body)
	logger := log.Debug().Int("statusCode", resp.StatusCode).Str("duration", time.Since(start).String())
	if !strings.HasSuffix(rawURL, "screenshot") {
		// avoid printing screenshot data
		logger.Str("response", string(rawResp))
	}
	logger.Msg("get driver agent response")
	if err != nil {
		return nil, err
	}

	if err = rawResp.checkErr(); err != nil {
		if resp.StatusCode == http.StatusOK {
			return rawResp, nil
		}
		return nil, err
	}

	return
}

func (wd *Driver) resetUIA2Driver() error {
	ud, err := NewUIADriver(NewCapabilities(), wd.urlPrefix.String())
	if err != nil {
		return err
	}
	wd.client = ud.client
	wd.sessionId = ud.sessionId
	return nil
}

func (wd *Driver) uia2HttpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (rawResp rawResponse, err error) {
	disableRetryBool := len(disableRetry) > 0 && disableRetry[0]
	for retryCount := 1; retryCount <= 5; retryCount++ {
		rawResp, err = wd.httpRequest(method, rawURL, rawBody)
		if err == nil || disableRetryBool {
			return
		}
		// wait for UIA2 server to resume automatically
		time.Sleep(3 * time.Second)
		oldSessionID := wd.sessionId
		if err = wd.resetUIA2Driver(); err != nil {
			log.Err(err).Msgf("failed to reset uia2 driver, retry count: %v", retryCount)
			continue
		}
		log.Debug().Str("new session", wd.sessionId).Str("old session", oldSessionID).Msgf("successful to reset uia2 driver, retry count: %v", retryCount)
		if oldSessionID != "" {
			rawURL = strings.Replace(rawURL, oldSessionID, wd.sessionId, 1)
		}
	}
	return
}

func (wd *Driver) uia2HttpGET(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.uia2HttpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Driver) uia2HttpGETWithRetry(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.uia2HttpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil, true)
}

func (wd *Driver) uia2HttpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.uia2HttpRequest(http.MethodPost, wd.concatURL(nil, pathElem...), bsJSON)
}

func (wd *Driver) uia2HttpDELETE(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.uia2HttpRequest(http.MethodDelete, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Driver) resetWDASession() (err error) {
	capabilities := NewCapabilities()
	capabilities.WithDefaultAlertAction(AlertActionAccept)

	// [[FBRoute POST:@"/session"].withoutSession respondWithTarget:self action:@selector(handleCreateSession:)]
	data := make(map[string]interface{})
	data["capabilities"] = map[string]interface{}{"alwaysMatch": capabilities}

	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session"); err != nil {
		return err
	}
	var sessionInfo SessionInfo
	if sessionInfo, err = rawResp.valueConvertToSessionInfo(); err != nil {
		return err
	}
	wd.sessionId = sessionInfo.SessionId
	return
}

func (wd *Driver) resetWDADriver() error {
	capabilities := NewCapabilities()
	capabilities.WithDefaultAlertAction(AlertActionAccept)

	wdaDriver, err := NewWDADriver(capabilities, WDALocalPort, WDALocalMjpegPort)
	if err != nil {
		return err
	}
	wd.client = wdaDriver.client
	wd.sessionId = wdaDriver.sessionId
	return nil
}

func (wd *Driver) wdaHttpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (rawResp rawResponse, err error) {
	disableRetryBool := len(disableRetry) > 0 && disableRetry[0]
	for retryCount := 1; retryCount <= 5; retryCount++ {
		rawResp, err = wd.httpRequest(method, rawURL, rawBody)
		if err == nil || disableRetryBool {
			return
		}
		// TODO: polling WDA to check if resumed automatically
		time.Sleep(5 * time.Second)
		oldSessionID := wd.sessionId
		if err = wd.resetWDASession(); err != nil {
			log.Err(err).Msgf("failed to reset wda driver, retry count: %v", retryCount)
			continue
		}
		log.Debug().Str("new session", wd.sessionId).Str("old session", oldSessionID).Msgf("successful to reset wda driver, retry count: %v", retryCount)
		if oldSessionID != "" {
			rawURL = strings.Replace(rawURL, oldSessionID, wd.sessionId, 1)
		}
	}
	return
}

func (wd *Driver) wdaHttpGET(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.wdaHttpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil)
}

func (wd *Driver) wdaHttpGETWithRetry(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.wdaHttpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil, true)
}

func (wd *Driver) wdaHttpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.wdaHttpRequest(http.MethodPost, wd.concatURL(nil, pathElem...), bsJSON)
}

func (wd *Driver) wdaHttpDELETE(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.wdaHttpRequest(http.MethodDelete, wd.concatURL(nil, pathElem...), nil)
}

func convertToHTTPClient(conn net.Conn) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return conn, nil
			},
		},
		Timeout: 0,
	}
}
