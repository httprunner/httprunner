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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
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
	localPort, err := strconv.Atoi(os.Getenv("WDA_LOCAL_PORT"))
	if err != nil {
		localPort, err = builtin.GetFreePort()
		if err != nil {
			return 0, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}
		// forward local port to device
		if err = wd.Device.Forward(localPort, wd.Device.Options.WDAPort); err != nil {
			return 0, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("forward tcp port failed: %v", err))
		}
	} else {
		log.Info().Int("WDA_LOCAL_PORT", localPort).Msg("reuse WDA local port")
	}
	return localPort, nil
}

func (wd *WDADriver) getMjpegLocalPort() (int, error) {
	localMjpegPort, err := strconv.Atoi(os.Getenv("WDA_LOCAL_MJPEG_PORT"))
	if err != nil {
		localMjpegPort, err = builtin.GetFreePort()
		if err != nil {
			return 0, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}
		if err = wd.Device.Forward(localMjpegPort, wd.Device.Options.WDAMjpegPort); err != nil {
			return 0, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("forward tcp port failed: %v", err))
		}
	} else {
		log.Info().Int("WDA_LOCAL_MJPEG_PORT", localMjpegPort).
			Msg("reuse WDA local mjpeg port")
	}
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
		return 0, errors.Wrap(code.MobileUIDriverError,
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
	_, err = wd.Session.POST(data, "/wings/apps/launch")
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("wda launch failed: %v", err))
	}
	return nil
}

func (wd *WDADriver) AppLaunchUnattached(bundleId string) (err error) {
	log.Info().Str("bundleId", bundleId).Msg("WDADriver.AppLaunchUnattached")
	// [[FBRoute POST:@"/wda/apps/launchUnattached"].withoutSession respondWithTarget:self action:@selector(handleLaunchUnattachedApp:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.Session.POST(data, "/wda/apps/launchUnattached")
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("wda launchUnattached failed: %v", err))
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

func (wd *WDADriver) SetPasteboard(contentType types.PasteboardType, content string) (err error) {
	// [[FBRoute POST:@"/wda/setPasteboard"] respondWithTarget:self action:@selector(handleSetPasteboard:)]
	data := map[string]interface{}{
		"contentType": contentType,
		"content":     base64.StdEncoding.EncodeToString([]byte(content)),
	}
	_, err = wd.Session.POST(data, "/wda/setPasteboard")
	return
}

func (wd *WDADriver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	// [[FBRoute POST:@"/wda/getPasteboard"] respondWithTarget:self action:@selector(handleGetPasteboard:)]
	data := map[string]interface{}{"contentType": contentType}
	var rawResp DriverRawResponse
	if rawResp, err = wd.Session.POST(data, "/wda/getPasteboard"); err != nil {
		return nil, err
	}
	if raw, err = rawResp.ValueDecodeAsBase64(); err != nil {
		return nil, err
	}
	return
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
