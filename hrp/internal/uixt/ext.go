package uixt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type MobileMethod string

const (
	AppInstall          MobileMethod = "install"
	AppUninstall        MobileMethod = "uninstall"
	AppStart            MobileMethod = "app_start"
	AppLaunch           MobileMethod = "app_launch"            // 等待 app 打开并堵塞到 app 首屏加载完成，可以传入 app 的启动参数、环境变量
	AppLaunchUnattached MobileMethod = "app_launch_unattached" // 只负责通知打开 app，不堵塞等待，不可传入启动参数
	AppTerminate        MobileMethod = "app_terminate"
	AppStop             MobileMethod = "app_stop"
	CtlScreenShot       MobileMethod = "screenshot"
	CtlSleep            MobileMethod = "sleep"
	CtlStartCamera      MobileMethod = "camera_start" // alias for app_launch camera
	CtlStopCamera       MobileMethod = "camera_stop"  // alias for app_terminate camera
	RecordStart         MobileMethod = "record_start"
	RecordStop          MobileMethod = "record_stop"

	// UI validation
	SelectorName       string = "ui_name"
	SelectorLabel      string = "ui_label"
	SelectorOCR        string = "ui_ocr"
	SelectorImage      string = "ui_image"
	AssertionExists    string = "exists"
	AssertionNotExists string = "not_exists"

	// UI handling
	ACTION_Home        MobileMethod = "home"
	ACTION_TapXY       MobileMethod = "tap_xy"
	ACTION_TapByOCR    MobileMethod = "tap_ocr"
	ACTION_TapByCV     MobileMethod = "tap_cv"
	ACTION_Tap         MobileMethod = "tap"
	ACTION_DoubleTapXY MobileMethod = "double_tap_xy"
	ACTION_DoubleTap   MobileMethod = "double_tap"
	ACTION_Swipe       MobileMethod = "swipe"
	ACTION_Input       MobileMethod = "input"

	// custom actions
	ACTION_SwipeToTapApp  MobileMethod = "swipe_to_tap_app"  // swipe left & right to find app and tap
	ACTION_SwipeToTapText MobileMethod = "swipe_to_tap_text" // swipe up & down to find text and tap
)

type MobileAction struct {
	Method MobileMethod `json:"method,omitempty" yaml:"method,omitempty"`
	Params interface{}  `json:"params,omitempty" yaml:"params,omitempty"`

	Identifier          string `json:"identifier,omitempty" yaml:"identifier,omitempty"`                     // used to identify the action in log
	MaxRetryTimes       int    `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"`           // max retry times
	Timeout             int    `json:"timeout,omitempty" yaml:"timeout,omitempty"`                           // TODO: wait timeout in seconds for mobile action
	IgnoreNotFoundError bool   `json:"ignore_NotFoundError,omitempty" yaml:"ignore_NotFoundError,omitempty"` // ignore error if target element not found
}

type ActionOption func(o *MobileAction)

func WithIdentifier(identifier string) ActionOption {
	return func(o *MobileAction) {
		o.Identifier = identifier
	}
}

func WithMaxRetryTimes(maxRetryTimes int) ActionOption {
	return func(o *MobileAction) {
		o.MaxRetryTimes = maxRetryTimes
	}
}

func WithTimeout(timeout int) ActionOption {
	return func(o *MobileAction) {
		o.Timeout = timeout
	}
}

func WithIgnoreNotFoundError(ignoreError bool) ActionOption {
	return func(o *MobileAction) {
		o.IgnoreNotFoundError = ignoreError
	}
}

// TemplateMatchMode is the type of the template matching operation.
type TemplateMatchMode int

type CVArgs struct {
	matchMode TemplateMatchMode
	threshold float64
}

type CVOption func(*CVArgs)

func WithTemplateMatchMode(mode TemplateMatchMode) CVOption {
	return func(args *CVArgs) {
		args.matchMode = mode
	}
}

func WithThreshold(threshold float64) CVOption {
	return func(args *CVArgs) {
		args.threshold = threshold
	}
}

type DriverExt struct {
	Driver          WebDriver
	windowSize      Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool
	scale           float64
	host            string
	StartTime       time.Time // used to associate screenshots name
	ScreenShots     []string  // save screenshots path

	CVArgs
}

func extend(driver WebDriver) (dExt *DriverExt, err error) {
	dExt = &DriverExt{Driver: driver}
	dExt.doneMjpegStream = make(chan bool, 1)

	// get device window size
	dExt.windowSize, err = dExt.Driver.WindowSize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get windows size")
	}

	if dExt.scale, err = dExt.Driver.Scale(); err != nil {
		return nil, err
	}

	return dExt, nil
}

func (dExt *DriverExt) takeScreenShot() (raw *bytes.Buffer, err error) {
	// 优先使用 MJPEG 流进行截图，性能最优
	// 如果 MJPEG 流未开启，则使用 WebDriver 的截图接口
	if dExt.frame != nil {
		return dExt.frame, nil
	}
	if raw, err = dExt.Driver.Screenshot(); err != nil {
		log.Error().Err(err).Msgf("screenshot failed: %v", err)
		return nil, err
	}
	return
}

// saveScreenShot saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) saveScreenShot(raw *bytes.Buffer, fileName string) (string, error) {
	img, format, err := image.Decode(raw)
	if err != nil {
		return "", errors.Wrap(err, "decode screenshot image failed")
	}

	dir, _ := os.Getwd()
	screenshotsDir := filepath.Join(dir, "screenshots")
	if err = os.MkdirAll(screenshotsDir, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "create screenshots directory failed")
	}
	screenshotPath := filepath.Join(screenshotsDir,
		fmt.Sprintf("%s.%s", fileName, format))

	file, err := os.Create(screenshotPath)
	if err != nil {
		return "", errors.Wrap(err, "create screenshot image file failed")
	}
	defer func() {
		_ = file.Close()
	}()

	switch format {
	case "png":
		err = png.Encode(file, img)
	case "jpeg":
		err = jpeg.Encode(file, img, nil)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return "", errors.Wrap(err, "encode screenshot image failed")
	}

	return screenshotPath, nil
}

// ScreenShot takes screenshot and saves image file to $CWD/screenshots/ folder
func (dExt *DriverExt) ScreenShot(fileName string) (string, error) {
	raw, err := dExt.takeScreenShot()
	if err != nil {
		return "", errors.Wrap(err, "screenshot by WDA failed")
	}

	return dExt.saveScreenShot(raw, fileName)
}

// isPathExists returns true if path exists, whether path is file or dir
func isPathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (dExt *DriverExt) FindUIElement(param string) (ele WebElement, err error) {
	var selector BySelector
	if strings.HasPrefix(param, "/") {
		// xpath
		selector = BySelector{
			XPath: param,
		}
	} else {
		// name
		selector = BySelector{
			LinkText: NewElementAttribute().WithName(param),
		}
	}

	return dExt.Driver.FindElement(selector)
}

func (dExt *DriverExt) FindUIRectInUIKit(search string) (x, y, width, height float64, err error) {
	// click on text, using OCR
	if !isPathExists(search) {
		return dExt.FindTextByOCR(search)
	}
	// click on image, using opencv
	return dExt.FindImageRectInUIKit(search)
}

func (dExt *DriverExt) MappingToRectInUIKit(rect image.Rectangle) (x, y, width, height float64) {
	x, y = float64(rect.Min.X)/dExt.scale, float64(rect.Min.Y)/dExt.scale
	width, height = float64(rect.Dx())/dExt.scale, float64(rect.Dy())/dExt.scale
	return
}

func (dExt *DriverExt) PerformTouchActions(touchActions *TouchActions) error {
	return dExt.Driver.PerformAppiumTouchActions(touchActions)
}

func (dExt *DriverExt) PerformActions(actions *W3CActions) error {
	return dExt.Driver.PerformW3CActions(actions)
}

func (dExt *DriverExt) IsNameExist(name string) bool {
	selector := BySelector{
		LinkText: NewElementAttribute().WithName(name),
	}
	_, err := dExt.Driver.FindElement(selector)
	return err == nil
}

func (dExt *DriverExt) IsLabelExist(label string) bool {
	selector := BySelector{
		LinkText: NewElementAttribute().WithLabel(label),
	}
	_, err := dExt.Driver.FindElement(selector)
	return err == nil
}

func (dExt *DriverExt) IsOCRExist(text string) bool {
	_, _, _, _, err := dExt.FindTextByOCR(text)
	return err == nil
}

func (dExt *DriverExt) IsImageExist(text string) bool {
	_, _, _, _, err := dExt.FindImageRectInUIKit(text)
	return err == nil
}

var errActionNotImplemented = errors.New("UI action not implemented")

func (dExt *DriverExt) DoAction(action MobileAction) error {
	log.Info().Str("method", string(action.Method)).Interface("params", action.Params).Msg("start iOS UI action")

	switch action.Method {
	case AppInstall:
		// TODO
		return errActionNotImplemented
	case AppLaunch:
		if bundleId, ok := action.Params.(string); ok {
			return dExt.Driver.AppLaunch(bundleId)
		}
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			AppLaunch, action.Params)
	case AppLaunchUnattached:
		if bundleId, ok := action.Params.(string); ok {
			return dExt.Driver.AppLaunchUnattached(bundleId)
		}
		return fmt.Errorf("invalid %s params, should be bundleId(string), got %v",
			AppLaunchUnattached, action.Params)
	case ACTION_SwipeToTapApp:
		if appName, ok := action.Params.(string); ok {
			var x, y, width, height float64
			findApp := func(d *DriverExt) error {
				var err error
				x, y, width, height, err = d.FindTextByOCR(appName)
				return err
			}
			foundAppAction := func(d *DriverExt) error {
				// click app to launch
				return d.Driver.TapFloat(x+width*0.5, y+height*0.5-20)
			}

			// go to home screen
			if err := dExt.Driver.Homescreen(); err != nil {
				return errors.Wrap(err, "go to home screen failed")
			}

			// swipe to first screen
			for i := 0; i < 5; i++ {
				dExt.SwipeRight()
			}

			// default to retry 5 times
			if action.MaxRetryTimes == 0 {
				action.MaxRetryTimes = 5
			}
			// swipe next screen until app found
			return dExt.SwipeUntil("left", findApp, foundAppAction, action.MaxRetryTimes)
		}
		return fmt.Errorf("invalid %s params, should be app name(string), got %v",
			ACTION_SwipeToTapApp, action.Params)
	case ACTION_SwipeToTapText:
		if text, ok := action.Params.(string); ok {
			var x, y, width, height float64
			findText := func(d *DriverExt) error {
				var err error
				x, y, width, height, err = d.FindTextByOCR(text)
				return err
			}
			foundTextAction := func(d *DriverExt) error {
				// tap text
				return d.Driver.TapFloat(x+width*0.5, y+height*0.5)
			}

			// default to retry 10 times
			if action.MaxRetryTimes == 0 {
				action.MaxRetryTimes = 10
			}
			// swipe until live room found
			return dExt.SwipeUntil("up", findText, foundTextAction, action.MaxRetryTimes)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			ACTION_SwipeToTapText, action.Params)
	case AppTerminate:
		if bundleId, ok := action.Params.(string); ok {
			success, err := dExt.Driver.AppTerminate(bundleId)
			if err != nil {
				return errors.Wrap(err, "failed to terminate app")
			}
			if !success {
				log.Warn().Str("bundleId", bundleId).Msg("app was not running")
			}
			return nil
		}
		return fmt.Errorf("app_terminate params should be bundleId(string), got %v", action.Params)
	case ACTION_Home:
		return dExt.Driver.Homescreen()
	case ACTION_TapXY:
		if location, ok := action.Params.([]float64); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			return dExt.TapXY(location[0], location[1], action.Identifier)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapXY, action.Params)
	case ACTION_Tap:
		if param, ok := action.Params.(string); ok {
			return dExt.Tap(param, action.Identifier, action.IgnoreNotFoundError)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_Tap, action.Params)
	case ACTION_TapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			return dExt.TapByOCR(ocrText, action.Identifier, action.IgnoreNotFoundError)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByOCR, action.Params)
	case ACTION_TapByCV:
		if imagePath, ok := action.Params.(string); ok {
			return dExt.TapByCV(imagePath, action.Identifier, action.IgnoreNotFoundError)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByCV, action.Params)
	case ACTION_DoubleTapXY:
		if location, ok := action.Params.([]float64); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			return dExt.DoubleTapXY(location[0], location[1])
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTapXY, action.Params)
	case ACTION_DoubleTap:
		if param, ok := action.Params.(string); ok {
			return dExt.DoubleTap(param)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTap, action.Params)
	case ACTION_Swipe:
		if positions, ok := action.Params.([]float64); ok {
			// relative fromX, fromY, toX, toY of window size: [0.5, 0.9, 0.5, 0.1]
			if len(positions) != 4 {
				return fmt.Errorf("invalid swipe params [fromX, fromY, toX, toY]: %v", positions)
			}
			return dExt.SwipeRelative(
				positions[0], positions[1], positions[2], positions[3], action.Identifier)
		}
		if direction, ok := action.Params.(string); ok {
			return dExt.SwipeTo(direction, action.Identifier)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_Swipe, action.Params)
	case ACTION_Input:
		// input text on current active element
		// append \n to send text with enter
		// send \b\b\b to delete 3 chars
		param := fmt.Sprintf("%v", action.Params)
		return dExt.Driver.SendKeys(param)
	case CtlSleep:
		if param, ok := action.Params.(json.Number); ok {
			seconds, _ := param.Float64()
			time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(float64); ok {
			time.Sleep(time.Duration(param*1000) * time.Millisecond)
			return nil
		}
		return fmt.Errorf("invalid sleep params: %v(%T)", action.Params, action.Params)
	case CtlScreenShot:
		// take snapshot
		log.Info().Msg("take snapshot for current screen")
		screenshotPath, err := dExt.ScreenShot(fmt.Sprintf("%d_screenshot_%d",
			dExt.StartTime.Unix(), time.Now().Unix()))
		if err != nil {
			return errors.Wrap(err, "take screenshot failed")
		}
		log.Info().Str("path", screenshotPath).Msg("take screenshot")
		dExt.ScreenShots = append(dExt.ScreenShots, screenshotPath)
		return err
	case CtlStartCamera:
		// start camera, alias for app_launch com.apple.camera
		return dExt.Driver.AppLaunch("com.apple.camera")
	case CtlStopCamera:
		// stop camera, alias for app_terminate com.apple.camera
		success, err := dExt.Driver.AppTerminate("com.apple.camera")
		if err != nil {
			return errors.Wrap(err, "failed to terminate camera")
		}
		if !success {
			log.Warn().Msg("camera was not running")
		}
		return nil
	}
	return nil
}

func (dExt *DriverExt) DoValidation(check, assert, expected string, message ...string) bool {
	var exists bool
	if assert == AssertionExists {
		exists = true
	} else {
		exists = false
	}
	var result bool
	switch check {
	case SelectorName:
		result = (dExt.IsNameExist(expected) == exists)
	case SelectorLabel:
		result = (dExt.IsLabelExist(expected) == exists)
	case SelectorOCR:
		result = (dExt.IsOCRExist(expected) == exists)
	case SelectorImage:
		result = (dExt.IsImageExist(expected) == exists)
	}

	if !result {
		if message == nil {
			message = []string{""}
		}
		log.Error().
			Str("assert", assert).
			Str("expect", expected).
			Str("msg", message[0]).
			Msg("validate UI failed")
		return false
	}

	log.Info().
		Str("assert", assert).
		Str("expect", expected).
		Msg("validate UI success")
	return true
}
