package uixt

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
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

const (
	defaultBightInsightPort = 8000
	defaultDouyinServerPort = 32921
)

func NewStubIOSDriver(device *IOSDevice) (driver *StubIOSDriver, err error) {
	localStubPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}

	if err = device.forward(localStubPort, defaultBightInsightPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	localServerPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = device.forward(localServerPort, defaultDouyinServerPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	host := "localhost"
	timeout := 10 * time.Second
	driver = &StubIOSDriver{}
	driver.device = device
	driver.bightInsightPrefix = fmt.Sprintf("http://%s:%d", host, localStubPort)
	driver.serverPrefix = fmt.Sprintf("http://%s:%d", host, localServerPort)
	driver.timeout = timeout
	driver.client = &http.Client{
		Timeout: time.Second * 10, // 设置超时时间为 10 秒
	}

	return driver, nil
}

type StubIOSDriver struct {
	*WDADriver

	bightInsightPrefix string
	serverPrefix       string
	timeout            time.Duration
	device             *IOSDevice
}

func (s *StubIOSDriver) setUpWda() (err error) {
	if s.WDADriver == nil {
		capabilities := option.NewCapabilities()
		capabilities.WithDefaultAlertAction(option.AlertActionAccept)
		driver, err := s.device.NewHTTPDriver(capabilities)
		if err != nil {
			log.Error().Err(err).Msg("stub driver failed to init wda driver")
			return err
		}
		s.WDADriver = driver.(*WDADriver)
	}
	return nil
}

// NewSession starts a new session and returns the DriverSession.
func (s *StubIOSDriver) NewSession(capabilities option.Capabilities) (Session, error) {
	err := s.setUpWda()
	if err != nil {
		return Session{}, err
	}
	return s.WDADriver.NewSession(capabilities)
}

// DeleteSession Kills application associated with that session and removes session
//  1. alertsMonitor disable
//  2. testedApplicationBundleId terminate
func (s *StubIOSDriver) DeleteSession() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.DeleteSession()
}

func (s *StubIOSDriver) Status() (types.DeviceStatus, error) {
	err := s.setUpWda()
	if err != nil {
		return types.DeviceStatus{}, err
	}
	return s.WDADriver.Status()
}

func (s *StubIOSDriver) DeviceInfo() (types.DeviceInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return types.DeviceInfo{}, err
	}
	return s.WDADriver.DeviceInfo()
}

func (s *StubIOSDriver) Location() (types.Location, error) {
	err := s.setUpWda()
	if err != nil {
		return types.Location{}, err
	}
	return s.WDADriver.Location()
}

func (s *StubIOSDriver) BatteryInfo() (types.BatteryInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return types.BatteryInfo{}, err
	}
	return s.WDADriver.BatteryInfo()
}

// WindowSize Return the width and height in portrait mode.
// when getting the window size in wda/ui2/adb, if the device is in landscape mode,
// the width and height will be reversed.
func (s *StubIOSDriver) WindowSize() (types.Size, error) {
	err := s.setUpWda()
	if err != nil {
		return types.Size{}, err
	}
	return s.WDADriver.WindowSize()
}

func (s *StubIOSDriver) Screen() (ai.Screen, error) {
	err := s.setUpWda()
	if err != nil {
		return ai.Screen{}, err
	}
	return s.WDADriver.Screen()
}

func (s *StubIOSDriver) Scale() (float64, error) {
	err := s.setUpWda()
	if err != nil {
		return 0, err
	}
	return s.WDADriver.Scale()
}

// Homescreen Forces the device under test to switch to the home screen
func (s *StubIOSDriver) Homescreen() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Homescreen()
}

func (s *StubIOSDriver) Unlock() (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Unlock()
}

// AppLaunch Launch an application with given bundle identifier in scope of current session.
// !This method is only available since Xcode9 SDK
func (s *StubIOSDriver) AppLaunch(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.AppLaunch(packageName)
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (s *StubIOSDriver) AppTerminate(packageName string) (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.WDADriver.AppTerminate(packageName)
}

// GetForegroundApp returns current foreground app package name and activity name
func (s *StubIOSDriver) GetForegroundApp() (app types.AppInfo, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.AppInfo{}, err
	}
	return s.WDADriver.GetForegroundApp()
}

// StartCamera Starts a new camera for recording
func (s *StubIOSDriver) StartCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StartCamera()
}

// StopCamera Stops the camera for recording
func (s *StubIOSDriver) StopCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StopCamera()
}

func (s *StubIOSDriver) Orientation() (orientation types.Orientation, err error) {
	err = s.setUpWda()
	if err != nil {
		return types.OrientationPortrait, err
	}
	return s.WDADriver.Orientation()
}

// Tap Sends a tap event at the coordinate.
func (s *StubIOSDriver) Tap(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Tap(x, y, opts...)
}

// DoubleTap Sends a double tap event at the coordinate.
func (s *StubIOSDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.DoubleTap(x, y, opts...)
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (s *StubIOSDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TouchAndHold(x, y, opts...)
}

// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
func (s *StubIOSDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Drag(fromX, fromY, toX, toY, opts...)
}

// SetPasteboard Sets data to the general pasteboard
func (s *StubIOSDriver) SetPasteboard(contentType types.PasteboardType, content string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SetPasteboard(contentType, content)
}

// GetPasteboard Gets the data contained in the general pasteboard.
//
//	It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
func (s *StubIOSDriver) GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.GetPasteboard(contentType)
}

func (s *StubIOSDriver) SetIme(ime string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SetIme(ime)
}

// SendKeys Types a string into active element. There must be element with keyboard focus,
// otherwise an error is raised.
// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
func (s *StubIOSDriver) SendKeys(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SendKeys(text, opts...)
}

// Input works like SendKeys
func (s *StubIOSDriver) Input(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Input(text, opts...)
}

func (s *StubIOSDriver) Clear(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Clear(packageName)
}

// PressButton Presses the corresponding hardware button on the device
func (s *StubIOSDriver) PressButton(devBtn types.DeviceButton) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressButton(devBtn)
}

// PressBack Presses the back button
func (s *StubIOSDriver) PressBack(opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressBack(opts...)
}

func (s *StubIOSDriver) PressKeyCode(keyCode KeyCode) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressKeyCode(keyCode)
}

func (s *StubIOSDriver) Screenshot() (*bytes.Buffer, error) {
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

func (s *StubIOSDriver) TapByText(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TapByText(text, opts...)
}

func (s *StubIOSDriver) TapByTexts(actions ...TapTextAction) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TapByTexts(actions...)
}

// AccessibleSource Return application elements accessibility tree
func (s *StubIOSDriver) AccessibleSource() (string, error) {
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
func (s *StubIOSDriver) HealthCheck() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.HealthCheck()
}

func (s *StubIOSDriver) GetAppiumSettings() (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.GetAppiumSettings()
}

func (s *StubIOSDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.SetAppiumSettings(settings)
}

func (s *StubIOSDriver) IsHealthy() (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.WDADriver.IsHealthy()
}

// triggers the log capture and returns the log entries
func (s *StubIOSDriver) StartCaptureLog(identifier ...string) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StartCaptureLog(identifier...)
}

func (s *StubIOSDriver) StopCaptureLog() (result interface{}, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.StopCaptureLog()
}

func (s *StubIOSDriver) GetDriverResults() []*DriverRequests {
	err := s.setUpWda()
	if err != nil {
		return nil
	}
	return s.WDADriver.GetDriverResults()
}

func (s *StubIOSDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := s.Request(http.MethodGet, fmt.Sprintf("%s/source?format=json&onlyWeb=false", s.bightInsightPrefix), []byte{})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (s *StubIOSDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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
	resp, err := s.Request(http.MethodPost, fmt.Sprintf("%s/host/login/account/", s.serverPrefix), bsJSON)
	if err != nil {
		return info, err
	}
	res, err := resp.valueConvertToJsonObject()
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
	resp, err := s.Request(http.MethodGet, fmt.Sprintf("%s/host/loginout/", s.serverPrefix), []byte{})
	if err != nil {
		return err
	}
	res, err := resp.valueConvertToJsonObject()
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

func (s *StubIOSDriver) TearDown() error {
	s.client.CloseIdleConnections()
	return nil
}

func (s *StubIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	resp, err := s.Request(http.MethodGet, fmt.Sprintf("%s/host/app/info/", s.serverPrefix), []byte{})
	if err != nil {
		return info, err
	}
	res, err := resp.valueConvertToJsonObject()
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

func (s *StubIOSDriver) GetSession() *Session {
	return s.Session
}
