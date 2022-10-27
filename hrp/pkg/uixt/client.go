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
