package uixt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type stubIOSDriver struct {
	bightInsightPrefix string
	serverPrefix       string
	timeout            time.Duration
	Driver
	*wdaDriver
	device *IOSDevice
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
	driver.Driver.client = &http.Client{
		Timeout: time.Second * 10, // 设置超时时间为 10 秒
	}
	return driver, nil
}

func (s *stubIOSDriver) setUpWda() (err error) {
	if s.wdaDriver == nil {
		capabilities := NewCapabilities()
		capabilities.WithDefaultAlertAction(AlertActionAccept)
		driver, err := s.device.NewHTTPDriver(capabilities)
		if err != nil {
			log.Error().Err(err).Msg("stub driver failed to init wda driver")
			return err
		}
		s.wdaDriver = driver.(*wdaDriver)
	}
	return nil
}

// NewSession starts a new session and returns the SessionInfo.
func (s *stubIOSDriver) NewSession(capabilities Capabilities) (SessionInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return SessionInfo{}, err
	}
	return s.wdaDriver.NewSession(capabilities)
}

// DeleteSession Kills application associated with that session and removes session
//  1. alertsMonitor disable
//  2. testedApplicationBundleId terminate
func (s *stubIOSDriver) DeleteSession() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.DeleteSession()
}

func (s *stubIOSDriver) Status() (DeviceStatus, error) {
	err := s.setUpWda()
	if err != nil {
		return DeviceStatus{}, err
	}
	return s.wdaDriver.Status()
}

func (s *stubIOSDriver) DeviceInfo() (DeviceInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return DeviceInfo{}, err
	}
	return s.wdaDriver.DeviceInfo()
}

func (s *stubIOSDriver) Location() (Location, error) {
	err := s.setUpWda()
	if err != nil {
		return Location{}, err
	}
	return s.wdaDriver.Location()
}

func (s *stubIOSDriver) BatteryInfo() (BatteryInfo, error) {
	err := s.setUpWda()
	if err != nil {
		return BatteryInfo{}, err
	}
	return s.wdaDriver.BatteryInfo()
}

// WindowSize Return the width and height in portrait mode.
// when getting the window size in wda/ui2/adb, if the device is in landscape mode,
// the width and height will be reversed.
func (s *stubIOSDriver) WindowSize() (Size, error) {
	err := s.setUpWda()
	if err != nil {
		return Size{}, err
	}
	return s.wdaDriver.WindowSize()
}

func (s *stubIOSDriver) Screen() (Screen, error) {
	err := s.setUpWda()
	if err != nil {
		return Screen{}, err
	}
	return s.wdaDriver.Screen()
}

func (s *stubIOSDriver) Scale() (float64, error) {
	err := s.setUpWda()
	if err != nil {
		return 0, err
	}
	return s.wdaDriver.Scale()
}

// GetTimestamp returns the timestamp of the mobile device
func (s *stubIOSDriver) GetTimestamp() (timestamp int64, err error) {
	err = s.setUpWda()
	if err != nil {
		return 0, err
	}
	return s.wdaDriver.GetTimestamp()
}

// Homescreen Forces the device under test to switch to the home screen
func (s *stubIOSDriver) Homescreen() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Homescreen()
}

func (s *stubIOSDriver) Unlock() (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Unlock()
}

// AppLaunch Launch an application with given bundle identifier in scope of current session.
// !This method is only available since Xcode9 SDK
func (s *stubIOSDriver) AppLaunch(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.AppLaunch(packageName)
}

// AppTerminate Terminate an application with the given package name.
// Either `true` if the app has been successfully terminated or `false` if it was not running
func (s *stubIOSDriver) AppTerminate(packageName string) (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.wdaDriver.AppTerminate(packageName)
}

// GetForegroundApp returns current foreground app package name and activity name
func (s *stubIOSDriver) GetForegroundApp() (app AppInfo, err error) {
	err = s.setUpWda()
	if err != nil {
		return AppInfo{}, err
	}
	return s.wdaDriver.GetForegroundApp()
}

// AssertForegroundApp returns nil if the given package and activity are in foreground
func (s *stubIOSDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.AssertForegroundApp(packageName, activityType...)
}

// StartCamera Starts a new camera for recording
func (s *stubIOSDriver) StartCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.StartCamera()
}

// StopCamera Stops the camera for recording
func (s *stubIOSDriver) StopCamera() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.StopCamera()
}

func (s *stubIOSDriver) Orientation() (orientation Orientation, err error) {
	err = s.setUpWda()
	if err != nil {
		return OrientationPortrait, err
	}
	return s.wdaDriver.Orientation()
}

// Tap Sends a tap event at the coordinate.
func (s *stubIOSDriver) Tap(x, y float64, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Tap(x, y, options...)
}

// DoubleTap Sends a double tap event at the coordinate.
func (s *stubIOSDriver) DoubleTap(x, y float64, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.DoubleTap(x, y, options...)
}

// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
//
//	second: The default value is 1
func (s *stubIOSDriver) TouchAndHold(x, y float64, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.TouchAndHold(x, y, options...)
}

// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
func (s *stubIOSDriver) Drag(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Drag(fromX, fromY, toX, toY, options...)
}

// SetPasteboard Sets data to the general pasteboard
func (s *stubIOSDriver) SetPasteboard(contentType PasteboardType, content string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.SetPasteboard(contentType, content)
}

// GetPasteboard Gets the data contained in the general pasteboard.
//
//	It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
func (s *stubIOSDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.GetPasteboard(contentType)
}

func (s *stubIOSDriver) SetIme(ime string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.SetIme(ime)
}

// SendKeys Types a string into active element. There must be element with keyboard focus,
// otherwise an error is raised.
// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
func (s *stubIOSDriver) SendKeys(text string, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.SendKeys(text, options...)
}

// Input works like SendKeys
func (s *stubIOSDriver) Input(text string, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Input(text, options...)
}

func (s *stubIOSDriver) Clear(packageName string) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.Clear(packageName)
}

// PressButton Presses the corresponding hardware button on the device
func (s *stubIOSDriver) PressButton(devBtn DeviceButton) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.PressButton(devBtn)
}

// PressBack Presses the back button
func (s *stubIOSDriver) PressBack(options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.PressBack(options...)
}

func (s *stubIOSDriver) PressKeyCode(keyCode KeyCode) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.PressKeyCode(keyCode)
}

func (s *stubIOSDriver) Screenshot() (*bytes.Buffer, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.Screenshot()
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

func (s *stubIOSDriver) TapByText(text string, options ...ActionOption) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.TapByText(text, options...)
}

func (s *stubIOSDriver) TapByTexts(actions ...TapTextAction) error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.TapByTexts(actions...)
}

// AccessibleSource Return application elements accessibility tree
func (s *stubIOSDriver) AccessibleSource() (string, error) {
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
func (s *stubIOSDriver) HealthCheck() error {
	err := s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.HealthCheck()
}

func (s *stubIOSDriver) GetAppiumSettings() (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.GetAppiumSettings()
}

func (s *stubIOSDriver) SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error) {
	err := s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.SetAppiumSettings(settings)
}

func (s *stubIOSDriver) IsHealthy() (bool, error) {
	err := s.setUpWda()
	if err != nil {
		return false, err
	}
	return s.wdaDriver.IsHealthy()
}

// triggers the log capture and returns the log entries
func (s *stubIOSDriver) StartCaptureLog(identifier ...string) (err error) {
	err = s.setUpWda()
	if err != nil {
		return err
	}
	return s.wdaDriver.StartCaptureLog(identifier...)
}

func (s *stubIOSDriver) StopCaptureLog() (result interface{}, err error) {
	err = s.setUpWda()
	if err != nil {
		return nil, err
	}
	return s.wdaDriver.StopCaptureLog()
}

func (s *stubIOSDriver) GetDriverResults() []*DriverResult {
	err := s.setUpWda()
	if err != nil {
		return nil
	}
	return s.wdaDriver.GetDriverResults()
}

func (s *stubIOSDriver) Source(srcOpt ...SourceOption) (string, error) {
	resp, err := s.Driver.Request(http.MethodGet, fmt.Sprintf("%s/source?format=json&onlyWeb=false", s.bightInsightPrefix), []byte{})
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
	resp, err := s.Driver.Request(http.MethodPost, fmt.Sprintf("%s/host/login/account/", s.serverPrefix), bsJSON)
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
	resp, err := s.Driver.Request(http.MethodGet, fmt.Sprintf("%s/host/loginout/", s.serverPrefix), []byte{})
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
	s.Driver.client.CloseIdleConnections()
	return nil
}

func (s *stubIOSDriver) getLoginAppInfo(packageName string) (info AppLoginInfo, err error) {
	resp, err := s.Driver.Request(http.MethodGet, fmt.Sprintf("%s/host/app/info/", s.serverPrefix), []byte{})
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

func (s *stubIOSDriver) GetSession() *DriverSession {
	return &s.Driver.session
}
