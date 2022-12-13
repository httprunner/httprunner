package uixt

import (
	"bytes"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

type wdaDriver struct {
	Driver

	// default port
	defaultConn gidevice.InnerConn

	// mjpeg port
	mjpegUSBConn  gidevice.InnerConn // via USB
	mjpegHTTPConn net.Conn           // via HTTP
	mjpegClient   *http.Client
}

func (wd *wdaDriver) GetMjpegClient() *http.Client {
	return wd.mjpegClient
}

func (wd *wdaDriver) Close() error {
	if wd.defaultConn != nil {
		wd.defaultConn.Close()
	}
	if wd.mjpegUSBConn != nil {
		wd.mjpegUSBConn.Close()
	}

	if wd.mjpegClient != nil {
		wd.mjpegClient.CloseIdleConnections()
	}
	return wd.mjpegHTTPConn.Close()
}

func (wd *wdaDriver) NewSession(capabilities Capabilities) (sessionInfo SessionInfo, err error) {
	// [[FBRoute POST:@"/session"].withoutSession respondWithTarget:self action:@selector(handleCreateSession:)]
	data := make(map[string]interface{})
	if len(capabilities) == 0 {
		data["capabilities"] = make(map[string]interface{})
	} else {
		data["capabilities"] = map[string]interface{}{"alwaysMatch": capabilities}
	}

	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session"); err != nil {
		return SessionInfo{}, err
	}
	if sessionInfo, err = rawResp.valueConvertToSessionInfo(); err != nil {
		return SessionInfo{}, err
	}
	wd.sessionId = sessionInfo.SessionId
	return
}

func (wd *wdaDriver) ActiveSession() (sessionInfo SessionInfo, err error) {
	// [[FBRoute GET:@""] respondWithTarget:self action:@selector(handleGetActiveSession:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId); err != nil {
		return SessionInfo{}, err
	}
	if sessionInfo, err = rawResp.valueConvertToSessionInfo(); err != nil {
		return SessionInfo{}, err
	}
	return
}

func (wd *wdaDriver) DeleteSession() (err error) {
	// [[FBRoute DELETE:@""] respondWithTarget:self action:@selector(handleDeleteSession:)]
	_, err = wd.httpDELETE("/session", wd.sessionId)
	return
}

func (wd *wdaDriver) Status() (deviceStatus DeviceStatus, err error) {
	// [[FBRoute GET:@"/status"].withoutSession respondWithTarget:self action:@selector(handleGetStatus:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/status"); err != nil {
		return DeviceStatus{}, err
	}
	reply := new(struct{ Value struct{ DeviceStatus } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return DeviceStatus{}, err
	}
	deviceStatus = reply.Value.DeviceStatus
	return
}

func (wd *wdaDriver) DeviceInfo() (deviceInfo DeviceInfo, err error) {
	// [[FBRoute GET:@"/wda/device/info"] respondWithTarget:self action:@selector(handleGetDeviceInfo:)]
	// [[FBRoute GET:@"/wda/device/info"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/device/info"); err != nil {
		return DeviceInfo{}, err
	}
	reply := new(struct{ Value struct{ DeviceInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return DeviceInfo{}, err
	}
	deviceInfo = reply.Value.DeviceInfo
	return
}

func (wd *wdaDriver) Location() (location Location, err error) {
	// [[FBRoute GET:@"/wda/device/location"] respondWithTarget:self action:@selector(handleGetLocation:)]
	// [[FBRoute GET:@"/wda/device/location"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/device/location"); err != nil {
		return Location{}, err
	}
	reply := new(struct{ Value struct{ Location } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Location{}, err
	}
	location = reply.Value.Location
	return
}

func (wd *wdaDriver) BatteryInfo() (batteryInfo BatteryInfo, err error) {
	// [[FBRoute GET:@"/wda/batteryInfo"] respondWithTarget:self action:@selector(handleGetBatteryInfo:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/batteryInfo"); err != nil {
		return BatteryInfo{}, err
	}
	reply := new(struct{ Value struct{ BatteryInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return BatteryInfo{}, err
	}
	batteryInfo = reply.Value.BatteryInfo
	return
}

func (wd *wdaDriver) WindowSize() (size Size, err error) {
	// [[FBRoute GET:@"/window/size"] respondWithTarget:self action:@selector(handleGetWindowSize:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/window/size"); err != nil {
		return Size{}, err
	}
	reply := new(struct{ Value struct{ Size } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Size{}, err
	}
	size = reply.Value.Size
	return
}

func (wd *wdaDriver) Screen() (screen Screen, err error) {
	// [[FBRoute GET:@"/wda/screen"] respondWithTarget:self action:@selector(handleGetScreen:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/screen"); err != nil {
		return Screen{}, err
	}
	reply := new(struct{ Value struct{ Screen } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Screen{}, err
	}
	screen = reply.Value.Screen
	return
}

func (wd *wdaDriver) Scale() (float64, error) {
	screen, err := wd.Screen()
	if err != nil {
		return 0, err
	}
	return screen.Scale, nil
}

func (wd *wdaDriver) ActiveAppInfo() (info AppInfo, err error) {
	// [[FBRoute GET:@"/wda/activeAppInfo"] respondWithTarget:self action:@selector(handleActiveAppInfo:)]
	// [[FBRoute GET:@"/wda/activeAppInfo"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/activeAppInfo"); err != nil {
		return AppInfo{}, err
	}
	reply := new(struct{ Value struct{ AppInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return AppInfo{}, err
	}
	info = reply.Value.AppInfo
	return
}

func (wd *wdaDriver) ActiveAppsList() (appsList []AppBaseInfo, err error) {
	// [[FBRoute GET:@"/wda/apps/list"] respondWithTarget:self action:@selector(handleGetActiveAppsList:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/apps/list"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value []AppBaseInfo })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	appsList = reply.Value
	return
}

func (wd *wdaDriver) AppState(bundleId string) (runState AppState, err error) {
	// [[FBRoute POST:@"/wda/apps/state"] respondWithTarget:self action:@selector(handleSessionAppState:)]
	data := map[string]interface{}{"bundleId": bundleId}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/apps/state"); err != nil {
		return 0, err
	}
	reply := new(struct{ Value AppState })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return 0, err
	}
	runState = reply.Value
	_ = rawResp
	return
}

func (wd *wdaDriver) IsLocked() (locked bool, err error) {
	// [[FBRoute GET:@"/wda/locked"] respondWithTarget:self action:@selector(handleIsLocked:)]
	// [[FBRoute GET:@"/wda/locked"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/locked"); err != nil {
		return false, err
	}
	if locked, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (wd *wdaDriver) Unlock() (err error) {
	// [[FBRoute POST:@"/wda/unlock"] respondWithTarget:self action:@selector(handleUnlock:)]
	// [[FBRoute POST:@"/wda/unlock"].withoutSession
	_, err = wd.httpPOST(nil, "/session", wd.sessionId, "/wda/unlock")
	return
}

func (wd *wdaDriver) Lock() (err error) {
	// [[FBRoute POST:@"/wda/lock"] respondWithTarget:self action:@selector(handleLock:)]
	// [[FBRoute POST:@"/wda/lock"].withoutSession
	_, err = wd.httpPOST(nil, "/session", wd.sessionId, "/wda/lock")
	return
}

func (wd *wdaDriver) Homescreen() (err error) {
	// [[FBRoute POST:@"/wda/homescreen"].withoutSession respondWithTarget:self action:@selector(handleHomescreenCommand:)]
	_, err = wd.httpPOST(nil, "/wda/homescreen")
	return
}

func (wd *wdaDriver) AlertText() (text string, err error) {
	// [[FBRoute GET:@"/alert/text"] respondWithTarget:self action:@selector(handleAlertGetTextCommand:)]
	// [[FBRoute GET:@"/alert/text"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/alert/text"); err != nil {
		return "", err
	}
	if text, err = rawResp.valueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (wd *wdaDriver) AlertButtons() (btnLabels []string, err error) {
	// [[FBRoute GET:@"/wda/alert/buttons"] respondWithTarget:self action:@selector(handleGetAlertButtonsCommand:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/alert/buttons"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value []string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	btnLabels = reply.Value
	return
}

func (wd *wdaDriver) AlertAccept(label ...string) (err error) {
	// [[FBRoute POST:@"/alert/accept"] respondWithTarget:self action:@selector(handleAlertAcceptCommand:)]
	// [[FBRoute POST:@"/alert/accept"].withoutSession
	data := make(map[string]interface{})
	if len(label) != 0 && label[0] != "" {
		data["name"] = label[0]
	}
	_, err = wd.httpPOST(data, "/alert/accept")
	return
}

func (wd *wdaDriver) AlertDismiss(label ...string) (err error) {
	// [[FBRoute POST:@"/alert/dismiss"] respondWithTarget:self action:@selector(handleAlertDismissCommand:)]
	// [[FBRoute POST:@"/alert/dismiss"].withoutSession
	data := make(map[string]interface{})
	if len(label) != 0 && label[0] != "" {
		data["name"] = label[0]
	}
	_, err = wd.httpPOST(data, "/alert/dismiss")
	return
}

func (wd *wdaDriver) AlertSendKeys(text string) (err error) {
	// [[FBRoute POST:@"/alert/text"] respondWithTarget:self action:@selector(handleAlertSetTextCommand:)]
	data := map[string]interface{}{"value": strings.Split(text, "")}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/alert/text")
	return
}

func (wd *wdaDriver) AppLaunch(bundleId string, launchOpt ...AppLaunchOption) (err error) {
	// [[FBRoute POST:@"/wda/apps/launch"] respondWithTarget:self action:@selector(handleSessionAppLaunch:)]
	data := make(map[string]interface{})
	if len(launchOpt) != 0 {
		data = launchOpt[0]
	}
	data["bundleId"] = bundleId
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/apps/launch")
	return
}

func (wd *wdaDriver) AppLaunchUnattached(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/launchUnattached"].withoutSession respondWithTarget:self action:@selector(handleLaunchUnattachedApp:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.httpPOST(data, "/wda/apps/launchUnattached")
	return
}

func (wd *wdaDriver) AppTerminate(bundleId string) (successful bool, err error) {
	// [[FBRoute POST:@"/wda/apps/terminate"] respondWithTarget:self action:@selector(handleSessionAppTerminate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/apps/terminate"); err != nil {
		return false, err
	}
	if successful, err = rawResp.valueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (wd *wdaDriver) AppActivate(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/activate"] respondWithTarget:self action:@selector(handleSessionAppActivate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/apps/activate")
	return
}

func (wd *wdaDriver) AppDeactivate(second float64) (err error) {
	// [[FBRoute POST:@"/wda/deactivateApp"] respondWithTarget:self action:@selector(handleDeactivateAppCommand:)]
	if second < 3 {
		second = 3.0
	}
	data := map[string]interface{}{"duration": second}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/deactivateApp")
	return
}

func (wd *wdaDriver) AppAuthReset(resource ProtectedResource) (err error) {
	// [[FBRoute POST:@"/wda/resetAppAuth"] respondWithTarget:self action:@selector(handleResetAppAuth:)]
	data := map[string]interface{}{"resource": resource}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/resetAppAuth")
	return
}

func (wd *wdaDriver) Tap(x, y int, options ...DataOption) error {
	return wd.TapFloat(float64(x), float64(y), options...)
}

func (wd *wdaDriver) TapFloat(x, y float64, options ...DataOption) (err error) {
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	// new data options in post data for extra WDA configurations
	newData := NewData(data, options...)

	_, err = wd.httpPOST(newData, "/session", wd.sessionId, "/wda/tap/0")
	return
}

func (wd *wdaDriver) DoubleTap(x, y int) error {
	return wd.DoubleTapFloat(float64(x), float64(y))
}

func (wd *wdaDriver) DoubleTapFloat(x, y float64) (err error) {
	// [[FBRoute POST:@"/wda/doubleTap"] respondWithTarget:self action:@selector(handleDoubleTapCoordinate:)]
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/doubleTap")
	return
}

func (wd *wdaDriver) TouchAndHold(x, y int, second ...float64) error {
	return wd.TouchAndHoldFloat(float64(x), float64(y), second...)
}

func (wd *wdaDriver) TouchAndHoldFloat(x, y float64, second ...float64) (err error) {
	// [[FBRoute POST:@"/wda/touchAndHold"] respondWithTarget:self action:@selector(handleTouchAndHoldCoordinate:)]
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	if len(second) == 0 || second[0] <= 0 {
		second = []float64{1.0}
	}
	data["duration"] = second[0]
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/touchAndHold")
	return
}

func (wd *wdaDriver) Drag(fromX, fromY, toX, toY int, options ...DataOption) error {
	return wd.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (wd *wdaDriver) DragFloat(fromX, fromY, toX, toY float64, options ...DataOption) (err error) {
	// [[FBRoute POST:@"/wda/dragfromtoforduration"] respondWithTarget:self action:@selector(handleDragCoordinate:)]
	data := map[string]interface{}{
		"fromX": fromX,
		"fromY": fromY,
		"toX":   toX,
		"toY":   toY,
	}

	// new data options in post data for extra WDA configurations
	newData := NewData(data, options...)

	_, err = wd.httpPOST(newData, "/session", wd.sessionId, "/wda/dragfromtoforduration")
	return
}

func (wd *wdaDriver) Swipe(fromX, fromY, toX, toY int, options ...DataOption) error {
	return wd.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (wd *wdaDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...DataOption) error {
	return wd.DragFloat(fromX, fromY, toX, toY, options...)
}

func (wd *wdaDriver) ForceTouch(x, y int, pressure float64, second ...float64) error {
	return wd.ForceTouchFloat(float64(x), float64(y), pressure, second...)
}

func (wd *wdaDriver) ForceTouchFloat(x, y, pressure float64, second ...float64) error {
	if len(second) == 0 || second[0] <= 0 {
		second = []float64{1.0}
	}
	actions := NewTouchActions().
		Press(
			NewTouchActionPress().WithXYFloat(x, y).WithPressure(pressure)).
		Wait(second[0]).
		Release()
	return wd.PerformAppiumTouchActions(actions)
}

func (wd *wdaDriver) PerformW3CActions(actions *W3CActions) (err error) {
	// [[FBRoute POST:@"/actions"] respondWithTarget:self action:@selector(handlePerformW3CTouchActions:)]
	data := map[string]interface{}{"actions": actions}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/actions")
	return
}

func (wd *wdaDriver) PerformAppiumTouchActions(touchActs *TouchActions) (err error) {
	// [[FBRoute POST:@"/wda/touch/perform"] respondWithTarget:self action:@selector(handlePerformAppiumTouchActions:)]
	// [[FBRoute POST:@"/wda/touch/multi/perform"]
	data := map[string]interface{}{"actions": touchActs}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/touch/multi/perform")
	return
}

func (wd *wdaDriver) SetPasteboard(contentType PasteboardType, content string) (err error) {
	// [[FBRoute POST:@"/wda/setPasteboard"] respondWithTarget:self action:@selector(handleSetPasteboard:)]
	data := map[string]interface{}{
		"contentType": contentType,
		"content":     base64.StdEncoding.EncodeToString([]byte(content)),
	}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/setPasteboard")
	return
}

func (wd *wdaDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	// [[FBRoute POST:@"/wda/getPasteboard"] respondWithTarget:self action:@selector(handleGetPasteboard:)]
	data := map[string]interface{}{"contentType": contentType}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/getPasteboard"); err != nil {
		return nil, err
	}
	if raw, err = rawResp.valueDecodeAsBase64(); err != nil {
		return nil, err
	}
	return
}

func (wd *wdaDriver) SendKeys(text string, options ...DataOption) (err error) {
	// [[FBRoute POST:@"/wda/keys"] respondWithTarget:self action:@selector(handleKeys:)]
	data := map[string]interface{}{"value": strings.Split(text, "")}

	// new data options in post data for extra WDA configurations
	newData := NewData(data, options...)

	_, err = wd.httpPOST(newData, "/session", wd.sessionId, "/wda/keys")
	return
}

func (wd *wdaDriver) Input(text string, options ...DataOption) (err error) {
	return wd.SendKeys(text, options...)
}

func (wd *wdaDriver) KeyboardDismiss(keyNames ...string) (err error) {
	// [[FBRoute POST:@"/wda/keyboard/dismiss"] respondWithTarget:self action:@selector(handleDismissKeyboardCommand:)]
	if len(keyNames) == 0 {
		keyNames = []string{"return"}
	}
	data := map[string]interface{}{"keyNames": keyNames}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/keyboard/dismiss")
	return
}

// PressBack simulates a short press on the BACK button.
func (wd *wdaDriver) PressBack(options ...DataOption) (err error) {
	windowSize, err := wd.WindowSize()
	if err != nil {
		return
	}

	data := map[string]interface{}{
		"fromX": float64(windowSize.Width) * 0,
		"fromY": float64(windowSize.Height) * 0.5,
		"toX":   float64(windowSize.Width) * 0.6,
		"toY":   float64(windowSize.Height) * 0.5,
	}

	// new data options in post data for extra WDA configurations
	newData := NewData(data, options...)

	_, err = wd.httpPOST(newData, "/session", wd.sessionId, "/wda/dragfromtoforduration")
	return
}

func (wd *wdaDriver) PressButton(devBtn DeviceButton) (err error) {
	// [[FBRoute POST:@"/wda/pressButton"] respondWithTarget:self action:@selector(handlePressButtonCommand:)]
	data := map[string]interface{}{"name": devBtn}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/pressButton")
	return
}

func (wd *wdaDriver) IOHIDEvent(pageID EventPageID, usageID EventUsageID, duration ...float64) (err error) {
	// [[FBRoute POST:@"/wda/performIoHidEvent"] respondWithTarget:self action:@selector(handlePeformIOHIDEvent:)]
	if len(duration) == 0 || duration[0] <= 0 {
		duration = []float64{0.005}
	}
	data := map[string]interface{}{
		"page":     pageID,
		"usage":    usageID,
		"duration": duration[0],
	}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/performIoHidEvent")
	return
}

func (wd *wdaDriver) StartCamera() (err error) {
	// start camera, alias for app_launch com.apple.camera
	return wd.AppLaunch("com.apple.camera")
}

func (wd *wdaDriver) StopCamera() (err error) {
	// stop camera, alias for app_terminate com.apple.camera
	success, err := wd.AppTerminate("com.apple.camera")
	if err != nil {
		return errors.Wrap(err, "failed to terminate camera")
	}
	if !success {
		log.Warn().Msg("camera was not running")
	}
	return nil
}

func (wd *wdaDriver) ExpectNotification(notifyName string, notifyType NotificationType, second ...int) (err error) {
	// [[FBRoute POST:@"/wda/expectNotification"] respondWithTarget:self action:@selector(handleExpectNotification:)]
	if len(second) == 0 {
		second = []int{60}
	}
	data := map[string]interface{}{
		"name":    notifyName,
		"type":    notifyType,
		"timeout": second[0],
	}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/expectNotification")
	return
}

func (wd *wdaDriver) SiriActivate(text string) (err error) {
	// [[FBRoute POST:@"/wda/siri/activate"] respondWithTarget:self action:@selector(handleActivateSiri:)]
	data := map[string]interface{}{"text": text}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/siri/activate")
	return
}

func (wd *wdaDriver) SiriOpenUrl(url string) (err error) {
	// [[FBRoute POST:@"/url"] respondWithTarget:self action:@selector(handleOpenURL:)]
	data := map[string]interface{}{"url": url}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/url")
	return
}

func (wd *wdaDriver) Orientation() (orientation Orientation, err error) {
	// [[FBRoute GET:@"/orientation"] respondWithTarget:self action:@selector(handleGetOrientation:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/orientation"); err != nil {
		return "", err
	}
	reply := new(struct{ Value Orientation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	orientation = reply.Value
	return
}

func (wd *wdaDriver) SetOrientation(orientation Orientation) (err error) {
	// [[FBRoute POST:@"/orientation"] respondWithTarget:self action:@selector(handleSetOrientation:)]
	data := map[string]interface{}{"orientation": orientation}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/orientation")
	return
}

func (wd *wdaDriver) Rotation() (rotation Rotation, err error) {
	// [[FBRoute GET:@"/rotation"] respondWithTarget:self action:@selector(handleGetRotation:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/rotation"); err != nil {
		return Rotation{}, err
	}
	reply := new(struct{ Value Rotation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Rotation{}, err
	}
	rotation = reply.Value
	return
}

func (wd *wdaDriver) SetRotation(rotation Rotation) (err error) {
	// [[FBRoute POST:@"/rotation"] respondWithTarget:self action:@selector(handleSetRotation:)]
	_, err = wd.httpPOST(rotation, "/session", wd.sessionId, "/rotation")
	return
}

func (wd *wdaDriver) MatchTouchID(isMatch bool) (err error) {
	// [FBRoute POST:@"/wda/touch_id"]
	data := map[string]interface{}{"match": isMatch}
	_, err = wd.httpPOST(data, "/session", wd.sessionId, "/wda/touch_id")
	return
}

func (wd *wdaDriver) ActiveElement() (element WebElement, err error) {
	// [[FBRoute GET:@"/element/active"] respondWithTarget:self action:@selector(handleGetActiveElement:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/element/active"); err != nil {
		return nil, err
	}
	var elementID string
	if elementID, err = rawResp.valueConvertToElementID(); err != nil {
		return nil, err
	}
	element = &wdaElement{parent: wd, id: elementID}
	return
}

func (wd *wdaDriver) FindElement(by BySelector) (element WebElement, err error) {
	// [[FBRoute POST:@"/element"] respondWithTarget:self action:@selector(handleFindElement:)]
	using, value := by.getUsingAndValue()
	data := map[string]interface{}{
		"using": using,
		"value": value,
	}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.sessionId, "/element"); err != nil {
		return nil, err
	}
	var elementID string
	if elementID, err = rawResp.valueConvertToElementID(); err != nil {
		if errors.Is(err, errNoSuchElement) {
			return nil, fmt.Errorf("%w: unable to find an element using '%s', value '%s'", err, using, value)
		}
		return nil, err
	}
	element = &wdaElement{parent: wd, id: elementID}
	return
}

func (wd *wdaDriver) FindElements(by BySelector) (elements []WebElement, err error) {
	// [[FBRoute POST:@"/elements"] respondWithTarget:self action:@selector(handleFindElements:)]
	using, value := by.getUsingAndValue()
	data := map[string]interface{}{
		"using": using,
		"value": value,
	}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.sessionId, "/elements"); err != nil {
		return nil, err
	}
	var elementIDs []string
	if elementIDs, err = rawResp.valueConvertToElementIDs(); err != nil {
		if errors.Is(err, errNoSuchElement) {
			return nil, fmt.Errorf("%w: unable to find an element using '%s', value '%s'", err, using, value)
		}
		return nil, err
	}
	elements = make([]WebElement, len(elementIDs))
	for i := range elementIDs {
		elements[i] = &wdaElement{parent: wd, id: elementIDs[i]}
	}
	return
}

func (wd *wdaDriver) Screenshot() (raw *bytes.Buffer, err error) {
	// [[FBRoute GET:@"/screenshot"] respondWithTarget:self action:@selector(handleGetScreenshot:)]
	// [[FBRoute GET:@"/screenshot"].withoutSession respondWithTarget:self action:@selector(handleGetScreenshot:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/screenshot"); err != nil {
		return nil, errors.Wrap(code.IOSScreenShotError,
			fmt.Sprintf("get WDA screenshot data failed: %v", err))
	}

	if raw, err = rawResp.valueDecodeAsBase64(); err != nil {
		return nil, errors.Wrap(code.IOSScreenShotError,
			fmt.Sprintf("decode WDA screenshot data failed: %v", err))
	}
	return
}

func (wd *wdaDriver) Source(srcOpt ...SourceOption) (source string, err error) {
	// [[FBRoute GET:@"/source"] respondWithTarget:self action:@selector(handleGetSourceCommand:)]
	// [[FBRoute GET:@"/source"].withoutSession
	tmp, _ := url.Parse(wd.concatURL(nil, "/session", wd.sessionId))
	toJsonRaw := false
	if len(srcOpt) != 0 {
		q := tmp.Query()
		for k, val := range srcOpt[0] {
			v := val.(string)
			q.Set(k, v)
			if k == "format" && v == "json" {
				toJsonRaw = true
			}
		}
		tmp.RawQuery = q.Encode()
	}

	var rawResp rawResponse
	if rawResp, err = wd.httpRequest(http.MethodGet, wd.concatURL(tmp, "/source"), nil); err != nil {
		return "", nil
	}
	if toJsonRaw {
		var jr builtinJSON.RawMessage
		if jr, err = rawResp.valueConvertToJsonRawMessage(); err != nil {
			return "", err
		}
		return string(jr), nil
	}
	if source, err = rawResp.valueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (wd *wdaDriver) AccessibleSource() (source string, err error) {
	// [[FBRoute GET:@"/wda/accessibleSource"] respondWithTarget:self action:@selector(handleGetAccessibleSourceCommand:)]
	// [[FBRoute GET:@"/wda/accessibleSource"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/wda/accessibleSource"); err != nil {
		return "", err
	}
	var jr builtinJSON.RawMessage
	if jr, err = rawResp.valueConvertToJsonRawMessage(); err != nil {
		return "", err
	}
	source = string(jr)
	return
}

func (wd *wdaDriver) HealthCheck() (err error) {
	// [[FBRoute GET:@"/wda/healthcheck"].withoutSession respondWithTarget:self action:@selector(handleGetHealthCheck:)]
	_, err = wd.httpGET("/wda/healthcheck")
	return
}

func (wd *wdaDriver) GetAppiumSettings() (settings map[string]interface{}, err error) {
	// [[FBRoute GET:@"/appium/settings"] respondWithTarget:self action:@selector(handleGetSettings:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.sessionId, "/appium/settings"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	settings = reply.Value
	return
}

func (wd *wdaDriver) SetAppiumSettings(settings map[string]interface{}) (ret map[string]interface{}, err error) {
	// [[FBRoute POST:@"/appium/settings"] respondWithTarget:self action:@selector(handleSetSettings:)]
	data := map[string]interface{}{"settings": settings}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.sessionId, "/appium/settings"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	ret = reply.Value
	return
}

func (wd *wdaDriver) IsHealthy() (healthy bool, err error) {
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/health"); err != nil {
		return false, err
	}
	if string(rawResp) != "I-AM-ALIVE" {
		return false, nil
	}
	return true, nil
}

func (wd *wdaDriver) WdaShutdown() (err error) {
	_, err = wd.httpGET("/wda/shutdown")
	return
}

func (wd *wdaDriver) WaitWithTimeoutAndInterval(condition Condition, timeout, interval time.Duration) error {
	startTime := time.Now()
	for {
		done, err := condition(wd)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		if elapsed := time.Since(startTime); elapsed > timeout {
			return fmt.Errorf("timeout after %v", elapsed)
		}
		time.Sleep(interval)
	}
}

func (wd *wdaDriver) WaitWithTimeout(condition Condition, timeout time.Duration) error {
	return wd.WaitWithTimeoutAndInterval(condition, timeout, DefaultWaitInterval)
}

func (wd *wdaDriver) Wait(condition Condition) error {
	return wd.WaitWithTimeoutAndInterval(condition, DefaultWaitTimeout, DefaultWaitInterval)
}

func (wd *wdaDriver) triggerWDALog(data map[string]interface{}) (rawResp []byte, err error) {
	// [[FBRoute POST:@"/gtf/automation/log"].withoutSession respondWithTarget:self action:@selector(handleAutomationLog:)]
	return wd.httpPOST(data, "/gtf/automation/log")
}

func (wd *wdaDriver) StartCaptureLog(identifier ...string) error {
	log.Info().Msg("start WDA log recording")
	if identifier == nil {
		identifier = []string{""}
	}
	data := map[string]interface{}{"action": "start", "type": 2, "identifier": identifier[0]}
	_, err := wd.triggerWDALog(data)
	if err != nil {
		return errors.Wrap(code.IOSCaptureLogError,
			fmt.Sprintf("start WDA log recording failed: %v", err))
	}

	return nil
}

type wdaResponse struct {
	Value     interface{} `json:"value"`
	SessionID string      `json:"sessionId"`
}

func (wd *wdaDriver) StopCaptureLog() (result interface{}, err error) {
	log.Info().Msg("stop log recording")
	data := map[string]interface{}{"action": "stop"}
	rawResp, err := wd.triggerWDALog(data)
	if err != nil {
		log.Error().Err(err).Bytes("rawResp", rawResp).Msg("failed to get WDA logs")
		return "", errors.Wrap(code.IOSCaptureLogError,
			fmt.Sprintf("get WDA logs failed: %v", err))
	}
	reply := new(wdaResponse)
	if err = json.Unmarshal(rawResp, reply); err != nil {
		log.Error().Err(err).Bytes("rawResp", rawResp).Msg("failed to json.Unmarshal WDA logs")
		return reply, errors.Wrap(code.IOSCaptureLogError,
			fmt.Sprintf("json.Unmarshal WDA logs failed: %v", err))
	}
	log.Info().Interface("value", reply.Value).Msg("get WDA log response")
	return reply.Value, nil
}
