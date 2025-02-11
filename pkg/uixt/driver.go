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
	Setup() error
	TearDown() error

	// session
	InitSession(capabilities option.Capabilities) error
	GetSession() *Session
	DeleteSession() error

	// device info and status
	Status() (types.DeviceStatus, error)
	DeviceInfo() (types.DeviceInfo, error)
	BatteryInfo() (types.BatteryInfo, error)
	ForegroundInfo() (app types.AppInfo, err error)
	WindowSize() (types.Size, error)
	ScreenShot(opts ...option.ActionOption) (*bytes.Buffer, error)
	ScreenRecord(duration time.Duration) (videoPath string, err error)
	Source(srcOpt ...option.SourceOption) (string, error)
	Orientation() (orientation types.Orientation, err error)
	Rotation() (rotation types.Rotation, err error)

	// config
	SetRotation(rotation types.Rotation) error
	SetIme(ime string) error

	// actions
	Home() error
	Unlock() error
	Back() error
	// tap
	TapXY(x, y float64, opts ...option.ActionOption) error       // by percentage
	TapAbsXY(x, y float64, opts ...option.ActionOption) error    // by absolute coordinate
	DoubleTapXY(x, y float64, opts ...option.ActionOption) error // by percentage
	TouchAndHold(x, y float64, opts ...option.ActionOption) error
	TapByText(text string, opts ...option.ActionOption) error // TODO: remove
	TapByTexts(actions ...TapTextAction) error                // TODO: remove
	// swipe
	Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error
	Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error // by percentage
	// input
	Input(text string, opts ...option.ActionOption) error
	Backspace(count int, opts ...option.ActionOption) error

	// app related
	AppLaunch(packageName string) error
	AppTerminate(packageName string) (bool, error)
	AppClear(packageName string) error

	// triggers the log capture and returns the log entries
	StartCaptureLog(identifier ...string) error
	StopCaptureLog() (result interface{}, err error)
}

func NewXTDriver(driver IDriver, opts ...ai.AIServiceOption) *XTDriver {
	services := ai.NewAIService(opts...)
	driverExt := &XTDriver{
		IDriver:    driver,
		CVService:  services.ICVService,
		LLMService: services.ILLMService,
	}
	return driverExt
}

var _ IDriverExt = (*XTDriver)(nil)

// XTDriver = IDriver + AI
type IDriverExt interface {
	GetScreenResult(opts ...option.ActionOption) (screenResult *ScreenResult, err error)
	GetScreenTexts(opts ...option.ActionOption) (ocrTexts ai.OCRTexts, err error)

	// tap with AI
	TapByOCR(text string, opts ...option.ActionOption) error
	TapByCV(opts ...option.ActionOption) error // TODO: refactor

	CheckPopup() (popup *PopupInfo, err error)
	ClosePopupsHandler() error

	DoAction(action MobileAction) error
	DoValidation(check, assert, expected string, message ...string) error
}

type XTDriver struct {
	IDriver
	CVService  ai.ICVService  // OCR/CV
	LLMService ai.ILLMService // LLM
}

func (dExt *XTDriver) Setup() error {
	// unlock device screen
	err := dExt.Unlock()
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

func (dExt *XTDriver) assertForegroundApp(appName, assert string) error {
	app, err := dExt.ForegroundInfo()
	if err != nil {
		log.Warn().Err(err).Msg("get foreground app failed, skip app assertion")
		return nil // Notice: ignore error when get foreground app failed
	}

	switch assert {
	case AssertionEqual:
		if app.PackageName != appName {
			return errors.Wrap(err, "assert foreground app equal failed")
		}
	case AssertionNotEqual:
		if app.PackageName == appName {
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
