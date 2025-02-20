package driver_ext

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gorilla/websocket"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

func (wd *StubBrowserDriver) Drag(fromX, fromY, toX, toY float64, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"from_x": fromX,
		"from_y": fromY,
		"to_x":   toX,
		"to_y":   toY,
	}

	actionOptions := option.NewActionOptions(options...)

	if actionOptions.Duration > 0 {
		data["duration"] = actionOptions.Duration
	}

	_, err = wd.httpPOST(data, wd.sessionId, "ui/drag")
	return
}

func (wd *StubBrowserDriver) AppLaunch(packageName string) (err error) {
	data := map[string]interface{}{
		"url": packageName,
	}

	_, err = wd.httpPOST(data, wd.sessionId, "ui/page_launch")
	return
}

func (wd *StubBrowserDriver) DeleteSession() (err error) {

	url := wd.concatURL("context", wd.sessionId)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Timeout: 60 * time.Second, // 设置超时时间为5秒
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	rawResp, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	var result CreateBrowserResponse
	if err = json.Unmarshal(rawResp, &result); err != nil {
		return err
	}

	if result.Code != 0 {
		return errors.New(result.Message)
	}
	return nil
}

func (wd *StubBrowserDriver) Scroll(delta int) (err error) {
	data := map[string]interface{}{
		"delta": delta,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/scroll")
	return err
}

func (wd *StubBrowserDriver) CreateNetListener() (*websocket.Conn, error) {
	webSocketUrl := "ws://localhost:8093/websocket_net_listen"
	c, _, err := websocket.DefaultDialer.Dial(webSocketUrl, nil)
	if err != nil {
		return nil, err
	}
	// 发送消息
	initMessage := fmt.Sprintf(`{
    "type":"create_net_listener",
    "context_id":"%v"
	}`, wd.sessionId)
	err = c.WriteMessage(websocket.TextMessage, []byte(initMessage))
	return c, nil
}

func (wd *StubBrowserDriver) ClosePage(pageIndex int) (err error) {
	data := map[string]interface{}{
		"page_index": pageIndex,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/page_close")
	return err
}

func (wd *StubBrowserDriver) HoverBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/hover")
	return err
}

func (wd *StubBrowserDriver) tapBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/tap")
	return err
}

func (wd *StubBrowserDriver) RightClick(x, y int) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/right_click")
	return err
}

func (wd *StubBrowserDriver) RightclickbySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/right_click")
	return err
}

func (wd *StubBrowserDriver) GetElementTextBySelector(selector string, options ...option.ActionOption) (text string, err error) {
	actionOptions := option.NewActionOptions(options...)
	uri := "ui/element_text?selector=" + selector
	if actionOptions.Index > 0 {
		uri = uri + "&element_index=" + fmt.Sprintf("%v", actionOptions.Index)
	}
	resp, err := wd.httpGet(http.MethodGet, wd.sessionId, uri)
	if err != nil {
		return "", err
	}
	data := resp.Data.(map[string]interface{})
	return data["text"].(string), nil
}

func (wd *StubBrowserDriver) GetPageUrl(options ...option.ActionOption) (text string, err error) {
	uri := "ui/page_url"
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		uri = uri + "?page_index=" + fmt.Sprintf("%v", actionOptions.Index)
	}
	resp, err := wd.httpGet(http.MethodGet, wd.sessionId, uri)
	if err != nil {
		return "", err
	}
	data := resp.Data.(map[string]interface{})
	return data["url"].(string), nil
}

func (wd *StubBrowserDriver) IsElementExistBySelector(selector string) (bool, error) {
	resp, err := wd.httpGet(wd.sessionId, "ui/element_exist", "?selector=", selector)
	if err != nil {
		return false, err
	}
	data := resp.Data.(map[string]interface{})
	return data["exist"].(bool), nil
}

func (wd *StubBrowserDriver) Hover(x, y float64) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/hover")
	return err
}

func (wd *StubBrowserDriver) DoubleTapXY(x, y float64, option ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/double_tap")
	return err
}

func (wd *StubBrowserDriver) Input(text string, option ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"text": text,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/input")
	return err
}

// Source Return application elements tree
func (wd *StubBrowserDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := wd.httpGet(http.MethodGet, wd.sessionId, "stub/source")

	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(resp.Data)

	if err != nil {
		return "", err
	}

	return string(jsonData), err
}

func (wd *StubBrowserDriver) ScreenShot(options ...option.ActionOption) (*bytes.Buffer, error) {
	resp, err := wd.httpGet(http.MethodGet, wd.sessionId, "screenshot")
	if err != nil {
		return nil, err
	}
	data := resp.Data.(map[string]interface{})
	screenshotBase64 := data["screenshot"].(string)
	screenRaw, err := base64.StdEncoding.DecodeString(screenshotBase64)
	res := bytes.NewBuffer(screenRaw)

	return res, err
}

func (wd *StubBrowserDriver) httpPOST(data interface{}, pathElem ...string) (response *WebAgentResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}

	return wd.httpRequest(http.MethodPost, wd.concatURL(pathElem...), bsJSON)
}

func (wd *StubBrowserDriver) httpGet(data interface{}, pathElem ...string) (response *WebAgentResponse, err error) {

	return wd.httpRequest(http.MethodGet, wd.concatURL(pathElem...), nil)
}

func (wd *StubBrowserDriver) concatURL(elem ...string) string {
	tmp, _ := url.Parse(wd.urlPrefix.String())
	commonPath := path.Join(append([]string{wd.urlPrefix.Path}, "api/v1/")...)
	tmp.Path = path.Join(append([]string{commonPath}, elem...)...)
	return tmp.String()
}

func (wd *StubBrowserDriver) httpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (response *WebAgentResponse, err error) {
	req, err := http.NewRequest(method, rawURL, bytes.NewBuffer(rawBody))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	// 新建http client
	client := &http.Client{
		Timeout: 60 * time.Second, // 设置超时时间为5秒
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	rawResp, err := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	// 将结果解析为 JSON
	var result WebAgentResponse
	if err = json.Unmarshal(rawResp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		log.Info().Msgf("%v", result.Message)
		return nil, errors.New(result.Message)
	}

	if err != nil {
		return nil, err
	}
	return &result, err
}

func (wd *StubBrowserDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	data := map[string]interface{}{
		"url":        packageName,
		"web_cookie": password,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "stub/login")

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

func (wd *StubBrowserDriver) WindowSize() (types.Size, error) {
	resp, err := wd.httpGet(http.MethodGet, wd.sessionId, "window_size")
	if err != nil {
		return types.Size{}, err
	}
	data := resp.Data.(map[string]interface{})
	width := data["width"]
	height := data["height"]
	return types.Size{
		Width:  int(width.(float64)),
		Height: int(height.(float64)),
	}, nil
}

func (wd *StubBrowserDriver) TapFloat(x, y float64, options ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(options...)
	duration := 0.1
	if actionOptions.Duration > 0 {
		duration = actionOptions.Duration
	}
	data := map[string]interface{}{
		"x":        x,
		"y":        y,
		"duration": duration,
	}
	_, err := wd.httpPOST(data, wd.sessionId, "ui/tap")
	return err
}

// DoubleTap Sends a double tap event at the coordinate.
func (wd *StubBrowserDriver) DoubleTap(x, y float64, options ...option.ActionOption) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err := wd.httpPOST(data, wd.sessionId, "ui/double_tap")
	return err
}
func (wd *StubBrowserDriver) UploadFile(x, y float64, FileUrl, FileFormat string) (err error) {
	data := map[string]interface{}{
		"x":           x,
		"y":           y,
		"file_url":    FileUrl,
		"file_format": FileFormat,
	}
	_, err = wd.httpPOST(data, wd.sessionId, "ui/upload")
	return err
}
