package driver_ext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/pkg/errors"
)

const BROWSER_LOCAL_ADDRESS = "localhost:8093"

type WebAgentResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
	Result  interface{} `json:"result"`
}

type CreateBrowserResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    BrowserInfo `json:"data"`
}

type StubBrowserDriver struct {
	*uixt.BrowserWebDriver
	urlPrefix *url.URL
	sessionId string
	scale     float64
}

type BrowserInfo struct {
	ContextId string `json:"context_id"`
}

func CreateBrowser(timeout int) (browserInfo *BrowserInfo, err error) {
	data := map[string]interface{}{
		"timeout": timeout,
	}

	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}

	rawURL := "http://" + BROWSER_LOCAL_ADDRESS + "/api/v1/create_browser"
	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBuffer(bsJSON))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 30 * time.Second, // 设置超时时间为5秒
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	rawResp, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var result CreateBrowserResponse
	if err = json.Unmarshal(rawResp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Message)
	}

	return &result.Data, nil
}

func NewStubBrowserDriver(browserId string) (driver *StubBrowserDriver, err error) {
	BrowserWebDriver, err := uixt.NewBrowserWebDriver(browserId)
	if err != nil {
		return nil, errors.Wrap(err, "create browser session failed")
	}
	driver = &StubBrowserDriver{
		BrowserWebDriver: BrowserWebDriver,
	}
	driver.sessionId = browserId
	if err != nil {
		return nil, fmt.Errorf("adb forward: %w", err)
	}
	return driver, nil
}

// Source Return application elements tree
func (wd *StubBrowserDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := wd.BrowserWebDriver.HttpGet(http.MethodGet, wd.sessionId, "stub/source")

	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(resp.Data)

	if err != nil {
		return "", err
	}

	return string(jsonData), err
}

func (wd *StubBrowserDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	data := map[string]interface{}{
		"url":        packageName,
		"web_cookie": password,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "stub/login")

	if err != nil {
		return info, err
	}
	loginSuccss := AppLoginInfo{
		IsLogin: true,
		Uid:     wd.sessionId,
		Did:     password,
	}
	return loginSuccss, err
}
