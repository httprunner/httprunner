package uixt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/utf7"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func NewUIA2Driver(device *AndroidDevice) (*UIA2Driver, error) {
	log.Info().Interface("device", device).Msg("init android UIA2 driver")
	adbDriver, err := NewADBDriver(device)
	if err != nil {
		return nil, err
	}
	driver := &UIA2Driver{
		ADBDriver: adbDriver,
	}

	// setup driver
	if err := driver.Setup(); err != nil {
		return nil, err
	}

	// register driver session reset handler
	driver.Session.RegisterResetHandler(driver.Setup)

	return driver, nil
}

type UIA2Driver struct {
	*ADBDriver

	// cache to avoid repeated query
	windowSize types.Size
}

func (ud *UIA2Driver) Setup() error {
	localPort, err := ud.Device.Forward(ud.Device.Options.UIA2Port)
	if err != nil {
		return errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				localPort, ud.Device.Options.UIA2Port, err))
	}
	err = ud.Session.SetupPortForward(localPort)
	if err != nil {
		return err
	}
	ud.Session.SetBaseURL(
		fmt.Sprintf("http://forward-to-%d:%d/wd/hub",
			localPort, ud.Device.Options.UIA2Port))

	// uiautomator2 server must be started before

	// check uiautomator server package installed
	if !ud.Device.IsPackageInstalled(ud.Device.Options.UIA2ServerPackageName) {
		return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
			"%s not installed", ud.Device.Options.UIA2ServerPackageName)
	}
	if !ud.Device.IsPackageInstalled(ud.Device.Options.UIA2ServerTestPackageName) {
		return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
			"%s not installed", ud.Device.Options.UIA2ServerTestPackageName)
	}

	// TODO: check uiautomator server package running
	// if dev.IsPackageRunning(UIA2ServerPackageName) {
	// 	return nil
	// }

	// start uiautomator2 server
	// Todo: keep-alive
	go func() {
		if err := ud.startUIA2Server(); err != nil {
			log.Fatal().Err(err).Msg("start UIA2 failed")
		}
	}()
	time.Sleep(5 * time.Second) // wait for uiautomator2 server start

	// create new session
	err = ud.InitSession(nil)
	if err != nil {
		return err
	}
	return nil
}

func (ud *UIA2Driver) TearDown() error {
	log.Warn().Msg("TearDown not implemented in UIA2Driver")
	return nil
}

func (ud *UIA2Driver) InitSession(capabilities option.Capabilities) (err error) {
	// register(postHandler, new InitSession("/wd/hub/session"))
	var rawResp DriverRawResponse
	data := make(map[string]interface{})
	if len(capabilities) == 0 {
		data["capabilities"] = make(map[string]interface{})
	} else {
		data["capabilities"] = map[string]interface{}{"alwaysMatch": capabilities}
	}
	if rawResp, err = ud.Session.POST(data, "/session"); err != nil {
		return err
	}
	reply := new(struct{ Value struct{ SessionId string } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return err
	}
	ud.Session.ID = reply.Value.SessionId
	return nil
}

func (ud *UIA2Driver) DeleteSession() (err error) {
	if ud.Session.ID == "" {
		return nil
	}
	urlStr := fmt.Sprintf("/session/%s", ud.Session.ID)
	if _, err = ud.Session.DELETE(urlStr); err == nil {
		ud.Session.ID = ""
	}

	return err
}

func (ud *UIA2Driver) Status() (deviceStatus types.DeviceStatus, err error) {
	// register(getHandler, new Status("/wd/hub/status"))
	var rawResp DriverRawResponse
	// Notice: use Driver.GET instead of httpGET to avoid loop calling
	if rawResp, err = ud.Session.GET("/status"); err != nil {
		return types.DeviceStatus{Ready: false}, err
	}
	reply := new(struct {
		Value struct {
			// Message string
			Ready bool
		}
	})
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.DeviceStatus{Ready: false}, err
	}
	return types.DeviceStatus{Ready: true}, nil
}

func (ud *UIA2Driver) DeviceInfo() (deviceInfo types.DeviceInfo, err error) {
	// register(getHandler, new GetDeviceInfo("/wd/hub/session/:sessionId/appium/device/info"))
	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/appium/device/info", ud.Session.ID)
	if rawResp, err = ud.Session.GET(urlStr); err != nil {
		return types.DeviceInfo{}, err
	}
	reply := new(struct{ Value struct{ types.DeviceInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.DeviceInfo{}, err
	}
	deviceInfo = reply.Value.DeviceInfo
	return
}

func (ud *UIA2Driver) BatteryInfo() (batteryInfo types.BatteryInfo, err error) {
	// register(getHandler, new GetBatteryInfo("/wd/hub/session/:sessionId/appium/device/battery_info"))
	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/appium/device/battery_info", ud.Session.ID)
	if rawResp, err = ud.Session.GET(urlStr); err != nil {
		return types.BatteryInfo{}, err
	}
	reply := new(struct{ Value struct{ types.BatteryInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.BatteryInfo{}, err
	}
	if reply.Value.Level == -1 || reply.Value.Status == -1 {
		return reply.Value.BatteryInfo, errors.New("cannot be retrieved from the system")
	}
	batteryInfo = reply.Value.BatteryInfo
	return
}

func (ud *UIA2Driver) WindowSize() (size types.Size, err error) {
	// register(getHandler, new GetDeviceSize("/wd/hub/session/:sessionId/window/:windowHandle/size"))
	if !ud.windowSize.IsNil() {
		// use cached window size
		return ud.windowSize, nil
	}

	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/window/:windowHandle/size", ud.Session.ID)
	if rawResp, err = ud.Session.GET(urlStr); err != nil {
		return types.Size{}, errors.Wrap(err, "get window size failed by UIA2 request")
	}
	reply := new(struct{ Value struct{ types.Size } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.Size{}, errors.Wrap(err, "get window size failed by UIA2 response")
	}
	size = reply.Value.Size

	// check orientation
	orientation, err := ud.Orientation()
	if err != nil {
		log.Warn().Err(err).Msgf("window size get orientation failed, use default orientation")
		orientation = types.OrientationPortrait
	}
	if orientation != types.OrientationPortrait {
		size.Width, size.Height = size.Height, size.Width
	}

	ud.windowSize = size // cache window size
	return size, nil
}

// Back simulates a short press on the BACK button.
func (ud *UIA2Driver) Back() (err error) {
	log.Info().Msg("UIA2Driver.Back")
	// register(postHandler, new PressBack("/wd/hub/session/:sessionId/back"))
	urlStr := fmt.Sprintf("/session/%s/back", ud.Session.ID)
	_, err = ud.Session.POST(nil, urlStr)
	return
}

func (ud *UIA2Driver) PressKeyCodes(keyCode KeyCode, metaState KeyMeta, flags ...KeyFlag) (err error) {
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
	urlStr := fmt.Sprintf("/session/%s/appium/device/press_keycode", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return
}

func (ud *UIA2Driver) Orientation() (orientation types.Orientation, err error) {
	// [[FBRoute GET:@"/orientation"] respondWithTarget:self action:@selector(handleGetOrientation:)]
	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/orientation", ud.Session.ID)
	if rawResp, err = ud.Session.GET(urlStr); err != nil {
		return "", err
	}
	reply := new(struct{ Value types.Orientation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	orientation = reply.Value
	return
}

func (ud *UIA2Driver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("UIA2Driver.DoubleTap")
	var err error
	x, y, err = handlerDoubleTap(ud, x, y, opts...)
	if err != nil {
		return err
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
					map[string]interface{}{"type": "pointerUp", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pointerDown", "duration": 0, "button": 0},
					map[string]interface{}{"type": "pointerUp", "duration": 0, "button": 0},
				},
			},
		},
	}

	urlStr := fmt.Sprintf("/session/%s/actions/tap", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return err
}

func (ud *UIA2Driver) TapXY(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("UIA2Driver.TapXY")
	// register(postHandler, new Tap("/wd/hub/session/:sessionId/appium/tap"))
	absX, absY, err := convertToAbsolutePoint(ud, x, y)
	if err != nil {
		return err
	}
	return ud.TapAbsXY(absX, absY, opts...)
}

func (ud *UIA2Driver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("UIA2Driver.TapAbsXY")
	// register(postHandler, new Tap("/wd/hub/session/:sessionId/appium/tap"))

	var err error
	x, y, err = handlerTapAbsXY(ud, x, y, opts...)
	if err != nil {
		return err
	}

	actionOptions := option.NewActionOptions(opts...)
	duration := 100.0
	if actionOptions.PressDuration > 0 {
		duration = actionOptions.PressDuration * 1000 // convert to ms
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
	option.MergeOptions(data, opts...)

	urlStr := fmt.Sprintf("/session/%s/actions/tap", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return err
}

func (ud *UIA2Driver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	log.Info().Float64("x", x).Float64("y", y).Msg("UIA2Driver.TouchAndHold")
	actionOptions := option.NewActionOptions(opts...)
	x, y = actionOptions.ApplyTapOffset(x, y)
	duration := actionOptions.Duration
	if duration == 0 {
		duration = 1.0
	}
	// register(postHandler, new TouchLongClick("/wd/hub/session/:sessionId/touch/longclick"))
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"x":        x,
			"y":        y,
			"duration": int(duration * 1000),
		},
	}
	urlStr := fmt.Sprintf("/session/%s/touch/longclick", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return
}

// Drag performs a swipe from one coordinate to another coordinate. You can control
// the smoothness and speed of the swipe by specifying the number of steps.
// Each step execution is throttled to 5 milliseconds per step, so for a 100
// steps, the swipe will take around 0.5 seconds to complete.
func (ud *UIA2Driver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	log.Info().Float64("fromX", fromX).Float64("fromY", fromY).
		Float64("toX", toX).Float64("toY", toY).Msg("UIA2Driver.Drag")

	var err error
	fromX, fromY, toX, toY, err = handlerDrag(ud, fromX, fromY, toX, toY, opts...)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"startX": fromX,
		"startY": fromY,
		"endX":   toX,
		"endY":   toY,
	}
	option.MergeOptions(data, opts...)

	// register(postHandler, new Drag("/wd/hub/session/:sessionId/touch/drag"))
	urlStr := fmt.Sprintf("/session/%s/touch/drag", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return err
}

// Swipe performs a swipe from one coordinate to another using the number of steps
// to determine smoothness and speed. Each step execution is throttled to 5ms
// per step. So for a 100 steps, the swipe will take about 1/2 second to complete.
//
//	`steps` is the number of move steps sent to the system
func (ud *UIA2Driver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	// register(postHandler, new Swipe("/wd/hub/session/:sessionId/touch/perform"))
	log.Info().Float64("fromX", fromX).Float64("fromY", fromY).
		Float64("toX", toX).Float64("toY", toY).Msg("UIA2Driver.Swipe")
	var err error
	fromX, fromY, toX, toY, err = handlerSwipe(ud, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	actionOptions := option.NewActionOptions(opts...)
	duration := 200.0
	if actionOptions.PressDuration > 0 {
		duration = actionOptions.PressDuration * 1000 // ms
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
	option.MergeOptions(data, opts...)

	urlStr := fmt.Sprintf("/session/%s/actions/swipe", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return err
}

func (ud *UIA2Driver) SetPasteboard(contentType types.PasteboardType, content string) (err error) {
	log.Info().Str("contentType", string(contentType)).
		Str("content", content).Msg("UIA2Driver.SetPasteboard")
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
	urlStr := fmt.Sprintf("/session/%s/appium/device/set_clipboard", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return
}

func (ud *UIA2Driver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	if len(contentType) == 0 {
		contentType = types.PasteboardTypePlaintext
	}
	// register(postHandler, new GetClipboard("/wd/hub/session/:sessionId/appium/device/get_clipboard"))
	data := map[string]interface{}{
		"contentType": contentType[0],
	}
	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/appium/device/get_clipboard", ud.Session.ID)
	if rawResp, err = ud.Session.POST(data, urlStr); err != nil {
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
func (ud *UIA2Driver) Input(text string, opts ...option.ActionOption) (err error) {
	log.Info().Str("text", text).Msg("UIA2Driver.Input")
	// register(postHandler, new SendKeysToElement("/wd/hub/session/:sessionId/keys"))
	// https://github.com/appium/appium-uiautomator2-server/blob/master/app/src/main/java/io/appium/uiautomator2/handler/SendKeysToElement.java#L76-L85
	err = ud.SendUnicodeKeys(text, opts...)
	if err == nil {
		return nil
	}

	data := map[string]interface{}{
		"text": text,
	}
	option.MergeOptions(data, opts...)
	urlStr := fmt.Sprintf("/session/%s/keys", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return
}

func (ud *UIA2Driver) SendUnicodeKeys(text string, opts ...option.ActionOption) (err error) {
	log.Info().Str("text", text).Msg("UIA2Driver.SendUnicodeKeys")
	// If the Unicode IME is not installed, fall back to the old interface.
	// There might be differences in the tracking schemes across different phones, and it is pending further verification.
	// In release version: without the Unicode IME installed, the test cannot execute.
	if !ud.IsUnicodeIMEInstalled() {
		return fmt.Errorf("appium unicode ime not installed")
	}
	currentIme, err := ud.ADBDriver.GetIme()
	if err != nil {
		return
	}
	if currentIme != option.UnicodeImePackageName {
		defer func() {
			_ = ud.ADBDriver.SetIme(currentIme)
		}()
		err = ud.ADBDriver.SetIme(option.UnicodeImePackageName)
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
	err = ud.SendActionKey(encodedStr, opts...)
	return
}

func (ud *UIA2Driver) SendActionKey(text string, opts ...option.ActionOption) (err error) {
	log.Info().Str("text", text).Msg("UIA2Driver.SendActionKey")
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
	option.MergeOptions(data, opts...)

	urlStr := fmt.Sprintf("/session/%s/actions/keys", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return
}

func (ud *UIA2Driver) Rotation() (rotation types.Rotation, err error) {
	// register(getHandler, new GetRotation("/wd/hub/session/:sessionId/rotation"))
	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/rotation", ud.Session.ID)
	if rawResp, err = ud.Session.GET(urlStr); err != nil {
		return types.Rotation{}, err
	}
	reply := new(struct{ Value types.Rotation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.Rotation{}, err
	}

	rotation = reply.Value
	return
}

func (ud *UIA2Driver) ScreenShot(opts ...option.ActionOption) (raw *bytes.Buffer, err error) {
	// https://bytedance.larkoffice.com/docx/C8qEdmSHnoRvMaxZauocMiYpnLh
	// ui2截图受内存影响，改为adb截图
	return ud.ADBDriver.ScreenShot(opts...)
}

func (ud *UIA2Driver) Source(srcOpt ...option.SourceOption) (source string, err error) {
	// register(getHandler, new Source("/wd/hub/session/:sessionId/source"))
	var rawResp DriverRawResponse
	urlStr := fmt.Sprintf("/session/%s/source", ud.Session.ID)
	if rawResp, err = ud.Session.GET(urlStr); err != nil {
		return "", err
	}
	reply := new(struct{ Value string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}

	source = reply.Value
	return
}

func (ud *UIA2Driver) startUIA2Server() error {
	const maxRetries = 20
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Info().Str("package", ud.Device.Options.UIA2ServerTestPackageName).
			Int("attempt", attempt).Msg("start uiautomator server")
		// $ adb shell am instrument -w $UIA2ServerTestPackageName
		// -w: wait for instrumentation to finish before returning.
		// Required for test runners.
		out, err := ud.Device.RunShellCommand("am", "instrument", "-w",
			ud.Device.Options.UIA2ServerTestPackageName)
		if err != nil {
			log.Error().Err(err).Int("retryCount", maxRetries).Msg("start uiautomator server failed, retrying...")
		}
		if strings.Contains(out, "Process crashed") {
			log.Error().Msg("uiautomator server crashed, retrying...")
		}
	}

	return errors.Wrapf(code.MobileUIDriverAppCrashed,
		"uiautomator server crashed %d times", maxRetries)
}

func (ud *UIA2Driver) stopUIA2Server() error {
	_, err := ud.Device.RunShellCommand("am", "force-stop",
		ud.Device.Options.UIA2ServerPackageName)
	return err
}
