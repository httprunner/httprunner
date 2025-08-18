package uixt

import (
	"bytes"
	"context"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/internal/simulation"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func NewWDADriver(device *IOSDevice) (*WDADriver, error) {
	log.Info().Interface("device", device).Msg("init ios WDA driver")
	driver := &WDADriver{
		Device:  device,
		Session: NewDriverSession(),
	}

	// setup driver
	if err := driver.Setup(); err != nil {
		return nil, err
	}

	// register driver session reset handler
	driver.Session.RegisterResetHandler(driver.Setup)

	return driver, nil
}

type WDADriver struct {
	Device  *IOSDevice
	Session *DriverSession

	// cache to avoid repeated query
	windowSize types.Size
	scale      float64

	mjpegClient *http.Client
	mjpegUrl    string
}

func (wd *WDADriver) getLocalPort() (int, error) {
	localPort, err := wd.Device.Forward(wd.Device.Options.WDAPort)
	if err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	return localPort, nil
}

func (wd *WDADriver) getMjpegLocalPort() (int, error) {
	localMjpegPort, err := wd.Device.Forward(wd.Device.Options.WDAMjpegPort)
	if err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	log.Info().Int("WDA_LOCAL_MJPEG_PORT", localMjpegPort).
		Msg("reuse WDA local mjpeg port")
	return localMjpegPort, nil
}

func (wd *WDADriver) initMjpegClient() error {
	host := "localhost"
	localMjpegPort, err := wd.getMjpegLocalPort()
	if err != nil {
		return err
	}
	mjpegHTTPConn, err := net.Dial(
		"tcp",
		net.JoinHostPort(host, fmt.Sprintf("%d", localMjpegPort)),
	)
	if err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	wd.mjpegClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return mjpegHTTPConn, nil
			},
		},
		Timeout: 30 * time.Second,
	}
	wd.mjpegUrl = fmt.Sprintf("http://%s:%d", host, localMjpegPort)
	return nil
}

func (wd *WDADriver) GetMjpegClient() *http.Client {
	return wd.mjpegClient
}

func (wd *WDADriver) Setup() error {
	localPort, err := wd.getLocalPort()
	if err != nil {
		return err
	}
	err = wd.Session.SetupPortForward(localPort)
	if err != nil {
		return err
	}

	// Store base URL for building full URLs
	baseURL := fmt.Sprintf("http://localhost:%d", localPort)
	wd.Session.SetBaseURL(baseURL)
	// create new session
	if err := wd.InitSession(nil); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}

	return nil
}

func (wd *WDADriver) TearDown() error {
	return nil
}

func (wd *WDADriver) InitSession(capabilities option.Capabilities) error {
	// [[FBRoute POST:@"/session"].withoutSession respondWithTarget:self action:@selector(handleCreateSession:)]
	data := make(map[string]interface{})
	if len(capabilities) == 0 {
		data["capabilities"] = make(map[string]interface{})
	} else {
		data["capabilities"] = map[string]interface{}{"alwaysMatch": capabilities}
	}
	// No longer need to track session ID
	return nil
}

func (wd *WDADriver) DeleteSession() (err error) {
	if wd.mjpegClient != nil {
		wd.mjpegClient.CloseIdleConnections()
	}

	wd.Session.client.CloseIdleConnections()
	return
}

func (wd *WDADriver) Status() (deviceStatus types.DeviceStatus, err error) {
	// [[FBRoute GET:@"/status"].withoutSession respondWithTarget:self action:@selector(handleGetStatus:)]
	var rawResp DriverRawResponse
	// Notice: use Driver.GET instead of httpGET to avoid loop calling
	if rawResp, err = wd.Session.GET("/status"); err != nil {
		return types.DeviceStatus{}, err
	}
	reply := new(struct{ Value struct{ types.DeviceStatus } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.DeviceStatus{}, err
	}
	deviceStatus = reply.Value.DeviceStatus
	return
}

func (wd *WDADriver) GetDevice() IDevice {
	return wd.Device
}

func (wd *WDADriver) DeviceInfo() (deviceInfo types.DeviceInfo, err error) {
	// [[FBRoute GET:@"/wda/device/info"] respondWithTarget:self action:@selector(handleGetDeviceInfo:)]
	// [[FBRoute GET:@"/wda/device/info"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/device/info"); err != nil {
		return types.DeviceInfo{}, err
	}
	reply := new(struct{ Value struct{ types.DeviceInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.DeviceInfo{}, err
	}
	deviceInfo = reply.Value.DeviceInfo
	return
}

func (wd *WDADriver) Location() (location types.Location, err error) {
	// [[FBRoute GET:@"/wda/device/location"] respondWithTarget:self action:@selector(handleGetLocation:)]
	// [[FBRoute GET:@"/wda/device/location"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/device/location"); err != nil {
		return types.Location{}, err
	}
	reply := new(struct{ Value struct{ types.Location } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.Location{}, err
	}
	location = reply.Value.Location
	return
}

func (wd *WDADriver) BatteryInfo() (batteryInfo types.BatteryInfo, err error) {
	// [[FBRoute GET:@"/wda/batteryInfo"] respondWithTarget:self action:@selector(handleGetBatteryInfo:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/batteryInfo"); err != nil {
		return types.BatteryInfo{}, err
	}
	reply := new(struct{ Value struct{ types.BatteryInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.BatteryInfo{}, err
	}
	batteryInfo = reply.Value.BatteryInfo
	return
}

func (wd *WDADriver) WindowSize() (size types.Size, err error) {
	// [[FBRoute GET:@"/window/size"] respondWithTarget:self action:@selector(handleGetWindowSize:)]
	if !wd.windowSize.IsNil() {
		// use cached window size
		return wd.windowSize, nil
	}

	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wings/window/size"); err != nil {
		return types.Size{}, errors.Wrap(err, "get window size failed by WDA request")
	}
	reply := new(struct{ Value struct{ types.Size } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.Size{}, errors.Wrap(err, "get window size failed by WDA response")
	}
	size = reply.Value.Size
	scale, err := wd.Scale()
	if err != nil {
		return types.Size{}, errors.Wrap(err, "get window size scale failed")
	}
	size.Height = size.Height * int(scale)
	size.Width = size.Width * int(scale)

	wd.windowSize = size // cache window size
	return wd.windowSize, nil
}

func (wd *WDADriver) Scale() (float64, error) {
	if !builtin.IsZeroFloat64(wd.scale) {
		return wd.scale, nil
	}
	screen, err := wd.Screen()
	if err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get screen info failed: %v", err))
	}
	return screen.Scale, nil
}

type Screen struct {
	types.Size
	Scale float64 `json:"scale"`
}

func (wd *WDADriver) Screen() (screen Screen, err error) {
	// [[FBRoute GET:@"/wings/window/size"] respondWithTarget:self action:@selector(handleGetScreen:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wings/window/size"); err != nil {
		return Screen{}, err
	}
	reply := new(struct{ Value struct{ Screen } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return Screen{}, err
	}
	screen = reply.Value.Screen
	return
}

func (wd *WDADriver) ScreenShot(opts ...option.ActionOption) (raw *bytes.Buffer, err error) {
	// [[FBRoute GET:@"/screenshot"] respondWithTarget:self action:@selector(handleGetScreenshot:)]
	// [[FBRoute GET:@"/screenshot"].withoutSession respondWithTarget:self action:@selector(handleGetScreenshot:)]
	rawResp, err := wd.Session.GET("/screenshot")
	if err != nil {
		return nil, errors.Wrap(code.DeviceScreenShotError,
			fmt.Sprintf("WDA screenshot failed %v", err))
	}
	raw, err = rawResp.ValueDecodeAsBase64()
	if err != nil {
		return nil, errors.Wrap(code.DeviceScreenShotError,
			fmt.Sprintf("decode WDA screenshot data failed: %v", err))
	}
	return raw, nil
}

func (wd *WDADriver) toScale(x float64) (float64, error) {
	if wd.scale == 0 {
		// not setup yet
		var err error
		if wd.scale, err = wd.Scale(); err != nil || wd.scale == 0 {
			log.Error().Err(err).Msg("get screen scale failed")
			return 0, err
		}
	}
	return x / wd.scale, nil
}

func (wd *WDADriver) ActiveAppInfo() (info types.AppInfo, err error) {
	// [[FBRoute GET:@"/wda/activeAppInfo"] respondWithTarget:self action:@selector(handleActiveAppInfo:)]
	// [[FBRoute GET:@"/wda/activeAppInfo"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wings/activeAppInfo"); err != nil {
		return types.AppInfo{}, err
	}
	reply := new(struct{ Value struct{ types.AppInfo } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.AppInfo{}, err
	}
	info = reply.Value.AppInfo
	return
}

func (wd *WDADriver) ActiveAppsList() (appsList []types.AppBaseInfo, err error) {
	// [[FBRoute GET:@"/wda/apps/list"] respondWithTarget:self action:@selector(handleGetActiveAppsList:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/apps/list"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value []types.AppBaseInfo })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	appsList = reply.Value
	return
}

func (wd *WDADriver) AppState(bundleId string) (runState types.AppState, err error) {
	log.Warn().Str("bundleId", bundleId).Msg("WDADriver.AppState not implemented")
	return 0, nil
}

func (wd *WDADriver) IsLocked() (locked bool, err error) {
	// [[FBRoute GET:@"/wda/locked"] respondWithTarget:self action:@selector(handleIsLocked:)]
	// [[FBRoute GET:@"/wda/locked"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/locked"); err != nil {
		return false, err
	}
	if locked, err = rawResp.ValueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (wd *WDADriver) Unlock() (err error) {
	// [[FBRoute POST:@"/wda/unlock"] respondWithTarget:self action:@selector(handleUnlock:)]
	// [[FBRoute POST:@"/wda/unlock"].withoutSession
	_, err = wd.Session.POST(nil, "/wda/unlock")
	return
}

func (wd *WDADriver) Lock() (err error) {
	// [[FBRoute POST:@"/wda/lock"] respondWithTarget:self action:@selector(handleLock:)]
	// [[FBRoute POST:@"/wda/lock"].withoutSession
	_, err = wd.Session.POST(nil, "/wda/lock")
	return
}

func (wd *WDADriver) Home() (err error) {
	// [[FBRoute POST:@"/wda/homescreen"].withoutSession respondWithTarget:self action:@selector(handleHomescreenCommand:)]
	_, err = wd.Session.POST(nil, "/wda/homescreen")
	return
}

func (wd *WDADriver) AlertText() (text string, err error) {
	// [[FBRoute GET:@"/alert/text"] respondWithTarget:self action:@selector(handleAlertGetTextCommand:)]
	// [[FBRoute GET:@"/alert/text"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/alert/text"); err != nil {
		return "", err
	}
	if text, err = rawResp.ValueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (wd *WDADriver) AlertButtons() (btnLabels []string, err error) {
	// [[FBRoute GET:@"/wda/alert/buttons"] respondWithTarget:self action:@selector(handleGetAlertButtonsCommand:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/alert/buttons"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value []string })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	btnLabels = reply.Value
	return
}

func (wd *WDADriver) AlertAccept(label ...string) (err error) {
	// [[FBRoute POST:@"/alert/accept"] respondWithTarget:self action:@selector(handleAlertAcceptCommand:)]
	// [[FBRoute POST:@"/alert/accept"].withoutSession
	data := make(map[string]interface{})
	if len(label) != 0 && label[0] != "" {
		data["name"] = label[0]
	}
	_, err = wd.Session.POST(data, "/alert/accept")
	return
}

func (wd *WDADriver) AlertDismiss(label ...string) (err error) {
	// [[FBRoute POST:@"/alert/dismiss"] respondWithTarget:self action:@selector(handleAlertDismissCommand:)]
	// [[FBRoute POST:@"/alert/dismiss"].withoutSession
	data := make(map[string]interface{})
	if len(label) != 0 && label[0] != "" {
		data["name"] = label[0]
	}
	_, err = wd.Session.POST(data, "/alert/dismiss")
	return
}

func (wd *WDADriver) AlertSendKeys(text string) (err error) {
	// [[FBRoute POST:@"/alert/text"] respondWithTarget:self action:@selector(handleAlertSetTextCommand:)]
	data := map[string]interface{}{"value": strings.Split(text, "")}
	_, err = wd.Session.POST(data, "/alert/text")
	return
}

func (wd *WDADriver) AppLaunch(bundleId string) (err error) {
	log.Info().Str("bundleId", bundleId).Msg("WDADriver.AppLaunch")
	// [[FBRoute POST:@"/wda/apps/launch"] respondWithTarget:self action:@selector(handleSessionAppLaunch:)]
	data := make(map[string]interface{})
	data["bundleId"] = bundleId
	data["environment"] = map[string]interface{}{
		"SHOW_EXPLORER": "NO",
	}
	// 超时两分钟
	_, err = wd.Session.POST(data, "/wings/apps/launch", option.WithTimeout(120))
	if err != nil {
		// Check for untrusted certificate error
		errMsg := err.Error()
		if strings.Contains(errMsg, "has not been explicitly trusted by the user") ||
			strings.Contains(errMsg, "invalid code signature") ||
			strings.Contains(errMsg, "inadequate entitlements") {
			return errors.Wrap(code.DeviceUntrustedCertError, "App certificate not trusted: "+bundleId)
		}
		return errors.Wrap(err, "wda app launch failed")
	}
	return nil
}

func (wd *WDADriver) AppLaunchUnattached(bundleId string) (err error) {
	log.Info().Str("bundleId", bundleId).Msg("WDADriver.AppLaunchUnattached")
	// [[FBRoute POST:@"/wda/apps/launchUnattached"].withoutSession respondWithTarget:self action:@selector(handleLaunchUnattachedApp:)]]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.Session.POST(data, "/wda/apps/launchUnattached")
	if err != nil {
		// Check for untrusted certificate error
		errMsg := err.Error()
		if strings.Contains(errMsg, "has not been explicitly trusted by the user") ||
			strings.Contains(errMsg, "invalid code signature") ||
			strings.Contains(errMsg, "inadequate entitlements") {
			return errors.Wrap(code.DeviceUntrustedCertError, "App certificate not trusted: "+bundleId)
		}
		return errors.Wrap(err, "wda app launchUnattached failed")
	}
	return nil
}

func (wd *WDADriver) AppTerminate(bundleId string) (successful bool, err error) {
	log.Info().Str("bundleId", bundleId).Msg("WDADriver.AppTerminate")
	// [[FBRoute POST:@"/wda/apps/terminate"] respondWithTarget:self action:@selector(handleSessionAppTerminate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.POST(data, "/wings/apps/terminate"); err != nil {
		return false, err
	}
	if successful, err = rawResp.ValueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (wd *WDADriver) AppActivate(bundleId string) (err error) {
	log.Info().Str("bundleId", bundleId).Msg("WDADriver.AppActivate")
	// [[FBRoute POST:@"/wda/apps/activate"] respondWithTarget:self action:@selector(handleSessionAppActivate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.Session.POST(data, "/wings/apps/activate")
	return
}

func (wd *WDADriver) AppDeactivate(second float64) (err error) {
	log.Warn().Float64("second", second).Msg("WDADriver.AppDeactivate not implemented")
	return nil
}

func (wd *WDADriver) ForegroundInfo() (appInfo types.AppInfo, err error) {
	activeAppInfo, err := wd.ActiveAppInfo()
	appInfo.BundleId = activeAppInfo.BundleId
	if err != nil {
		return appInfo, err
	}
	apps, err := wd.Device.ListApps(ApplicationTypeAny)
	if err != nil {
		return appInfo, err
	}
	for _, app := range apps {
		if app.CFBundleIdentifier == activeAppInfo.BundleId {
			appInfo.Name = app.CFBundleDisplayName
			appInfo.AppName = app.CFBundleName
			appInfo.BundleId = app.CFBundleIdentifier
			appInfo.PackageName = app.CFBundleIdentifier
			appInfo.VersionName = app.CFBundleShortVersionString
			appInfo.VersionCode = app.CFBundleVersion
			return appInfo, err
		}
	}
	return appInfo, err
}

func (wd *WDADriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("WDADriver.TapXY")
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]
	absX, absY, err := convertToAbsolutePoint(wd, x, y)
	if err != nil {
		return err
	}
	return wd.TapAbsXY(absX, absY, opts...)
}

func (wd *WDADriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("WDADriver.TapAbsXY")
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]

	var err error
	actionOptions := option.NewActionOptions(opts...)
	x, y, err = preHandler_TapAbsXY(wd, actionOptions, x, y)
	if err != nil {
		return err
	}
	defer postHandler(wd, option.ACTION_TapAbsXY, actionOptions)

	if x, err = wd.toScale(x); err != nil {
		return err
	}
	if y, err = wd.toScale(y); err != nil {
		return err
	}

	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	option.MergeOptions(data, opts...)

	_, err = wd.Session.POST(data, "/wings/interaction/tap")
	return err
}

func (wd *WDADriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("WDADriver.DoubleTap")
	// [[FBRoute POST:@"/wda/doubleTap"] respondWithTarget:self action:@selector(handleDoubleTapCoordinate:)]

	var err error
	actionOptions := option.NewActionOptions(opts...)
	x, y, err = preHandler_DoubleTap(wd, actionOptions, x, y)
	if err != nil {
		return err
	}
	defer postHandler(wd, option.ACTION_DoubleTapXY, actionOptions)

	if x, err = wd.toScale(x); err != nil {
		return err
	}
	if y, err = wd.toScale(y); err != nil {
		return err
	}
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.Session.POST(data, "/wings/interaction/doubleTap")
	return err
}

// FIXME: hold not work
func (wd *WDADriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	log.Info().Float64("x", x).Float64("y", y).Msg("WDADriver.TouchAndHold")
	actionOptions := option.NewActionOptions(opts...)
	x, y = actionOptions.ApplyTapOffset(x, y)
	if actionOptions.Duration == 0 {
		opts = append(opts, option.WithPressDuration(1))
	}
	return wd.TapXY(x, y, opts...)
}

func (wd *WDADriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	log.Info().Float64("fromX", fromX).Float64("fromY", fromY).
		Float64("toX", toX).Float64("toY", toY).Msg("WDADriver.Drag")
	// [[FBRoute POST:@"/wda/dragfromtoforduration"] respondWithTarget:self action:@selector(handleDragCoordinate:)]

	var err error
	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err = preHandler_Drag(wd, actionOptions, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	defer postHandler(wd, option.ACTION_Drag, actionOptions)

	if fromX, err = wd.toScale(fromX); err != nil {
		return err
	}
	if fromY, err = wd.toScale(fromY); err != nil {
		return err
	}
	if toX, err = wd.toScale(toX); err != nil {
		return err
	}
	if toY, err = wd.toScale(toY); err != nil {
		return err
	}
	data := map[string]interface{}{
		"fromX": builtin.RoundToOneDecimal(fromX),
		"fromY": builtin.RoundToOneDecimal(fromY),
		"toX":   builtin.RoundToOneDecimal(toX),
		"toY":   builtin.RoundToOneDecimal(toY),
	}
	option.MergeOptions(data, opts...)
	// wda 50 version
	_, err = wd.Session.POST(data, "/wings/interaction/drag")
	return err
}

func (wd *WDADriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	return wd.Drag(fromX, fromY, toX, toY, opts...)
}

// TouchByEvents performs a complex swipe using a sequence of touch events with pressure and size data
func (wd *WDADriver) TouchByEvents(events []types.TouchEvent, opts ...option.ActionOption) error {
	log.Info().Int("eventCount", len(events)).Msg("WDADriver.SwipeSimulator")

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

	fromX, fromY, toX, toY, err := preHandler_Swipe(wd, option.ACTION_SwipeCoordinate, actionOptions,
		startX, startY, endX, endY)
	if err != nil {
		return err
	}
	defer postHandler(wd, option.ACTION_SwipeCoordinate, actionOptions)

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

		if x, err = wd.toScale(x); err != nil {
			return err
		}
		if y, err = wd.toScale(y); err != nil {
			return err
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

	_, err = wd.Session.POST(data, "/wings/actions")
	return err
}

// SIMSwipeWithDirection 向指定方向滑动任意距离
// direction: 滑动方向 ("up", "down", "left", "right")
// fromX, fromY: 起始坐标
// simMinDistance, simMaxDistance: 距离范围，如果相等则为固定距离，否则为随机距离
func (wd *WDADriver) SIMSwipeWithDirection(direction string, fromX, fromY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) error {
	absStartX, absStartY, err := convertToAbsolutePoint(wd, fromX, fromY)
	if err != nil {
		return err
	}
	// 获取设备型号和配置参数
	deviceModel := "iphone"
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Str("direction", direction).
		Float64("startX", absStartX).Float64("startY", absStartY).
		Float64("minDistance", simMinDistance).Float64("maxDistance", simMaxDistance).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("WDADriver.SIMSwipeWithDirection")

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
	return wd.TouchByEvents(events, opts...)
}

// SIMSwipeInArea 在指定区域内向指定方向滑动任意距离
// direction: 滑动方向 ("up", "down", "left", "right")
// simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY: 区域范围(相对坐标)
// simMinDistance, simMaxDistance: 距离范围，如果相等则为固定距离，否则为随机距离
func (wd *WDADriver) SIMSwipeInArea(direction string, simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) error {
	// 转换区域坐标为绝对坐标
	absAreaStartX, absAreaStartY, err := convertToAbsolutePoint(wd, simAreaStartX, simAreaStartY)
	if err != nil {
		return err
	}
	absAreaEndX, absAreaEndY, err := convertToAbsolutePoint(wd, simAreaEndX, simAreaEndY)
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
	deviceModel := "iphone"
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Str("direction", direction).
		Float64("areaStartX", absAreaStartX).Float64("areaStartY", absAreaStartY).
		Float64("areaEndX", absAreaEndX).Float64("areaEndY", absAreaEndY).
		Float64("minDistance", simMinDistance).Float64("maxDistance", simMaxDistance).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("WDADriver.SIMSwipeInArea")

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
	return wd.TouchByEvents(events, opts...)
}

// SIMSwipeFromPointToPoint 指定起始点和结束点进行滑动
// fromX, fromY: 起始坐标(相对坐标)
// toX, toY: 结束坐标(相对坐标)
func (wd *WDADriver) SIMSwipeFromPointToPoint(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	// 转换起始点和结束点为绝对坐标
	absStartX, absStartY, err := convertToAbsolutePoint(wd, fromX, fromY)
	if err != nil {
		return err
	}
	absEndX, absEndY, err := convertToAbsolutePoint(wd, toX, toY)
	if err != nil {
		return err
	}

	// 获取设备型号和配置参数
	deviceModel := "iphone"
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Float64("startX", absStartX).Float64("startY", absStartY).
		Float64("endX", absEndX).Float64("endY", absEndY).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("WDADriver.SIMSwipeFromPointToPoint")

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
	return wd.TouchByEvents(events, opts...)
}

// SIMClickAtPoint 点击相对坐标
// x, y: 点击坐标(相对坐标)
func (wd *WDADriver) SIMClickAtPoint(x, y float64, opts ...option.ActionOption) error {
	// 转换为绝对坐标
	absX, absY, err := convertToAbsolutePoint(wd, x, y)
	if err != nil {
		return err
	}

	// 获取设备型号和配置参数
	deviceModel := "iphone"
	deviceParams := simulation.GetRandomDeviceParams(deviceModel)

	log.Info().Float64("x", absX).Float64("y", absY).
		Str("deviceModel", deviceModel).
		Int("deviceID", deviceParams.DeviceID).
		Float64("pressure", deviceParams.Pressure).
		Float64("size", deviceParams.Size).
		Msg("WDADriver.SIMClickAtPoint")

	// 导入点击仿真库
	clickSimulator := simulation.NewClickSimulatorAPI(nil)

	// 使用点击仿真算法生成触摸事件序列
	events, err := clickSimulator.GenerateClickEvents(
		absX, absY, deviceParams.DeviceID, deviceParams.Pressure, deviceParams.Size)
	if err != nil {
		return fmt.Errorf("generate click events failed: %v", err)
	}

	// 执行触摸事件序列
	return wd.TouchByEvents(events, opts...)
}

func (wd *WDADriver) SetPasteboard(contentType types.PasteboardType, content string) (err error) {
	// [[FBRoute POST:@"/wda/setPasteboard"] respondWithTarget:self action:@selector(handleSetPasteboard:)]
	data := map[string]interface{}{
		"contentType": contentType,
		"content":     base64.StdEncoding.EncodeToString([]byte(content)),
	}
	_, err = wd.Session.POST(data, "/wda/setPasteboard")
	return
}

func (wd *WDADriver) GetPasteboard() (content string, err error) {
	// [[FBRoute POST:@"/wda/getPasteboard"] respondWithTarget:self action:@selector(handleGetPasteboard:)]
	err = wd.AppLaunch("com.gtf.wda.runner.xctrunner")
	if err != nil {
		return "", errors.Wrap(err, "GetPasteboard failed. WDA app not launched")
	}
	data := map[string]interface{}{}
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.POST(data, "/wda/getPasteboard"); err != nil {
		return "", err
	}
	var raw *bytes.Buffer
	if raw, err = rawResp.ValueDecodeAsBase64(); err != nil {
		return "", err
	}
	return string(raw.Bytes()), nil
}

func (wd *WDADriver) SetIme(ime string) error {
	return types.ErrDriverNotImplemented
}

func (wd *WDADriver) Input(text string, opts ...option.ActionOption) (err error) {
	log.Info().Str("text", text).Msg("WDADriver.Input")
	// [[FBRoute POST:@"/wda/keys"] respondWithTarget:self action:@selector(handleKeys:)]
	data := map[string]interface{}{"value": strings.Split(text, "")}
	option.MergeOptions(data, opts...)
	_, err = wd.Session.POST(data, "/wings/interaction/keys")
	return
}

// SIMInput 仿真输入函数，模拟人类分批输入行为
// 将文本智能分割，英文单词和数字保持完整，中文按1-2个字符分割
func (wd *WDADriver) SIMInput(text string, opts ...option.ActionOption) error {
	log.Info().Str("text", text).Msg("WDADriver.SIMInput")

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
		// 使用Input进行输入（内部已包含Session.POST请求）
		segmentErr := wd.Input(segment.Text, opts...)
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
		data := map[string]interface{}{"value": strings.Split(text, "")}
		option.MergeOptions(data, opts...)
		_, err := wd.Session.POST(data, "/wings/interaction/keys")
		return err
	}
	log.Info().Int("totalSegments", response.Metrics.TotalSegments).
		Int("actualDelayMs", response.Metrics.TotalDelayMs).
		Msg("SIMInput completed successfully")

	return nil
}

func (wd *WDADriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	log.Info().Int("count", count).Msg("WDADriver.Backspace")
	if count == 0 {
		return nil
	}
	data := map[string]interface{}{"count": count}
	option.MergeOptions(data, opts...)
	_, err = wd.Session.POST(data, "/gtf/interaction/input/backspace")
	return
}

func (wd *WDADriver) AppClear(packageName string) error {
	return types.ErrDriverNotImplemented
}

// Back simulates a short press on the BACK button.
func (wd *WDADriver) Back() (err error) {
	log.Info().Msg("WDADriver.Back")
	return wd.Swipe(0, 0.5, 0.6, 0.5)
}

func (wd *WDADriver) PressButton(button types.DeviceButton) (err error) {
	// [[FBRoute POST:@"/wda/pressButton"] respondWithTarget:self action:@selector(handlePressButtonCommand:)]

	if button == types.DeviceButtonEnter {
		return wd.Input("\n")
	}

	data := map[string]interface{}{"name": button}
	_, err = wd.Session.POST(data, "/wda/pressButton")
	return
}

func (wd *WDADriver) Orientation() (orientation types.Orientation, err error) {
	// [[FBRoute GET:@"/orientation"] respondWithTarget:self action:@selector(handleGetOrientation:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/orientation"); err != nil {
		return "", err
	}
	reply := new(struct{ Value types.Orientation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return "", err
	}
	orientation = reply.Value
	return
}

func (wd *WDADriver) SetOrientation(orientation types.Orientation) (err error) {
	// [[FBRoute POST:@"/orientation"] respondWithTarget:self action:@selector(handleSetOrientation:)]
	data := map[string]interface{}{"orientation": orientation}
	_, err = wd.Session.POST(data, "/orientation")
	return
}

func (wd *WDADriver) Rotation() (rotation types.Rotation, err error) {
	// [[FBRoute GET:@"/rotation"] respondWithTarget:self action:@selector(handleGetRotation:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/rotation"); err != nil {
		return types.Rotation{}, err
	}
	reply := new(struct{ Value types.Rotation })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return types.Rotation{}, err
	}
	rotation = reply.Value
	return
}

func (wd *WDADriver) SetRotation(rotation types.Rotation) (err error) {
	// [[FBRoute POST:@"/rotation"] respondWithTarget:self action:@selector(handleSetRotation:)]
	_, err = wd.Session.POST(rotation, "/rotation")
	return
}

func (wd *WDADriver) Source(srcOpt ...option.SourceOption) (source string, err error) {
	// [[FBRoute GET:@"/source"] respondWithTarget:self action:@selector(handleGetSourceCommand:)]
	// [[FBRoute GET:@"/source"].withoutSession

	options := option.NewSourceOptions(srcOpt...)
	query := options.Query()
	if len(query) > 0 {
		query = "?" + query
	}
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/source" + query); err != nil {
		return "", err
	}
	// json format
	if options.Format == option.SourceFormatJSON {
		var jr builtinJSON.RawMessage
		if jr, err = rawResp.ValueConvertToJsonRawMessage(); err != nil {
			return "", err
		}
		return string(jr), nil
	}

	// xml/description format
	if source, err = rawResp.ValueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (wd *WDADriver) AccessibleSource() (source string, err error) {
	// [[FBRoute GET:@"/wda/accessibleSource"] respondWithTarget:self action:@selector(handleGetAccessibleSourceCommand:)]
	// [[FBRoute GET:@"/wda/accessibleSource"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/wda/accessibleSource"); err != nil {
		return "", err
	}
	var jr builtinJSON.RawMessage
	if jr, err = rawResp.ValueConvertToJsonRawMessage(); err != nil {
		return "", err
	}
	source = string(jr)
	return
}

func (wd *WDADriver) HealthCheck() (err error) {
	// [[FBRoute GET:@"/wda/healthcheck"].withoutSession respondWithTarget:self action:@selector(handleGetHealthCheck:)]
	_, err = wd.Session.GET("/wda/healthcheck")
	return
}

func (wd *WDADriver) IsHealthy() (healthy bool, err error) {
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/health"); err != nil {
		return false, err
	}
	if string(rawResp) != "I-AM-ALIVE" {
		return false, nil
	}
	return true, nil
}

func (wd *WDADriver) GetAppiumSettings() (settings map[string]interface{}, err error) {
	// [[FBRoute GET:@"/appium/settings"] respondWithTarget:self action:@selector(handleGetSettings:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.GET("/appium/settings"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	settings = reply.Value
	return
}

func (wd *WDADriver) SetAppiumSettings(settings map[string]interface{}) (ret map[string]interface{}, err error) {
	// [[FBRoute POST:@"/appium/settings"] respondWithTarget:self action:@selector(handleSetSettings:)]
	data := map[string]interface{}{"settings": settings}
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.POST(data, "/appium/settings"); err != nil {
		return nil, err
	}
	reply := new(struct{ Value map[string]interface{} })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}
	ret = reply.Value
	return
}

func (wd *WDADriver) WdaShutdown() (err error) {
	_, err = wd.Session.GET("/wda/shutdown")
	return
}

func (wd *WDADriver) triggerWDALog(data map[string]interface{}) (rawResp []byte, err error) {
	// [[FBRoute POST:@"/gtf/automation/log"].withoutSession respondWithTarget:self action:@selector(handleAutomationLog:)]
	return wd.Session.POST(data, "/gtf/automation/log")
}

func (wd *WDADriver) ScreenRecord(opts ...option.ActionOption) (videoPath string, err error) {
	log.Info().Msg("WDADriver.ScreenRecord")
	err = wd.initMjpegClient()
	if err != nil {
		return "", err
	}
	timestamp := time.Now().Format("20060102_150405") + fmt.Sprintf("_%03d", time.Now().UnixNano()/1e6%1000)
	fileName := filepath.Join(config.GetConfig().ScreenShotsPath(), fmt.Sprintf("%s.mp4", timestamp))

	options := option.NewActionOptions(opts...)
	duration := time.Duration(options.Duration * float64(time.Second))

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	// ffmpeg 命令
	cmd := exec.Command(
		"ffmpeg",
		"-use_wallclock_as_timestamps", "1",
		"-f", "mjpeg",
		"-y",
		"-r", "10",
		"-i", wd.mjpegUrl,
		"-c:v", "libx264",
		"-vf", "pad=width=ceil(iw/2)*2:height=ceil(ih/2)*2",
		fileName,
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	// 启动命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting ffmpeg command:", err)
		return "", err
	}
	timer := time.After(duration)

	done := make(chan error)
	go func() {
		// 等待 ffmpeg 命令执行完毕
		done <- cmd.Wait()
	}()
	select {
	case <-timer:
		// 超时，停止 ffmpeg 进程
		fmt.Println("Time is up, stopping ffmpeg command...")
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			fmt.Println("Error killing ffmpeg process:", err)
		}
	case err := <-done:
		// ffmpeg 正常结束
		if err != nil {
			fmt.Println("FFmpeg finished with error:", err)
		} else {
			fmt.Println("FFmpeg finished successfully")
		}
	}
	return filepath.Abs(fileName)
}

func (wd *WDADriver) StartCaptureLog(identifier ...string) error {
	log.Info().Msg("start WDA log recording")
	if identifier == nil {
		identifier = []string{""}
	}
	data := map[string]interface{}{"action": "start", "type": 2, "identifier": identifier[0]}
	_, err := wd.triggerWDALog(data)
	if err != nil {
		return errors.Wrap(code.DeviceCaptureLogError,
			fmt.Sprintf("start WDA log recording failed: %v", err))
	}

	return nil
}

func (wd *WDADriver) PushImage(localPath string) error {
	log.Info().Str("localPath", localPath).Msg("WDADriver.PushImage")
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	imageBytes, err := io.ReadAll(localFile)
	data := map[string]interface{}{
		"file_name": path.Base(localPath),
		"file_data": base64.StdEncoding.EncodeToString(imageBytes),
	}
	if err != nil {
		return err
	}

	_, err = wd.Session.POST(data, "/wings/albums/add")
	return err
}

func (wd *WDADriver) PullImages(localDir string) error {
	log.Warn().Msg("PullImages not implemented in WDADriver")
	return nil
}

func (wd *WDADriver) ClearImages() error {
	log.Info().Msg("WDADriver.ClearImages")
	data := map[string]interface{}{}

	_, err := wd.Session.POST(data, "/wings/albums/clear")
	return err
}

func (wd *WDADriver) PushFile(localPath string, remoteDir string) error {
	log.Warn().Msg("PushFile not implemented in WDADriver")
	return nil
}

func (wd *WDADriver) PullFiles(localDir string, remoteDirs ...string) error {
	log.Warn().Msg("PullFiles not implemented in WDADriver")
	return nil
}

func (wd *WDADriver) ClearFiles(paths ...string) error {
	log.Warn().Msg("ClearFiles not implemented in WDADriver")
	return nil
}

type wdaResponse struct {
	Status    int         `json:"status"`
	SessionID string      `json:"sessionId"`
	Value     interface{} `json:"value"`
}

func (wd *WDADriver) StopCaptureLog() (result interface{}, err error) {
	log.Info().Msg("stop log recording")
	data := map[string]interface{}{"action": "stop"}
	rawResp, err := wd.triggerWDALog(data)
	if err != nil {
		log.Error().Err(err).Bytes("rawResp", rawResp).Msg("failed to get WDA logs")
		return "", errors.Wrap(code.DeviceCaptureLogError,
			fmt.Sprintf("get WDA logs failed: %v", err))
	}
	reply := new(wdaResponse)
	if err = json.Unmarshal(rawResp, reply); err != nil {
		log.Error().Err(err).Bytes("rawResp", rawResp).Msg("failed to json.Unmarshal WDA logs")
		return reply, errors.Wrap(code.DeviceCaptureLogError,
			fmt.Sprintf("json.Unmarshal WDA logs failed: %v", err))
	}
	log.Info().Interface("value", reply.Value).Msg("get WDA log response")
	return reply.Value, nil
}

func (wd *WDADriver) GetSession() *DriverSession {
	return wd.Session
}

func (wd *WDADriver) HoverBySelector(selector string, options ...option.ActionOption) (err error) {
	return err
}

func (wd *WDADriver) TapBySelector(text string, opts ...option.ActionOption) error {
	return nil
}

func (wd *WDADriver) SecondaryClick(x, y float64) (err error) {
	return err
}

func (wd *WDADriver) SecondaryClickBySelector(selector string, options ...option.ActionOption) (err error) {
	return err
}
