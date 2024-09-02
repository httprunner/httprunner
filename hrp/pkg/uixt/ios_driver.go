package uixt

import (
	"bytes"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
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

func (wd *wdaDriver) resetSession() error {
	capabilities := NewCapabilities()
	capabilities.WithDefaultAlertAction(AlertActionAccept)

	_, err := wd.NewSession(capabilities)
	if err != nil {
		return err
	}
	return nil
}

func (wd *wdaDriver) httpRequest(method string, rawURL string, rawBody []byte, disableRetry ...bool) (rawResp rawResponse, err error) {
	disableRetryBool := len(disableRetry) > 0 && disableRetry[0]
	for retryCount := 1; retryCount <= 5; retryCount++ {
		rawResp, err = wd.Driver.httpRequest(method, rawURL, rawBody)
		if err == nil || disableRetryBool {
			return
		}
		// TODO: polling WDA to check if resumed automatically
		time.Sleep(5 * time.Second)
		oldSessionID := wd.session.ID
		if err2 := wd.resetSession(); err2 != nil {
			log.Err(err2).Msgf("failed to reset wda driver, retry count: %v", retryCount)
			continue
		}
		log.Debug().Str("new session", wd.session.ID).Str("old session", oldSessionID).Msgf("successful to reset wda driver, retry count: %v", retryCount)
		if oldSessionID != "" {
			rawURL = strings.Replace(rawURL, oldSessionID, wd.session.ID, 1)
		}
	}
	return
}

func (wd *wdaDriver) httpGET(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.httpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil)
}

func (wd *wdaDriver) httpGETWithRetry(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.httpRequest(http.MethodGet, wd.concatURL(nil, pathElem...), nil, true)
}

func (wd *wdaDriver) httpPOST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.httpRequest(http.MethodPost, wd.concatURL(nil, pathElem...), bsJSON)
}

func (wd *wdaDriver) httpDELETE(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.httpRequest(http.MethodDelete, wd.concatURL(nil, pathElem...), nil)
}

func (wd *wdaDriver) GetMjpegClient() *http.Client {
	return wd.mjpegClient
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
	if rawResp, err = wd.Driver.httpPOST(data, "/session"); err != nil {
		return SessionInfo{}, err
	}
	if sessionInfo, err = rawResp.valueConvertToSessionInfo(); err != nil {
		return SessionInfo{}, err
	}
	wd.Driver.session.Init()
	wd.Driver.session.ID = sessionInfo.SessionId
	return
}

func (wd *wdaDriver) DeleteSession() (err error) {
	if wd.defaultConn != nil {
		wd.defaultConn.Close()
	}
	if wd.mjpegUSBConn != nil {
		wd.mjpegUSBConn.Close()
	}

	if wd.mjpegClient != nil {
		wd.mjpegClient.CloseIdleConnections()
	}
	if wd.mjpegHTTPConn != nil {
		wd.mjpegHTTPConn.Close()
	}

	// [[FBRoute DELETE:@""] respondWithTarget:self action:@selector(handleDeleteSession:)]
	_, err = wd.httpDELETE("/session", wd.session.ID)
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/device/info"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/device/location"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/batteryInfo"); err != nil {
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
	if wd.windowSize != nil {
		size = *wd.windowSize
	} else {
		var rawResp rawResponse
		if rawResp, err = wd.httpGET("/session", wd.session.ID, "/window/size"); err != nil {
			return Size{}, errors.Wrap(err, "get window size failed with wda")
		}
		reply := new(struct{ Value struct{ Size } })
		if err = json.Unmarshal(rawResp, reply); err != nil {
			return Size{}, err
		}
		size = reply.Value.Size
		scale, err := wd.Scale()
		if err != nil {
			return Size{}, errors.Wrap(err, "get window size scale failed")
		}
		size.Height = size.Height * int(scale)
		size.Width = size.Width * int(scale)
		wd.windowSize = &size
	}

	// check orientation
	orientation, err := wd.Orientation()
	if err != nil {
		log.Warn().Err(err).Msgf("window size get orientation failed, use default orientation")
		orientation = OrientationPortrait
	}
	if orientation != OrientationPortrait {
		size.Width, size.Height = size.Height, size.Width
	}
	return
}

func (wd *wdaDriver) Screen() (screen Screen, err error) {
	// [[FBRoute GET:@"/wda/screen"] respondWithTarget:self action:@selector(handleGetScreen:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/screen"); err != nil {
		return Screen{}, err
	}
	reply := new(struct{ Value struct{ Screen } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Screen{}, err
	}
	screen = reply.Value.Screen
	return
}

func (wd *wdaDriver) GetTimestamp() (timestamp int64, err error) {
	return 0, errors.Wrap(errDriverNotImplemented,
		"GetTimestamp not implemented for ios")
}

func (wd *wdaDriver) Scale() (float64, error) {
	if !builtin.IsZeroFloat64(wd.scale) {
		return wd.scale, nil
	}
	screen, err := wd.Screen()
	if err != nil {
		return 0, errors.Wrap(code.MobileUIDriverError,
			fmt.Sprintf("get screen info failed: %v", err))
	}
	return screen.Scale, nil
}

func (wd *wdaDriver) toScale(x float64) float64 {
	return x / wd.scale
}

func (wd *wdaDriver) ActiveAppInfo() (info AppInfo, err error) {
	// [[FBRoute GET:@"/wda/activeAppInfo"] respondWithTarget:self action:@selector(handleActiveAppInfo:)]
	// [[FBRoute GET:@"/wda/activeAppInfo"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/activeAppInfo"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/apps/list"); err != nil {
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
	if rawResp, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/apps/state"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/locked"); err != nil {
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
	_, err = wd.httpPOST(nil, "/session", wd.session.ID, "/wda/unlock")
	return
}

func (wd *wdaDriver) Lock() (err error) {
	// [[FBRoute POST:@"/wda/lock"] respondWithTarget:self action:@selector(handleLock:)]
	// [[FBRoute POST:@"/wda/lock"].withoutSession
	_, err = wd.httpPOST(nil, "/session", wd.session.ID, "/wda/lock")
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/alert/text"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/alert/buttons"); err != nil {
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
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/alert/text")
	return
}

func (wd *wdaDriver) AppLaunch(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/launch"] respondWithTarget:self action:@selector(handleSessionAppLaunch:)]
	data := make(map[string]interface{})
	data["bundleId"] = bundleId
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/apps/launch")
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("wda launch failed: %v", err))
	}
	return nil
}

func (wd *wdaDriver) AppLaunchUnattached(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/launchUnattached"].withoutSession respondWithTarget:self action:@selector(handleLaunchUnattachedApp:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.httpPOST(data, "/wda/apps/launchUnattached")
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("wda launchUnattached failed: %v", err))
	}
	return nil
}

func (wd *wdaDriver) AppTerminate(bundleId string) (successful bool, err error) {
	// [[FBRoute POST:@"/wda/apps/terminate"] respondWithTarget:self action:@selector(handleSessionAppTerminate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/apps/terminate"); err != nil {
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
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/apps/activate")
	return
}

func (wd *wdaDriver) AppDeactivate(second float64) (err error) {
	// [[FBRoute POST:@"/wda/deactivateApp"] respondWithTarget:self action:@selector(handleDeactivateAppCommand:)]
	if second < 3 {
		second = 3.0
	}
	data := map[string]interface{}{"duration": second}
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/deactivateApp")
	return
}

func (wd *wdaDriver) GetForegroundApp() (app AppInfo, err error) {
	// appInfo, err := wd.ActiveAppInfo()
	// if err != nil {
	// 	return AppInfo{}, err
	// }

	// app = AppInfo{
	// 	AppBaseInfo: AppBaseInfo{
	// 		PackageName: appInfo.BundleId,
	// 		Activity:    "",
	// 	},
	// }
	return AppInfo{}, errors.Wrap(errDriverNotImplemented,
		"GetForegroundApp not implemented for ios")
}

func (wd *wdaDriver) AssertForegroundApp(bundleId string, viewControllerType ...string) error {
	log.Warn().Str("bundleId", bundleId).
		Strs("viewControllerType", viewControllerType).
		Msg("ios view controller assertion not implemented, skip")
	return nil
}

func (wd *wdaDriver) Tap(x, y int, options ...ActionOption) error {
	return wd.TapFloat(float64(x), float64(y), options...)
}

func (wd *wdaDriver) TapFloat(x, y float64, options ...ActionOption) (err error) {
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]
	actionOptions := NewActionOptions(options...)

	x = wd.toScale(x)
	y = wd.toScale(y)
	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}
	x += actionOptions.getRandomOffset()
	y += actionOptions.getRandomOffset()

	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	// update data options in post data for extra WDA configurations
	actionOptions.updateData(data)

	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/tap/0")
	return
}

func (wd *wdaDriver) DoubleTap(x, y int) error {
	return wd.DoubleTapFloat(float64(x), float64(y))
}

func (wd *wdaDriver) DoubleTapFloat(x, y float64) (err error) {
	// [[FBRoute POST:@"/wda/doubleTap"] respondWithTarget:self action:@selector(handleDoubleTapCoordinate:)]
	data := map[string]interface{}{
		"x": wd.toScale(x),
		"y": wd.toScale(y),
	}
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/doubleTap")
	return
}

func (wd *wdaDriver) TouchAndHold(x, y int, second ...float64) error {
	return wd.TouchAndHoldFloat(float64(x), float64(y), second...)
}

func (wd *wdaDriver) TouchAndHoldFloat(x, y float64, second ...float64) (err error) {
	// [[FBRoute POST:@"/wda/touchAndHold"] respondWithTarget:self action:@selector(handleTouchAndHoldCoordinate:)]
	data := map[string]interface{}{
		"x": wd.toScale(x),
		"y": wd.toScale(y),
	}
	if len(second) == 0 || second[0] <= 0 {
		second = []float64{1.0}
	}
	data["duration"] = second[0]
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/touchAndHold")
	return
}

func (wd *wdaDriver) Drag(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return wd.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (wd *wdaDriver) DragFloat(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	// [[FBRoute POST:@"/wda/dragfromtoforduration"] respondWithTarget:self action:@selector(handleDragCoordinate:)]
	actionOptions := NewActionOptions(options...)

	fromX = wd.toScale(fromX)
	fromY = wd.toScale(fromY)
	toX = wd.toScale(toX)
	toY = wd.toScale(toY)
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
		"fromX": fromX,
		"fromY": fromY,
		"toX":   toX,
		"toY":   toY,
	}

	// update data options in post data for extra WDA configurations
	actionOptions.updateData(data)
	// wda 43 version
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/drag")
	return
}

func (wd *wdaDriver) Swipe(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return wd.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (wd *wdaDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	return wd.DragFloat(fromX, fromY, toX, toY, options...)
}

func (wd *wdaDriver) SetPasteboard(contentType PasteboardType, content string) (err error) {
	// [[FBRoute POST:@"/wda/setPasteboard"] respondWithTarget:self action:@selector(handleSetPasteboard:)]
	data := map[string]interface{}{
		"contentType": contentType,
		"content":     base64.StdEncoding.EncodeToString([]byte(content)),
	}
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/setPasteboard")
	return
}

func (wd *wdaDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	// [[FBRoute POST:@"/wda/getPasteboard"] respondWithTarget:self action:@selector(handleGetPasteboard:)]
	data := map[string]interface{}{"contentType": contentType}
	var rawResp rawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/getPasteboard"); err != nil {
		return nil, err
	}
	if raw, err = rawResp.valueDecodeAsBase64(); err != nil {
		return nil, err
	}
	return
}

func (wd *wdaDriver) SetIme(ime string) error {
	return errDriverNotImplemented
}

func (wd *wdaDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return errDriverNotImplemented
}

func (wd *wdaDriver) SendKeys(text string, options ...ActionOption) (err error) {
	// [[FBRoute POST:@"/wda/keys"] respondWithTarget:self action:@selector(handleKeys:)]
	actionOptions := NewActionOptions(options...)
	data := map[string]interface{}{"value": strings.Split(text, "")}

	// new data options in post data for extra WDA configurations
	actionOptions.updateData(data)

	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/keys")
	return
}

func (wd *wdaDriver) Input(text string, options ...ActionOption) (err error) {
	return wd.SendKeys(text, options...)
}

func (wd *wdaDriver) Clear(packageName string) error {
	return errDriverNotImplemented
}

// PressBack simulates a short press on the BACK button.
func (wd *wdaDriver) PressBack(options ...ActionOption) (err error) {
	actionOptions := NewActionOptions(options...)

	windowSize, err := wd.WindowSize()
	if err != nil {
		return
	}
	fromX := wd.toScale(float64(windowSize.Width) * 0)
	fromY := wd.toScale(float64(windowSize.Height) * 0.5)
	toX := wd.toScale(float64(windowSize.Width) * 0.6)
	toY := wd.toScale(float64(windowSize.Height) * 0.5)
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
		"fromX": fromX,
		"fromY": fromY,
		"toX":   toX,
		"toY":   toY,
	}

	// update data options in post data for extra WDA configurations
	actionOptions.updateData(data)

	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/dragfromtoforduration")
	return
}

func (wd *wdaDriver) PressButton(devBtn DeviceButton) (err error) {
	// [[FBRoute POST:@"/wda/pressButton"] respondWithTarget:self action:@selector(handlePressButtonCommand:)]
	data := map[string]interface{}{"name": devBtn}
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/wda/pressButton")
	return
}

func (wd *wdaDriver) LoginNoneUI(packageName, phoneNumber string, captcha string) error {
	return errDriverNotImplemented
}

func (wd *wdaDriver) LogoutNoneUI(packageName string) error {
	return errDriverNotImplemented
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

func (wd *wdaDriver) Orientation() (orientation Orientation, err error) {
	// [[FBRoute GET:@"/orientation"] respondWithTarget:self action:@selector(handleGetOrientation:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/orientation"); err != nil {
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
	_, err = wd.httpPOST(data, "/session", wd.session.ID, "/orientation")
	return
}

func (wd *wdaDriver) Rotation() (rotation Rotation, err error) {
	// [[FBRoute GET:@"/rotation"] respondWithTarget:self action:@selector(handleGetRotation:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/rotation"); err != nil {
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
	_, err = wd.httpPOST(rotation, "/session", wd.session.ID, "/rotation")
	return
}

func (wd *wdaDriver) Screenshot() (raw *bytes.Buffer, err error) {
	// [[FBRoute GET:@"/screenshot"] respondWithTarget:self action:@selector(handleGetScreenshot:)]
	// [[FBRoute GET:@"/screenshot"].withoutSession respondWithTarget:self action:@selector(handleGetScreenshot:)]
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/screenshot"); err != nil {
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
	tmp, _ := url.Parse(wd.concatURL(nil, "/session", wd.session.ID))
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

func (wd *wdaDriver) TapByText(text string, options ...ActionOption) error {
	return errDriverNotImplemented
}

func (wd *wdaDriver) TapByTexts(actions ...TapTextAction) error {
	return errDriverNotImplemented
}

func (wd *wdaDriver) AccessibleSource() (source string, err error) {
	// [[FBRoute GET:@"/wda/accessibleSource"] respondWithTarget:self action:@selector(handleGetAccessibleSourceCommand:)]
	// [[FBRoute GET:@"/wda/accessibleSource"].withoutSession
	var rawResp rawResponse
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/wda/accessibleSource"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.session.ID, "/appium/settings"); err != nil {
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
	if rawResp, err = wd.httpPOST(data, "/session", wd.session.ID, "/appium/settings"); err != nil {
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

func (ud *wdaDriver) GetSession() *DriverSession {
	return &ud.Driver.session
}

type rawResponse []byte

func (r rawResponse) checkErr() (err error) {
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

func (r rawResponse) valueConvertToString() (s string, err error) {
	reply := new(struct{ Value string })
	if err = json.Unmarshal(r, reply); err != nil {
		return "", errors.Wrapf(err, "json.Unmarshal failed, rawResponse: %s", string(r))
	}
	s = reply.Value
	return
}

func (r rawResponse) valueConvertToBool() (b bool, err error) {
	reply := new(struct{ Value bool })
	if err = json.Unmarshal(r, reply); err != nil {
		return false, err
	}
	b = reply.Value
	return
}

func (r rawResponse) valueConvertToSessionInfo() (sessionInfo SessionInfo, err error) {
	reply := new(struct{ Value struct{ SessionInfo } })
	if err = json.Unmarshal(r, reply); err != nil {
		return SessionInfo{}, err
	}
	sessionInfo = reply.Value.SessionInfo
	return
}

func (r rawResponse) valueConvertToJsonRawMessage() (raw builtinJSON.RawMessage, err error) {
	reply := new(struct{ Value builtinJSON.RawMessage })
	if err = json.Unmarshal(r, reply); err != nil {
		return nil, err
	}
	raw = reply.Value
	return
}

func (r rawResponse) valueConvertToJsonObject() (obj map[string]interface{}, err error) {
	if err = json.Unmarshal(r, &obj); err != nil {
		return nil, err
	}
	return
}

func (r rawResponse) valueDecodeAsBase64() (raw *bytes.Buffer, err error) {
	str, err := r.valueConvertToString()
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
