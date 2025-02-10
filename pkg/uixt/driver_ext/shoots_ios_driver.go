package driver_ext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

const (
	defaultBightInsightPort = 8000
	defaultDouyinServerPort = 32921
)

func NewShootsIOSDriver(device *uixt.IOSDevice) (driver *ShootsIOSDriver, err error) {
	localShootsPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}

	if err = device.Forward(localShootsPort, defaultBightInsightPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	localServerPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = device.Forward(localServerPort, defaultDouyinServerPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	host := "localhost"
	timeout := 10 * time.Second
	driver = &ShootsIOSDriver{}
	driver.device = device
	driver.bightInsightPrefix = fmt.Sprintf("http://%s:%d", host, localShootsPort)
	driver.serverPrefix = fmt.Sprintf("http://%s:%d", host, localServerPort)
	driver.timeout = timeout

	return driver, nil
}

type ShootsIOSDriver struct {
	*uixt.WDADriver

	bightInsightPrefix string
	serverPrefix       string
	timeout            time.Duration
	device             *uixt.IOSDevice
}

func (s *ShootsIOSDriver) setUpWda() (err error) {
	if s.WDADriver == nil {
		capabilities := option.NewCapabilities()
		capabilities.WithDefaultAlertAction(option.AlertActionAccept)
		driver, err := s.device.NewHTTPDriver(capabilities)
		if err != nil {
			log.Error().Err(err).Msg("failed to init WDA driver for shoots IOS")
			return err
		}
		s.WDADriver = driver.(*uixt.WDADriver)
	}
	return nil
}

// InitSession starts a new session and returns the DriverSession.
func (s *ShootsIOSDriver) InitSession(capabilities option.Capabilities) error {
	return s.WDADriver.InitSession(capabilities)
}

// DeleteSession Kills application associated with that session and removes session
//  1. alertsMonitor disable
//  2. testedApplicationBundleId terminate
func (s *ShootsIOSDriver) DeleteSession() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.DeleteSession()
}

func (s *ShootsIOSDriver) Status() (types.DeviceStatus, error) {
	err := s.setUpWda()
	if err != nil {
		return types.DeviceStatus{}, err
	}
	return s.WDADriver.Status()
}

func (s *ShootsIOSDriver) DeviceInfo() (types.DeviceInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return types.DeviceInfo{}, err
	}
	return s.WDADriver.DeviceInfo()
}

func (s *ShootsIOSDriver) BatteryInfo() (types.BatteryInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return types.BatteryInfo{}, err
	}
	return s.WDADriver.BatteryInfo()
}

// WindowSize Return the width and height in portrait mode.
// when getting the window size in wda/ui2/adb, if the device is in landscape mode,
// the width and height will be reversed.
func (s *ShootsIOSDriver) WindowSize() (types.Size, error) {
	err := s.setUpWda()
	if err != nil {
		return types.Size{}, err
	}
	return s.WDADriver.WindowSize()
}

func (s *ShootsIOSDriver) Screen() (ai.Screen, error) {
	err := s.setUpWda()
	if err != nil {
		return ai.Screen{}, err
	}
	return s.WDADriver.Screen()
}

func (s *ShootsIOSDriver) Scale() (float64, error) {
	err := s.setUpWda()
	if err != nil {
		return 0, err
	}
	return s.WDADriver.Scale()
}

// Homescreen Forces the device under test to switch to the home screen
func (s *ShootsIOSDriver) Homescreen() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Homescreen()
}

func (s *ShootsIOSDriver) Unlock() (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Unlock()
}

// AppLaunch Launch an application with given bundle identifier in scope of current session.
// !This method is only available since Xcode9 SDK
func (s *ShootsIOSDriver) AppLaunch(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.AppLaunch(packageName)
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (s *ShootsIOSDriver) AppTerminate(packageName string) (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.WDADriver.AppTerminate(packageName)
}

// GetForegroundApp returns current foreground app package name and activity name
func (s *ShootsIOSDriver) GetForegroundApp() (app types.AppInfo, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.AppInfo{}, err
	}
	return s.WDADriver.GetForegroundApp()
}

// StartCamera Starts a new camera for recording
func (s *ShootsIOSDriver) StartCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StartCamera()
}

// StopCamera Stops the camera for recording
func (s *ShootsIOSDriver) StopCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StopCamera()
}

func (s *ShootsIOSDriver) Orientation() (orientation types.Orientation, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.OrientationPortrait, err
	}
	return s.WDADriver.Orientation()
}

// Tap Sends a tap event at the coordinate.
func (s *ShootsIOSDriver) Tap(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Tap(x, y, opts...)
}

// DoubleTap Sends a double tap event at the coordinate.
func (s *ShootsIOSDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.DoubleTap(x, y, opts...)
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (s *ShootsIOSDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TouchAndHold(x, y, opts...)
}

// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
func (s *ShootsIOSDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Drag(fromX, fromY, toX, toY, opts...)
}

// SetPasteboard Sets data to the general pasteboard
func (s *ShootsIOSDriver) SetPasteboard(contentType types.PasteboardType, content string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SetPasteboard(contentType, content)
}

// GetPasteboard Gets the data contained in the general pasteboard.
//
//	It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
func (s *ShootsIOSDriver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.GetPasteboard(contentType)
}

func (s *ShootsIOSDriver) SetIme(ime string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SetIme(ime)
}

// SendKeys Types a string into active element. There must be element with keyboard focus,
// otherwise an error is raised.
// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
func (s *ShootsIOSDriver) SendKeys(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SendKeys(text, opts...)
}

// Input works like SendKeys
func (s *ShootsIOSDriver) Input(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Input(text, opts...)
}

func (s *ShootsIOSDriver) Clear(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Clear(packageName)
}

// PressButton Presses the corresponding hardware button on the device
func (s *ShootsIOSDriver) PressButton(devBtn types.DeviceButton) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressButton(devBtn)
}

// PressBack Presses the back button
func (s *ShootsIOSDriver) PressBack(opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressBack(opts...)
}

func (s *ShootsIOSDriver) PressKeyCode(keyCode uixt.KeyCode) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressKeyCode(keyCode)
}

func (s *ShootsIOSDriver) Screenshot() (*bytes.Buffer, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.Screenshot()
	//screenshotService, err := instruments.NewScreenshotService(s.device.d)
	//if err != nil {
	//	log.Error().Err(err).Msg("Starting screenshot service failed")
	//	return nil, err
	//}
	//defer screenshotService.Close()
	//
	//imageBytes, err := screenshotService.TakeScreenshot()
	//if err != nil {
	//	log.Error().Err(err).Msg("failed to task screenshot")
	//	return nil, err
	//}
	//return bytes.NewBuffer(imageBytes), nil
}

func (s *ShootsIOSDriver) TapByText(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TapByText(text, opts...)
}

func (s *ShootsIOSDriver) TapByTexts(actions ...uixt.TapTextAction) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TapByTexts(actions...)
}

// AccessibleSource Return application elements accessibility tree
func (s *ShootsIOSDriver) AccessibleSource() (string, error) {
	err := s.setUpWda()
	if err != nil {
		return "", err
	}
	return s.WDADriver.AccessibleSource()
}

// HealthCheck Health check might modify simulator state so it should only be called in-between testing sessions
//
//	Checks health of XCTest by:
//	1) Querying application for some elements,
//	2) Triggering some device events.
func (s *ShootsIOSDriver) HealthCheck() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.HealthCheck()
}

func (s *ShootsIOSDriver) GetAppiumSettings() (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.GetAppiumSettings()
}

func (s *ShootsIOSDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.SetAppiumSettings(settings)
}

func (s *ShootsIOSDriver) IsHealthy() (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.WDADriver.IsHealthy()
}

// triggers the log capture and returns the log entries
func (s *ShootsIOSDriver) StartCaptureLog(identifier ...string) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StartCaptureLog(identifier...)
}

func (s *ShootsIOSDriver) StopCaptureLog() (result interface{}, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.StopCaptureLog()
}

func (s *ShootsIOSDriver) GetDriverResults() []*uixt.DriverRequests {
	err := s.setUpWda()
	if err != nil {
		return nil
	}
	return s.WDADriver.GetDriverResults()
}

func (s *ShootsIOSDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := s.Session.Request(http.MethodGet, fmt.Sprintf("%s/source?format=json&onlyWeb=false", s.bightInsightPrefix), []byte{})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (s *ShootsIOSDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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
	bsJSON, err := json.Marshal(params)
	if err != nil {
		return info, err
	}
	resp, err := s.Session.Request(http.MethodPost, fmt.Sprintf("%s/host/login/account/", s.serverPrefix), bsJSON)
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

func (s *ShootsIOSDriver) LogoutNoneUI(packageName string) error {
	resp, err := s.Session.Request(http.MethodGet, fmt.Sprintf("%s/host/loginout/", s.serverPrefix), []byte{})
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

func (s *ShootsIOSDriver) TearDown() error {
	s.WDADriver.TearDown()
	return nil
}

func (s *ShootsIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	resp, err := s.Session.Request(http.MethodGet, fmt.Sprintf("%s/host/app/info/", s.serverPrefix), []byte{})
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

func (s *ShootsIOSDriver) GetSession() *uixt.Session {
	return s.Session
}
