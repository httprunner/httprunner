package uixt

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

type BrowserWebDriver struct {
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

func NewBrowserWebDriver(browserId string) (driver *BrowserWebDriver, err error) {
	log.Info().Msg("init NewBrowserWebDriver driver")
	driver = new(BrowserWebDriver)
	driver.urlPrefix = &url.URL{}
	driver.urlPrefix.Host = BROWSER_LOCAL_ADDRESS
	driver.urlPrefix.Scheme = "http"
	driver.scale = 1.0
	if err != nil {
		return nil, errors.Wrap(err, "create browser session failed")
	}
	driver.sessionId = browserId
	if err != nil {
		return nil, fmt.Errorf("adb forward: %w", err)
	}
	return driver, nil
}

func (wd *BrowserWebDriver) Drag(fromX, fromY, toX, toY float64, options ...option.ActionOption) (err error) {
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

	_, err = wd.HttpPOST(data, wd.sessionId, "ui/drag")
	return
}

func (wd *BrowserWebDriver) AppLaunch(packageName string) (err error) {
	data := map[string]interface{}{
		"url": packageName,
	}

	_, err = wd.HttpPOST(data, wd.sessionId, "ui/page_launch")
	return
}

func (wd *BrowserWebDriver) DeleteSession() (err error) {

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

func (wd *BrowserWebDriver) Scroll(delta int) (err error) {
	data := map[string]interface{}{
		"delta": delta,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/scroll")
	return err
}

func (wd *BrowserWebDriver) CreateNetListener() (*websocket.Conn, error) {
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

func (wd *BrowserWebDriver) ClosePage(pageIndex int) (err error) {
	data := map[string]interface{}{
		"page_index": pageIndex,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/page_close")
	return err
}

func (wd *BrowserWebDriver) HoverBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/hover")
	return err
}

func (wd *BrowserWebDriver) tapBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/tap")
	return err
}

func (wd *BrowserWebDriver) RightClick(x, y float64) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/right_click")
	return err
}

func (wd *BrowserWebDriver) RightclickbySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/right_click")
	return err
}

func (wd *BrowserWebDriver) GetElementTextBySelector(selector string, options ...option.ActionOption) (text string, err error) {
	actionOptions := option.NewActionOptions(options...)
	uri := "ui/element_text?selector=" + selector
	if actionOptions.Index > 0 {
		uri = uri + "&element_index=" + fmt.Sprintf("%v", actionOptions.Index)
	}
	resp, err := wd.HttpGet(http.MethodGet, wd.sessionId, uri)
	if err != nil {
		return "", err
	}
	data := resp.Data.(map[string]interface{})
	return data["text"].(string), nil
}

func (wd *BrowserWebDriver) GetPageUrl(options ...option.ActionOption) (text string, err error) {
	uri := "ui/page_url"
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		uri = uri + "?page_index=" + fmt.Sprintf("%v", actionOptions.Index)
	}
	resp, err := wd.HttpGet(http.MethodGet, wd.sessionId, uri)
	if err != nil {
		return "", err
	}
	data := resp.Data.(map[string]interface{})
	return data["url"].(string), nil
}

func (wd *BrowserWebDriver) IsElementExistBySelector(selector string) (bool, error) {
	resp, err := wd.HttpGet(wd.sessionId, "ui/element_exist", "?selector=", selector)
	if err != nil {
		return false, err
	}
	data := resp.Data.(map[string]interface{})
	return data["exist"].(bool), nil
}

func (wd *BrowserWebDriver) Hover(x, y float64) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/hover")
	return err
}

func (wd *BrowserWebDriver) DoubleTapXY(x, y float64, option ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/double_tap")
	return err
}

func (wd *BrowserWebDriver) Input(text string, option ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"text": text,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/input")
	return err
}

// Source Return application elements tree
func (wd *BrowserWebDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := wd.HttpGet(http.MethodGet, wd.sessionId, "stub/source")

	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(resp.Data)

	if err != nil {
		return "", err
	}

	return string(jsonData), err
}

func (wd *BrowserWebDriver) ScreenShot(options ...option.ActionOption) (*bytes.Buffer, error) {
	resp, err := wd.HttpGet(http.MethodGet, wd.sessionId, "screenshot")
	if err != nil {
		return nil, err
	}
	data := resp.Data.(map[string]interface{})
	screenshotBase64 := data["screenshot"].(string)
	screenRaw, err := base64.StdEncoding.DecodeString(screenshotBase64)
	res := bytes.NewBuffer(screenRaw)

	return res, err
}

func (wd *BrowserWebDriver) HttpPOST(data interface{}, pathElem ...string) (response *WebAgentResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}

	return wd.httpRequest(http.MethodPost, wd.concatURL(pathElem...), bsJSON)
}

func (wd *BrowserWebDriver) HttpGet(data interface{}, pathElem ...string) (response *WebAgentResponse, err error) {

	return wd.httpRequest(http.MethodGet, wd.concatURL(pathElem...), nil)
}

func (wd *BrowserWebDriver) concatURL(elem ...string) string {
	tmp, _ := url.Parse(wd.urlPrefix.String())
	commonPath := path.Join(append([]string{wd.urlPrefix.Path}, "api/v1/")...)
	tmp.Path = path.Join(append([]string{commonPath}, elem...)...)
	return tmp.String()
}

func (wd *BrowserWebDriver) httpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (response *WebAgentResponse, err error) {
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

func (wd *BrowserWebDriver) Status() (deviceStatus types.DeviceStatus, err error) {
	log.Warn().Msg("Status not implemented in ADBDriver")
	return
}

func (wd *BrowserWebDriver) DeviceInfo() (deviceInfo types.DeviceInfo, err error) {
	log.Warn().Msg("DeviceInfo not implemented in ADBDriver")
	return
}

func (wd *BrowserWebDriver) BatteryInfo() (batteryInfo types.BatteryInfo, err error) {
	log.Warn().Msg("BatteryInfo not implemented in ADBDriver")
	return
}

func (wd *BrowserWebDriver) WindowSize() (types.Size, error) {
	resp, err := wd.HttpGet(http.MethodGet, wd.sessionId, "window_size")
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

func (wd *BrowserWebDriver) Screen() (Screen, error) {
	return Screen{}, errors.New("not support")
}
func (wd *BrowserWebDriver) Scale() (float64, error) {
	return 0, errors.New("not support")
}

// GetTimestamp returns the timestamp of the mobile device
func (wd *BrowserWebDriver) GetTimestamp() (timestamp int64, err error) {
	return 0, errors.New("not support")
}

// Homescreen Forces the device under test to switch to the home screen
func (wd *BrowserWebDriver) Homescreen() error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) Unlock() (err error) {
	return errors.New("not support")
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (wd *BrowserWebDriver) AppTerminate(packageName string) (bool, error) {
	return false, errors.New("not support")
}

// AssertForegroundApp returns nil if the given package and activity are in foreground
func (wd *BrowserWebDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) Back() error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) AppClear(packageName string) error {
	return errors.New("not support")
}
func (wd *BrowserWebDriver) ClearImages() error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) PushImage(localPath string) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) Orientation() (orientation types.Orientation, err error) {
	log.Warn().Msg("Orientation not implemented in ADBDriver")
	return
}

// Tap Sends a tap event at the coordinate.
func (wd *BrowserWebDriver) Tap(x, y int, options ...option.ActionOption) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) TapFloat(x, y float64, options ...option.ActionOption) error {
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
	_, err := wd.HttpPOST(data, wd.sessionId, "ui/tap")
	return err
}

// DoubleTap Sends a double tap event at the coordinate.
func (wd *BrowserWebDriver) DoubleTap(x, y float64, options ...option.ActionOption) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err := wd.HttpPOST(data, wd.sessionId, "ui/double_tap")
	return err
}
func (wd *BrowserWebDriver) UploadFile(x, y float64, FileUrl, FileFormat string) (err error) {
	data := map[string]interface{}{
		"x":           x,
		"y":           y,
		"file_url":    FileUrl,
		"file_format": FileFormat,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/upload")
	return err
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (wd *BrowserWebDriver) TouchAndHold(x, y float64, options ...option.ActionOption) error {
	return errors.New("not support")
}

// Swipe works like Drag, but `pressForDuration` value is 0
func (wd *BrowserWebDriver) Swipe(fromX, fromY, toX, toY float64, options ...option.ActionOption) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...option.ActionOption) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) SetIme(ime string) error {
	return errors.New("not support")
}

// SendKeys Types a string into active element. There must be element with keyboard focus,
// otherwise an error is raised.
// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
func (wd *BrowserWebDriver) SendKeys(text string, options ...option.ActionOption) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) Clear(packageName string) error {
	return errors.New("not support")
}
func (wd *BrowserWebDriver) Setup() error {
	return nil
}

func (wd *BrowserWebDriver) GetDevice() IDevice {
	return nil
}

func (wd *BrowserWebDriver) ForegroundInfo() (app types.AppInfo, err error) {
	return
}

// PressBack Presses the back button
func (wd *BrowserWebDriver) PressBack(options ...option.ActionOption) error {
	_, err := wd.HttpPOST(map[string]interface{}{}, wd.sessionId, "ui/back")
	return err
}

func (wd *BrowserWebDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) Backspace(count int, options ...option.ActionOption) (err error) {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) LogoutNoneUI(packageName string) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) TapByText(text string, options ...option.ActionOption) error {
	return errors.New("not support")
}

// AccessibleSource Return application elements accessibility tree
func (wd *BrowserWebDriver) AccessibleSource() (string, error) {
	return "", errors.New("not support")
}

// HealthCheck Health check might modify simulator state so it should only be called in-between testing sessions
//
//	Checks health of XCTest by:
//	1) Querying application for some elements,
//	2) Triggering some device events.
func (wd *BrowserWebDriver) HealthCheck() error {
	return errors.New("not support")
}
func (wd *BrowserWebDriver) GetAppiumSettings() (map[string]interface{}, error) {
	return nil, errors.New("not support")
}
func (wd *BrowserWebDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	return nil, errors.New("not support")
}

func (wd *BrowserWebDriver) IsHealthy() (bool, error) {
	return false, errors.New("not support")
}

// triggers the log capture and returns the log entries
func (wd *BrowserWebDriver) StartCaptureLog(identifier ...string) (err error) {
	return errors.New("not support")
}
func (wd *BrowserWebDriver) StopCaptureLog() (result interface{}, err error) {
	return nil, errors.New("not support")
}
func (wd *BrowserWebDriver) RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error) {
	return "", errors.New("not support")
}

func (wd *BrowserWebDriver) TearDown() error {
	return nil
}

func (wd *BrowserWebDriver) InitSession(capabilities option.Capabilities) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) GetSession() *DriverSession {
	return nil
}

func (wd *BrowserWebDriver) ScreenRecord(duration time.Duration) (videoPath string, err error) {
	return
}

func (wd *BrowserWebDriver) Rotation() (rotation types.Rotation, err error) {
	return
}

func (wd *BrowserWebDriver) SetRotation(rotation types.Rotation) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) Home() error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	return errors.New("not support")
}

func (wd *BrowserWebDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	return wd.TapFloat(x, y, opts...)
}
