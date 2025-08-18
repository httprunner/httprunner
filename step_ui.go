package hrp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type MobileUI struct {
	OSType              string `json:"os_type,omitempty" yaml:"os_type,omitempty"` // mobile device os type
	Serial              string `json:"serial,omitempty" yaml:"serial,omitempty"`   // mobile device serial number
	option.MobileAction `yaml:",inline"`
	Actions             []option.MobileAction `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// StepMobile implements IStep interface.
type StepMobile struct {
	StepConfig
	Mobile  *MobileUI `json:"mobile,omitempty" yaml:"mobile,omitempty"`
	Android *MobileUI `json:"android,omitempty" yaml:"android,omitempty"`
	Harmony *MobileUI `json:"harmony,omitempty" yaml:"harmony,omitempty"`
	IOS     *MobileUI `json:"ios,omitempty" yaml:"ios,omitempty"`
	Browser *MobileUI `json:"browser,omitempty" yaml:"browser,omitempty"`
	cache   *MobileUI // used for caching
}

// uniform interface for all types of mobile systems
func (s *StepMobile) obj() *MobileUI {
	if s.cache != nil {
		return s.cache
	}

	if s.IOS != nil {
		s.cache = s.IOS
		s.cache.OSType = string(StepTypeIOS)
		return s.cache
	} else if s.Harmony != nil {
		s.cache = s.Harmony
		s.cache.OSType = string(StepTypeHarmony)
		return s.cache
	} else if s.Android != nil {
		s.cache = s.Android
		s.cache.OSType = string(StepTypeAndroid)
		return s.cache
	} else if s.Browser != nil {
		s.cache = s.Browser
		s.cache.OSType = string(stepTypeBrowser)
		return s.cache
	} else if s.Mobile != nil {
		s.cache = s.Mobile
		return s.cache
	}

	panic("no mobile device config")
}

func (s *StepMobile) Serial(serial string) *StepMobile {
	s.obj().Serial = serial
	return s
}

func (s *StepMobile) Log(actionName option.ActionName) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method: option.ACTION_LOG,
		Params: actionName,
	})
	return s
}

func (s *StepMobile) InstallApp(path string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method: option.ACTION_AppInstall,
		Params: path,
	})
	return s
}

func (s *StepMobile) WebLoginNoneUI(packageName, phoneNumber string, captcha, password string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method: option.ACTION_WebLoginNoneUI,
		Params: []string{packageName, phoneNumber, captcha, password},
	})
	return s
}

func (s *StepMobile) AppLaunch(bundleId string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method: option.ACTION_AppLaunch,
		Params: bundleId,
	})
	return s
}

func (s *StepMobile) AppTerminate(bundleId string) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method: option.ACTION_AppTerminate,
		Params: bundleId,
	})
	return s
}

func (s *StepMobile) Home() *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method: option.ACTION_Home,
		Params: nil,
	})
	return s
}

// TapXY taps the point {X,Y}
// if X<1 & Y<1, {X,Y} will be considered as percentage
// else, X & Y will be considered as absolute coordinates
func (s *StepMobile) TapXY(x, y float64, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_TapXY,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapAbsXY taps the point {X,Y}, X & Y is absolute coordinates
func (s *StepMobile) TapAbsXY(x, y float64, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_TapAbsXY,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByOCR taps on the target element by OCR recognition
func (s *StepMobile) TapByOCR(ocrText string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_TapByOCR,
		Params:  ocrText,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByCV taps on the target element by CV recognition
func (s *StepMobile) TapByCV(imagePath string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_TapByCV,
		Params:  imagePath,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// TapByUITypes taps on the target element specified by uiTypes, the higher the uiTypes, the higher the priority
func (s *StepMobile) TapByUITypes(opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_TapByCV,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// StartToGoal do goal-oriented actions with VLM
func (s *StepMobile) StartToGoal(prompt string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_StartToGoal,
		Params:  prompt,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// AIAction do actions with VLM
func (s *StepMobile) AIAction(prompt string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_AIAction,
		Params:  prompt,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// AIQuery query information from screen using VLM
func (s *StepMobile) AIQuery(prompt string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_Query,
		Params:  prompt,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// DoubleTapXY double taps the point {X,Y}, X & Y is percentage of coordinates
func (s *StepMobile) DoubleTapXY(x, y float64, opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method:  option.ACTION_DoubleTapXY,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(opts...),
	})
	return s
}

func (s *StepMobile) Back() *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_Back,
		Params:  nil,
		Options: nil,
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Swipe drags from [sx, sy] to [ex, ey]
func (s *StepMobile) Swipe(sx, sy, ex, ey float64, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeCoordinate,
		Params:  []float64{sx, sy, ex, ey},
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeUp(opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "up",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeDown(opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "down",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeLeft(opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "left",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeRight(opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeDirection,
		Params:  "right",
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SIMSwipeWithDirection performs simulated swipe in specified direction with random distance
func (s *StepMobile) SIMSwipeWithDirection(direction string, fromX, fromY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) *StepMobile {
	// Create params map for SIMSwipeWithDirection
	params := map[string]interface{}{
		"direction":        direction,
		"from_x":           fromX,
		"from_y":           fromY,
		"sim_min_distance": simMinDistance,
		"sim_max_distance": simMaxDistance,
	}

	action := option.MobileAction{
		Method:  option.ACTION_SIMSwipeDirection,
		Params:  params,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SIMSwipeInArea performs simulated swipe in specified area with direction and random distance
func (s *StepMobile) SIMSwipeInArea(direction string, simAreaStartX, simAreaStartY, simAreaEndX, simAreaEndY, simMinDistance, simMaxDistance float64, opts ...option.ActionOption) *StepMobile {
	// Create params map for SIMSwipeInArea
	params := map[string]interface{}{
		"direction":        direction,
		"sim_area_start_x": simAreaStartX,
		"sim_area_start_y": simAreaStartY,
		"sim_area_end_x":   simAreaEndX,
		"sim_area_end_y":   simAreaEndY,
		"sim_min_distance": simMinDistance,
		"sim_max_distance": simMaxDistance,
	}

	action := option.MobileAction{
		Method:  option.ACTION_SIMSwipeInArea,
		Params:  params,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SIMSwipeFromPointToPoint performs simulated swipe from point to point
func (s *StepMobile) SIMSwipeFromPointToPoint(fromX, fromY, toX, toY float64, opts ...option.ActionOption) *StepMobile {
	// Create params map for SIMSwipeFromPointToPoint
	params := map[string]interface{}{
		"from_x": fromX,
		"from_y": fromY,
		"to_x":   toX,
		"to_y":   toY,
	}

	action := option.MobileAction{
		Method:  option.ACTION_SIMSwipeFromPointToPoint,
		Params:  params,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SIMClickAtPoint performs simulated click at specified point
func (s *StepMobile) SIMClickAtPoint(x, y float64, opts ...option.ActionOption) *StepMobile {
	// Create params map for SIMClickAtPoint
	params := map[string]interface{}{
		"x": x,
		"y": y,
	}

	action := option.MobileAction{
		Method:  option.ACTION_SIMClickAtPoint,
		Params:  params,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SIMInput performs simulated text input with intelligent segmentation
func (s *StepMobile) SIMInput(text string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SIMInput,
		Params:  text,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapApp(appName string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeToTapApp,
		Params:  appName,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapText(text string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeToTapText,
		Params:  text,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SwipeToTapTexts(texts interface{}, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SwipeToTapTexts,
		Params:  texts,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SecondaryClick(x, y float64, options ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SecondaryClick,
		Params:  []float64{x, y},
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SecondaryClickBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SecondaryClickBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) HoverBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_HoverBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) TapBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_TapBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) WebCloseTab(idx int, options ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_WebCloseTab,
		Params:  idx,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) GetElementTextBySelector(selector string, options ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_GetElementTextBySelector,
		Params:  selector,
		Options: option.NewActionOptions(options...),
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) Input(text string, opts ...option.ActionOption) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_Input,
		Params:  text,
		Options: option.NewActionOptions(opts...),
	}

	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// Sleep specify sleep seconds after last action
func (s *StepMobile) Sleep(nSeconds float64, startTime ...time.Time) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_Sleep,
		Params:  nSeconds,
		Options: nil,
	}
	if len(startTime) > 0 {
		action.Params = uixt.SleepConfig{
			StartTime: startTime[0],
			Seconds:   nSeconds,
		}
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

func (s *StepMobile) SleepMS(nMilliseconds int64, startTime ...time.Time) *StepMobile {
	action := option.MobileAction{
		Method:  option.ACTION_SleepMS,
		Params:  nMilliseconds,
		Options: nil,
	}
	if len(startTime) > 0 {
		action.Params = uixt.SleepConfig{
			StartTime:    startTime[0],
			Milliseconds: nMilliseconds,
		}
	}
	s.obj().Actions = append(s.obj().Actions, action)
	return s
}

// SleepRandom specify random sleeping seconds after last action
// params have two different kinds:
// 1. [min, max] : min and max are float64 time range boundaries
// 2. [min1, max1, weight1, min2, max2, weight2, ...] : weight is the probability of the time range
func (s *StepMobile) SleepRandom(params ...float64) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method:  option.ACTION_SleepRandom,
		Params:  params,
		Options: nil,
	})
	return s
}

func (s *StepMobile) EndToEndDelay(opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method:  option.ACTION_EndToEndDelay,
		Params:  nil,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

func (s *StepMobile) ScreenShot(opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method:  option.ACTION_ScreenShot,
		Params:  nil,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

func (s *StepMobile) ClosePopups(opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method:  option.ACTION_ClosePopups,
		Params:  nil,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

// EnableAutoPopupHandler enables auto popup handler for this step.
func (s *StepMobile) EnableAutoPopupHandler() *StepMobile {
	s.AutoPopupHandler = true
	return s
}

func (s *StepMobile) Call(name string, fn func(), opts ...option.ActionOption) *StepMobile {
	s.obj().Actions = append(s.obj().Actions, option.MobileAction{
		Method:  option.ACTION_CallFunction,
		Params:  name, // function description
		Fn:      fn,
		Options: option.NewActionOptions(opts...),
	})
	return s
}

// Validate switches to step validation.
func (s *StepMobile) Validate() *StepMobileUIValidation {
	return &StepMobileUIValidation{
		StepMobile: s,
		Validators: make([]interface{}, 0),
	}
}

func (s *StepMobile) Name() string {
	return s.StepName
}

func (s *StepMobile) Type() StepType {
	osType := s.obj().OSType
	if osType != "" {
		return StepType(osType)
	}
	return StepType("mobile")
}

func (s *StepMobile) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepMobile) Run(r *SessionRunner) (*StepResult, error) {
	return runStepMobileUI(r, s)
}

// StepMobileUIValidation implements IStep interface.
type StepMobileUIValidation struct {
	*StepMobile
	Validators []interface{} `json:"validate,omitempty" yaml:"validate,omitempty"`
}

func (s *StepMobileUIValidation) AssertNameExists(expectedName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorName,
		Assert: option.AssertionExists,
		Expect: expectedName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute name [%s] not found", expectedName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertNameNotExists(expectedName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorName,
		Assert: option.AssertionNotExists,
		Expect: expectedName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute name [%s] should not exist", expectedName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertLabelExists(expectedLabel string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorLabel,
		Assert: option.AssertionExists,
		Expect: expectedLabel,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute label [%s] not found", expectedLabel)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertLabelNotExists(expectedLabel string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorLabel,
		Assert: option.AssertionNotExists,
		Expect: expectedLabel,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("attribute label [%s] should not exist", expectedLabel)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertOCRExists(expectedText string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorOCR,
		Assert: option.AssertionExists,
		Expect: expectedText,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("ocr text [%s] not found", expectedText)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertOCRNotExists(expectedText string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorOCR,
		Assert: option.AssertionNotExists,
		Expect: expectedText,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("ocr text [%s] should not exist", expectedText)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertImageExists(expectedImagePath string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorImage,
		Assert: option.AssertionExists,
		Expect: expectedImagePath,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("cv image [%s] not found", expectedImagePath)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertImageNotExists(expectedImagePath string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorImage,
		Assert: option.AssertionNotExists,
		Expect: expectedImagePath,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("cv image [%s] should not exist", expectedImagePath)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertAI(prompt string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorAI,
		Assert: option.AssertionAI,
		Expect: prompt,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("assert ai prompt [%s] failed", prompt)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertAppInForeground(packageName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorForegroundApp,
		Assert: option.AssertionEqual,
		Expect: packageName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("app [%s] should be in foreground", packageName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) AssertAppNotInForeground(packageName string, msg ...string) *StepMobileUIValidation {
	v := Validator{
		Check:  option.SelectorForegroundApp,
		Assert: option.AssertionNotEqual,
		Expect: packageName,
	}
	if len(msg) > 0 {
		v.Message = msg[0]
	} else {
		v.Message = fmt.Sprintf("app [%s] should not be in foreground", packageName)
	}
	s.Validators = append(s.Validators, v)
	return s
}

func (s *StepMobileUIValidation) Name() string {
	return s.StepName
}

func (s *StepMobileUIValidation) Type() StepType {
	return s.StepMobile.Type() + stepTypeSuffixValidation
}

func (s *StepMobileUIValidation) Config() *StepConfig {
	// Get the original StepConfig from embedded StepMobile
	config := &s.StepMobile.StepConfig
	// Sync validators to the StepConfig
	config.Validators = s.Validators
	return config
}

func (s *StepMobileUIValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepMobileUI(r, s)
}

func runStepMobileUI(s *SessionRunner, step IStep) (stepResult *StepResult, err error) {
	start := time.Now()
	stepResult = &StepResult{
		Name:        step.Name(),
		StepType:    step.Type(),
		Success:     false,
		ContentSize: 0,
		StartTime:   start.UnixMilli(),
	}

	var stepVariables map[string]interface{}
	var stepValidators []interface{}
	var stepAutoPopupHandler bool
	var stepIgnorePopup bool

	var mobileStep *MobileUI
	switch stepMobile := step.(type) {
	case *StepMobile:
		mobileStep = stepMobile.obj()
		stepVariables = stepMobile.Variables
		stepAutoPopupHandler = stepMobile.AutoPopupHandler
		stepIgnorePopup = stepMobile.IgnorePopup
	case *StepMobileUIValidation:
		mobileStep = stepMobile.obj()
		stepVariables = stepMobile.Variables
		stepValidators = stepMobile.Validators
		stepAutoPopupHandler = stepMobile.StepMobile.AutoPopupHandler
		stepIgnorePopup = stepMobile.StepMobile.IgnorePopup
	default:
		return stepResult, errors.New("invalid mobile UI step type")
	}

	// report GA event
	go sdk.SendGA4Event("hrp_run_ui", map[string]interface{}{
		"osType": mobileStep.OSType,
	})

	// init wda/uia/hdc driver
	config := uixt.DriverCacheConfig{
		Platform: mobileStep.OSType,
		Serial:   mobileStep.Serial,
	}

	// Extract AI service options from global configuration
	if s.caseRunner != nil && s.caseRunner.Config != nil {
		globalConfig := s.caseRunner.Config.Get()
		if globalConfig != nil && globalConfig.AIOptions != nil {
			config.AIOptions = globalConfig.AIOptions.Options()
		}
	}

	uiDriver, err := uixt.GetOrCreateXTDriver(config)
	if err != nil {
		return stepResult, err
	}

	defer func() {
		attachments := uixt.Attachments{}
		if err != nil {
			attachments["error"] = err.Error()

			// save foreground app
			startTime := time.Now()
			actionResult := &ActionResult{
				MobileAction: option.MobileAction{
					Method: option.ACTION_GetForegroundApp,
					Params: "[ForDebug] check foreground app",
				},
				StartTime: startTime.UnixMilli(),
			}
			sessionData, err1 := uiDriver.ExecuteAction(
				context.Background(), actionResult.MobileAction)
			if err1 != nil {
				actionResult.Error = err1.Error()
				log.Warn().Err(err1).Msg("get foreground app failed, ignore")
			}
			actionResult.Elapsed = time.Since(startTime).Milliseconds()
			actionResult.SessionData = sessionData
			stepResult.Actions = append(stepResult.Actions, actionResult)
		}

		// Get session data and add to attachments, clear session for next step
		if uiDriver != nil {
			sessionData := uiDriver.GetSession().GetData(true) // clear session after getting data
			if len(sessionData.ScreenResults) > 0 {
				attachments["screen_results"] = sessionData.ScreenResults
				log.Debug().Int("count", len(sessionData.ScreenResults)).
					Str("step", step.Name()).Msg("added screen results to step attachments")
			}
		}

		var config *TConfig
		if s.caseRunner != nil && s.caseRunner.Config != nil {
			config = s.caseRunner.Config.Get()
		}
		// automatic handling of pop-up windows on each step finished, default to disabled
		// priority: config ignore_popup > step ignore_popup > config auto_popup_handler > step auto_popup_handler
		shouldHandlePopup := false
		if s.ignorePopup(mobileStep.OSType) {
			shouldHandlePopup = false
		} else if stepIgnorePopup {
			// step level config, keep for compatibility
			shouldHandlePopup = false
		} else if config != nil && config.AutoPopupHandler {
			// testcase level config has higher priority
			shouldHandlePopup = true
		} else if stepAutoPopupHandler {
			// step level config
			shouldHandlePopup = true
		}

		if shouldHandlePopup && uiDriver != nil {
			startTime := time.Now()
			actionResult := &ActionResult{
				MobileAction: option.MobileAction{
					Method: option.ACTION_ClosePopups,
					Params: "[ForDebug] close popups handler",
				},
				StartTime: startTime.UnixMilli(),
			}
			sessionData, err2 := uiDriver.ExecuteAction(
				context.Background(), actionResult.MobileAction)
			if err2 != nil {
				actionResult.Error = err2.Error()
				log.Warn().Err(err2).Str("step", step.Name()).Msg("auto handle popup failed")
			}
			actionResult.Elapsed = time.Since(startTime).Milliseconds()
			actionResult.SessionData = sessionData
			stepResult.Actions = append(stepResult.Actions, actionResult)
		}

		stepResult.Attachments = attachments
		stepResult.Elapsed = time.Since(start).Milliseconds()
	}()

	// run actions
	for _, action := range mobileStep.Actions {
		select {
		case <-s.caseRunner.hrpRunner.caseTimeoutTimer.C:
			log.Warn().Msg("case timeout in mobile UI runner, abort running")
			return stepResult, errors.Wrap(code.TimeoutError, "mobile UI runner case timeout")
		case <-s.caseRunner.hrpRunner.interruptSignal:
			log.Warn().Msg("interrupted in mobile UI runner, abort running")
			return stepResult, errors.Wrap(code.InterruptError, "mobile UI runner interrupted")
		default:
			actionStartTime := time.Now()
			// Parse action params first for variable substitution
			if action.Params, err = s.caseRunner.parser.Parse(action.Params, stepVariables); err != nil {
				if !code.IsErrorPredefined(err) {
					err = errors.Wrap(code.ParseError,
						fmt.Sprintf("parse action params failed: %v", err))
				}
				return stepResult, err
			}

			// Create ActionResult with parsed params for accurate reporting
			actionResult := &ActionResult{
				MobileAction: action,                      // Now contains parsed params
				StartTime:    actionStartTime.UnixMilli(), // action start time
			}

			// Apply global configuration from testcase config
			if s.caseRunner != nil && s.caseRunner.Config != nil {
				config := s.caseRunner.Config.Get()
				if config != nil {
					if action.Options == nil {
						action.Options = &option.ActionOptions{}
					}

					// Apply global AntiRisk configuration
					if config.AntiRisk && !action.Options.AntiRisk {
						action.Options.AntiRisk = true
					}

					// Apply global LLM service configuration for AI actions
					if config.AIOptions != nil && (action.Method == option.ACTION_AIAction || action.Method == option.ACTION_StartToGoal ||
						action.Method == option.ACTION_AIAssert || action.Method == option.ACTION_Query) {
						if config.AIOptions.LLMService != "" && action.Options.LLMService == "" {
							action.Options.LLMService = string(config.AIOptions.LLMService)
							log.Debug().Str("action", string(action.Method)).
								Str("llmService", action.Options.LLMService).
								Msg("Applied global LLM service config to action")
						}
						if config.AIOptions.CVService != "" && action.Options.CVService == "" {
							action.Options.CVService = string(config.AIOptions.CVService)
							log.Debug().Str("action", string(action.Method)).
								Str("cvService", action.Options.CVService).
								Msg("Applied global CV service config to action")
						}
					}
				}
			}

			// stat uixt action
			if action.Method == option.ACTION_LOG {
				log.Info().Interface("action", action.Params).Msg("stat uixt action")
				actionMethod := option.ActionName(action.Params.(string))
				s.summary.Stat.Actions[actionMethod]++
				continue
			}

			// call custom function
			if action.Method == option.ACTION_CallFunction {
				if funcDesc, ok := action.Params.(string); ok {
					err := Call(funcDesc, action.Fn, action.GetOptions()...)
					if err != nil {
						return stepResult, err
					}
				}
				continue
			}

			// call MCP tool to execute action with cancellable context
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)

			// Create a goroutine to monitor for interrupt signals and timeouts
			go func() {
				select {
				case <-s.caseRunner.hrpRunner.interruptSignal:
					log.Warn().Msg("cancelling action due to interrupt signal")
					cancel(code.InterruptError)
				case <-s.caseRunner.hrpRunner.caseTimeoutTimer.C:
					log.Warn().Msg("cancelling action due to case timeout")
					cancel(code.TimeoutError)
				case <-ctx.Done():
					// Context already cancelled
				}
			}()

			// handle start_to_goal AI action
			if action.Method == option.ACTION_StartToGoal {
				planningResults, err := uiDriver.StartToGoal(ctx,
					action.Params.(string), action.GetOptions()...)
				actionResult.Elapsed = time.Since(actionStartTime).Milliseconds()
				actionResult.Plannings = planningResults
				stepResult.Actions = append(stepResult.Actions, actionResult)
				if err != nil {
					actionResult.Error = err.Error()
					if !code.IsErrorPredefined(err) {
						err = errors.Wrap(code.MobileUIDriverError, err.Error())
					}
					return stepResult, err
				}
				continue
			}

			// handle AI operations (ai_action, ai_query, ai_assert) with unified result storage
			if action.Method == option.ACTION_AIAction || action.Method == option.ACTION_Query || action.Method == option.ACTION_AIAssert {
				var aiResult *uixt.AIExecutionResult
				var err error

				prompt := action.Params.(string)
				switch action.Method {
				case option.ACTION_AIAction:
					aiResult, err = uiDriver.AIAction(ctx, prompt, action.GetOptions()...)
				case option.ACTION_Query:
					aiResult, err = uiDriver.AIQuery(prompt, action.GetOptions()...)
				case option.ACTION_AIAssert:
					aiResult, err = uiDriver.AIAssert(prompt, action.GetOptions()...)
				}

				actionResult.Elapsed = time.Since(actionStartTime).Milliseconds()
				actionResult.AIResult = aiResult
				stepResult.Actions = append(stepResult.Actions, actionResult)
				if err != nil {
					actionResult.Error = err.Error()
					if !code.IsErrorPredefined(err) {
						err = errors.Wrap(code.MobileUIDriverError, err.Error())
					}
					return stepResult, err
				}
				continue
			}
			if action.Method == option.ACTION_GetPasteboard {
				content, err := uiDriver.GetPasteboard()
				if err != nil {
					actionResult.Error = err.Error()
					if !code.IsErrorPredefined(err) {
						err = errors.Wrap(code.MobileUIDriverError, err.Error())
					}
					return stepResult, err
				}
				actionResult.ExtraData = content
				stepResult.Actions = append(stepResult.Actions, actionResult)
				continue
			}

			// handle other non-AI actions
			sessionData, err := uiDriver.ExecuteAction(ctx, action)
			actionResult.Elapsed = time.Since(actionStartTime).Milliseconds()
			actionResult.SessionData = sessionData
			stepResult.Actions = append(stepResult.Actions, actionResult)
			if err != nil {
				actionResult.Error = err.Error()
				if !code.IsErrorPredefined(err) {
					err = errors.Wrap(code.MobileUIDriverError, err.Error())
				}
				return stepResult, err
			}
		}
	}

	// validate
	validateResults, err := validateUI(uiDriver, stepValidators, s.caseRunner.parser, stepVariables)
	if len(validateResults) > 0 {
		// Always save validation results if any exist, regardless of success or failure
		sessionData := &SessionData{
			Validators: validateResults,
		}
		stepResult.Data = sessionData
	}
	if err != nil {
		// Handle validation error after saving results
		if !code.IsErrorPredefined(err) {
			err = errors.Wrap(code.MobileUIValidationError, err.Error())
		}
		return stepResult, err
	}

	stepResult.Success = true
	return stepResult, nil
}

func validateUI(ud *uixt.XTDriver, iValidators []interface{}, parser *Parser, stepVariables map[string]interface{}) (validateResults []*ValidationResult, err error) {
	// Parse all validators for variable substitution
	parsedValidators, err := parseStepValidators(iValidators, parser, stepVariables)
	if err != nil {
		return nil, err
	}

	// Execute validation for each parsed validator
	for _, validator := range parsedValidators {
		// Debug: print validator details
		log.Debug().
			Str("check", validator.Check).
			Str("assert", validator.Assert).
			Interface("expect", validator.Expect).
			Str("message", validator.Message).
			Msg("processing validator")

		validationResult := &ValidationResult{
			Validator:   validator, // Use parsed validator for accurate reporting
			CheckResult: "fail",
		}

		// Check if this is a UI validator or AI assert validator
		if !strings.HasPrefix(validator.Check, "ui_") && validator.Assert != "ai_assert" {
			validationResult.CheckResult = "skip"
			log.Warn().Interface("validator", validator).Msg("skip validator")
			validateResults = append(validateResults, validationResult)
			continue
		}

		// Validate expected value type
		expected, ok := validator.Expect.(string)
		if !ok {
			return nil, errors.New("validator expect should be string")
		}

		// Perform validation
		validationResult.AIResult, err = ud.DoValidation(
			validator.Check, validator.Assert, expected, validator.Message)
		if err != nil {
			// Add the failed validation result to the list before returning error
			validateResults = append(validateResults, validationResult)
			return validateResults, errors.Wrap(err, "step validation failed")
		}

		validationResult.CheckResult = "pass"
		validateResults = append(validateResults, validationResult)
	}

	return validateResults, nil
}

// parseStepValidators parses all validators for variable substitution
func parseStepValidators(iValidators []interface{}, parser *Parser, stepVariables map[string]interface{}) ([]Validator, error) {
	var parsedValidators []Validator

	for _, iValidator := range iValidators {
		validator, ok := iValidator.(Validator)
		if !ok {
			return nil, errors.New("validator type error")
		}

		parsedValidator := validator

		// Parse Expect field for variable substitution
		if expectedStr, ok := validator.Expect.(string); ok {
			if parsedExpected, err := parser.Parse(expectedStr, stepVariables); err != nil {
				return nil, errors.Wrap(err, "failed to parse validator expect field")
			} else {
				parsedValidator.Expect = parsedExpected
			}
		}

		// Parse Message field for variable substitution
		if validator.Message != "" {
			if parsedMessage, err := parser.Parse(validator.Message, stepVariables); err != nil {
				return nil, errors.Wrap(err, "failed to parse validator message field")
			} else {
				if msgStr, ok := parsedMessage.(string); ok {
					parsedValidator.Message = msgStr
				}
			}
		}

		parsedValidators = append(parsedValidators, parsedValidator)
	}

	return parsedValidators, nil
}
