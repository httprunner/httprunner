package uixt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/electricbubble/gadb"
)

type uiaDriver struct {
	Driver

	adbDevice gadb.Device
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
	if driver.sessionId, err = driver.NewSession(capabilities); err != nil {
		return nil, err
	}
	return
}

func (d *uiaDriver) NewSession(capabilities Capabilities) (sessionID string, err error) {
	// register(postHandler, new NewSession("/wd/hub/session"))
	var rawResp rawResponse
	data := map[string]interface{}{"capabilities": capabilities}
	if rawResp, err = d.httpPOST(data, "/session"); err != nil {
		return "", err
	}
	reply := new(struct{ Value struct{ SessionId string } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	sessionID = reply.Value.SessionId
	// d.sessionIdCache[sessionID] = true
	return
}

func (d *uiaDriver) Quit() (err error) {
	// register(deleteHandler, new DeleteSession("/wd/hub/session/:sessionId"))
	if d.sessionId == "" {
		return nil
	}
	if _, err = d.httpDELETE("/session", d.sessionId); err == nil {
		d.sessionId = ""
	}

	return err
}

func (d *uiaDriver) ActiveSessionID() string {
	return d.sessionId
}

func (d *uiaDriver) SessionIDs() (sessionIDs []string, err error) {
	// register(getHandler, new GetSessions("/wd/hub/sessions"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/sessions"); err != nil {
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

func (d *uiaDriver) SessionDetails() (scrollData map[string]interface{}, err error) {
	// register(getHandler, new GetSessionDetails("/wd/hub/session/:sessionId"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	scrollData = reply.Value
	return
}

func (d *uiaDriver) Status() (ready bool, err error) {
	// register(getHandler, new Status("/wd/hub/status"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/status"); err != nil {
		return false, err
	}
	reply := new(struct {
		Value struct {
			// Message string
			Ready bool
		}
	})
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return false, err
	}
	ready = reply.Value.Ready
	return
}

// Screenshot grab device screenshot
func (d *uiaDriver) Screenshot() (raw *bytes.Buffer, err error) {
	// register(getHandler, new CaptureScreenshot("/wd/hub/session/:sessionId/screenshot"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "screenshot"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	var decodeStr []byte
	if decodeStr, err = base64.StdEncoding.DecodeString(reply.Value); err != nil {
		return nil, err
	}

	raw = bytes.NewBuffer(decodeStr)
	return
}

func (d *uiaDriver) Orientation() (orientation Orientation, err error) {
	// register(getHandler, new GetOrientation("/wd/hub/session/:sessionId/orientation"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "orientation"); err != nil {
		return "", err
	}
	reply := new(struct{ Value Orientation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}

	orientation = reply.Value
	return
}

func (d *uiaDriver) Rotation() (rotation Rotation, err error) {
	// register(getHandler, new GetRotation("/wd/hub/session/:sessionId/rotation"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "rotation"); err != nil {
		return Rotation{}, err
	}
	reply := new(struct{ Value Rotation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Rotation{}, err
	}

	rotation = reply.Value
	return
}

// DeviceSize get window size of the device
func (d *uiaDriver) DeviceSize() (deviceSize Size, err error) {
	// register(getHandler, new GetDeviceSize("/wd/hub/session/:sessionId/window/:windowHandle/size"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "window/:windowHandle/size"); err != nil {
		return Size{}, err
	}
	reply := new(struct{ Value Size })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Size{}, err
	}

	deviceSize = reply.Value
	return
}

// Source get page source
func (d *uiaDriver) Source() (sXML string, err error) {
	// register(getHandler, new Source("/wd/hub/session/:sessionId/source"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "source"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}

	sXML = reply.Value
	return
}

// StatusBarHeight get status bar height of the device
func (d *uiaDriver) StatusBarHeight() (height int, err error) {
	// register(getHandler, new GetSystemBars("/wd/hub/session/:sessionId/appium/device/system_bars"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "appium/device/system_bars"); err != nil {
		return 0, err
	}
	reply := new(struct{ Value struct{ StatusBar int } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return 0, err
	}

	height = reply.Value.StatusBar
	return
}

func (d *uiaDriver) check() error {
	if d.adbDevice.Serial() == "" {
		return errors.New("adb daemon: the device is not ready")
	}
	return nil
}

// Dispose corresponds to the command:
//  adb -s $serial forward --remove $localPort
func (d *uiaDriver) Dispose() (err error) {
	if err = d.check(); err != nil {
		return err
	}
	if d.localPort == 0 {
		return nil
	}
	return d.adbDevice.ForwardKill(d.localPort)
}

func (d *uiaDriver) ActiveAppActivity() (appActivity string, err error) {
	if err = d.check(); err != nil {
		return "", err
	}

	var sOutput string
	if sOutput, err = d.adbDevice.RunShellCommand("dumpsys activity activities | grep mResumedActivity"); err != nil {
		return "", err
	}
	re := regexp.MustCompile(`\{(.+?)\}`)
	if !re.MatchString(sOutput) {
		return "", fmt.Errorf("active app activity: %s", strings.TrimSpace(sOutput))
	}
	fields := strings.Fields(re.FindStringSubmatch(sOutput)[1])
	appActivity = fields[2]
	return
}

func (d *uiaDriver) ActiveAppPackageName() (appPackageName string, err error) {
	var activity string
	if activity, err = d.ActiveAppActivity(); err != nil {
		return "", err
	}
	appPackageName = strings.Split(activity, "/")[0]
	return
}

func (d *uiaDriver) AppLaunch(appPackageName string, waitForComplete ...AndroidBySelector) (err error) {
	if err = d.check(); err != nil {
		return err
	}

	var sOutput string
	if sOutput, err = d.adbDevice.RunShellCommand("monkey -p", appPackageName, "-c android.intent.category.LAUNCHER 1"); err != nil {
		return err
	}
	if strings.Contains(sOutput, "monkey aborted") {
		return fmt.Errorf("app launch: %s", strings.TrimSpace(sOutput))
	}

	if len(waitForComplete) != 0 {
		var ce error
		exists := func(d *uiaDriver) (bool, error) {
			for i := range waitForComplete {
				_, ce = d.FindElement(waitForComplete[i])
				if ce == nil {
					return true, nil
				}
			}
			return false, nil
		}
		if err = d.WaitWithTimeoutAndInterval(exists, 45, 1.5); err != nil {
			return fmt.Errorf("app launch (waitForComplete): %s: %w", err.Error(), ce)
		}
	}
	return
}

func (d *uiaDriver) AppTerminate(appPackageName string) (err error) {
	if err = d.check(); err != nil {
		return err
	}

	_, err = d.adbDevice.RunShellCommand("am force-stop", appPackageName)
	return
}

func (d *uiaDriver) AppInstall(apkPath string, reinstall ...bool) (err error) {
	if err = d.check(); err != nil {
		return err
	}

	apkName := filepath.Base(apkPath)
	if !strings.HasSuffix(strings.ToLower(apkName), ".apk") {
		return fmt.Errorf("apk file must have an extension of '.apk': %s", apkPath)
	}

	var apkFile *os.File
	if apkFile, err = os.Open(apkPath); err != nil {
		return fmt.Errorf("apk file: %w", err)
	}

	remotePath := path.Join(DeviceTempPath, apkName)
	if err = d.adbDevice.PushFile(apkFile, remotePath); err != nil {
		return fmt.Errorf("apk push: %w", err)
	}

	var shellOutput string
	if len(reinstall) != 0 && reinstall[0] {
		shellOutput, err = d.adbDevice.RunShellCommand("pm install", "-r", remotePath)
	} else {
		shellOutput, err = d.adbDevice.RunShellCommand("pm install", remotePath)
	}

	if err != nil {
		return fmt.Errorf("apk install: %w", err)
	}

	if !strings.Contains(shellOutput, "Success") {
		return fmt.Errorf("apk installed: %s", shellOutput)
	}

	return
}

func (d *uiaDriver) AppUninstall(appPackageName string, keepDataAndCache ...bool) (err error) {
	if err = d.check(); err != nil {
		return err
	}

	var shellOutput string
	if len(keepDataAndCache) != 0 && keepDataAndCache[0] {
		shellOutput, err = d.adbDevice.RunShellCommand("pm uninstall", "-k", appPackageName)
	} else {
		shellOutput, err = d.adbDevice.RunShellCommand("pm uninstall", appPackageName)
	}

	if err != nil {
		return fmt.Errorf("apk uninstall: %w", err)
	}

	if !strings.Contains(shellOutput, "Success") {
		return fmt.Errorf("apk uninstalled: %s", shellOutput)
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

func (d *uiaDriver) BatteryInfo() (info BatteryInfo, err error) {
	// register(getHandler, new GetBatteryInfo("/wd/hub/session/:sessionId/appium/device/battery_info"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "appium/device/battery_info"); err != nil {
		return BatteryInfo{}, err
	}
	reply := new(struct{ Value BatteryInfo })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return BatteryInfo{}, err
	}

	info = reply.Value
	if info.Level == -1 || info.Status == -1 {
		return info, errors.New("cannot be retrieved from the system")
	}
	return
}

func (d *uiaDriver) GetAppiumSettings() (settings map[string]interface{}, err error) {
	// register(getHandler, new GetSettings("/wd/hub/session/:sessionId/appium/settings"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "appium/settings"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	settings = reply.Value
	return
}

// DeviceScaleRatio get device pixel ratio
func (d *uiaDriver) DeviceScaleRatio() (scale float64, err error) {
	// register(getHandler, new GetDevicePixelRatio("/wd/hub/session/:sessionId/appium/device/pixel_ratio"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "appium/device/pixel_ratio"); err != nil {
		return 0, err
	}
	reply := new(struct{ Value float64 })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return 0, err
	}

	scale = reply.Value
	return
}

type (
	AndroidDeviceInfo struct {
		// ANDROID_ID A 64-bit number (as a hex string) that is uniquely generated when the user
		// first sets up the device and should remain constant for the lifetime of the user's device. The value
		// may change if a factory reset is performed on the device.
		AndroidID string `json:"androidId"`
		// Build.MANUFACTURER value
		Manufacturer string `json:"manufacturer"`
		// Build.MODEL value
		Model string `json:"model"`
		// Build.BRAND value
		Brand string `json:"brand"`
		// Current running OS's API VERSION
		APIVersion string `json:"apiVersion"`
		// The current version string, for example "1.0" or "3.4b5"
		PlatformVersion string `json:"platformVersion"`
		// the name of the current celluar network carrier
		CarrierName string `json:"carrierName"`
		// the real size of the default display
		RealDisplaySize string `json:"realDisplaySize"`
		// The logical density of the display in Density Independent Pixel units.
		DisplayDensity int `json:"displayDensity"`
		// available networks
		Networks []networkInfo `json:"networks"`
		// current system locale
		Locale string `json:"locale"`
		// current system timezone
		// e.g. "Asia/Tokyo", "America/Caracas", "Asia/Shanghai"
		TimeZone  string `json:"timeZone"`
		Bluetooth struct {
			State string `json:"state"`
		} `json:"bluetooth"`
	}
	networkCapabilities struct {
		TransportTypes            string `json:"transportTypes"`
		NetworkCapabilities       string `json:"networkCapabilities"`
		LinkUpstreamBandwidthKbps int    `json:"linkUpstreamBandwidthKbps"`
		LinkDownBandwidthKbps     int    `json:"linkDownBandwidthKbps"`
		SignalStrength            int    `json:"signalStrength"`
		SSID                      string `json:"SSID"`
	}
	networkInfo struct {
		Type          int                 `json:"type"`
		TypeName      string              `json:"typeName"`
		Subtype       int                 `json:"subtype"`
		SubtypeName   string              `json:"subtypeName"`
		IsConnected   bool                `json:"isConnected"`
		DetailedState string              `json:"detailedState"`
		State         string              `json:"state"`
		ExtraInfo     string              `json:"extraInfo"`
		IsAvailable   bool                `json:"isAvailable"`
		IsRoaming     bool                `json:"isRoaming"`
		IsFailover    bool                `json:"isFailover"`
		Capabilities  networkCapabilities `json:"capabilities"`
	}
)

func (d *uiaDriver) DeviceInfo() (info AndroidDeviceInfo, err error) {
	// register(getHandler, new GetDeviceInfo("/wd/hub/session/:sessionId/appium/device/info"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "appium/device/info"); err != nil {
		return AndroidDeviceInfo{}, err
	}
	reply := new(struct{ Value AndroidDeviceInfo })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return AndroidDeviceInfo{}, err
	}

	info = reply.Value
	return
}

// AlertText get text of the on-screen dialog
func (d *uiaDriver) AlertText() (text string, err error) {
	// register(getHandler, new GetAlertText("/wd/hub/session/:sessionId/alert/text"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "alert/text"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}

	text = reply.Value
	return
}

// Tap perform a click at arbitrary coordinates specified
func (d *uiaDriver) Tap(x, y int) (err error) {
	return d.TapFloat(float64(x), float64(y))
}

func (d *uiaDriver) TapFloat(x, y float64) (err error) {
	// register(postHandler, new Tap("/wd/hub/session/:sessionId/appium/tap"))
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "appium/tap")
	return
}

func (d *uiaDriver) TapPoint(point Point) (err error) {
	return d.Tap(point.X, point.Y)
}

func (d *uiaDriver) TapPointF(point PointF) (err error) {
	return d.TapFloat(point.X, point.Y)
}

func (d *uiaDriver) _swipe(startX, startY, endX, endY interface{}, steps int, elementID ...string) (err error) {
	// register(postHandler, new Swipe("/wd/hub/session/:sessionId/touch/perform"))
	data := map[string]interface{}{
		"startX": startX,
		"startY": startY,
		"endX":   endX,
		"endY":   endY,
		"steps":  steps,
	}
	if len(elementID) != 0 {
		data["elementId"] = elementID[0]
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/perform")
	return
}

// Swipe performs a swipe from one coordinate to another using the number of steps
// to determine smoothness and speed. Each step execution is throttled to 5ms
// per step. So for a 100 steps, the swipe will take about 1/2 second to complete.
//  `steps` is the number of move steps sent to the system
func (d *uiaDriver) Swipe(startX, startY, endX, endY int, steps ...int) (err error) {
	return d.SwipeFloat(float64(startX), float64(startY), float64(endX), float64(endY), steps...)
}

func (d *uiaDriver) SwipeFloat(startX, startY, endX, endY float64, steps ...int) (err error) {
	if len(steps) == 0 {
		steps = []int{12}
	}
	return d._swipe(startX, startY, endX, endY, steps[0])
}

func (d *uiaDriver) SwipePoint(startPoint, endPoint Point, steps ...int) (err error) {
	return d.Swipe(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y, steps...)
}

func (d *uiaDriver) SwipePointF(startPoint, endPoint PointF, steps ...int) (err error) {
	return d.SwipeFloat(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y, steps...)
}

func (d *uiaDriver) _drag(data map[string]interface{}) (err error) {
	// register(postHandler, new Drag("/wd/hub/session/:sessionId/touch/drag"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/drag")
	return
}

// Drag performs a swipe from one coordinate to another coordinate. You can control
// the smoothness and speed of the swipe by specifying the number of steps.
// Each step execution is throttled to 5 milliseconds per step, so for a 100
// steps, the swipe will take around 0.5 seconds to complete.
func (d *uiaDriver) Drag(startX, startY, endX, endY int, steps ...int) (err error) {
	return d.DragFloat(float64(startX), float64(startY), float64(endX), float64(endY), steps...)
}

func (d *uiaDriver) DragFloat(startX, startY, endX, endY float64, steps ...int) error {
	if len(steps) == 0 {
		steps = []int{12}
	}
	data := map[string]interface{}{
		"startX": startX,
		"startY": startY,
		"endX":   endX,
		"endY":   endY,
		"steps":  steps[0],
	}
	return d._drag(data)
}

func (d *uiaDriver) DragPoint(startPoint Point, endPoint Point, steps ...int) error {
	return d.Drag(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y, steps...)
}

func (d *uiaDriver) DragPointF(startPoint PointF, endPoint PointF, steps ...int) (err error) {
	return d.DragFloat(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y, steps...)
}

func (d *uiaDriver) TouchLongClick(x, y int, duration ...float64) (err error) {
	if len(duration) == 0 {
		duration = []float64{1.0}
	}
	// register(postHandler, new TouchLongClick("/wd/hub/session/:sessionId/touch/longclick"))
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"x":        x,
			"y":        y,
			"duration": int(duration[0] * 1000),
		},
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/longclick")
	return
}

func (d *uiaDriver) TouchLongClickPoint(point Point, duration ...float64) (err error) {
	return d.TouchLongClick(point.X, point.Y, duration...)
}

func (d *uiaDriver) SendKeys(text string, isReplace ...bool) (err error) {
	if len(isReplace) == 0 {
		isReplace = []bool{true}
	}
	// register(postHandler, new SendKeysToElement("/wd/hub/session/:sessionId/keys"))
	// https://github.com/appium/appium-uiautomator2-server/blob/master/app/src/main/java/io/appium/uiautomator2/handler/SendKeysToElement.java#L76-L85
	data := map[string]interface{}{
		"text":    text,
		"replace": isReplace[0],
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "keys")
	return
}

// PressBack simulates a short press on the BACK button.
func (d *uiaDriver) PressBack() (err error) {
	// register(postHandler, new PressBack("/wd/hub/session/:sessionId/back"))
	_, err = d.httpPOST(nil, "/session", d.sessionId, "back")
	return
}

// public class KeyCodeModel extends BaseModel {
//    @RequiredField
//    public Integer keycode;
//    public Integer metastate;
//    public Integer flags;
// }
func (d *uiaDriver) LongPressKeyCode(keyCode KeyCode, metaState KeyMeta, flags ...KeyFlag) (err error) {
	if len(flags) == 0 {
		flags = []KeyFlag{KFFromSystem}
	}
	data := map[string]interface{}{
		"keycode":   keyCode,
		"metastate": metaState,
		"flags":     flags[0],
	}
	// register(postHandler, new LongPressKeyCode("/wd/hub/session/:sessionId/appium/device/long_press_keycode"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "/appium/device/long_press_keycode")
	return
}

func (d *uiaDriver) _pressKeyCode(keyCode KeyCode, metaState KeyMeta, flags ...KeyFlag) (err error) {
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
	_, err = d.httpPOST(data, "/session", d.sessionId, "appium/device/press_keycode")
	return
}

func (d *uiaDriver) PressKeyCode(keyCode KeyCode, metaState KeyMeta, flags ...KeyFlag) (err error) {
	if len(flags) == 0 {
		flags = []KeyFlag{KFFromSystem}
	}
	return d._pressKeyCode(keyCode, metaState, KFFromSystem)
}

// PressKeyCodeAsync simulates a short press using a key code.
func (d *uiaDriver) PressKeyCodeAsync(keyCode KeyCode, metaState ...KeyMeta) (err error) {
	if len(metaState) == 0 {
		metaState = []KeyMeta{KMEmpty}
	}
	return d._pressKeyCode(keyCode, metaState[0])
}

func (d *uiaDriver) TouchDown(x, y int) (err error) {
	// register(postHandler, new TouchDown("/wd/hub/session/:sessionId/touch/down"))
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"x": x,
			"y": y,
		},
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/down")
	return
}

func (d *uiaDriver) TouchDownPoint(point Point) error {
	return d.TouchDown(point.X, point.Y)
}

func (d *uiaDriver) TouchUp(x, y int) (err error) {
	// register(postHandler, new TouchUp("/wd/hub/session/:sessionId/touch/up"))
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"x": x,
			"y": y,
		},
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/up")
	return
}

func (d *uiaDriver) TouchUpPoint(point Point) error {
	return d.TouchUp(point.X, point.Y)
}

func (d *uiaDriver) TouchMove(x, y int) (err error) {
	// register(postHandler, new TouchMove("/wd/hub/session/:sessionId/touch/move"))
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"x": x,
			"y": y,
		},
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/move")
	return
}

func (d *uiaDriver) TouchMovePoint(point Point) error {
	return d.TouchMove(point.X, point.Y)
}

// OpenNotification opens the notification shade.
func (d *uiaDriver) OpenNotification() (err error) {
	// register(postHandler, new OpenNotification("/wd/hub/session/:sessionId/appium/device/open_notifications"))
	_, err = d.httpPOST(nil, "/session", d.sessionId, "appium/device/open_notifications")
	return
}

func (d *uiaDriver) _flick(data map[string]interface{}) (err error) {
	// register(postHandler, new Flick("/wd/hub/session/:sessionId/touch/flick"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/flick")
	return
}

func (d *uiaDriver) Flick(xSpeed, ySpeed int) (err error) {
	data := map[string]interface{}{
		"xspeed": xSpeed,
		"yspeed": ySpeed,
	}
	if xSpeed == 0 && ySpeed == 0 {
		return errors.New("both 'xSpeed' and 'ySpeed' cannot be zero")
	}

	return d._flick(data)
}

func (d *uiaDriver) _scrollTo(method, selector string, maxSwipes int, elementID ...string) (err error) {
	// register(postHandler, new ScrollTo("/wd/hub/session/:sessionId/touch/scroll"))
	params := map[string]interface{}{
		"strategy": method,
		"selector": selector,
	}
	if maxSwipes > 0 {
		params["maxSwipes"] = maxSwipes
	}
	data := map[string]interface{}{"params": params}
	if len(elementID) != 0 {
		data["origin"] = map[string]string{
			legacyWebElementIdentifier: elementID[0],
			webElementIdentifier:       elementID[0],
		}
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "touch/scroll")
	return
}

func (d *uiaDriver) ScrollTo(by AndroidBySelector, maxSwipes ...int) (err error) {
	if len(maxSwipes) == 0 {
		maxSwipes = []int{0}
	}
	method, selector := by.getMethodAndSelector()
	return d._scrollTo(method, selector, maxSwipes[0])
}

type W3CMouseButtonType int

const (
	MBTLeft   W3CMouseButtonType = 0
	MBTMiddle W3CMouseButtonType = 1
	MBTRight  W3CMouseButtonType = 2
)

func (g *W3CGestures) PointerDown(button ...W3CMouseButtonType) *W3CGestures {
	if len(button) == 0 {
		button = []W3CMouseButtonType{MBTLeft}
	}
	*g = append(*g, _newW3CGesture().pointerDown(int(button[0])))
	return g
}

func (g *W3CGestures) PointerUp(button ...W3CMouseButtonType) *W3CGestures {
	if len(button) == 0 {
		button = []W3CMouseButtonType{MBTLeft}
	}
	*g = append(*g, _newW3CGesture().pointerUp(int(button[0])))
	return g
}

type W3CPointerMoveType string

const (
	PMTViewport W3CPointerMoveType = "viewport"
	PMTPointer  W3CPointerMoveType = "pointer"
)

func (g *W3CGestures) PointerMove(x, y float64, origin interface{}, duration float64, pressure, size float64) *W3CGestures {
	val := ""
	switch v := origin.(type) {
	case string:
		val = v
	case W3CPointerMoveType:
		val = string(v)
	case *uiaElement:
		val = v.id
	default:
		val = string(PMTViewport)
	}
	*g = append(*g, _newW3CGesture().pointerMove(x, y, val, duration, pressure, size))
	return g
}

func (g *W3CGestures) PointerMoveTo(x, y float64, duration ...float64) *W3CGestures {
	if len(duration) == 0 || duration[0] < 0 {
		duration = []float64{0.5}
	}
	*g = append(*g, _newW3CGesture().pointerMove(x, y, string(PMTViewport), duration[0]*1000))
	return g
}

func (g *W3CGestures) PointerMoveRelative(x, y float64, duration ...float64) *W3CGestures {
	if len(duration) == 0 || duration[0] < 0 {
		duration = []float64{0.5}
	}
	*g = append(*g, _newW3CGesture().pointerMove(x, y, string(PMTPointer), duration[0]*1000))
	return g
}

func (g *W3CGestures) PointerMouseOver(x, y float64, element *uiaElement, duration ...float64) *W3CGestures {
	if len(duration) == 0 || duration[0] < 0 {
		duration = []float64{0.5}
	}
	*g = append(*g, _newW3CGesture().pointerMove(x, y, element.id, duration[0]*1000))
	return g
}

type W3CAction map[string]interface{}

type W3CActionType string

const (
	_         W3CActionType = "none"
	ATKey     W3CActionType = "key"
	ATPointer W3CActionType = "pointer"
)

type W3CPointerType string

const (
	PTMouse W3CPointerType = "mouse"
	PTPen   W3CPointerType = "pen"
	PTTouch W3CPointerType = "touch"
)

func NewW3CAction(actionType W3CActionType, gestures *W3CGestures, pointerType ...W3CPointerType) W3CAction {
	w3cAction := make(W3CAction)
	w3cAction["type"] = actionType
	w3cAction["actions"] = gestures
	if actionType != ATPointer {
		return w3cAction
	}

	if len(pointerType) == 0 {
		pointerType = []W3CPointerType{PTTouch}
	}
	type W3CItemParameters struct {
		PointerType W3CPointerType `json:"pointerType"`
	}
	w3cAction["parameters"] = W3CItemParameters{PointerType: pointerType[0]}
	return w3cAction
}

func (d *uiaDriver) PerformW3CActions(action W3CAction, acts ...W3CAction) (err error) {
	var actionId uint64 = 1
	acts = append([]W3CAction{action}, acts...)
	for i := range acts {
		item := acts[i]
		item["id"] = strconv.FormatUint(actionId, 10)
		actionId++
		acts[i] = item
	}
	data := map[string]interface{}{
		"actions": acts,
	}
	// register(postHandler, new W3CActions("/wd/hub/session/:sessionId/actions"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "/actions")
	return
}

type ClipDataType string

const ClipDataTypePlaintext ClipDataType = "PLAINTEXT"

func (d *uiaDriver) GetClipboard(contentType ...ClipDataType) (content string, err error) {
	if len(contentType) == 0 {
		contentType = []ClipDataType{ClipDataTypePlaintext}
	}
	// register(postHandler, new GetClipboard("/wd/hub/session/:sessionId/appium/device/get_clipboard"))
	data := map[string]interface{}{
		"contentType": contentType[0],
	}
	var rawResp rawResponse
	if rawResp, err = d.httpPOST(data, "/session", d.sessionId, "appium/device/get_clipboard"); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}

	content = reply.Value
	if data, err := base64.StdEncoding.DecodeString(content); err != nil {
		return content, err
	} else {
		content = string(data)
	}
	return
}

func (d *uiaDriver) SetClipboard(contentType ClipDataType, content string, label ...string) (err error) {
	lbl := content
	if len(label) != 0 {
		lbl = label[0]
	}
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
	_, err = d.httpPOST(data, "/session", d.sessionId, "appium/device/set_clipboard")
	return
}

func (d *uiaDriver) AlertAccept(buttonLabel ...string) (err error) {
	data := map[string]interface{}{
		"buttonLabel": nil,
	}
	if len(buttonLabel) != 0 {
		data["buttonLabel"] = buttonLabel[0]
	}
	// register(postHandler, new AcceptAlert("/wd/hub/session/:sessionId/alert/accept"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "alert/accept")
	return
}

func (d *uiaDriver) AlertDismiss(buttonLabel ...string) (err error) {
	data := map[string]interface{}{
		"buttonLabel": nil,
	}
	if len(buttonLabel) != 0 {
		data["buttonLabel"] = buttonLabel[0]
	}
	// register(postHandler, new DismissAlert("/wd/hub/session/:sessionId/alert/dismiss"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "alert/dismiss")
	return
}

func (d *uiaDriver) SetAppiumSettings(settings map[string]interface{}) (err error) {
	data := map[string]interface{}{
		"settings": settings,
	}
	// register(postHandler, new UpdateSettings("/wd/hub/session/:sessionId/appium/settings"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "appium/settings")
	return
}

func (d *uiaDriver) SetOrientation(orientation Orientation) (err error) {
	data := map[string]interface{}{
		"orientation": orientation,
	}
	// register(postHandler, new SetOrientation("/wd/hub/session/:sessionId/orientation"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "orientation")
	return
}

// SetRotation
//  `x` and `y` are ignored. We only care about `z`
//  0/90/180/270
func (d *uiaDriver) SetRotation(rotation Rotation) (err error) {
	data := map[string]interface{}{
		"z": rotation.Z,
	}
	// register(postHandler, new SetRotation("/wd/hub/session/:sessionId/rotation"))
	_, err = d.httpPOST(data, "/session", d.sessionId, "rotation")
	return
}

type NetworkType int

const (
	NetworkTypeWifi NetworkType = 2

	// NetworkTypeNone NetworkType = iota
	// NetworkTypeAirplane
	// NetworkTypeWifi
	// _
	// NetworkTypeData
	// _
	// NetworkTypeAll
)

// NetworkConnection always turn on
func (d *uiaDriver) NetworkConnection(networkType NetworkType) (err error) {
	// register(postHandler, new NetworkConnection("/wd/hub/session/:sessionId/network_connection"))
	data := map[string]interface{}{
		"type": networkType,
	}
	_, err = d.httpPOST(data, "/session", d.sessionId, "network_connection")
	return
}

func (d *uiaDriver) _findElements(method, selector string, elementID ...string) (elements []*uiaElement, err error) {
	// register(postHandler, new FindElements("/wd/hub/session/:sessionId/elements"))
	data := map[string]interface{}{
		"strategy": method,
		"selector": selector,
	}
	if len(elementID) != 0 {
		data["context"] = elementID[0]
	}
	var rawResp rawResponse
	if rawResp, err = d.httpPOST(data, "/session", d.sessionId, "/elements"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value []map[string]string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	if len(reply.Value) == 0 {
		return nil, fmt.Errorf("no such element: unable to find an element using '%s', value '%s'", method, selector)
	}
	elements = make([]*uiaElement, len(reply.Value))
	for i, elem := range reply.Value {
		var id string
		if id = elementIDFromValue(elem); id == "" {
			return nil, fmt.Errorf("invalid element returned: %+v", reply)
		}
		elements[i] = &uiaElement{parent: d, id: id}
	}
	return
}

func (d *uiaDriver) _findElement(method, selector string, elementID ...string) (elem *uiaElement, err error) {
	// register(postHandler, new FindElement("/wd/hub/session/:sessionId/element"))
	data := map[string]interface{}{
		"strategy": method,
		"selector": selector,
	}
	if len(elementID) != 0 {
		data["context"] = elementID[0]
	}
	var rawResp rawResponse
	if rawResp, err = d.httpPOST(data, "/session", d.sessionId, "/element"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	if len(reply.Value) == 0 {
		return nil, fmt.Errorf("no such element: unable to find an element using '%s', value '%s'", method, selector)
	}
	var id string
	if id = elementIDFromValue(reply.Value); id == "" {
		return nil, fmt.Errorf("invalid element returned: %+v", reply)
	}
	elem = &uiaElement{parent: d, id: id}
	return
}

func (d *uiaDriver) FindElements(by AndroidBySelector) (elements []*uiaElement, err error) {
	return d._findElements(by.getMethodAndSelector())
}

func (d *uiaDriver) FindElement(by AndroidBySelector) (elem *uiaElement, err error) {
	return d._findElement(by.getMethodAndSelector())
}

func (d *uiaDriver) ActiveElement() (elem *uiaElement, err error) {
	// register(getHandler, new ActiveElement("/wd/hub/session/:sessionId/element/active"))
	var rawResp rawResponse
	if rawResp, err = d.httpGET("/session", d.sessionId, "/element/active"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	if len(reply.Value) == 0 {
		return nil, errors.New("no such element")
	}
	var id string
	if id = elementIDFromValue(reply.Value); id == "" {
		return nil, fmt.Errorf("invalid element returned: %+v", reply)
	}
	elem = &uiaElement{parent: d, id: id}
	return
}

type AndroidCondition func(d *uiaDriver) (bool, error)

func (d *uiaDriver) _waitWithTimeoutAndInterval(condition AndroidCondition, timeout, interval time.Duration) (err error) {
	startTime := time.Now()
	for {
		done, err := condition(d)
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

// WaitWithTimeoutAndInterval waits for the condition to evaluate to true.
func (d *uiaDriver) WaitWithTimeoutAndInterval(condition AndroidCondition, timeout, interval float64) (err error) {
	dTimeout := time.Millisecond * time.Duration(timeout*1000)
	dInterval := time.Millisecond * time.Duration(interval*1000)
	return d._waitWithTimeoutAndInterval(condition, dTimeout, dInterval)
}

// WaitWithTimeout works like WaitWithTimeoutAndInterval, but with default polling interval.
func (d *uiaDriver) WaitWithTimeout(condition AndroidCondition, timeout float64) error {
	dTimeout := time.Millisecond * time.Duration(timeout*1000)
	return d._waitWithTimeoutAndInterval(condition, dTimeout, DefaultWaitInterval)
}

// Wait works like WaitWithTimeoutAndInterval, but using the default timeout and polling interval.
func (d *uiaDriver) Wait(condition AndroidCondition) error {
	return d._waitWithTimeoutAndInterval(condition, DefaultWaitTimeout, DefaultWaitInterval)
}
