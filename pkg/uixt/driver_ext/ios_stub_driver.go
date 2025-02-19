package driver_ext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
	"net/url"
	"os"
	"time"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type StubIOSDriver struct {
	Device              *uixt.IOSDevice
	Session             *uixt.DriverSession
	wdaDriver           *uixt.WDADriver
	timeout             time.Duration
	douyinUrlPrefix     string
	douyinLiteUrlPrefix string
}

const (
	IOSDouyinPort           = 32921
	IOSDouyinLitePort       = 33461
	defaultBightInsightPort = 8000
)

func NewStubIOSDriver(dev *uixt.IOSDevice) (*StubIOSDriver, error) {
	// lazy setup WDA
	dev.Options.LazySetup = true

	wdaDriver, err := uixt.NewWDADriver(dev)
	if err != nil {
		return nil, err
	}
	driver := &StubIOSDriver{
		Device:  dev,
		timeout: 10 * time.Second,
		Session: uixt.NewDriverSession(),
	}

	// setup driver
	if err := driver.Setup(); err != nil {
		return nil, err
	}

	return driver, nil
}

func (s *StubIOSDriver) Setup() error {
	localPort, err := s.getLocalPort()
	if err != nil {
		return err
	}
	s.Session.SetBaseURL(fmt.Sprintf("http://127.0.0.1:%d", localPort))

	localDouyinPort, err := builtin.GetFreePort()
	if err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = s.Device.Forward(localDouyinPort, IOSDouyinPort); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	s.douyinUrlPrefix = fmt.Sprintf("http://127.0.0.1:%d", localDouyinPort)

	localDouyinLitePort, err := builtin.GetFreePort()
	if err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = s.Device.Forward(localDouyinLitePort, IOSDouyinLitePort); err != nil {
		return errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	s.douyinLiteUrlPrefix = fmt.Sprintf("http://127.0.0.1:%d", localDouyinLitePort)
	return nil
}

func (s *StubIOSDriver) setUpWda() (err error) {
	if s.wdaDriver == nil {
		driver, err := uixt.NewWDADriver(s.Device)
		if err != nil {
			log.Error().Err(err).Msg("stub driver failed to init wda driver")
			return err
		}
		s.wdaDriver = driver
	}
	return nil
}

func (s *StubIOSDriver) getLocalPort() (int, error) {
	localStubPort, err := builtin.GetFreePort()
	if err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = s.Device.Forward(localStubPort, defaultBightInsightPort); err != nil {
		return 0, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	return localStubPort, nil
}

func (s *StubIOSDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := s.Session.GET("/source?format=json&onlyWeb=false")
	if err != nil {
		log.Error().Err(err).Msg("get source err")
		return "", nil
	}
	return string(resp), nil
}

func (s *StubIOSDriver) OpenUrl(urlStr string, opts ...option.ActionOption) (err error) {
	targetUrl := fmt.Sprintf("/openURL?url=%s", url.QueryEscape(urlStr))
	_, err = s.Session.GET(targetUrl)
	if err != nil {
		log.Error().Err(err).Msg("get source err")
		return nil
	}
	return nil
}

func (s *StubIOSDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	appInfo, err := s.ForegroundInfo()
	if err != nil {
		return info, err
	}
	if appInfo.BundleId == "com.ss.iphone.ugc.AwemeInhouse" || appInfo.BundleId == "com.ss.iphone.ugc.awemeinhouse.lite" {
		return s.LoginDouyin(appInfo.BundleId, phoneNumber, captcha, password)
	} else if appInfo.BundleId == "com.ss.iphone.InHouse.article.Video" {
		return s.LoginXigua(appInfo.BundleId, phoneNumber, captcha, password)
	} else {
		return info, fmt.Errorf("not support app")
	}
}

func (s *StubIOSDriver) LoginXigua(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	loginSchema := ""
	if captcha != "" {
		loginSchema = fmt.Sprintf("snssdk32://local_channel_autologin?login_type=1&account=%s&smscode=%s", phoneNumber, captcha)
	} else if password != "" {
		loginSchema = fmt.Sprintf("snssdk32://local_channel_autologin?login_type=2&account=%s&password=%s", phoneNumber, password)
	} else {
		return info, fmt.Errorf("password and capcha is empty")
	}
	info.IsLogin = true
	return info, s.OpenUrl(loginSchema)
}

func (s *StubIOSDriver) LoginDouyin(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	params := map[string]interface{}{
		"phone": phoneNumber,
	}
	if captcha != "" {
		params["captcha"] = captcha
	} else if password != "" {
		params["password"] = password
	} else {
		return info, fmt.Errorf("password and capcha is empty")
	}
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/login/account/"
	resp, err := s.Session.POST(params, fullUrl)
	if err != nil {
		return info, err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return info, err
	}
	log.Info().Msgf("%v", res)
	// {'isSuccess': True, 'data': '登录成功', 'code': 0}
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to logout %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return info, err
	}
	time.Sleep(20 * time.Second)
	info, err = s.getLoginAppInfo(packageName)
	if err != nil || !info.IsLogin {
		return info, fmt.Errorf("falied to login %v", info)
	}
	return info, nil
}

func (s *StubIOSDriver) LogoutNoneUI(packageName string) error {
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return err
	}
	fullUrl := urlPrefix + "/host/loginout/"
	resp, err := s.Session.GET(fullUrl)
	if err != nil {
		return err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to logout %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return err
	}
	time.Sleep(10 * time.Second)
	return nil
}

func (s *StubIOSDriver) EnableDevtool(packageName string, enable bool) (err error) {
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return err
	}
	fullUrl := urlPrefix + "/host/devtool/enable"

	params := map[string]interface{}{
		"enable": enable,
	}
	resp, err := s.Session.POST(params, fullUrl)
	if err != nil {
		return err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to enable devtool %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return err
	}
	return nil
}

func (s *StubIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	urlPrefix, err := s.getUrlPrefix(packageName)
	if err != nil {
		return info, err
	}
	fullUrl := urlPrefix + "/host/app/info/"

	resp, err := s.Session.GET(fullUrl)
	if err != nil {
		return info, err
	}
	res, err := resp.ValueConvertToJsonObject()
	if err != nil {
		return info, err
	}
	log.Info().Msgf("%v", res)
	if res["isSuccess"] != true {
		err = fmt.Errorf("falied to get is login %s", res["data"])
		log.Err(err).Msgf("%v", res)
		return info, err
	}
	err = json.Unmarshal([]byte(res["data"].(string)), &info)
	if err != nil {
		return info, err
	}
	return info, nil
}

func (s *StubIOSDriver) getUrlPrefix(packageName string) (urlPrefix string, err error) {
	if packageName == "com.ss.iphone.ugc.AwemeInhouse" {
		urlPrefix = s.douyinUrlPrefix
	} else if packageName == "com.ss.iphone.ugc.awemeinhouse.lite" {
		urlPrefix = s.douyinLiteUrlPrefix
	} else {
		return "", fmt.Errorf("not support app %s", packageName)
	}
	return urlPrefix, nil
}

func (s *StubIOSDriver) TearDown() error {
	if s.wdaDriver != nil {
		_ = s.wdaDriver.TearDown()
	}
	return nil
}

// NewSession starts a new session and returns the SessionInfo.
func (s *StubIOSDriver) InitSession(capabilities option.Capabilities) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.InitSession(capabilities)
}

// DeleteSession Kills application associated with that session and removes session
//  1. alertsMonitor disable
//  2. testedApplicationBundleId terminate
func (s *StubIOSDriver) DeleteSession() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.DeleteSession()
}

func (s *StubIOSDriver) Status() (types.DeviceStatus, error) {
	err := s.setUpWda()
	if err != nil {
		return types.DeviceStatus{}, err
	}
	return s.wdaDriver.Status()
}

func (s *StubIOSDriver) GetDevice() uixt.IDevice {
	return s.Device
}

func (s *StubIOSDriver) DeviceInfo() (types.DeviceInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return types.DeviceInfo{}, err
	}
	return s.wdaDriver.DeviceInfo()
}

func (s *StubIOSDriver) Location() (types.Location, error) {
	err := s.setUpWda()
	if err != nil {
		return types.Location{}, err
	}
	return s.wdaDriver.Location()
}

func (s *StubIOSDriver) BatteryInfo() (types.BatteryInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return types.BatteryInfo{}, err
	}
	return s.wdaDriver.BatteryInfo()
}

// WindowSize Return the width and height in portrait mode.
// when getting the window size in wda/ui2/adb, if the device is in landscape mode,
// the width and height will be reversed.
func (s *StubIOSDriver) WindowSize() (types.Size, error) {
	err := s.setUpWda()
	if err != nil {
		return types.Size{}, err
	}
	return s.wdaDriver.WindowSize()
}

func (s *StubIOSDriver) Screen() (uixt.Screen, error) {
	err := s.setUpWda()
	if err != nil {
		return uixt.Screen{}, err
	}
	return s.wdaDriver.Screen()
}

func (s *StubIOSDriver) Scale() (float64, error) {
	err := s.setUpWda()
	if err != nil {
		return 0, err
	}
	return s.wdaDriver.Scale()
}

// Homescreen Forces the device under test to switch to the home screen
func (s *StubIOSDriver) Home() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Home()
}

func (s *StubIOSDriver) Unlock() (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Unlock()
}

// AppLaunch Launch an application with given bundle identifier in scope of current session.
// !This method is only available since Xcode9 SDK
func (s *StubIOSDriver) AppLaunch(packageName string) error {
	_ = s.EnableDevtool(packageName, true)
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.AppLaunch(packageName)
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (s *StubIOSDriver) AppTerminate(packageName string) (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.wdaDriver.AppTerminate(packageName)
}

// GetForegroundApp returns current foreground app package name and activity name
func (s *StubIOSDriver) ForegroundInfo() (appInfo types.AppInfo, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.AppInfo{}, err
	}
	return s.wdaDriver.ForegroundInfo()
}

func (s *StubIOSDriver) Orientation() (orientation types.Orientation, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.OrientationPortrait, err
	}
	return s.wdaDriver.Orientation()
}

func (s *StubIOSDriver) Rotation() (rotation types.Rotation, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.Rotation{}, err
	}
	return s.wdaDriver.Rotation()
}

func (s *StubIOSDriver) SetRotation(rotation types.Rotation) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.SetRotation(rotation)
}

// Tap Sends a tap event at the coordinate.
func (s *StubIOSDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.TapXY(x, y, opts...)
}

func (s *StubIOSDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.TapAbsXY(x, y, opts...)
}

// DoubleTap Sends a double tap event at the coordinate.
func (s *StubIOSDriver) DoubleTapXY(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.DoubleTapXY(x, y, opts...)
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (s *StubIOSDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.TouchAndHold(x, y, opts...)
}

// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
func (s *StubIOSDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Drag(fromX, fromY, toX, toY, opts...)
}

// Swipe works like Drag, but `pressForDuration` value is 0
func (s *StubIOSDriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Swipe(fromX, fromY, toX, toY, opts...)
}

// SetPasteboard Sets data to the general pasteboard
func (s *StubIOSDriver) SetPasteboard(contentType types.PasteboardType, content string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.SetPasteboard(contentType, content)
}

// GetPasteboard Gets the data contained in the general pasteboard.
//
//	It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
func (s *StubIOSDriver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.GetPasteboard(contentType)
}

func (s *StubIOSDriver) SetIme(ime string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.SetIme(ime)
}

// Input works like SendKeys
func (s *StubIOSDriver) Input(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Input(text, opts...)
}

func (s *StubIOSDriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Backspace(count, opts...)
}

func (s *StubIOSDriver) AppClear(packageName string) error {
	return types.ErrDriverNotImplemented
}

func (s *StubIOSDriver) Back() (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Back()
}

// PressButton Presses the corresponding hardware button on the device
func (s *StubIOSDriver) PressButton(devBtn types.DeviceButton) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.PressButton(devBtn)
}

func (s *StubIOSDriver) ScreenShot(opts ...option.ActionOption) (*bytes.Buffer, error) {
	if os.Getenv("WINGS_LOCAL") == "true" {
		return s.Device.ScreenShot()
	}
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.ScreenShot()
}

// AccessibleSource Return application elements accessibility tree
func (s *StubIOSDriver) AccessibleSource() (string, error) {
	err := s.setUpWda()
	if err != nil {
		return "", err
	}
	return s.wdaDriver.AccessibleSource()
}

// HealthCheck Health check might modify simulator state so it should only be called in-between testing sessions
//
//	Checks health of XCTest by:
//	1) Querying application for some elements,
//	2) Triggering some device events.
func (s *StubIOSDriver) HealthCheck() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.HealthCheck()
}

func (s *StubIOSDriver) GetAppiumSettings() (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.GetAppiumSettings()
}

func (s *StubIOSDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.SetAppiumSettings(settings)
}

func (s *StubIOSDriver) IsHealthy() (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.wdaDriver.IsHealthy()
}

func (s *StubIOSDriver) ScreenRecord(duration time.Duration) (videoPath string, err error) {
	err = s.setUpWda()
	if err != nil {
		return "", err
	}
	return s.wdaDriver.ScreenRecord(duration)
}

func (s *StubIOSDriver) PushImage(localPath string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.PushImage(localPath)
}

func (s *StubIOSDriver) ClearImages() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.ClearImages()
}

// triggers the log capture and returns the log entries
func (s *StubIOSDriver) StartCaptureLog(identifier ...string) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.StartCaptureLog(identifier...)
}

func (s *StubIOSDriver) StopCaptureLog() (result interface{}, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.StopCaptureLog()
}

func (s *StubIOSDriver) GetSession() *uixt.DriverSession {
	err := s.setUpWda()
	if err != nil {
		return nil
	}
	return s.wdaDriver.Session
}
