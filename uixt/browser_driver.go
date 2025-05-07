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
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
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

type BrowserDriver struct {
	urlPrefix *url.URL
	sessionId string
	Session   *DriverSession
}

type BrowserInfo struct {
	ContextId string `json:"context_id"`
}

func CreateBrowser(timeout int, width, height int) (browserInfo *BrowserInfo, err error) {
	data := map[string]interface{}{
		"timeout": timeout,
	}

	if width > 0 && height > 0 {
		data["size"] = map[string]int{
			"width":  width,
			"height": height,
		}
	}

	var bsJSON []byte = nil
	if bsJSON, err = json.Marshal(data); err != nil {
		return nil, err
	}

	rawURL := "http://" + BROWSER_LOCAL_ADDRESS + "/api/v1/create_browser"
	req, err := http.NewRequest(http.MethodPost, rawURL, bytes.NewBuffer(bsJSON))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	rawResp, _ := io.ReadAll(resp.Body)
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

func NewBrowserDriver(device *BrowserDevice) (driver *BrowserDriver, err error) {
	log.Info().Msg("init NewBrowserDriver driver")
	driver = new(BrowserDriver)
	driver.urlPrefix = &url.URL{}
	driver.urlPrefix.Host = BROWSER_LOCAL_ADDRESS
	driver.urlPrefix.Scheme = "http"
	driver.sessionId = device.UUID()
	driver.Session = NewDriverSession()
	driver.Session.ID = driver.sessionId
	return driver, nil
}

func (wd *BrowserDriver) Setup() error {
	err := wd.Session.SetupPortForward(8093)
	if err != nil {
		return err
	}
	wd.Session.SetBaseURL(BROWSER_LOCAL_ADDRESS)
	return nil
}

func (wd *BrowserDriver) Drag(fromX, fromY, toX, toY float64, options ...option.ActionOption) (err error) {
	fromX, fromY, toX, toY, err = handlerDrag(wd, fromX, fromY, toX, toY, options...)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"from_x": fromX,
		"from_y": fromY,
		"to_x":   toX,
		"to_y":   toY,
	}
	actionOptions := option.NewActionOptions(options...)

	if actionOptions.Duration > 0 {
		data["duration"] = actionOptions.Duration
	} else {
		data["duration"] = 0.5
	}

	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/drag"))
	return
}

func (wd *BrowserDriver) AppLaunch(packageName string) (err error) {
	data := map[string]interface{}{
		"url": packageName,
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/page_launch"))
	return err
}

func (wd *BrowserDriver) DeleteSession() (err error) {
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

	rawResp, _ := io.ReadAll(resp.Body)
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

func (wd *BrowserDriver) Scroll(delta int) (err error) {
	data := map[string]interface{}{
		"delta": delta,
	}
	_, err = wd.HttpPOST(data, wd.sessionId, "ui/scroll")
	return err
}

func (wd *BrowserDriver) CreateNetListener() (*websocket.Conn, error) {
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
	return c, err
}

func (wd *BrowserDriver) ClosePage(pageIndex int) (err error) {
	data := map[string]interface{}{
		"page_index": pageIndex,
	}

	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/page_close"))
	return err
}

func (wd *BrowserDriver) HoverBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/hover"))
	return err
}

func (wd *BrowserDriver) TapBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/tap"))
	return err
}

func (wd *BrowserDriver) RightClick(x, y float64) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/right_click"))
	return err
}

func (wd *BrowserDriver) RightClickBySelector(selector string, options ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"selector": selector,
	}
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		data["element_index"] = actionOptions.Index
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/right_click"))
	return err
}

func (wd *BrowserDriver) GetElementTextBySelector(selector string, options ...option.ActionOption) (text string, err error) {
	actionOptions := option.NewActionOptions(options...)
	baseURL := fmt.Sprintf("http://%s/api/v1/%s/element_text", BROWSER_LOCAL_ADDRESS, wd.sessionId)

	// 使用 url.Values 构建查询参数
	params := url.Values{}
	params.Add("selector", selector)
	if actionOptions.Index > 0 {
		params.Add("element_index", fmt.Sprintf("%v", actionOptions.Index))
	}

	// 构建完整的 URL
	uri := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 使用 http.Get 直接请求
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result WebAgentResponse
	if err = json.Unmarshal(rawResp, &result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", errors.New(result.Message)
	}

	data := result.Data.(map[string]interface{})
	if data["text"] == nil {
		return "", nil
	}
	return data["text"].(string), nil
}

func (wd *BrowserDriver) GetPageUrl(options ...option.ActionOption) (text string, err error) {
	uri := "ui/page_url"
	actionOptions := option.NewActionOptions(options...)
	if actionOptions.Index > 0 {
		uri = uri + "?page_index=" + fmt.Sprintf("%v", actionOptions.Index)
	}
	resp, err := wd.Session.GET(wd.concatURL(wd.sessionId, uri))
	if err != nil {
		return "", err
	}
	data, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return "", err
	}
	data = data["data"].(map[string]interface{})
	return data["url"].(string), nil
}

func (wd *BrowserDriver) IsElementExistBySelector(selector string) (bool, error) {
	resp, err := wd.Session.GET(wd.concatURL("ui/element_exist", "?selector=", selector))
	if err != nil {
		return false, err
	}
	data, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return false, err
	}
	data = data["data"].(map[string]interface{})
	return data["exist"].(bool), nil
}

func (wd *BrowserDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (success bool, err error) {
	data := map[string]interface{}{
		"url":        packageName,
		"web_cookie": password,
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "stub/login"))
	if err != nil {
		return false, err
	}
	return true, err
}

func (wd *BrowserDriver) Hover(x, y float64) (err error) {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/hover"))
	return err
}

func (wd *BrowserDriver) Input(text string, option ...option.ActionOption) (err error) {
	data := map[string]interface{}{
		"text": text,
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/input"))
	return err
}

// Source Return application elements tree
func (wd *BrowserDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := wd.Session.GET(wd.concatURL(wd.sessionId, "stub/source"))
	if err != nil {
		return "", err
	}

	return resp.ValueConvertToString()
}

func (wd *BrowserDriver) ScreenShot(options ...option.ActionOption) (*bytes.Buffer, error) {
	resp, err := wd.Session.GET(wd.concatURL(wd.sessionId, "screenshot"))
	if err != nil {
		return nil, err
	}

	// 将结果解析为 JSON
	var result WebAgentResponse
	if err = json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		log.Info().Msgf("%v", result.Message)
		return nil, errors.New(result.Message)
	}

	data := result.Data.(map[string]interface{})
	screenshotBase64 := data["screenshot"].(string)
	screenRaw, err := base64.StdEncoding.DecodeString(screenshotBase64)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(screenRaw), nil
}

func (wd *BrowserDriver) HttpPOST(data interface{}, pathElem ...string) (response *WebAgentResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}

	return wd.httpRequest(http.MethodPost, wd.concatURL(pathElem...), bsJSON)
}

func (wd *BrowserDriver) HttpGet(data interface{}, pathElem ...string) (response *WebAgentResponse, err error) {
	return wd.httpRequest(http.MethodGet, wd.concatURL(pathElem...), nil)
}

func (wd *BrowserDriver) concatURL(elem ...string) string {
	tmp, _ := url.Parse(wd.urlPrefix.String())
	commonPath := path.Join(append([]string{wd.urlPrefix.Path}, "api/v1/")...)
	tmp.Path = path.Join(append([]string{commonPath}, elem...)...)
	return tmp.String()
}

func (wd *BrowserDriver) httpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (response *WebAgentResponse, err error) {
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

	rawResp, _ := io.ReadAll(resp.Body)
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

func (wd *BrowserDriver) Status() (deviceStatus types.DeviceStatus, err error) {
	log.Warn().Msg("Status not implemented in ADBDriver")
	return
}

func (wd *BrowserDriver) DeviceInfo() (deviceInfo types.DeviceInfo, err error) {
	log.Warn().Msg("DeviceInfo not implemented in ADBDriver")
	return
}

func (wd *BrowserDriver) BatteryInfo() (batteryInfo types.BatteryInfo, err error) {
	log.Warn().Msg("BatteryInfo not implemented in ADBDriver")
	return
}

func (wd *BrowserDriver) WindowSize() (types.Size, error) {
	resp, err := wd.Session.GET(wd.concatURL(wd.sessionId, "window_size"))
	if err != nil {
		return types.Size{}, err
	}

	data, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return types.Size{}, err
	}
	data = data["data"].(map[string]interface{})
	width := data["width"]
	height := data["height"]
	return types.Size{
		Width:  int(width.(float64)),
		Height: int(height.(float64)),
	}, nil
}

func (wd *BrowserDriver) Screen() (Screen, error) {
	return Screen{}, errors.New("not support")
}

func (wd *BrowserDriver) Scale() (float64, error) {
	return 1, nil
}

// GetTimestamp returns the timestamp of the mobile device
func (wd *BrowserDriver) GetTimestamp() (timestamp int64, err error) {
	return 0, errors.New("not support")
}

// Homescreen Forces the device under test to switch to the home screen
func (wd *BrowserDriver) Homescreen() error {
	return nil
}

func (wd *BrowserDriver) Unlock() (err error) {
	return nil
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (wd *BrowserDriver) AppTerminate(packageName string) (bool, error) {
	return true, wd.DeleteSession()
}

// AssertForegroundApp returns nil if the given package and activity are in foreground
func (wd *BrowserDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	return nil
}

func (wd *BrowserDriver) Back() error {
	return wd.PressBack()
}

func (wd *BrowserDriver) AppClear(packageName string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) PushImage(localPath string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) PullImages(localDir string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) ClearImages() error {
	return errors.New("not support")
}

func (wd *BrowserDriver) PushFile(localPath string, remoteDir string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) PullFiles(localDir string, remoteDirs ...string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) ClearFiles(paths ...string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) Orientation() (orientation types.Orientation, err error) {
	log.Warn().Msg("Orientation not implemented in ADBDriver")
	return
}

// Tap Sends a tap event at the coordinate.
func (wd *BrowserDriver) Tap(x, y float64, options ...option.ActionOption) error {
	return wd.TapFloat(x, y, options...)
}

func (wd *BrowserDriver) TapFloat(x, y float64, opts ...option.ActionOption) error {
	var err error
	x, y, err = handlerTapAbsXY(wd, x, y, opts...)
	if err != nil {
		return err
	}

	actionOptions := option.NewActionOptions(opts...)
	duration := 0.1
	if actionOptions.Duration > 0 {
		duration = actionOptions.Duration
	}

	data := map[string]interface{}{
		"x":        x,
		"y":        y,
		"duration": duration,
	}

	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/tap"))
	return err
}

// DoubleTap Sends a double tap event at the coordinate.
func (wd *BrowserDriver) DoubleTap(x, y float64, options ...option.ActionOption) error {
	var err error
	x, y, err = handlerDoubleTap(wd, x, y, options...)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}

	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/double_tap"))
	return err
}

func (wd *BrowserDriver) UploadFile(x, y float64, FileUrl, FileFormat string) (err error) {
	data := map[string]interface{}{
		"x":           x,
		"y":           y,
		"file_url":    FileUrl,
		"file_format": FileFormat,
	}
	_, err = wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/upload"))
	return err
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (wd *BrowserDriver) TouchAndHold(x, y float64, options ...option.ActionOption) error {
	return errors.New("not support")
}

// Swipe works like Drag, but `pressForDuration` value is 0
func (wd *BrowserDriver) Swipe(fromX, fromY, toX, toY float64, options ...option.ActionOption) error {
	log.Info().Msg(fmt.Sprintf("swipe from (%v, %v) to (%v, %v)", fromX, fromY, toX, toY))
	return wd.Drag(fromX, fromY, toX, toY, options...)
}

func (wd *BrowserDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...option.ActionOption) error {
	return wd.Drag(fromX, fromY, toX, toY, options...)
}

func (wd *BrowserDriver) SetIme(ime string) error {
	return errors.New("not support")
}

// SendKeys Types a string into active element. There must be element with keyboard focus,
// otherwise an error is raised.
// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
func (wd *BrowserDriver) SendKeys(text string, options ...option.ActionOption) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) Clear(packageName string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) GetDevice() IDevice {
	return nil
}

func (wd *BrowserDriver) ForegroundInfo() (app types.AppInfo, err error) {
	return
}

// PressBack Presses the back button
func (wd *BrowserDriver) PressBack(options ...option.ActionOption) error {
	_, err := wd.HttpPOST(map[string]interface{}{}, wd.sessionId, "ui/back")
	return err
}

func (wd *BrowserDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return errors.New("not support")
}

func (wd *BrowserDriver) Backspace(count int, options ...option.ActionOption) (err error) {
	return errors.New("not support")
}

func (wd *BrowserDriver) LogoutNoneUI(packageName string) error {
	return errors.New("not support")
}

func (wd *BrowserDriver) TapByText(text string, options ...option.ActionOption) error {
	return errors.New("not support")
}

// AccessibleSource Return application elements accessibility tree
func (wd *BrowserDriver) AccessibleSource() (string, error) {
	return "", errors.New("not support")
}

// HealthCheck Health check might modify simulator state so it should only be called in-between testing sessions
//
//	Checks health of XCTest by:
//	1) Querying application for some elements,
//	2) Triggering some device events.
func (wd *BrowserDriver) HealthCheck() error {
	return errors.New("not support")
}

func (wd *BrowserDriver) GetAppiumSettings() (map[string]interface{}, error) {
	return nil, errors.New("not support")
}

func (wd *BrowserDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	return nil, errors.New("not support")
}

func (wd *BrowserDriver) IsHealthy() (bool, error) {
	return false, errors.New("not support")
}

// triggers the log capture and returns the log entries
func (wd *BrowserDriver) StartCaptureLog(identifier ...string) (err error) {
	return errors.New("not support")
}

func (wd *BrowserDriver) StopCaptureLog() (result interface{}, err error) {
	return nil, nil
}

func (wd *BrowserDriver) RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error) {
	return "", errors.New("not support")
}

func (wd *BrowserDriver) TearDown() error {
	return nil
}

func (wd *BrowserDriver) InitSession(capabilities option.Capabilities) error {
	return nil
}

func (wd *BrowserDriver) GetSession() *DriverSession {
	return wd.Session
}

func (wd *BrowserDriver) ScreenRecord(opts ...option.ActionOption) (videoPath string, err error) {
	return
}

func (wd *BrowserDriver) Rotation() (rotation types.Rotation, err error) {
	return
}

func (wd *BrowserDriver) SetRotation(rotation types.Rotation) error {
	return nil
}

func (wd *BrowserDriver) Home() error {
	return nil
}

func (wd *BrowserDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err := wd.Session.POST(data, wd.concatURL(wd.sessionId, "ui/double_tap"))
	return err
}

func (wd *BrowserDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	return wd.TapFloat(x, y, opts...)
}
