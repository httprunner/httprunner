package uixt

import (
	"bytes"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

func NewWDADriver(device *IOSDevice) (*WDADriver, error) {
	log.Info().Interface("device", device).Msg("init ios WDA driver")
	driver := &WDADriver{
		Device:  device,
		Session: &Session{},
	}
	driver.InitSession(nil)
	return driver, nil
}

type WDADriver struct {
	Device  *IOSDevice
	Session *Session

	mjpegHTTPConn net.Conn // via HTTP
	mjpegClient   *http.Client
	mjpegUrl      string
}

func (wd *WDADriver) resetSession() error {
	capabilities := option.NewCapabilities()
	capabilities.WithDefaultAlertAction(option.AlertActionAccept)

	data := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"alwaysMatch": capabilities,
		},
	}

	// Notice: use Driver.POST instead of httpPOST to avoid loop calling
	rawResp, err := wd.Session.POST(data, "/session")
	if err != nil {
		return err
	}
	sessionInfo, err := rawResp.ValueConvertToSessionInfo()
	if err != nil {
		return err
	}
	// update session ID
	wd.Session.ID = sessionInfo.ID
	return nil
}

func (wd *WDADriver) httpRequest(method string, rawURL string, rawBody []byte) (rawResp DriverRawResponse, err error) {
	retryInterval := 3 * time.Second
	for retryCount := 1; retryCount <= 3; retryCount++ {
		rawResp, err = wd.Session.Request(method, rawURL, rawBody)
		if err == nil {
			return
		}

		// check WDA server status
		status, err := wd.Status()
		if err != nil {
			log.Err(err).Msg("get WDA server status failed")
		} else if status.State != "success" {
			log.Warn().Interface("status", status).Msg("WDA server status is not success")
		} else {
			log.Info().Interface("status", status).Msg("get WDA server status")
		}

		retryInterval = retryInterval * 2
		time.Sleep(retryInterval)
		oldSessionID := wd.Session.ID
		if err2 := wd.resetSession(); err2 != nil {
			log.Err(err2).Msgf("failed to reset wda driver session, retry count: %v", retryCount)
			continue
		}
		log.Debug().Str("new session", wd.Session.ID).Str("old session", oldSessionID).
			Msgf("reset wda driver session successfully, retry count: %v", retryCount)
		if oldSessionID != "" {
			rawURL = strings.Replace(rawURL, oldSessionID, wd.Session.ID, 1)
		}
	}
	return
}

func (wd *WDADriver) httpGET(pathElem ...string) (rawResp DriverRawResponse, err error) {
	return wd.httpRequest(http.MethodGet, wd.Session.concatURL(nil, pathElem...), nil)
}

func (wd *WDADriver) httpPOST(data interface{}, pathElem ...string) (rawResp DriverRawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.httpRequest(http.MethodPost, wd.Session.concatURL(nil, pathElem...), bsJSON)
}

func (wd *WDADriver) httpDELETE(pathElem ...string) (rawResp DriverRawResponse, err error) {
	return wd.httpRequest(http.MethodDelete, wd.Session.concatURL(nil, pathElem...), nil)
}

func (wd *WDADriver) GetMjpegClient() *http.Client {
	return wd.mjpegClient
}

func (wd *WDADriver) InitSession(capabilities option.Capabilities) error {
	// [[FBRoute POST:@"/session"].withoutSession respondWithTarget:self action:@selector(handleCreateSession:)]
	data := make(map[string]interface{})
	if len(capabilities) == 0 {
		data["capabilities"] = make(map[string]interface{})
	} else {
		data["capabilities"] = map[string]interface{}{
			"alwaysMatch": capabilities,
		}
	}

	rawResp, err := wd.httpPOST(data, "/session")
	if err != nil {
		return err
	}
	sessionInfo, err := rawResp.ValueConvertToSessionInfo()
	if err != nil {
		return err
	}
	wd.Session = &sessionInfo
	return nil
}

func (wd *WDADriver) DeleteSession() (err error) {
	if wd.mjpegClient != nil {
		wd.mjpegClient.CloseIdleConnections()
	}
	if wd.mjpegHTTPConn != nil {
		wd.mjpegHTTPConn.Close()
	}

	// [[FBRoute DELETE:@""] respondWithTarget:self action:@selector(handleDeleteSession:)]
	_, err = wd.httpDELETE("/session", wd.Session.ID)
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
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/device/info"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/device/location"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/batteryInfo"); err != nil {
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
	if !wd.Session.windowSize.IsNil() {
		// use cached window size
		return wd.Session.windowSize, nil
	}

	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/window/size"); err != nil {
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

	wd.Session.windowSize = size // cache window size
	return wd.Session.windowSize, nil
}

func (wd *WDADriver) Scale() (float64, error) {
	if !builtin.IsZeroFloat64(wd.Session.scale) {
		return wd.Session.scale, nil
	}
	screen, err := wd.Screen()
	if err != nil {
		return 0, errors.Wrap(code.MobileUIDriverError,
			fmt.Sprintf("get screen info failed: %v", err))
	}
	return screen.Scale, nil
}

func (wd *WDADriver) Screen() (screen ai.Screen, err error) {
	// [[FBRoute GET:@"/wda/screen"] respondWithTarget:self action:@selector(handleGetScreen:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/screen"); err != nil {
		return ai.Screen{}, err
	}
	reply := new(struct{ Value struct{ ai.Screen } })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return ai.Screen{}, err
	}
	screen = reply.Value.Screen
	return
}

func (wd *WDADriver) ScreenShot(opts ...option.ActionOption) (raw *bytes.Buffer, err error) {
	// [[FBRoute GET:@"/screenshot"] respondWithTarget:self action:@selector(handleGetScreenshot:)]
	// [[FBRoute GET:@"/screenshot"].withoutSession respondWithTarget:self action:@selector(handleGetScreenshot:)]
	rawResp, err := wd.httpGET("/session", wd.Session.ID, "/screenshot")
	if err != nil {
		return nil, errors.Wrap(code.DeviceScreenShotError,
			fmt.Sprintf("WDA screenshot failed %v", err))
	}
	raw, err = rawResp.ValueDecodeAsBase64()
	if err != nil {
		return nil, errors.Wrap(code.DeviceScreenShotError,
			fmt.Sprintf("decode WDA screenshot data failed: %v", err))
	}

	actionOptions := option.NewActionOptions(opts...)
	if actionOptions.ScreenShotFileName != "" {
		// save screenshot to file
		path, err := saveScreenShot(raw, actionOptions.ScreenShotFileName)
		if err != nil {
			return nil, errors.Wrapf(code.DeviceScreenShotError,
				"save screenshot file failed %v", err)
		}
		log.Info().Str("path", path).Msg("screenshot saved")
	}

	return
}

func (wd *WDADriver) toScale(x float64) float64 {
	return x / wd.Session.scale
}

func (wd *WDADriver) ActiveAppInfo() (info types.AppInfo, err error) {
	// [[FBRoute GET:@"/wda/activeAppInfo"] respondWithTarget:self action:@selector(handleActiveAppInfo:)]
	// [[FBRoute GET:@"/wda/activeAppInfo"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/activeAppInfo"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/apps/list"); err != nil {
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
	// [[FBRoute POST:@"/wda/apps/state"] respondWithTarget:self action:@selector(handleSessionAppState:)]
	data := map[string]interface{}{"bundleId": bundleId}
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/apps/state"); err != nil {
		return 0, err
	}
	reply := new(struct{ Value types.AppState })
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return 0, err
	}
	runState = reply.Value
	_ = rawResp
	return
}

func (wd *WDADriver) IsLocked() (locked bool, err error) {
	// [[FBRoute GET:@"/wda/locked"] respondWithTarget:self action:@selector(handleIsLocked:)]
	// [[FBRoute GET:@"/wda/locked"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/locked"); err != nil {
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
	_, err = wd.httpPOST(nil, "/session", wd.Session.ID, "/wda/unlock")
	return
}

func (wd *WDADriver) Lock() (err error) {
	// [[FBRoute POST:@"/wda/lock"] respondWithTarget:self action:@selector(handleLock:)]
	// [[FBRoute POST:@"/wda/lock"].withoutSession
	_, err = wd.httpPOST(nil, "/session", wd.Session.ID, "/wda/lock")
	return
}

func (wd *WDADriver) Home() (err error) {
	// [[FBRoute POST:@"/wda/homescreen"].withoutSession respondWithTarget:self action:@selector(handleHomescreenCommand:)]
	_, err = wd.httpPOST(nil, "/wda/homescreen")
	return
}

func (wd *WDADriver) AlertText() (text string, err error) {
	// [[FBRoute GET:@"/alert/text"] respondWithTarget:self action:@selector(handleAlertGetTextCommand:)]
	// [[FBRoute GET:@"/alert/text"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/alert/text"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/alert/buttons"); err != nil {
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
	_, err = wd.httpPOST(data, "/alert/accept")
	return
}

func (wd *WDADriver) AlertDismiss(label ...string) (err error) {
	// [[FBRoute POST:@"/alert/dismiss"] respondWithTarget:self action:@selector(handleAlertDismissCommand:)]
	// [[FBRoute POST:@"/alert/dismiss"].withoutSession
	data := make(map[string]interface{})
	if len(label) != 0 && label[0] != "" {
		data["name"] = label[0]
	}
	_, err = wd.httpPOST(data, "/alert/dismiss")
	return
}

func (wd *WDADriver) AlertSendKeys(text string) (err error) {
	// [[FBRoute POST:@"/alert/text"] respondWithTarget:self action:@selector(handleAlertSetTextCommand:)]
	data := map[string]interface{}{"value": strings.Split(text, "")}
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/alert/text")
	return
}

func (wd *WDADriver) AppLaunch(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/launch"] respondWithTarget:self action:@selector(handleSessionAppLaunch:)]
	data := make(map[string]interface{})
	data["bundleId"] = bundleId
	data["environment"] = map[string]interface{}{
		"SHOW_EXPLORER": "NO",
	}
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/apps/launch")
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("wda launch failed: %v", err))
	}
	return nil
}

func (wd *WDADriver) AppLaunchUnattached(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/launchUnattached"].withoutSession respondWithTarget:self action:@selector(handleLaunchUnattachedApp:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.httpPOST(data, "/wda/apps/launchUnattached")
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("wda launchUnattached failed: %v", err))
	}
	return nil
}

func (wd *WDADriver) AppTerminate(bundleId string) (successful bool, err error) {
	// [[FBRoute POST:@"/wda/apps/terminate"] respondWithTarget:self action:@selector(handleSessionAppTerminate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/apps/terminate"); err != nil {
		return false, err
	}
	if successful, err = rawResp.ValueConvertToBool(); err != nil {
		return false, err
	}
	return
}

func (wd *WDADriver) AppActivate(bundleId string) (err error) {
	// [[FBRoute POST:@"/wda/apps/activate"] respondWithTarget:self action:@selector(handleSessionAppActivate:)]
	data := map[string]interface{}{"bundleId": bundleId}
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/apps/activate")
	return
}

func (wd *WDADriver) AppDeactivate(second float64) (err error) {
	// [[FBRoute POST:@"/wda/deactivateApp"] respondWithTarget:self action:@selector(handleDeactivateAppCommand:)]
	if second < 3 {
		second = 3.0
	}
	data := map[string]interface{}{"duration": second}
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/deactivateApp")
	return
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
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]
	absX, absY, err := convertToAbsolutePoint(wd, x, y)
	if err != nil {
		return err
	}
	return wd.TapAbsXY(absX, absY, opts...)
}

func (wd *WDADriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	// [[FBRoute POST:@"/wda/tap/:uuid"] respondWithTarget:self action:@selector(handleTap:)]
	data := map[string]interface{}{
		"x": wd.toScale(x),
		"y": wd.toScale(y),
	}
	_, err := wd.httpPOST(data, "/session", wd.Session.ID, "/wda/tap/0")
	return err
}

func (wd *WDADriver) DoubleTapXY(x, y float64, opts ...option.ActionOption) (err error) {
	// [[FBRoute POST:@"/wda/doubleTap"] respondWithTarget:self action:@selector(handleDoubleTapCoordinate:)]
	actionOptions := option.NewActionOptions(opts...)
	x = wd.toScale(x)
	y = wd.toScale(y)
	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}
	x += actionOptions.GetRandomOffset()
	y += actionOptions.GetRandomOffset()

	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/doubleTap")
	return
}

func (wd *WDADriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	actionOptions := option.NewActionOptions(opts...)
	if actionOptions.Duration == 0 {
		opts = append(opts, option.WithDuration(1))
	}
	return wd.TapXY(x, y, opts...)
}

func (wd *WDADriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	// [[FBRoute POST:@"/wda/dragfromtoforduration"] respondWithTarget:self action:@selector(handleDragCoordinate:)]
	var err error
	actionOptions := option.NewActionOptions(opts...)
	if !actionOptions.AbsCoordinate {
		fromX, fromY, toX, toY, err = convertToAbsoluteCoordinates(wd, fromX, fromY, toX, toY)
		if err != nil {
			return err
		}
	}
	fromX = wd.toScale(fromX)
	fromY = wd.toScale(fromY)
	toX = wd.toScale(toX)
	toY = wd.toScale(toY)
	data := map[string]interface{}{
		"fromX": math.Round(fromX*10) / 10,
		"fromY": math.Round(fromY*10) / 10,
		"toX":   math.Round(toX*10) / 10,
		"toY":   math.Round(toY*10) / 10,
	}
	if actionOptions.PressDuration > 0 {
		data["pressDuration"] = actionOptions.PressDuration
	}

	// update data options in post data for extra WDA configurations
	actionOptions.UpdateData(data)
	// wda 43 version
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/dragfromtoforduration")
	// _, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/drag")
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
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/setPasteboard")
	return
}

func (wd *WDADriver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	// [[FBRoute POST:@"/wda/getPasteboard"] respondWithTarget:self action:@selector(handleGetPasteboard:)]
	data := map[string]interface{}{"contentType": contentType}
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/getPasteboard"); err != nil {
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
	// [[FBRoute POST:@"/wda/keys"] respondWithTarget:self action:@selector(handleKeys:)]
	actionOptions := option.NewActionOptions(opts...)
	data := map[string]interface{}{"value": strings.Split(text, "")}

	// new data options in post data for extra WDA configurations
	actionOptions.UpdateData(data)

	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/keys")
	return
}

func (wd *WDADriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	if count == 0 {
		return nil
	}
	actionOptions := option.NewActionOptions(opts...)
	data := map[string]interface{}{"count": count}

	// new data options in post data for extra WDA configurations
	actionOptions.UpdateData(data)

	_, err = wd.httpPOST(data, "/gtf/interaction/input/backspace")
	return
}

func (wd *WDADriver) AppClear(packageName string) error {
	return types.ErrDriverNotImplemented
}

// Back simulates a short press on the BACK button.
func (wd *WDADriver) Back() (err error) {
	windowSize, err := wd.WindowSize()
	if err != nil {
		return
	}
	fromX := wd.toScale(float64(windowSize.Width) * 0)
	fromY := wd.toScale(float64(windowSize.Height) * 0.5)
	toX := wd.toScale(float64(windowSize.Width) * 0.6)
	toY := wd.toScale(float64(windowSize.Height) * 0.5)

	data := map[string]interface{}{
		"fromX": fromX,
		"fromY": fromY,
		"toX":   toX,
		"toY":   toY,
	}

	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/dragfromtoforduration")
	return
}

func (wd *WDADriver) PressButton(devBtn types.DeviceButton) (err error) {
	// [[FBRoute POST:@"/wda/pressButton"] respondWithTarget:self action:@selector(handlePressButtonCommand:)]
	data := map[string]interface{}{"name": devBtn}
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/wda/pressButton")
	return
}

func (wd *WDADriver) Orientation() (orientation types.Orientation, err error) {
	// [[FBRoute GET:@"/orientation"] respondWithTarget:self action:@selector(handleGetOrientation:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/orientation"); err != nil {
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
	_, err = wd.httpPOST(data, "/session", wd.Session.ID, "/orientation")
	return
}

func (wd *WDADriver) Rotation() (rotation types.Rotation, err error) {
	// [[FBRoute GET:@"/rotation"] respondWithTarget:self action:@selector(handleGetRotation:)]
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/rotation"); err != nil {
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
	_, err = wd.httpPOST(rotation, "/session", wd.Session.ID, "/rotation")
	return
}

func (wd *WDADriver) Source(srcOpt ...option.SourceOption) (source string, err error) {
	// [[FBRoute GET:@"/source"] respondWithTarget:self action:@selector(handleGetSourceCommand:)]
	// [[FBRoute GET:@"/source"].withoutSession
	tmp, _ := url.Parse(wd.Session.concatURL(nil, "/session", wd.Session.ID))
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

	var rawResp DriverRawResponse
	if rawResp, err = wd.httpRequest(http.MethodGet, wd.Session.concatURL(tmp, "/source"), nil); err != nil {
		return "", nil
	}
	if toJsonRaw {
		var jr builtinJSON.RawMessage
		if jr, err = rawResp.ValueConvertToJsonRawMessage(); err != nil {
			return "", err
		}
		return string(jr), nil
	}
	if source, err = rawResp.ValueConvertToString(); err != nil {
		return "", err
	}
	return
}

func (wd *WDADriver) TapByText(text string, opts ...option.ActionOption) error {
	return types.ErrDriverNotImplemented
}

func (wd *WDADriver) TapByTexts(actions ...TapTextAction) error {
	return types.ErrDriverNotImplemented
}

func (wd *WDADriver) AccessibleSource() (source string, err error) {
	// [[FBRoute GET:@"/wda/accessibleSource"] respondWithTarget:self action:@selector(handleGetAccessibleSourceCommand:)]
	// [[FBRoute GET:@"/wda/accessibleSource"].withoutSession
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/wda/accessibleSource"); err != nil {
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
	_, err = wd.httpGET("/wda/healthcheck")
	return
}

func (wd *WDADriver) IsHealthy() (healthy bool, err error) {
	var rawResp DriverRawResponse
	if rawResp, err = wd.httpGET("/health"); err != nil {
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
	if rawResp, err = wd.httpGET("/session", wd.Session.ID, "/appium/settings"); err != nil {
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
	if rawResp, err = wd.httpPOST(data, "/session", wd.Session.ID, "/appium/settings"); err != nil {
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
	_, err = wd.httpGET("/wda/shutdown")
	return
}

func (wd *WDADriver) triggerWDALog(data map[string]interface{}) (rawResp []byte, err error) {
	// [[FBRoute POST:@"/gtf/automation/log"].withoutSession respondWithTarget:self action:@selector(handleAutomationLog:)]
	return wd.httpPOST(data, "/gtf/automation/log")
}

func (wd *WDADriver) ScreenRecord(duration time.Duration) (videoPath string, err error) {
	timestamp := time.Now().Format("20060102_150405") + fmt.Sprintf("_%03d", time.Now().UnixNano()/1e6%1000)
	fileName := filepath.Join(config.ScreenShotsPath, fmt.Sprintf("%s.mp4", timestamp))

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
		"-i", "http://"+wd.mjpegUrl,
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

type wdaResponse struct {
	Value     interface{} `json:"value"`
	SessionID string      `json:"sessionId"`
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

func (wd *WDADriver) GetSession() *Session {
	return wd.Session
}

func (wd *WDADriver) Setup() error {
	return nil
}

func (wd *WDADriver) TearDown() error {
	wd.mjpegClient.CloseIdleConnections()
	wd.Session.client.CloseIdleConnections()
	return nil
}
