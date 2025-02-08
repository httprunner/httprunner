package uixt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

type stubIOSDriver struct {
	*WDADriver

	bightInsightPrefix string
	serverPrefix       string
	timeout            time.Duration
	device             *IOSDevice
}

func newStubIOSDriver(bightInsightAddr, serverAddr string, dev *IOSDevice, readTimeout ...time.Duration) (*stubIOSDriver, error) {
	timeout := 10 * time.Second
	if len(readTimeout) > 0 {
		timeout = readTimeout[0]
	}
	driver := new(stubIOSDriver)
	driver.device = dev
	driver.bightInsightPrefix = bightInsightAddr
	driver.serverPrefix = serverAddr
	driver.timeout = timeout
	driver.client = &http.Client{
		Timeout: time.Second * 10, // 设置超时时间为 10 秒
	}
	return driver, nil
}

func (s *stubIOSDriver) setUpWda() (err error) {
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
func (s *stubIOSDriver) NewSession(capabilities option.Capabilities) (Session, error) {
	err := s.setUpWda()
	if err != nil {
		return Session{}, err
	}
	return s.WDADriver.NewSession(capabilities)
}

// DeleteSession Kills application associated with that session and removes session
//  1. alertsMonitor disable
//  2. testedApplicationBundleId terminate
func (s *stubIOSDriver) DeleteSession() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.DeleteSession()
}

func (s *stubIOSDriver) Status() (DeviceStatus, error) {
	err := s.setUpWda()
	if err != nil {
		return DeviceStatus{}, err
	}
	return s.WDADriver.Status()
}

func (s *stubIOSDriver) DeviceInfo() (DeviceInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return DeviceInfo{}, err
	}
	return s.WDADriver.DeviceInfo()
}

func (s *stubIOSDriver) Location() (Location, error) {
	err := s.setUpWda()
	if err != nil {
		return Location{}, err
	}
	return s.WDADriver.Location()
}

func (s *stubIOSDriver) BatteryInfo() (BatteryInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return BatteryInfo{}, err
	}
	return s.WDADriver.BatteryInfo()
}

// WindowSize Return the width and height in portrait mode.
// when getting the window size in wda/ui2/adb, if the device is in landscape mode,
// the width and height will be reversed.
func (s *stubIOSDriver) WindowSize() (ai.Size, error) {
	err := s.setUpWda()
	if err != nil {
		return ai.Size{}, err
	}
	return s.WDADriver.WindowSize()
}

func (s *stubIOSDriver) Screen() (ai.Screen, error) {
	err := s.setUpWda()
	if err != nil {
		return ai.Screen{}, err
	}
	return s.WDADriver.Screen()
}

func (s *stubIOSDriver) Scale() (float64, error) {
	err := s.setUpWda()
	if err != nil {
		return 0, err
	}
	return s.WDADriver.Scale()
}

// Homescreen Forces the device under test to switch to the home screen
func (s *stubIOSDriver) Homescreen() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Homescreen()
}

func (s *stubIOSDriver) Unlock() (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Unlock()
}

// AppLaunch Launch an application with given bundle identifier in scope of current session.
// !This method is only available since Xcode9 SDK
func (s *stubIOSDriver) AppLaunch(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.AppLaunch(packageName)
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (s *stubIOSDriver) AppTerminate(packageName string) (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.WDADriver.AppTerminate(packageName)
}

// GetForegroundApp returns current foreground app package name and activity name
func (s *stubIOSDriver) GetForegroundApp() (app AppInfo, err error) {
	err = s.setUpWda()
	if err != nil {
		return AppInfo{}, err
	}
	return s.WDADriver.GetForegroundApp()
}

// StartCamera Starts a new camera for recording
func (s *stubIOSDriver) StartCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StartCamera()
}

// StopCamera Stops the camera for recording
func (s *stubIOSDriver) StopCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StopCamera()
}

func (s *stubIOSDriver) Orientation() (orientation Orientation, err error) {
	err = s.setUpWda()
	if err != nil {
		return OrientationPortrait, err
	}
	return s.WDADriver.Orientation()
}

// Tap Sends a tap event at the coordinate.
func (s *stubIOSDriver) Tap(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Tap(x, y, opts...)
}

// DoubleTap Sends a double tap event at the coordinate.
func (s *stubIOSDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.DoubleTap(x, y, opts...)
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (s *stubIOSDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TouchAndHold(x, y, opts...)
}

// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
func (s *stubIOSDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Drag(fromX, fromY, toX, toY, opts...)
}

// SetPasteboard Sets data to the general pasteboard
func (s *stubIOSDriver) SetPasteboard(contentType PasteboardType, content string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SetPasteboard(contentType, content)
}

// GetPasteboard Gets the data contained in the general pasteboard.
//
//	It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
func (s *stubIOSDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.GetPasteboard(contentType)
}

func (s *stubIOSDriver) SetIme(ime string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SetIme(ime)
}

// SendKeys Types a string into active element. There must be element with keyboard focus,
// otherwise an error is raised.
// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
func (s *stubIOSDriver) SendKeys(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.SendKeys(text, opts...)
}

// Input works like SendKeys
func (s *stubIOSDriver) Input(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Input(text, opts...)
}

func (s *stubIOSDriver) Clear(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.Clear(packageName)
}

// PressButton Presses the corresponding hardware button on the device
func (s *stubIOSDriver) PressButton(devBtn DeviceButton) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressButton(devBtn)
}

// PressBack Presses the back button
func (s *stubIOSDriver) PressBack(opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressBack(opts...)
}

func (s *stubIOSDriver) PressKeyCode(keyCode KeyCode) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.PressKeyCode(keyCode)
}

func (s *stubIOSDriver) Screenshot() (*bytes.Buffer, error) {
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

func (s *stubIOSDriver) TapByText(text string, opts ...option.ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TapByText(text, opts...)
}

func (s *stubIOSDriver) TapByTexts(actions ...TapTextAction) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.TapByTexts(actions...)
}

// AccessibleSource Return application elements accessibility tree
func (s *stubIOSDriver) AccessibleSource() (string, error) {
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
func (s *stubIOSDriver) HealthCheck() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.HealthCheck()
}

func (s *stubIOSDriver) GetAppiumSettings() (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.GetAppiumSettings()
}

func (s *stubIOSDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.SetAppiumSettings(settings)
}

func (s *stubIOSDriver) IsHealthy() (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.WDADriver.IsHealthy()
}

// triggers the log capture and returns the log entries
func (s *stubIOSDriver) StartCaptureLog(identifier ...string) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.WDADriver.StartCaptureLog(identifier...)
}

func (s *stubIOSDriver) StopCaptureLog() (result interface{}, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.WDADriver.StopCaptureLog()
}

func (s *stubIOSDriver) GetDriverResults() []*DriverRequests {
	err := s.setUpWda()
	if err != nil {
		return nil
	}
	return s.WDADriver.GetDriverResults()
}

func (s *stubIOSDriver) Source(srcOpt ...option.SourceOption) (string, error) {
	resp, err := s.Request(http.MethodGet, fmt.Sprintf("%s/source?format=json&onlyWeb=false", s.bightInsightPrefix), []byte{})
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (s *stubIOSDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
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

func (s *stubIOSDriver) LogoutNoneUI(packageName string) error {
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

func (s *stubIOSDriver) TearDown() error {
	s.client.CloseIdleConnections()
	return nil
}

func (s *stubIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
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

func (s *stubIOSDriver) GetSession() *Session {
	return s.Session
}
