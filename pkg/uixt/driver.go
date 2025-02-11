package uixt

import (
	"bytes"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/png"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

var (
	_ IDriver = (*ADBDriver)(nil)
	_ IDriver = (*UIA2Driver)(nil)
	_ IDriver = (*WDADriver)(nil)
	_ IDriver = (*HDCDriver)(nil)
)

// current implemeted driver: ADBDriver, UIA2Driver, WDADriver, HDCDriver
type IDriver interface {
	GetDevice() IDevice

	// session
	InitSession(capabilities option.Capabilities) error
	DeleteSession() error
	GetSession() *Session

	// device info and status
	Status() (types.DeviceStatus, error)
	DeviceInfo() (types.DeviceInfo, error)
	BatteryInfo() (types.BatteryInfo, error)
	WindowSize() (types.Size, error)
	Screen() (ai.Screen, error)
	Scale() (float64, error)

	// actions
	Homescreen() error
	Unlock() (err error)

	// AppLaunch Launch an application with given bundle identifier in scope of current session.
	// !This method is only available since Xcode9 SDK
	AppLaunch(packageName string) error
	// AppTerminate Terminate an application with the given package name.
	// Either `true` if the app has been successfully terminated or `false` if it was not running
	AppTerminate(packageName string) (bool, error)
	// GetForegroundApp returns current foreground app package name and activity name
	GetForegroundApp() (app types.AppInfo, err error)
	// AssertForegroundApp returns nil if the given package and activity are in foreground
	AssertForegroundApp(packageName string, activityType ...string) error

	// StartCamera Starts a new camera for recording
	StartCamera() error
	// StopCamera Stops the camera for recording
	StopCamera() error

	Orientation() (orientation types.Orientation, err error)

	SetRotation(rotation types.Rotation) (err error)
	Rotation() (rotation types.Rotation, err error)

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
	SetPasteboard(contentType types.PasteboardType, content string) error
	// GetPasteboard Gets the data contained in the general pasteboard.
	//  It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
	GetPasteboard(contentType types.PasteboardType) (raw *bytes.Buffer, err error)

	SetIme(ime string) error

	// SendKeys Types a string into active element. There must be element with keyboard focus,
	// otherwise an error is raised.
	// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
	SendKeys(text string, opts ...option.ActionOption) error

	// Input works like SendKeys
	Input(text string, opts ...option.ActionOption) error

	Clear(packageName string) error

	// PressButton Presses the corresponding hardware button on the device
	PressButton(devBtn types.DeviceButton) error

	// PressBack Presses the back button
	PressBack(opts ...option.ActionOption) error

	PressKeyCode(keyCode KeyCode) (err error)

	Backspace(count int, opts ...option.ActionOption) (err error)

	Screenshot() (*bytes.Buffer, error)

	// Source Return application elements tree
	Source(srcOpt ...option.SourceOption) (string, error)

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

	GetDriverResults() []*DriverRequests
	RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error)

	Setup() error
	TearDown() error
}

func NewXTDriver(driver IDriver, opts ...ai.AIServiceOption) *XTDriver {
	services := ai.NewAIService(opts...)
	driverExt := &XTDriver{
		Driver:     driver,
		CVService:  services.ICVService,
		LLMService: services.ILLMService,
	}
	return driverExt
}

var _ IDriverExt = (*XTDriver)(nil)

// XTDriver = IDriver + AI
type IDriverExt interface {
	GetDriver() IDriver // get original driver

	GetScreenResult(opts ...option.ActionOption) (screenResult *ScreenResult, err error)
	GetScreenTexts(opts ...option.ActionOption) (ocrTexts ai.OCRTexts, err error)
	GetScreenShot(fileName string) (raw *bytes.Buffer, path string, err error)

	// tap
	TapByOCR(ocrText string, opts ...option.ActionOption) error
	TapXY(x, y float64, opts ...option.ActionOption) error
	TapAbsXY(x, y float64, opts ...option.ActionOption) error
	TapOffset(param string, xOffset, yOffset float64, opts ...option.ActionOption) (err error)
	TapByUIDetection(opts ...option.ActionOption) error

	// swipe
	SwipeRelative(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error
	SwipeUp(opts ...option.ActionOption) error
	SwipeDown(opts ...option.ActionOption) error
	SwipeLeft(opts ...option.ActionOption) error
	SwipeRight(opts ...option.ActionOption) error

	SwipeToTapApp(appName string, opts ...option.ActionOption) error

	CheckPopup() (popup *PopupInfo, err error)
	ClosePopupsHandler() error

	DoAction(action MobileAction) (err error)
	DoValidation(check, assert, expected string, message ...string) (err error)
}

type XTDriver struct {
	Driver     IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM
}

func (dExt *XTDriver) GetDriver() IDriver {
	return dExt.Driver
}

func (dExt *XTDriver) Setup() error {
	// unlock device screen
	err := dExt.Driver.Unlock()
	if err != nil {
		log.Error().Err(err).Msg("unlock device screen failed")
		return err
	}

	return nil
}

func (dExt *XTDriver) assertOCR(text, assert string) error {
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

func (dExt *XTDriver) assertForegroundApp(appName, assert string) (err error) {
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

func (dExt *XTDriver) DoValidation(check, assert, expected string, message ...string) (err error) {
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

type DriverRawResponse []byte

func (r DriverRawResponse) CheckErr() (err error) {
	reply := new(struct {
		Value struct {
			Err        string `json:"error"`
			Message    string `json:"message"`
			Traceback  string `json:"traceback"`  // wda
			Stacktrace string `json:"stacktrace"` // uia
		}
	})
	if err = json.Unmarshal(r, reply); err != nil {
		return err
	}
	if reply.Value.Err != "" {
		errText := reply.Value.Message
		re := regexp.MustCompile(`{.+?=(.+?)}`)
		if re.MatchString(reply.Value.Message) {
			subMatch := re.FindStringSubmatch(reply.Value.Message)
			errText = subMatch[len(subMatch)-1]
		}
		return fmt.Errorf("%s: %s", reply.Value.Err, errText)
	}
	return
}

func (r DriverRawResponse) ValueConvertToString() (s string, err error) {
	reply := new(struct{ Value string })
	if err = json.Unmarshal(r, reply); err != nil {
		return "", errors.Wrapf(err, "json.Unmarshal failed, rawResponse: %s", string(r))
	}
	s = reply.Value
	return
}

func (r DriverRawResponse) ValueConvertToBool() (b bool, err error) {
	reply := new(struct{ Value bool })
	if err = json.Unmarshal(r, reply); err != nil {
		return false, err
	}
	b = reply.Value
	return
}

func (r DriverRawResponse) ValueConvertToSessionInfo() (sessionInfo Session, err error) {
	reply := new(struct{ Value struct{ Session } })
	if err = json.Unmarshal(r, reply); err != nil {
		return Session{}, err
	}
	sessionInfo = reply.Value.Session
	return
}

func (r DriverRawResponse) ValueConvertToJsonRawMessage() (raw builtinJSON.RawMessage, err error) {
	reply := new(struct{ Value builtinJSON.RawMessage })
	if err = json.Unmarshal(r, reply); err != nil {
		return nil, err
	}
	raw = reply.Value
	return
}

func (r DriverRawResponse) ValueConvertToJsonObject() (obj map[string]interface{}, err error) {
	if err = json.Unmarshal(r, &obj); err != nil {
		return nil, err
	}
	return
}

func (r DriverRawResponse) ValueDecodeAsBase64() (raw *bytes.Buffer, err error) {
	str, err := r.ValueConvertToString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to string")
	}
	decodeString, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64 string")
	}
	raw = bytes.NewBuffer(decodeString)
	return
}
