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
	"github.com/httprunner/httprunner/v5/internal/simulation"
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

	// Store base URL for building full URLs
	baseURL := fmt.Sprintf("http://forward-to-%d:%d/wd/hub",
		localPort, ud.Device.Options.UIA2Port)
	ud.Session.SetBaseURL(baseURL)

	// Notice: uiautomator2 server must be started before running test

	// check uiautomator server package installed
	if !ud.Device.IsPackageInstalled(ud.Device.Options.UIA2ServerPackageName) {
		return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
			"%s not installed", ud.Device.Options.UIA2ServerPackageName)
	}
	if !ud.Device.IsPackageInstalled(ud.Device.Options.UIA2ServerTestPackageName) {
		return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
			"%s not installed", ud.Device.Options.UIA2ServerTestPackageName)
	}

	// check uiautomator server package running
	if ud.Device.IsPackageRunning(ud.Device.Options.UIA2ServerTestPackageName) {
		log.Info().Str("package", ud.Device.Options.UIA2ServerTestPackageName).
			Msg("uiautomator2 server is already running, skip starting")
	} else {
		// start uiautomator2 server
		go func() {
			if err := ud.startUIA2Server(); err != nil {
				log.Fatal().Err(err).Msg("start UIA2 failed")
			}
		}()
		time.Sleep(5 * time.Second) // wait for uiautomator2 server start
	}

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
	actionOptions := option.NewActionOptions(opts...)
	x, y, err := preHandler_DoubleTap(ud, actionOptions, x, y)
	if err != nil {
		return err
	}
	defer postHandler(ud, option.ACTION_DoubleTapXY, actionOptions)

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
	actionOptions := option.NewActionOptions(opts...)
	x, y, err := preHandler_TapAbsXY(ud, actionOptions, x, y)
	if err != nil {
		return err
	}
	defer postHandler(ud, option.ACTION_TapAbsXY, actionOptions)

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

	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err := preHandler_Drag(ud, actionOptions, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	defer postHandler(ud, option.ACTION_Drag, actionOptions)

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

	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err := preHandler_Swipe(ud, option.ACTION_SwipeCoordinate, actionOptions, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	defer postHandler(ud, option.ACTION_SwipeCoordinate, actionOptions)

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

// TouchByEvents performs a complex swipe using a sequence of touch events with pressure and size data
func (ud *UIA2Driver) TouchByEvents(events []types.TouchEvent, opts ...option.ActionOption) error {
	log.Info().Int("eventCount", len(events)).Msg("UIA2Driver.SwipeSimulator")

	if len(events) == 0 {
		return fmt.Errorf("no touch events provided")
	}

	actionOptions := option.NewActionOptions(opts...)

	// Apply pre-handlers for the first and last events (start and end coordinates)
	firstEvent := events[0]
	lastEvent := events[len(events)-1]

	// Use rawX/rawY if available, otherwise fallback to X/Y for first event
	startX, startY := firstEvent.RawX, firstEvent.RawY
	if startX == 0 && startY == 0 {
		startX, startY = firstEvent.X, firstEvent.Y
	}

	// Use rawX/rawY if available, otherwise fallback to X/Y for last event
	endX, endY := lastEvent.RawX, lastEvent.RawY
	if endX == 0 && endY == 0 {
		endX, endY = lastEvent.X, lastEvent.Y
	}

	fromX, fromY, toX, toY, err := preHandler_Swipe(ud, option.ACTION_SwipeCoordinate, actionOptions,
		startX, startY, endX, endY)
	if err != nil {
		return err
	}
	defer postHandler(ud, option.ACTION_SwipeCoordinate, actionOptions)

	var actions []interface{}
	var prevEventTime int64

	for i, event := range events {
		var duration float64
		if i > 0 {
			// Calculate duration from previous event using EventTime (milliseconds)
			duration = float64(event.EventTime - prevEventTime)
		}
		prevEventTime = event.EventTime

		// Use rawX/rawY if available, otherwise fallback to X/Y
		x, y := event.RawX, event.RawY
		if x == 0 && y == 0 {
			// Fallback to X/Y if rawX/rawY are not set
			x, y = event.X, event.Y
		}

		// Apply coordinate transformation if it's the first or last event
		if i == 0 {
			x, y = fromX, fromY
		} else if i == len(events)-1 {
			x, y = toX, toY
		}

		var actionMap map[string]interface{}

		switch event.Action {
		case 0: // ACTION_DOWN
			actionMap = map[string]interface{}{
				"type":     "pointerDown",
				"duration": 0,
				"button":   0,
				"pressure": event.Pressure,
				"size":     event.Size,
			}
			// Add initial move to position before down
			if i == 0 {
				moveAction := map[string]interface{}{
					"type":     "pointerMove",
					"duration": 0,
					"x":        x,
					"y":        y,
					"origin":   "viewport",
					"pressure": event.Pressure,
					"size":     event.Size,
				}
				actions = append(actions, moveAction)
			}
		case 1: // ACTION_UP
			actionMap = map[string]interface{}{
				"type":     "pointerUp",
				"duration": 0,
				"button":   0,
				"pressure": event.Pressure,
				"size":     event.Size,
			}
		case 2: // ACTION_MOVE
			actionMap = map[string]interface{}{
				"type":     "pointerMove",
				"duration": duration,
				"x":        x,
				"y":        y,
				"origin":   "viewport",
				"pressure": event.Pressure,
				"size":     event.Size,
			}
		default:
			log.Warn().Int("action", event.Action).Msg("Unknown action type, skipping")
			continue
		}
		actions = append(actions, actionMap)
	}

	data := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"type":       "pointer",
				"parameters": map[string]string{"pointerType": "touch"},
				"id":         "touch",
				"actions":    actions,
			},
		},
	}
	option.MergeOptions(data, opts...)

	urlStr := fmt.Sprintf("/session/%s/actions/swipe", ud.Session.ID)
	_, err = ud.Session.POST(data, urlStr)
	return err
}

// SwipeWithDirection 向指定方向滑动任意距离
// direction: 滑动方向 ("up", "down", "left", "right")
// fromX, fromY: 起始坐标
// simMinDistance, simMaxDistance: 距离范围，如果相等则为固定距离，否则为随机距离
func (ud *UIA2Driver) SIMSwipeWithDirection(direction string, fromX, fromY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) error {
	absStartX, absStartY, err := convertToAbsolutePoint(ud, fromX, fromY)
	if err != nil {
		return err
	}
	// 获取设备型号和配置参数
	deviceModel, _ := ud.Device.Model()
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Str("direction", direction).
		Float64("startX", absStartX).Float64("startY", absStartY).
		Float64("minDistance", simMinDistance).Float64("maxDistance", simMaxDistance).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("UIA2Driver.SwipeWithDirection")

	// 导入滑动仿真库
	simulator := simulation.NewSlideSimulatorAPI(nil)

	// 转换方向字符串为Direction类型
	var slideDirection simulation.Direction
	switch direction {
	case "up":
		slideDirection = simulation.Up
	case "down":
		slideDirection = simulation.Down
	case "left":
		slideDirection = simulation.Left
	case "right":
		slideDirection = simulation.Right
	default:
		return fmt.Errorf("invalid direction: %s, must be one of: up, down, left, right", direction)
	}

	// 使用滑动仿真算法生成触摸事件序列
	events, err := simulator.GenerateSlideWithRandomDistance(
		absStartX, absStartY, slideDirection, simMinDistance, simMaxDistance,
		deviceParams.DeviceID, deviceParams.Pressure, deviceParams.Size)
	if err != nil {
		return fmt.Errorf("generate slide events failed: %v", err)
	}

	// 执行触摸事件序列
	return ud.TouchByEvents(events, opts...)
}

// SwipeInArea 在指定区域内向指定方向滑动任意距离
// direction: 滑动方向 ("up", "down", "left", "right")
// simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY: 区域范围(相对坐标)
// simMinDistance, simMaxDistance: 距离范围，如果相等则为固定距离，否则为随机距离
func (ud *UIA2Driver) SIMSwipeInArea(direction string, simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) error {
	// 转换区域坐标为绝对坐标
	absAreaStartX, absAreaStartY, err := convertToAbsolutePoint(ud, simAreaStartX, simAreaStartY)
	if err != nil {
		return err
	}
	absAreaEndX, absAreaEndY, err := convertToAbsolutePoint(ud, simAreaEndX, simAreaEndY)
	if err != nil {
		return err
	}

	// 确保区域坐标正确(start应该小于等于end)
	if absAreaStartX > absAreaEndX {
		absAreaStartX, absAreaEndX = absAreaEndX, absAreaStartX
	}
	if absAreaStartY > absAreaEndY {
		absAreaStartY, absAreaEndY = absAreaEndY, absAreaStartY
	}

	// 获取设备型号和配置参数
	deviceModel, _ := ud.Device.Model()
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Str("direction", direction).
		Float64("areaStartX", absAreaStartX).Float64("areaStartY", absAreaStartY).
		Float64("areaEndX", absAreaEndX).Float64("areaEndY", absAreaEndY).
		Float64("minDistance", simMinDistance).Float64("maxDistance", simMaxDistance).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("UIA2Driver.SwipeInArea")

	// 导入滑动仿真库
	simulator := simulation.NewSlideSimulatorAPI(nil)

	// 转换方向字符串为Direction类型
	var slideDirection simulation.Direction
	switch direction {
	case "up":
		slideDirection = simulation.Up
	case "down":
		slideDirection = simulation.Down
	case "left":
		slideDirection = simulation.Left
	case "right":
		slideDirection = simulation.Right
	default:
		return fmt.Errorf("invalid direction: %s, must be one of: up, down, left, right", direction)
	}

	// 使用滑动仿真算法生成区域内滑动的触摸事件序列
	events, err := simulator.GenerateSlideInArea(
		absAreaStartX, absAreaStartY, absAreaEndX, absAreaEndY,
		slideDirection, simMinDistance, simMaxDistance,
		deviceParams.DeviceID, deviceParams.Pressure, deviceParams.Size)
	if err != nil {
		return fmt.Errorf("generate slide in area events failed: %v", err)
	}

	// 执行触摸事件序列
	return ud.TouchByEvents(events, opts...)
}

// SwipeFromPointToPoint 指定起始点和结束点进行滑动
// fromX, fromY: 起始坐标(相对坐标)
// toX, toY: 结束坐标(相对坐标)
func (ud *UIA2Driver) SIMSwipeFromPointToPoint(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	// 转换起始点和结束点为绝对坐标
	absStartX, absStartY, err := convertToAbsolutePoint(ud, fromX, fromY)
	if err != nil {
		return err
	}
	absEndX, absEndY, err := convertToAbsolutePoint(ud, toX, toY)
	if err != nil {
		return err
	}

	// 获取设备型号和配置参数
	deviceModel, _ := ud.Device.Model()
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Float64("startX", absStartX).Float64("startY", absStartY).
		Float64("endX", absEndX).Float64("endY", absEndY).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("UIA2Driver.SwipeFromPointToPoint")

	// 导入滑动仿真库
	simulator := simulation.NewSlideSimulatorAPI(nil)

	// 使用滑动仿真算法生成点对点滑动的触摸事件序列
	events, err := simulator.GeneratePointToPointSlideEvents(
		absStartX, absStartY, absEndX, absEndY,
		deviceParams.DeviceID, deviceParams.Pressure, deviceParams.Size)
	if err != nil {
		return fmt.Errorf("generate point to point slide events failed: %v", err)
	}

	// 执行触摸事件序列
	return ud.TouchByEvents(events, opts...)
}

// ClickAtPoint 点击相对坐标
// x, y: 点击坐标(相对坐标)
func (ud *UIA2Driver) SIMClickAtPoint(x, y float64, opts ...option.ActionOption) error {
	// 转换为绝对坐标
	absX, absY, err := convertToAbsolutePoint(ud, x, y)
	if err != nil {
		return err
	}

	// 获取设备型号和配置参数
	deviceModel, _ := ud.Device.Model()
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Float64("x", absX).Float64("y", absY).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("UIA2Driver.ClickAtPoint")

	// 导入点击仿真库
	clickSimulator := simulation.NewClickSimulatorAPI(nil)

	// 使用点击仿真算法生成触摸事件序列
	events, err := clickSimulator.GenerateClickEvents(
		absX, absY, deviceParams.DeviceID, deviceParams.Pressure, deviceParams.Size)
	if err != nil {
		return fmt.Errorf("generate click events failed: %v", err)
	}

	// 执行触摸事件序列
	return ud.TouchByEvents(events, opts...)
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

// SIMInput 仿真输入函数，模拟人类分批输入行为
// 将文本智能分割，英文单词和数字保持完整，中文按1-2个字符分割
func (ud *UIA2Driver) SIMInput(text string, opts ...option.ActionOption) error {
	log.Info().Str("text", text).Msg("UIA2Driver.SIMInput")

	if text == "" {
		return nil
	}

	// 创建输入仿真器（使用默认配置）
	inputSimulator := simulation.NewInputSimulatorAPI(nil)

	// 生成输入片段（使用智能分割算法，所有参数使用默认值）
	inputReq := simulation.InputRequest{
		Text: text,
		// MinSegmentLen, MaxSegmentLen, MinDelayMs, MaxDelayMs 使用默认值
	}

	response := inputSimulator.GenerateInputSegments(inputReq)
	if !response.Success {
		return fmt.Errorf("failed to generate input segments: %s", response.Message)
	}

	log.Info().Int("segments", response.Metrics.TotalSegments).
		Int("totalDelayMs", response.Metrics.TotalDelayMs).
		Int("estimatedTimeMs", response.Metrics.EstimatedTimeMs).
		Msg("Input segments generated")

	// 逐个输入每个片段
	var segmentErrCnt int
	for _, segment := range response.Segments {
		// 使用SendUnicodeKeys进行输入（内部已包含Session.POST请求）
		segmentErr := ud.SendUnicodeKeys(segment.Text, opts...)
		if segmentErr != nil {
			segmentErrCnt++
			log.Info().Err(segmentErr).Int("segmentErrCnt", segmentErrCnt).
				Msg("segments err")
		}

		log.Debug().Str("segment", segment.Text).Int("index", segment.Index).
			Int("charLen", segment.CharLen).Msg("Successfully input segment")

		// 如果有延迟时间，则等待
		if segment.DelayMs > 0 {
			time.Sleep(time.Duration(segment.DelayMs) * time.Millisecond)

			log.Debug().Int("delayMs", segment.DelayMs).
				Msg("Delay between input segments")
		}
	}
	if segmentErrCnt > 0 {
		data := map[string]interface{}{
			"text": text,
		}
		option.MergeOptions(data, opts...)
		urlStr := fmt.Sprintf("/session/%s/keys", ud.Session.ID)
		_, err := ud.Session.POST(data, urlStr)
		return err
	}
	log.Info().Int("totalSegments", response.Metrics.TotalSegments).
		Int("actualDelayMs", response.Metrics.TotalDelayMs).
		Msg("SIMInput completed successfully")

	return nil
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
			log.Error().Str("output", out).Msg("uiautomator server crashed, retrying...")
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
