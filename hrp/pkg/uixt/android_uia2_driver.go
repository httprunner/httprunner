package uixt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/utf7"
)

var errDriverNotImplemented = errors.New("driver method not implemented")

type uiaDriver struct {
	adbDriver
}

func NewUIADriver(capabilities Capabilities, urlPrefix string) (driver *uiaDriver, err error) {
	log.Info().Msg("init uiautomator2 driver")
	if capabilities == nil {
		capabilities = NewCapabilities()
		capabilities.WithWaitForIdleTimeout(0)
	}
	driver = new(uiaDriver)
	if driver.urlPrefix, err = url.Parse(urlPrefix); err != nil {
		return nil, err
	}
	var localPort int
	{
		tmpURL, _ := url.Parse(driver.urlPrefix.String())
		hostname := tmpURL.Hostname()
		if strings.HasPrefix(hostname, forwardToPrefix) {
			localPort, _ = strconv.Atoi(strings.TrimPrefix(hostname, forwardToPrefix))
		}
	}
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return nil, fmt.Errorf("adb forward: %w", err)
	}
	driver.client = convertToHTTPClient(conn)

	session, err := driver.NewSession(capabilities)
	if err != nil {
		return nil, errors.Wrap(err, "create UIAutomator session failed")
	}
	driver.sessionId = session.SessionId
	return driver, nil
}

type BatteryStatus int

const (
	_                                  = iota
	BatteryStatusUnknown BatteryStatus = iota
	BatteryStatusCharging
	BatteryStatusDischarging
	BatteryStatusNotCharging
	BatteryStatusFull
)

func (bs BatteryStatus) String() string {
	switch bs {
	case BatteryStatusUnknown:
		return "unknown"
	case BatteryStatusCharging:
		return "charging"
	case BatteryStatusDischarging:
		return "discharging"
	case BatteryStatusNotCharging:
		return "not charging"
	case BatteryStatusFull:
		return "full"
	default:
		return fmt.Sprintf("unknown status code (%d)", bs)
	}
}

func (ud *uiaDriver) resetDriver() error {
	newUIADriver, err := NewUIADriver(NewCapabilities(), ud.urlPrefix.String())
	if err != nil {
		return err
	}
	ud.client = newUIADriver.client
	ud.sessionId = newUIADriver.sessionId
	return nil
}

func (ud *uiaDriver) httpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (rawResp rawResponse, err error) {
	disableRetryBool := len(disableRetry) > 0 && disableRetry[0]
	for retryCount := 1; retryCount <= 5; retryCount++ {
		rawResp, err = ud.Driver.httpRequest(method, rawURL, rawBody)
		if err == nil || disableRetryBool {
			return
		}
		// wait for UIA2 server to resume automatically
		time.Sleep(3 * time.Second)
		oldSessionID := ud.sessionId
		if err2 := ud.resetDriver(); err2 != nil {
			log.Err(err2).Msgf("failed to reset uia2 driver, retry count: %v", retryCount)
			continue
		}
		log.Debug().Str("new session", ud.sessionId).Str("old session", oldSessionID).Msgf("successful to reset uia2 driver, retry count: %v", retryCount)
		if oldSessionID != "" {
			rawURL = strings.Replace(rawURL, oldSessionID, ud.sessionId, 1)
		}
	}
	return
}

func (ud *uiaDriver) httpGET(pathElem ...string) (rawResp rawResponse, err error) {
	return ud.httpRequest(http.MethodGet, ud.concatURL(nil, pathElem...), nil)
}

func (ud *uiaDriver) httpGETWithRetry(pathElem ...string) (rawResp rawResponse, err error) {
	return ud.httpRequest(http.MethodGet, ud.concatURL(nil, pathElem...), nil, true)
}

func (ud *uiaDriver) httpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return ud.httpRequest(http.MethodPost, ud.concatURL(nil, pathElem...), bsJSON)
}

func (ud *uiaDriver) httpDELETE(pathElem ...string) (rawResp rawResponse, err error) {
	return ud.httpRequest(http.MethodDelete, ud.concatURL(nil, pathElem...), nil)
}

func (ud *uiaDriver) NewSession(capabilities Capabilities) (sessionInfo SessionInfo, err error) {
	// register(postHandler, new NewSession("/wd/hub/session"))
	var rawResp rawResponse
	data := make(map[string]interface{})
	if len(capabilities) == 0 {
		data["capabilities"] = make(map[string]interface{})
	} else {
		data["capabilities"] = map[string]interface{}{"alwaysMatch": capabilities}
	}
	if rawResp, err = ud.Driver.httpPOST(data, "/session"); err != nil {
		return SessionInfo{SessionId: ""}, err
	}
	reply := new(struct{ Value struct{ SessionId string } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return SessionInfo{SessionId: ""}, err
	}
	sessionID := reply.Value.SessionId
	// d.sessionIdCache[sessionID] = true
	return SessionInfo{SessionId: sessionID}, nil
}

func (ud *uiaDriver) DeleteSession() (err error) {
	if ud.sessionId == "" {
		return nil
	}
	if _, err = ud.httpDELETE("/session", ud.sessionId); err == nil {
		ud.sessionId = ""
	}

	return err
}

func (ud *uiaDriver) Status() (deviceStatus DeviceStatus, err error) {
	// register(getHandler, new Status("/wd/hub/status"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/status"); err != nil {
		return DeviceStatus{Ready: false}, err
	}
	reply := new(struct {
		Value struct {
			// Message string
			Ready bool
		}
	})
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return DeviceStatus{Ready: false}, err
	}
	return DeviceStatus{Ready: true}, nil
}

func (ud *uiaDriver) DeviceInfo() (deviceInfo DeviceInfo, err error) {
	// register(getHandler, new GetDeviceInfo("/wd/hub/session/:sessionId/appium/device/info"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "appium/device/info"); err != nil {
		return DeviceInfo{}, err
	}
	reply := new(struct{ Value struct{ DeviceInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return DeviceInfo{}, err
	}
	deviceInfo = reply.Value.DeviceInfo
	return
}

func (ud *uiaDriver) BatteryInfo() (batteryInfo BatteryInfo, err error) {
	// register(getHandler, new GetBatteryInfo("/wd/hub/session/:sessionId/appium/device/battery_info"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "appium/device/battery_info"); err != nil {
		return BatteryInfo{}, err
	}
	reply := new(struct{ Value struct{ BatteryInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return BatteryInfo{}, err
	}
	if reply.Value.Level == -1 || reply.Value.Status == -1 {
		return reply.Value.BatteryInfo, errors.New("cannot be retrieved from the system")
	}
	batteryInfo = reply.Value.BatteryInfo
	return
}

func (ud *uiaDriver) WindowSize() (size Size, err error) {
	// register(getHandler, new GetDeviceSize("/wd/hub/session/:sessionId/window/:windowHandle/size"))
	if ud.windowSize != nil {
		size = *ud.windowSize
	} else {
		var rawResp rawResponse
		if rawResp, err = ud.httpGET("/session", ud.sessionId, "window/:windowHandle/size"); err != nil {
			return Size{}, errors.Wrap(err, "get window size failed with uiautomator2")
		}
		reply := new(struct{ Value struct{ Size } })
		if err = json.Unmarshal(rawResp, reply); err != nil {
			return Size{}, err
		}
		size = reply.Value.Size
		ud.windowSize = &size
	}

	// check orientation
	orientation, err := ud.Orientation()
	if err != nil {
		log.Warn().Err(err).Msgf("window size get orientation failed, use default orientation")
		orientation = OrientationPortrait
	}
	if orientation != OrientationPortrait {
		size.Width, size.Height = size.Height, size.Width
	}
	return
}

// PressBack simulates a short press on the BACK button.
func (ud *uiaDriver) PressBack(options ...ActionOption) (err error) {
	// register(postHandler, new PressBack("/wd/hub/session/:sessionId/back"))
	_, err = ud.httpPOST(nil, "/session", ud.sessionId, "back")
	return
}

func (ud *uiaDriver) Homescreen() (err error) {
	return ud.PressKeyCodes(KCHome, KMEmpty)
}

func (ud *uiaDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return ud.PressKeyCodes(keyCode, KMEmpty)
}

func (ud *uiaDriver) PressKeyCodes(keyCode KeyCode, metaState KeyMeta, flags ...KeyFlag) (err error) {
	// register(postHandler, new PressKeyCodeAsync("/wd/hub/session/:sessionId/appium/device/press_keycode"))
	data := map[string]interface{}{
		"keycode": keyCode,
	}
	if metaState != KMEmpty {
		data["metastate"] = metaState
	}
	if len(flags) != 0 {
		data["flags"] = flags[0]
	}
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "appium/device/press_keycode")
	return
}

func (ud *uiaDriver) Orientation() (orientation Orientation, err error) {
	// [[FBRoute GET:@"/orientation"] respondWithTarget:self action:@selector(handleGetOrientation:)]
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "/orientation"); err != nil {
		return "", err
	}
	reply := new(struct{ Value Orientation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	orientation = reply.Value
	return
}

func (ud *uiaDriver) DoubleTap(x, y int) error {
	return ud.DoubleFloatTap(float64(x), float64(y))
}

func (ud *uiaDriver) DoubleFloatTap(x, y float64) error {
	data := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"type":       "pointer",
				"parameters": map[string]string{"pointerType": "touch"},
				"id":         "touch",
				"actions": []interface{}{
					map[string]interface{}{"type": "pointerMove", "duration": 0, "x": x, "y": y, "origin": "viewport"},
					map[string]interface{}{"type": "pointerDown", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pointerUp", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pointerDown", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pointerUp", "duration": 0, "button": 0},
				},
			},
		},
	}

	_, err := ud.httpPOST(data, "/session", ud.sessionId, "actions/tap")
	return err
}

func (ud *uiaDriver) Tap(x, y int, options ...ActionOption) error {
	return ud.TapFloat(float64(x), float64(y), options...)
}

func (ud *uiaDriver) TapFloat(x, y float64, options ...ActionOption) (err error) {
	// register(postHandler, new Tap("/wd/hub/session/:sessionId/appium/tap"))
	actionOptions := NewActionOptions(options...)

	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}
	x += actionOptions.getRandomOffset()
	y += actionOptions.getRandomOffset()

	duration := 100.0
	if actionOptions.PressDuration > 0 {
		duration = actionOptions.PressDuration * 1000
	}
	data := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"type":       "pointer",
				"parameters": map[string]string{"pointerType": "touch"},
				"id":         "touch",
				"actions": []interface{}{
					map[string]interface{}{"type": "pointerMove", "duration": 0, "x": x, "y": y, "origin": "viewport"},
					map[string]interface{}{"type": "pointerDown", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pause", "duration": duration},
					map[string]interface{}{"type": "pointerUp", "duration": 0, "button": 0},
				},
			},
		},
	}

	// update data options in post data for extra uiautomator configurations
	actionOptions.updateData(data)

	_, err = ud.httpPOST(data, "/session", ud.sessionId, "actions/tap")
	return err
}

func (ud *uiaDriver) TouchAndHold(x, y int, second ...float64) (err error) {
	return ud.TouchAndHoldFloat(float64(x), float64(y), second...)
}

func (ud *uiaDriver) TouchAndHoldFloat(x, y float64, second ...float64) (err error) {
	if len(second) == 0 {
		second = []float64{1.0}
	}
	// register(postHandler, new TouchLongClick("/wd/hub/session/:sessionId/touch/longclick"))
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"x":        x,
			"y":        y,
			"duration": int(second[0] * 1000),
		},
	}
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "touch/longclick")
	return
}

// Drag performs a swipe from one coordinate to another coordinate. You can control
// the smoothness and speed of the swipe by specifying the number of steps.
// Each step execution is throttled to 5 milliseconds per step, so for a 100
// steps, the swipe will take around 0.5 seconds to complete.
func (ud *uiaDriver) Drag(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return ud.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ud *uiaDriver) DragFloat(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	actionOptions := NewActionOptions(options...)
	if len(actionOptions.Offset) == 4 {
		fromX += float64(actionOptions.Offset[0])
		fromY += float64(actionOptions.Offset[1])
		toX += float64(actionOptions.Offset[2])
		toY += float64(actionOptions.Offset[3])
	}
	fromX += actionOptions.getRandomOffset()
	fromY += actionOptions.getRandomOffset()
	toX += actionOptions.getRandomOffset()
	toY += actionOptions.getRandomOffset()

	data := map[string]interface{}{
		"startX": fromX,
		"startY": fromY,
		"endX":   toX,
		"endY":   toY,
	}

	// update data options in post data for extra uiautomator configurations
	actionOptions.updateData(data)

	// register(postHandler, new Drag("/wd/hub/session/:sessionId/touch/drag"))
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "touch/drag")
	return
}

// Swipe performs a swipe from one coordinate to another using the number of steps
// to determine smoothness and speed. Each step execution is throttled to 5ms
// per step. So for a 100 steps, the swipe will take about 1/2 second to complete.
//
//	`steps` is the number of move steps sent to the system
func (ud *uiaDriver) Swipe(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return ud.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ud *uiaDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	// register(postHandler, new Swipe("/wd/hub/session/:sessionId/touch/perform"))
	actionOptions := NewActionOptions(options...)
	if len(actionOptions.Offset) == 4 {
		fromX += float64(actionOptions.Offset[0])
		fromY += float64(actionOptions.Offset[1])
		toX += float64(actionOptions.Offset[2])
		toY += float64(actionOptions.Offset[3])
	}
	fromX += actionOptions.getRandomOffset()
	fromY += actionOptions.getRandomOffset()
	toX += actionOptions.getRandomOffset()
	toY += actionOptions.getRandomOffset()

	duration := 200.0
	if actionOptions.PressDuration > 0 {
		duration = actionOptions.PressDuration * 1000
	}
	data := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"type":       "pointer",
				"parameters": map[string]string{"pointerType": "touch"},
				"id":         "touch",
				"actions": []interface{}{
					map[string]interface{}{"type": "pointerMove", "duration": 0, "x": fromX, "y": fromY, "origin": "viewport"},
					map[string]interface{}{"type": "pointerDown", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pointerMove", "duration": duration, "x": toX, "y": toY, "origin": "viewport"},
					map[string]interface{}{"type": "pointerUp", "duration": 0, "button": 0},
				},
			},
		},
	}

	// update data options in post data for extra uiautomator configurations
	actionOptions.updateData(data)

	_, err := ud.httpPOST(data, "/session", ud.sessionId, "actions/swipe")
	return err
}

func (ud *uiaDriver) SetPasteboard(contentType PasteboardType, content string) (err error) {
	lbl := content

	const defaultLabelLen = 10
	if len(lbl) > defaultLabelLen {
		lbl = lbl[:defaultLabelLen]
	}

	data := map[string]interface{}{
		"contentType": contentType,
		"label":       lbl,
		"content":     base64.StdEncoding.EncodeToString([]byte(content)),
	}
	// register(postHandler, new SetClipboard("/wd/hub/session/:sessionId/appium/device/set_clipboard"))
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "appium/device/set_clipboard")
	return
}

func (ud *uiaDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	if len(contentType) == 0 {
		contentType = PasteboardTypePlaintext
	}
	// register(postHandler, new GetClipboard("/wd/hub/session/:sessionId/appium/device/get_clipboard"))
	data := map[string]interface{}{
		"contentType": contentType[0],
	}
	var rawResp rawResponse
	if rawResp, err = ud.httpPOST(data, "/session", ud.sessionId, "appium/device/get_clipboard"); err != nil {
		return
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return
	}

	if data, err := base64.StdEncoding.DecodeString(reply.Value); err != nil {
		raw.Write([]byte(reply.Value))
	} else {
		raw.Write(data)
	}
	return
}

// SendKeys Android input does not support setting frequency.
func (ud *uiaDriver) SendKeys(text string, options ...ActionOption) (err error) {
	// register(postHandler, new SendKeysToElement("/wd/hub/session/:sessionId/keys"))
	// https://github.com/appium/appium-uiautomator2-server/blob/master/app/src/main/java/io/appium/uiautomator2/handler/SendKeysToElement.java#L76-L85
	actionOptions := NewActionOptions(options...)
	err = ud.SendUnicodeKeys(text, options...)
	if err != nil {
		data := map[string]interface{}{
			"text": text,
		}

		// new data options in post data for extra uiautomator configurations
		actionOptions.updateData(data)

		_, err = ud.httpPOST(data, "/session", ud.sessionId, "/keys")
	}
	return
}

func (ud *uiaDriver) SendUnicodeKeys(text string, options ...ActionOption) (err error) {
	// If the Unicode IME is not installed, fall back to the old interface.
	// There might be differences in the tracking schemes across different phones, and it is pending further verification.
	// In release version: without the Unicode IME installed, the test cannot execute.
	if !ud.IsUnicodeIMEInstalled() {
		return fmt.Errorf("appium unicode ime not installed")
	}
	currentIme, err := ud.adbDriver.GetIme()
	if err != nil {
		return
	}
	if currentIme != UnicodeImePackageName {
		defer func() {
			_ = ud.adbDriver.SetIme(currentIme)
		}()
		err = ud.adbDriver.SetIme(UnicodeImePackageName)
		if err != nil {
			log.Warn().Err(err).Msgf("set Unicode Ime failed")
			return
		}
	}
	encodedStr, err := utf7.Encoding.NewEncoder().String(text)
	if err != nil {
		log.Warn().Err(err).Msgf("encode text with modified utf7 failed")
		return
	}
	err = ud.SendActionKey(encodedStr, options...)
	return
}

func (ud *uiaDriver) SendActionKey(text string, options ...ActionOption) (err error) {
	actionOptions := NewActionOptions(options...)
	var actions []interface{}
	for i, c := range text {
		actions = append(actions, map[string]interface{}{"type": "keyDown", "value": string(c)},
			map[string]interface{}{"type": "keyUp", "value": string(c)})
		if i != len(text)-1 {
			actions = append(actions, map[string]interface{}{"type": "pause", "duration": 40})
		}
	}

	data := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"type":    "key",
				"id":      "key",
				"actions": actions,
			},
		},
	}

	// new data options in post data for extra uiautomator configurations
	actionOptions.updateData(data)
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "/actions/keys")
	return
}

func (ud *uiaDriver) Input(text string, options ...ActionOption) (err error) {
	return ud.SendKeys(text, options...)
}

func (ud *uiaDriver) Rotation() (rotation Rotation, err error) {
	// register(getHandler, new GetRotation("/wd/hub/session/:sessionId/rotation"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "rotation"); err != nil {
		return Rotation{}, err
	}
	reply := new(struct{ Value Rotation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Rotation{}, err
	}

	rotation = reply.Value
	return
}

func (ud *uiaDriver) Screenshot() (raw *bytes.Buffer, err error) {
	// https://bytedance.larkoffice.com/docx/C8qEdmSHnoRvMaxZauocMiYpnLh
	// ui2截图受内存影响，改为adb截图
	return ud.adbDriver.Screenshot()
}

func (ud *uiaDriver) Source(srcOpt ...SourceOption) (source string, err error) {
	// register(getHandler, new Source("/wd/hub/session/:sessionId/source"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "source"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}

	source = reply.Value
	return
}

func (ud *uiaDriver) sourceTree(srcOpt ...SourceOption) (sourceTree *Hierarchy, err error) {
	source, err := ud.Source()
	if err != nil {
		return
	}
	sourceTree = new(Hierarchy)
	err = xml.Unmarshal([]byte(source), sourceTree)
	if err != nil {
		return
	}
	return
}

func (ud *uiaDriver) TapByText(text string, options ...ActionOption) error {
	sourceTree, err := ud.sourceTree()
	if err != nil {
		return err
	}
	return ud.tapByTextUsingHierarchy(sourceTree, text, options...)
}

func (ud *uiaDriver) TapByTexts(actions ...TapTextAction) error {
	sourceTree, err := ud.sourceTree()
	if err != nil {
		return err
	}

	for _, action := range actions {
		err := ud.tapByTextUsingHierarchy(sourceTree, action.Text, action.Options...)
		if err != nil {
			return err
		}
	}
	return nil
}
