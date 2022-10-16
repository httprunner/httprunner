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
	"testing"
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
	ACTION_TapAbsXY    MobileMethod = "tap_abs_xy"
	ACTION_TapByOCR    MobileMethod = "tap_ocr"
	ACTION_TapByCV     MobileMethod = "tap_cv"
	ACTION_Tap         MobileMethod = "tap"
	ACTION_DoubleTapXY MobileMethod = "double_tap_xy"
	ACTION_DoubleTap   MobileMethod = "double_tap"
	ACTION_Swipe       MobileMethod = "swipe"
	ACTION_Input       MobileMethod = "input"

	// custom actions
	ACTION_SwipeToTapApp   MobileMethod = "swipe_to_tap_app"   // swipe left & right to find app and tap
	ACTION_SwipeToTapText  MobileMethod = "swipe_to_tap_text"  // swipe up & down to find text and tap
	ACTION_SwipeToTapTexts MobileMethod = "swipe_to_tap_texts" // swipe up & down to find text and tap
)

type MobileAction struct {
	Method MobileMethod `json:"method,omitempty" yaml:"method,omitempty"`
	Params interface{}  `json:"params,omitempty" yaml:"params,omitempty"`

	Identifier          string      `json:"identifier,omitempty" yaml:"identifier,omitempty"`                     // used to identify the action in log
	MaxRetryTimes       int         `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"`           // max retry times
	Direction           interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`                       // used by swipe to tap text or app
	Scope               []float64   `json:"scope,omitempty" yaml:"scope,omitempty"`                               // used by ocr to get text position in the scope
	Index               int         `json:"index,omitempty" yaml:"index,omitempty"`                               // index of the target element, should start from 1
	Timeout             int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                           // TODO: wait timeout in seconds for mobile action
	IgnoreNotFoundError bool        `json:"ignore_NotFoundError,omitempty" yaml:"ignore_NotFoundError,omitempty"` // ignore error if target element not found
	Text                string      `json:"text,omitempty" yaml:"text,omitempty"`
	ID                  string      `json:"id,omitempty" yaml:"id,omitempty"`
	Description         string      `json:"description,omitempty" yaml:"description,omitempty"`
}

type ActionOption func(o *MobileAction)

func WithIdentifier(identifier string) ActionOption {
	return func(o *MobileAction) {
		o.Identifier = identifier
	}
}

func WithIndex(index int) ActionOption {
	return func(o *MobileAction) {
		o.Index = index
	}
}

// WithDirection inputs direction (up, down, left, right)
func WithDirection(direction string) ActionOption {
	return func(o *MobileAction) {
		o.Direction = direction
	}
}

// WithCustomDirection inputs sx, sy, ex, ey
func WithCustomDirection(sx, sy, ex, ey float64) ActionOption {
	return func(o *MobileAction) {
		o.Direction = []float64{sx, sy, ex, ey}
	}
}

// WithScope inputs area of [(x1,y1), (x2,y2)]
func WithScope(x1, y1, x2, y2 float64) ActionOption {
	return func(o *MobileAction) {
		o.Scope = []float64{x1, y1, x2, y2}
	}
}

func WithText(text string) ActionOption {
	return func(o *MobileAction) {
		o.Text = text
	}
}

func WithID(id string) ActionOption {
	return func(o *MobileAction) {
		o.ID = id
	}
}

func WithDescription(description string) ActionOption {
	return func(o *MobileAction) {
		o.Description = description
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
	UUID            string // ios udid or android serial
	Driver          WebDriver
	windowSize      Size
	frame           *bytes.Buffer
	doneMjpegStream chan bool
	scale           float64
	StartTime       time.Time     // used to associate screenshots name
	ScreenShots     []string      // save screenshots path
	perfStop        chan struct{} // stop performance monitor
	perfData        []string      // save perf data

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

func (dExt *DriverExt) GetPerfData() []string {
	if dExt.perfStop == nil {
		return nil
	}
	close(dExt.perfStop)
	return dExt.perfData
}

func (dExt *DriverExt) takeScreenShot() (raw *bytes.Buffer, err error) {
	// wait for action done
	time.Sleep(500 * time.Millisecond)

	// iOS 优先使用 MJPEG 流进行截图，性能最优
	// 如果 MJPEG 流未开启，则使用 WebDriver 的截图接口
	if dExt.frame != nil {
		return dExt.frame, nil
	}
	if raw, err = dExt.Driver.Screenshot(); err != nil {
		log.Error().Err(err).Msg("takeScreenShot failed")
		return nil, err
	}
	return raw, nil
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
		return "", errors.Wrap(err, "screenshot failed")
	}

	path, err := dExt.saveScreenShot(raw, fileName)
	if err != nil {
		return "", errors.Wrap(err, "save screenshot failed")
	}
	return path, nil
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
	} else if strings.HasPrefix(param, "com.") {
		// name
		selector = BySelector{
			ResourceIdID: param,
		}
	} else {
		// name
		selector = BySelector{
			LinkText: NewElementAttribute().WithName(param),
		}
	}

	return dExt.Driver.FindElement(selector)
}

func (dExt *DriverExt) FindUIRectInUIKit(search string, index ...int) (x, y, width, height float64, err error) {
	// click on text, using OCR
	if !isPathExists(search) {
		return dExt.FindTextByOCR(search, WithCustomOption("index", index))
	}
	// click on image, using opencv
	return dExt.FindImageRectInUIKit(search, index...)
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
	log.Info().Str("method", string(action.Method)).Interface("params", action.Params).Msg("start UI action")

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
			if len(action.Scope) != 4 {
				action.Scope = []float64{0, 0, 1, 1}
			}

			var options []DataOption
			options = append(options,
				WithCustomOption("index", []int{action.Index}),
				WithCustomOption("scope", []int{
					int(action.Scope[0] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[1] * float64(dExt.windowSize.Height) * dExt.scale),
					int(action.Scope[2] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[3] * float64(dExt.windowSize.Height) * dExt.scale),
				}),
			)

			var point PointF
			findApp := func(d *DriverExt) error {
				var err error
				point, err = d.GetTextXY(appName, options...)
				return err
			}
			foundAppAction := func(d *DriverExt) error {
				// click app to launch
				return d.TapAbsXY(point.X, point.Y-25, action.Identifier)
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
			if len(action.Scope) != 4 {
				action.Scope = []float64{0, 0, 1, 1}
			}

			var options []DataOption
			options = append(options,
				WithCustomOption("index", []int{action.Index}),
				WithCustomOption("scope", []int{
					int(action.Scope[0] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[1] * float64(dExt.windowSize.Height) * dExt.scale),
					int(action.Scope[2] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[3] * float64(dExt.windowSize.Height) * dExt.scale),
				}),
			)

			var point PointF
			findText := func(d *DriverExt) error {
				var err error
				point, err = d.GetTextXY(text, options...)
				return err
			}
			foundTextAction := func(d *DriverExt) error {
				// tap text
				return d.TapAbsXY(point.X, point.Y, action.Identifier)
			}

			// default to retry 10 times
			if action.MaxRetryTimes == 0 {
				action.MaxRetryTimes = 10
			}

			if action.Direction != nil {
				return dExt.SwipeUntil(action.Direction, findText, foundTextAction, action.MaxRetryTimes)
			}
			// swipe until found
			return dExt.SwipeUntil("up", findText, foundTextAction, action.MaxRetryTimes)
		}
		return fmt.Errorf("invalid %s params, should be app text(string), got %v",
			ACTION_SwipeToTapText, action.Params)
	case ACTION_SwipeToTapTexts:
		if texts, ok := action.Params.([]interface{}); ok {
			var textList []string
			for _, t := range texts {
				textList = append(textList, t.(string))
			}
			action.Params = textList
		}
		if texts, ok := action.Params.([]string); ok {
			if len(action.Scope) != 4 {
				action.Scope = []float64{0, 0, 1, 1}
			}

			var options []DataOption
			options = append(options,
				WithCustomOption("scope", []int{
					int(action.Scope[0] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[1] * float64(dExt.windowSize.Height) * dExt.scale),
					int(action.Scope[2] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[3] * float64(dExt.windowSize.Height) * dExt.scale),
				}),
			)

			var point PointF
			findText := func(d *DriverExt) error {
				var err error
				points, err := d.GetTextXYs(texts, options...)
				if err != nil {
					return err
				}
				for _, point = range points {
					if point != (PointF{X: 0, Y: 0}) {
						return nil
					}
				}
				return errors.New("failed to find text position")
			}
			foundTextAction := func(d *DriverExt) error {
				// tap text
				return d.TapAbsXY(point.X, point.Y, action.Identifier)
			}

			// default to retry 10 times
			if action.MaxRetryTimes == 0 {
				action.MaxRetryTimes = 10
			}

			if action.Direction != nil {
				return dExt.SwipeUntil(action.Direction, findText, foundTextAction, action.MaxRetryTimes)
			}
			// swipe until found
			return dExt.SwipeUntil("up", findText, foundTextAction, action.MaxRetryTimes)
		}
		return fmt.Errorf("invalid %s params, should be app text([]string), got %v",
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
		if location, ok := action.Params.([]interface{}); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			x, _ := location[0].(float64)
			y, _ := location[1].(float64)
			return dExt.TapXY(x, y, action.Identifier)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapXY, action.Params)
	case ACTION_TapAbsXY:
		if location, ok := action.Params.([]interface{}); ok {
			// absolute coordinates x,y of window size: [100, 300]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			x, _ := location[0].(float64)
			y, _ := location[1].(float64)
			return dExt.TapAbsXY(x, y, action.Identifier)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapAbsXY, action.Params)
	case ACTION_Tap:
		if param, ok := action.Params.(string); ok {
			return dExt.Tap(param, action.Identifier, action.IgnoreNotFoundError, action.Index)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_Tap, action.Params)
	case ACTION_TapByOCR:
		if ocrText, ok := action.Params.(string); ok {
			if len(action.Scope) != 4 {
				action.Scope = []float64{0, 0, 1, 1}
			}

			var options []DataOption
			options = append(options,
				WithCustomOption("index", []int{action.Index}),
				WithCustomOption("scope", []int{
					int(action.Scope[0] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[1] * float64(dExt.windowSize.Height) * dExt.scale),
					int(action.Scope[2] * float64(dExt.windowSize.Width) * dExt.scale),
					int(action.Scope[3] * float64(dExt.windowSize.Height) * dExt.scale),
				}),
			)
			return dExt.TapByOCR(ocrText, action.Identifier, action.IgnoreNotFoundError, options...)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByOCR, action.Params)
	case ACTION_TapByCV:
		if imagePath, ok := action.Params.(string); ok {
			return dExt.TapByCV(imagePath, action.Identifier, action.IgnoreNotFoundError, action.Index)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_TapByCV, action.Params)
	case ACTION_DoubleTapXY:
		if location, ok := action.Params.([]interface{}); ok {
			// relative x,y of window size: [0.5, 0.5]
			if len(location) != 2 {
				return fmt.Errorf("invalid tap location params: %v", location)
			}
			x, _ := location[0].(float64)
			y, _ := location[1].(float64)
			return dExt.DoubleTapXY(x, y)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTapXY, action.Params)
	case ACTION_DoubleTap:
		if param, ok := action.Params.(string); ok {
			return dExt.DoubleTap(param)
		}
		return fmt.Errorf("invalid %s params: %v", ACTION_DoubleTap, action.Params)
	case ACTION_Swipe:
		if positions, ok := action.Params.([]interface{}); ok {
			// relative fromX, fromY, toX, toY of window size: [0.5, 0.9, 0.5, 0.1]
			if len(positions) != 4 {
				return fmt.Errorf("invalid swipe params [fromX, fromY, toX, toY]: %v", positions)
			}
			fromX, _ := positions[0].(float64)
			fromY, _ := positions[1].(float64)
			toX, _ := positions[2].(float64)
			toY, _ := positions[3].(float64)
			return dExt.SwipeRelative(fromX, fromY, toX, toY, action.Identifier)
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
		options := []DataOption{}
		if action.Text != "" {
			options = append(options, WithCustomOption("textview", action.Text))
		}
		if action.ID != "" {
			options = append(options, WithCustomOption("id", action.ID))
		}
		if action.Description != "" {
			options = append(options, WithCustomOption("description", action.Description))
		}
		if action.Identifier != "" {
			options = append(options, WithCustomOption("log", map[string]interface{}{
				"enable": true,
				"data":   action.Identifier,
			}))
		}
		return dExt.Driver.Input(param, options...)
	case CtlSleep:
		if param, ok := action.Params.(json.Number); ok {
			seconds, _ := param.Float64()
			time.Sleep(time.Duration(seconds*1000) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(float64); ok {
			time.Sleep(time.Duration(param*1000) * time.Millisecond)
			return nil
		} else if param, ok := action.Params.(int64); ok {
			time.Sleep(time.Duration(param) * time.Second)
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
		return dExt.Driver.StartCamera()
	case CtlStopCamera:
		return dExt.Driver.StopCamera()
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

func checkErr(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg, err)
		}
	}
}
