package uixt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
)

// See https://developer.android.com/reference/android/view/KeyEvent
const (
	KEYCODE_BACK      string = "4"
	KEYCODE_CAMERA    string = "27"
	KEYCODE_ALT_LEFT  string = "57"
	KEYCODE_ALT_RIGHT string = "58"
	KEYCODE_MENU      string = "82"
	KEYCODE_BREAK     string = "121"
	KEYCODE_ALL_APPS  string = "284"
)

var errDriverNotImplemented = errors.New("driver method not implemented")

type uiaDriver struct {
	Driver

	adbDevice gadb.Device
	logcat    *AdbLogcat
	localPort int
}

func NewUIADriver(capabilities Capabilities, urlPrefix string) (driver *uiaDriver, err error) {
	if capabilities == nil {
		capabilities = NewCapabilities()
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
	if session, err := driver.NewSession(capabilities); err != nil {
		return nil, err
	} else {
		driver.sessionId = session.SessionId
	}
	return
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

func (ud *uiaDriver) Close() (err error) {
	if ud.sessionId == "" {
		return nil
	}
	if _, err = ud.httpDELETE("/session", ud.sessionId); err == nil {
		ud.sessionId = ""
	}

	return err
}

func (ud *uiaDriver) NewSession(capabilities Capabilities) (sessionInfo SessionInfo, err error) {
	// register(postHandler, new NewSession("/wd/hub/session"))
	var rawResp rawResponse
	data := map[string]interface{}{"capabilities": capabilities}
	if rawResp, err = ud.httpPOST(data, "/session"); err != nil {
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

func (ud *uiaDriver) ActiveSession() (sessionInfo SessionInfo, err error) {
	// [[FBRoute GET:@""] respondWithTarget:self action:@selector(handleGetActiveSession:)]
	return SessionInfo{SessionId: ud.sessionId}, nil
}

func (ud *uiaDriver) SessionIDs() (sessionIDs []string, err error) {
	// register(getHandler, new GetSessions("/wd/hub/sessions"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/sessions"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value []struct{ SessionId string } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	sessionIDs = make([]string, len(reply.Value))
	for i := range reply.Value {
		sessionIDs[i] = reply.Value[i].SessionId
	}
	return
}

func (ud *uiaDriver) SessionDetails() (scrollData map[string]interface{}, err error) {
	// register(getHandler, new GetSessionDetails("/wd/hub/session/:sessionId"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	scrollData = reply.Value
	return
}

func (ud *uiaDriver) DeleteSession() (err error) {
	// TODO
	return errDriverNotImplemented
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

func (ud *uiaDriver) Location() (location Location, err error) {
	// TODO
	return location, errDriverNotImplemented
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
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "window/:windowHandle/size"); err != nil {
		return Size{}, err
	}
	reply := new(struct{ Value struct{ Size } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Size{}, err
	}
	size = reply.Value.Size
	return
}

func (ud *uiaDriver) Screen() (screen Screen, err error) {
	// TODO
	return screen, errDriverNotImplemented
}

func (ud *uiaDriver) Scale() (scale float64, err error) {
	return 1, nil
}

// PressBack simulates a short press on the BACK button.
func (ud *uiaDriver) PressBack(options ...DataOption) (err error) {
	// adb shell input keyevent 4
	_, err = ud.adbDevice.RunShellCommand("input", "keyevent", KEYCODE_BACK)
	if err == nil {
		return nil
	}
	// register(postHandler, new PressBack("/wd/hub/session/:sessionId/back"))
	_, err = ud.httpPOST(nil, "/session", ud.sessionId, "back")
	return
}

func (ud *uiaDriver) StartCamera() (err error) {
	if _, err = ud.adbDevice.RunShellCommand("rm", "-r", "/sdcard/DCIM/Camera"); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	var version string
	if version, err = ud.adbDevice.RunShellCommand("getprop", "ro.build.version.release"); err != nil {
		return err
	}
	if version == "11" || version == "12" {
		if _, err = ud.adbDevice.RunShellCommand("am", "start", "-a", "android.media.action.STILL_IMAGE_CAMERA"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ud.adbDevice.RunShellCommand("input", "swipe", "750", "1000", "250", "1000"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ud.adbDevice.RunShellCommand("input", "keyevent", KEYCODE_CAMERA); err != nil {
			return err
		}
		return
	} else {
		if _, err = ud.adbDevice.RunShellCommand("am", "start", "-a", "android.media.action.VIDEO_CAPTURE"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ud.adbDevice.RunShellCommand("input", "keyevent", KEYCODE_CAMERA); err != nil {
			return err
		}
		return
	}
}

func (ud *uiaDriver) StopCamera() (err error) {
	err = ud.PressBack()
	if err != nil {
		return err
	}
	err = ud.Homescreen()
	if err != nil {
		return err
	}

	// kill samsung shell command
	if _, err = ud.adbDevice.RunShellCommand("am", "force-stop", "com.sec.android.app.camera"); err != nil {
		return err
	}

	// kill other camera (huawei mi)
	if _, err = ud.adbDevice.RunShellCommand("am", "force-stop", "com.android.camera2"); err != nil {
		return err
	}
	return
}

func (ud *uiaDriver) Homescreen() (err error) {
	return ud.PressKeyCode(KCHome, KMEmpty)
}

func (ud *uiaDriver) PressKeyCode(keyCode KeyCode, metaState KeyMeta, flags ...KeyFlag) (err error) {
	// adb shell input keyevent <keyCode>
	_, err = ud.adbDevice.RunShellCommand(
		"input", "keyevent", fmt.Sprintf("%d", keyCode))
	if err == nil {
		return nil
	}

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

func (ud *uiaDriver) AppLaunch(bundleId string) (err error) {
	// 不指定 Activity 名称启动（启动主 Activity）
	// adb shell monkey -p <packagename> -c android.intent.category.LAUNCHER 1
	sOutput, err := ud.adbDevice.RunShellCommand(
		"monkey", "-p", bundleId, "-c", "android.intent.category.LAUNCHER", "1",
	)
	if err != nil {
		return err
	}
	if strings.Contains(sOutput, "monkey aborted") {
		return fmt.Errorf("app launch: %s", strings.TrimSpace(sOutput))
	}
	return nil
}

func (ud *uiaDriver) AppTerminate(bundleId string) (successful bool, err error) {
	// 强制停止应用，停止 <packagename> 相关的进程
	// adb shell am force-stop <packagename>
	_, err = ud.adbDevice.RunShellCommand("am", "force-stop", bundleId)
	return err == nil, err
}

func (ud *uiaDriver) Tap(x, y int, options ...DataOption) error {
	return ud.TapFloat(float64(x), float64(y), options...)
}

func (ud *uiaDriver) TapFloat(x, y float64, options ...DataOption) (err error) {
	// adb shell input tap x y
	_, err = ud.adbDevice.RunShellCommand(
		"input", "tap", fmt.Sprintf("%.1f", x), fmt.Sprintf("%.1f", y))
	if err == nil {
		return nil
	}

	// register(postHandler, new Tap("/wd/hub/session/:sessionId/appium/tap"))
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	// new data options in post data for extra uiautomator configurations
	newData := NewData(data, options...)

	_, err = ud.httpPOST(newData, "/session", ud.sessionId, "appium/tap")
	return
}

func (ud *uiaDriver) DoubleTap(x, y int) error {
	return ud.DoubleTapFloat(float64(x), float64(y))
}

func (ud *uiaDriver) DoubleTapFloat(x, y float64) (err error) {
	// TODO
	return errDriverNotImplemented
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
func (ud *uiaDriver) Drag(fromX, fromY, toX, toY int, options ...DataOption) error {
	return ud.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ud *uiaDriver) DragFloat(fromX, fromY, toX, toY float64, options ...DataOption) (err error) {
	data := map[string]interface{}{
		"startX": fromX,
		"startY": fromY,
		"endX":   toX,
		"endY":   toY,
	}

	// new data options in post data for extra uiautomator configurations
	newData := NewData(data, options...)

	// register(postHandler, new Drag("/wd/hub/session/:sessionId/touch/drag"))
	_, err = ud.httpPOST(newData, "/session", ud.sessionId, "touch/drag")
	return
}

// Swipe performs a swipe from one coordinate to another using the number of steps
// to determine smoothness and speed. Each step execution is throttled to 5ms
// per step. So for a 100 steps, the swipe will take about 1/2 second to complete.
//  `steps` is the number of move steps sent to the system
func (ud *uiaDriver) Swipe(fromX, fromY, toX, toY int, options ...DataOption) error {
	return ud.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ud *uiaDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...DataOption) error {
	// adb shell input swipe fromX fromY toX toY
	_, err := ud.adbDevice.RunShellCommand(
		"input", "swipe",
		fmt.Sprintf("%.1f", fromX), fmt.Sprintf("%.1f", fromY),
		fmt.Sprintf("%.1f", toX), fmt.Sprintf("%.1f", toY),
	)
	if err == nil {
		return nil
	}

	// register(postHandler, new Swipe("/wd/hub/session/:sessionId/touch/perform"))
	data := map[string]interface{}{
		"startX": fromX,
		"startY": fromY,
		"endX":   toX,
		"endY":   toY,
	}

	// new data options in post data for extra uiautomator configurations
	newData := NewData(data, options...)

	_, err = ud.httpPOST(newData, "/session", ud.sessionId, "touch/perform")
	return err
}

func (ud *uiaDriver) ForceTouch(x, y int, pressure float64, second ...float64) error {
	return ud.ForceTouchFloat(float64(x), float64(y), pressure, second...)
}

func (ud *uiaDriver) ForceTouchFloat(x, y, pressure float64, second ...float64) (err error) {
	// TODO
	return errDriverNotImplemented
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

func (ud *uiaDriver) SendKeys(text string, options ...DataOption) (err error) {
	// register(postHandler, new SendKeysToElement("/wd/hub/session/:sessionId/keys"))
	// https://github.com/appium/appium-uiautomator2-server/blob/master/app/src/main/java/io/appium/uiautomator2/handler/SendKeysToElement.java#L76-L85
	data := map[string]interface{}{
		"text": text,
	}
	// new data options in post data for extra uiautomator configurations
	newData := NewData(data, options...)

	_, err = ud.httpPOST(newData, "/session", ud.sessionId, "keys")
	return
}

func (ud *uiaDriver) Input(text string, options ...DataOption) (err error) {
	return ud.SendKeys(text, options...)
}

func (ud *uiaDriver) PressButton(devBtn DeviceButton) (err error) {
	// TODO
	return errDriverNotImplemented
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

func (ud *uiaDriver) SetRotation(rotation Rotation) (err error) {
	// TODO
	return errDriverNotImplemented
}

func (ud *uiaDriver) Screenshot() (raw *bytes.Buffer, err error) {
	// adb shell screencap -p
	resp, err := ud.adbDevice.RunShellCommandWithBytes(
		"screencap", "-p",
	)
	if err == nil {
		return bytes.NewBuffer(resp), nil
	}

	// register(getHandler, new CaptureScreenshot("/wd/hub/session/:sessionId/screenshot"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "screenshot"); err != nil {
		return nil, errors.Wrap(code.AndroidScreenShotError,
			fmt.Sprintf("get UIA screenshot data failed: %v", err))
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	var decodeStr []byte
	if decodeStr, err = base64.StdEncoding.DecodeString(reply.Value); err != nil {
		return nil, errors.Wrap(code.AndroidScreenShotError,
			fmt.Sprintf("decode UIA screenshot data failed: %v", err))
	}

	raw = bytes.NewBuffer(decodeStr)
	return
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

func (ud *uiaDriver) AccessibleSource() (source string, err error) {
	// TODO
	return source, errDriverNotImplemented
}

func (ud *uiaDriver) HealthCheck() (err error) {
	// TODO
	return errDriverNotImplemented
}

func (ud *uiaDriver) GetAppiumSettings() (settings map[string]interface{}, err error) {
	// register(getHandler, new GetSettings("/wd/hub/session/:sessionId/appium/settings"))
	var rawResp rawResponse
	if rawResp, err = ud.httpGET("/session", ud.sessionId, "appium/settings"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	settings = reply.Value
	return
}

func (ud *uiaDriver) SetAppiumSettings(settings map[string]interface{}) (ret map[string]interface{}, err error) {
	data := map[string]interface{}{
		"settings": settings,
	}
	// register(postHandler, new UpdateSettings("/wd/hub/session/:sessionId/appium/settings"))
	_, err = ud.httpPOST(data, "/session", ud.sessionId, "appium/settings")
	return
}

func (ud *uiaDriver) IsHealthy() (healthy bool, err error) {
	// TODO
	return healthy, errDriverNotImplemented
}

func (ud *uiaDriver) WaitWithTimeoutAndInterval(condition Condition, timeout, interval time.Duration) error {
	startTime := time.Now()
	for {
		done, err := condition(ud)
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

func (ud *uiaDriver) WaitWithTimeout(condition Condition, timeout time.Duration) error {
	return ud.WaitWithTimeoutAndInterval(condition, timeout, DefaultWaitInterval)
}

func (ud *uiaDriver) Wait(condition Condition) error {
	return ud.WaitWithTimeoutAndInterval(condition, DefaultWaitTimeout, DefaultWaitInterval)
}

func (ud *uiaDriver) StartCaptureLog(identifier ...string) (err error) {
	log.Info().Msg("start adb log recording")
	err = ud.logcat.CatchLogcat()
	if err != nil {
		err = errors.Wrap(code.AndroidCaptureLogError,
			fmt.Sprintf("start adb log recording failed: %v", err))
		return err
	}
	return nil
}

func (ud *uiaDriver) StopCaptureLog() (result interface{}, err error) {
	log.Info().Msg("stop adb log recording")
	err = ud.logcat.Stop()
	if err != nil {
		log.Error().Err(err).Msg("failed to get adb log recording")
		err = errors.Wrap(code.AndroidCaptureLogError,
			fmt.Sprintf("get adb log recording failed: %v", err))
		return "", err
	}
	content := ud.logcat.logBuffer.String()
	return ConvertPoints(content), nil
}
