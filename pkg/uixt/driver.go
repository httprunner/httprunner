package uixt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/httprunner/funplugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

// IWebDriver defines methods supported by IWebDriver drivers.
type IWebDriver interface {
	// NewSession starts a new session and returns the SessionInfo.
	NewSession(capabilities option.Capabilities) (SessionInfo, error)

	// DeleteSession Kills application associated with that session and removes session
	//  1) alertsMonitor disable
	//  2) testedApplicationBundleId terminate
	DeleteSession() error

	// GetSession returns session cache, including requests, screenshots, etc.
	GetSession() *DriverSession

	Status() (DeviceStatus, error)

	DeviceInfo() (DeviceInfo, error)

	// Location Returns device location data.
	//
	// It requires to configure location access permission by manual.
	// The response of 'latitude', 'longitude' and 'altitude' are always zero (0) without authorization.
	// 'authorizationStatus' indicates current authorization status. '3' is 'Always'.
	// https://developer.apple.com/documentation/corelocation/clauthorizationstatus
	//
	//  Settings -> Privacy -> Location Service -> WebDriverAgent-Runner -> Always
	//
	// The return value could be zero even if the permission is set to 'Always'
	// since the location service needs some time to update the location data.
	Location() (Location, error)
	BatteryInfo() (BatteryInfo, error)

	// WindowSize Return the width and height in portrait mode.
	// when getting the window size in wda/ui2/adb, if the device is in landscape mode,
	// the width and height will be reversed.
	WindowSize() (Size, error)
	Screen() (Screen, error)
	Scale() (float64, error)

	// GetTimestamp returns the timestamp of the mobile device
	GetTimestamp() (timestamp int64, err error)

	// Homescreen Forces the device under test to switch to the home screen
	Homescreen() error

	Unlock() (err error)

	// AppLaunch Launch an application with given bundle identifier in scope of current session.
	// !This method is only available since Xcode9 SDK
	AppLaunch(packageName string) error
	// AppTerminate Terminate an application with the given package name.
	// Either `true` if the app has been successfully terminated or `false` if it was not running
	AppTerminate(packageName string) (bool, error)
	// GetForegroundApp returns current foreground app package name and activity name
	GetForegroundApp() (app AppInfo, err error)
	// AssertForegroundApp returns nil if the given package and activity are in foreground
	AssertForegroundApp(packageName string, activityType ...string) error

	// StartCamera Starts a new camera for recording
	StartCamera() error
	// StopCamera Stops the camera for recording
	StopCamera() error

	Orientation() (orientation Orientation, err error)

	// Tap Sends a tap event at the coordinate.
	Tap(x, y float64, opts ...option.ActionOption) error

	// DoubleTap Sends a double tap event at the coordinate.
	DoubleTap(x, y float64, opts ...option.ActionOption) error

	// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
	//  second: The default value is 1
	TouchAndHold(x, y float64, opts ...option.ActionOption) error

	// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
	// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
	Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error

	// Swipe works like Drag, but `pressForDuration` value is 0
	Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error

	// SetPasteboard Sets data to the general pasteboard
	SetPasteboard(contentType PasteboardType, content string) error
	// GetPasteboard Gets the data contained in the general pasteboard.
	//  It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
	GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error)

	SetIme(ime string) error

	// SendKeys Types a string into active element. There must be element with keyboard focus,
	// otherwise an error is raised.
	// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
	SendKeys(text string, opts ...option.ActionOption) error

	// Input works like SendKeys
	Input(text string, opts ...option.ActionOption) error

	Clear(packageName string) error

	// PressButton Presses the corresponding hardware button on the device
	PressButton(devBtn DeviceButton) error

	// PressBack Presses the back button
	PressBack(opts ...option.ActionOption) error

	PressKeyCode(keyCode KeyCode) (err error)

	Backspace(count int, opts ...option.ActionOption) (err error)

	Screenshot() (*bytes.Buffer, error)

	// Source Return application elements tree
	Source(srcOpt ...option.SourceOption) (string, error)

	LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error)
	LogoutNoneUI(packageName string) error

	TapByText(text string, opts ...option.ActionOption) error
	TapByTexts(actions ...TapTextAction) error

	// AccessibleSource Return application elements accessibility tree
	AccessibleSource() (string, error)

	// HealthCheck Health check might modify simulator state so it should only be called in-between testing sessions
	//  Checks health of XCTest by:
	//  1) Querying application for some elements,
	//  2) Triggering some device events.
	HealthCheck() error
	GetAppiumSettings() (map[string]interface{}, error)
	SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error)

	IsHealthy() (bool, error)

	// triggers the log capture and returns the log entries
	StartCaptureLog(identifier ...string) (err error)
	StopCaptureLog() (result interface{}, err error)

	GetDriverResults() []*DriverResult
	RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error)

	TearDown() error
}

type SessionInfo struct {
	SessionId    string `json:"sessionId"`
	Capabilities struct {
		Device             string `json:"device"`
		BrowserName        string `json:"browserName"`
		SdkVersion         string `json:"sdkVersion"`
		CFBundleIdentifier string `json:"CFBundleIdentifier"`
	} `json:"capabilities"`
}

type DriverSession struct {
	ID string
	// cache uia2/wda request and response
	requests []*DriverResult
	// cache screenshot ocr results
	screenResults []*ScreenResult // list of actions
	// cache e2e delay
	e2eDelay []timeLog
}

func (d *DriverSession) addScreenResult(screenResult *ScreenResult) {
	d.screenResults = append(d.screenResults, screenResult)
}

func (d *DriverSession) addRequestResult(driverResult *DriverResult) {
	d.requests = append(d.requests, driverResult)
}

func (d *DriverSession) Reset() {
	d.screenResults = make([]*ScreenResult, 0)
	d.requests = make([]*DriverResult, 0)
	d.e2eDelay = nil
}

type Attachments map[string]interface{}

func (d *DriverSession) Get(withReset bool) Attachments {
	data := Attachments{
		"screen_results": d.screenResults,
	}
	if len(d.requests) != 0 {
		data["requests"] = d.requests
	}
	if d.e2eDelay != nil {
		data["e2e_results"] = d.e2eDelay
	}
	if withReset {
		d.Reset()
	}
	return data
}

type DriverResult struct {
	RequestMethod string    `json:"request_method"`
	RequestUrl    string    `json:"request_url"`
	RequestBody   string    `json:"request_body,omitempty"`
	RequestTime   time.Time `json:"request_time"`

	Success          bool   `json:"success"`
	ResponseStatus   int    `json:"response_status"`
	ResponseDuration int64  `json:"response_duration(ms)"` // ms
	ResponseBody     string `json:"response_body"`
	Error            string `json:"error,omitempty"`
}

type DriverClient struct {
	urlPrefix *url.URL
	client    *http.Client

	// cache to avoid repeated query
	scale         float64
	windowSize    Size
	driverResults []*DriverResult

	// cache session data
	session DriverSession
}

func (wd *DriverClient) concatURL(u *url.URL, elem ...string) string {
	var tmp *url.URL
	if u == nil {
		u = wd.urlPrefix
	}
	tmp, _ = url.Parse(u.String())
	tmp.Path = path.Join(append([]string{u.Path}, elem...)...)
	return tmp.String()
}

func (wd *DriverClient) GET(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.Request(http.MethodGet, wd.concatURL(nil, pathElem...), nil)
}

func (wd *DriverClient) POST(data interface{}, pathElem ...string) (rawResp rawResponse, err error) {
	var bsJSON []byte = nil
	if data != nil {
		if bsJSON, err = json.Marshal(data); err != nil {
			return nil, err
		}
	}
	return wd.Request(http.MethodPost, wd.concatURL(nil, pathElem...), bsJSON)
}

func (wd *DriverClient) DELETE(pathElem ...string) (rawResp rawResponse, err error) {
	return wd.Request(http.MethodDelete, wd.concatURL(nil, pathElem...), nil)
}

func (wd *DriverClient) Request(method string, rawURL string, rawBody []byte) (rawResp rawResponse, err error) {
	driverResult := &DriverResult{
		RequestMethod: method,
		RequestUrl:    rawURL,
		RequestBody:   string(rawBody),
	}

	defer func() {
		wd.session.addRequestResult(driverResult)

		var logger *zerolog.Event
		if err != nil {
			driverResult.Success = false
			driverResult.Error = err.Error()
			logger = log.Error().Bool("success", false).Err(err)
		} else {
			driverResult.Success = true
			logger = log.Debug().Bool("success", true)
		}

		logger = logger.Str("request_method", method).Str("request_url", rawURL).
			Str("request_body", string(rawBody))
		if !driverResult.RequestTime.IsZero() {
			logger = logger.Int64("request_time", driverResult.RequestTime.UnixMilli())
		}
		if driverResult.ResponseStatus != 0 {
			logger = logger.
				Int("response_status", driverResult.ResponseStatus).
				Int64("response_duration(ms)", driverResult.ResponseDuration).
				Str("response_body", driverResult.ResponseBody)
		}
		logger.Msg("request uixt driver")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, method, rawURL, bytes.NewBuffer(rawBody)); err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	driverResult.RequestTime = time.Now()
	var resp *http.Response
	if resp, err = wd.client.Do(req); err != nil {
		return nil, err
	}
	defer func() {
		// https://github.com/etcd-io/etcd/blob/v3.3.25/pkg/httputil/httputil.go#L16-L22
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	rawResp, err = io.ReadAll(resp.Body)
	duration := time.Since(driverResult.RequestTime)
	driverResult.ResponseDuration = duration.Milliseconds()
	driverResult.ResponseStatus = resp.StatusCode

	if strings.HasSuffix(rawURL, "screenshot") {
		// avoid printing screenshot data
		driverResult.ResponseBody = "OMITTED"
	} else {
		driverResult.ResponseBody = string(rawResp)
	}
	if err != nil {
		return nil, err
	}

	if err = rawResp.checkErr(); err != nil {
		if resp.StatusCode == http.StatusOK {
			return rawResp, nil
		}
		return nil, err
	}

	return
}

func convertToHTTPClient(conn net.Conn) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return conn, nil
			},
		},
		Timeout: 30 * time.Second,
	}
}

type DriverExt struct {
	Ctx          context.Context
	Device       IDevice
	Driver       IWebDriver
	ImageService IImageService // used to extract image data

	// funplugin
	plugin funplugin.IPlugin
}

func newDriverExt(device IDevice, driver IWebDriver, opts ...option.DriverOption) (dExt *DriverExt, err error) {
	driverOptions := option.NewDriverOptions(opts...)

	dExt = &DriverExt{
		Device: device,
		Driver: driver,
		plugin: driverOptions.Plugin,
	}

	if driverOptions.WithImageService {
		if dExt.ImageService, err = newVEDEMImageService(); err != nil {
			return nil, err
		}
	}
	if driverOptions.WithResultFolder {
		// create results directory
		if err = builtin.EnsureFolderExists(config.ResultsPath); err != nil {
			return nil, errors.Wrap(err, "create results directory failed")
		}
		if err = builtin.EnsureFolderExists(config.ScreenShotsPath); err != nil {
			return nil, errors.Wrap(err, "create screenshots directory failed")
		}
	}
	return dExt, nil
}

func (dExt *DriverExt) Init() error {
	// unlock device screen
	err := dExt.Driver.Unlock()
	if err != nil {
		log.Error().Err(err).Msg("unlock device screen failed")
		return err
	}

	return nil
}

func (dExt *DriverExt) assertOCR(text, assert string) error {
	var opts []option.ActionOption
	opts = append(opts, option.WithScreenShotFileName(fmt.Sprintf("assert_ocr_%s", text)))

	switch assert {
	case AssertionEqual:
		_, err := dExt.FindScreenText(text, opts...)
		if err != nil {
			return errors.Wrap(err, "assert ocr equal failed")
		}
	case AssertionNotEqual:
		_, err := dExt.FindScreenText(text, opts...)
		if err == nil {
			return errors.New("assert ocr not equal failed")
		}
	case AssertionExists:
		opts = append(opts, option.WithRegex(true))
		_, err := dExt.FindScreenText(text, opts...)
		if err != nil {
			return errors.Wrap(err, "assert ocr exists failed")
		}
	case AssertionNotExists:
		opts = append(opts, option.WithRegex(true))
		_, err := dExt.FindScreenText(text, opts...)
		if err == nil {
			return errors.New("assert ocr not exists failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *DriverExt) assertForegroundApp(appName, assert string) (err error) {
	err = dExt.Driver.AssertForegroundApp(appName)
	switch assert {
	case AssertionEqual:
		if err != nil {
			return errors.Wrap(err, "assert foreground app equal failed")
		}
	case AssertionNotEqual:
		if err == nil {
			return errors.New("assert foreground app not equal failed")
		}
	default:
		return fmt.Errorf("unexpected assert method %s", assert)
	}
	return nil
}

func (dExt *DriverExt) DoValidation(check, assert, expected string, message ...string) (err error) {
	switch check {
	case SelectorOCR:
		err = dExt.assertOCR(expected, assert)
	case SelectorForegroundApp:
		err = dExt.assertForegroundApp(expected, assert)
	}

	if err != nil {
		if message == nil {
			message = []string{""}
		}
		log.Error().Err(err).Str("assert", assert).Str("expect", expected).
			Str("msg", message[0]).Msg("validate failed")
		return err
	}

	log.Info().Str("assert", assert).Str("expect", expected).Msg("validate success")
	return nil
}
