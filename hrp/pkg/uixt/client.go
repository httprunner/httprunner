package uixt

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type DriverSession struct {
	ID string
	// cache uia2/wda request and response
	requests []*DriverResult
	// cache screenshot ocr results
	screenResults []*ScreenResult // list of actions
	// cache e2e delay
	e2eDelay []timeLog
}

func (d *DriverSession) addScreenResult(screenResult *ScreenResult) {
	d.screenResults = append(d.screenResults, screenResult)
}

func (d *DriverSession) addRequestResult(driverResult *DriverResult) {
	d.requests = append(d.requests, driverResult)
}

func (d *DriverSession) Reset() {
	d.screenResults = make([]*ScreenResult, 0)
	d.requests = make([]*DriverResult, 0)
	d.e2eDelay = nil
}

func (d *DriverSession) Get(withReset bool) map[string]interface{} {
	data := map[string]interface{}{
		"screen_results": d.screenResults,
	}
	if len(d.e2eDelay) != 0 {
		data["requests"] = d.requests
	}
	if d.e2eDelay != nil {
		data["e2e_results"] = d.e2eDelay
	}
	if withReset {
		d.Reset()
	}
	return data
}

type Driver struct {
	urlPrefix *url.URL
	client    *http.Client

	// cache to avoid repeated query
	scale      float64
	windowSize *Size

	// cache session data
	session DriverSession
}

type DriverResult struct {
	RequestUrl      string    `json:"request_driver_url"`
	RequestBody     string    `json:"request_driver_body,omitempty"`
	RequestDuration int64     `json:"request_driver_duration(ms)"` // ms
	RequestTime     time.Time `json:"request_driver_time"`
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
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	rawResp, err = io.ReadAll(resp.Body)
	duration := time.Since(start)
	driverResult := &DriverResult{
		RequestUrl:      rawURL,
		RequestBody:     string(rawBody),
		RequestDuration: duration.Milliseconds(),
		RequestTime:     time.Now(),
	}
	wd.session.addRequestResult(driverResult)
	logger := log.Debug().Int("statusCode", resp.StatusCode).Str("duration", duration.String())
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
