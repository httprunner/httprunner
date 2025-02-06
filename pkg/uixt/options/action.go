package options

import (
	"math/rand/v2"

	"github.com/httprunner/httprunner/v5/internal/builtin"
)

// (x1, y1) is the top left corner, (x2, y2) is the bottom right corner
// [x1, y1, x2, y2] in percentage of the screen
type Scope []float64

// [x1, y1, x2, y2] in absolute pixels
type AbsScope []int

func (s AbsScope) Option() ActionOption {
	return WithAbsScope(s[0], s[1], s[2], s[3])
}

type ActionOptions struct {
	// log
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"` // used to identify the action in log

	// control related
	MaxRetryTimes       int         `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"`           // max retry times
	IgnoreNotFoundError bool        `json:"ignore_NotFoundError,omitempty" yaml:"ignore_NotFoundError,omitempty"` // ignore error if target element not found
	Interval            float64     `json:"interval,omitempty" yaml:"interval,omitempty"`                         // interval between retries in seconds
	Duration            float64     `json:"duration,omitempty" yaml:"duration,omitempty"`                         // used to set duration of ios swipe action
	PressDuration       float64     `json:"press_duration,omitempty" yaml:"press_duration,omitempty"`             // used to set duration of ios swipe action
	Steps               int         `json:"steps,omitempty" yaml:"steps,omitempty"`                               // used to set steps of android swipe action
	Direction           interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`                       // used by swipe to tap text or app
	Timeout             int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                           // TODO: wait timeout in seconds for mobile action
	Frequency           int         `json:"frequency,omitempty" yaml:"frequency,omitempty"`

	// scope related
	Scope    Scope    `json:"scope,omitempty" yaml:"scope,omitempty"`
	AbsScope AbsScope `json:"abs_scope,omitempty" yaml:"abs_scope,omitempty"`

	Regex             bool  `json:"regex,omitempty" yaml:"regex,omitempty"`                             // use regex to match text
	Offset            []int `json:"offset,omitempty" yaml:"offset,omitempty"`                           // used to tap offset of point
	OffsetRandomRange []int `json:"offset_random_range,omitempty" yaml:"offset_random_range,omitempty"` // set random range [min, max] for tap/swipe points
	Index             int   `json:"index,omitempty" yaml:"index,omitempty"`                             // index of the target element
	MatchOne          bool  `json:"match_one,omitempty" yaml:"match_one,omitempty"`                     // match one of the targets if existed

	// set custiom options such as textview, id, description
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty"`

	// screenshot related
	ScreenShotWithOCR            bool     `json:"screenshot_with_ocr,omitempty" yaml:"screenshot_with_ocr,omitempty"`
	ScreenShotWithUpload         bool     `json:"screenshot_with_upload,omitempty" yaml:"screenshot_with_upload,omitempty"`
	ScreenShotWithLiveType       bool     `json:"screenshot_with_live_type,omitempty" yaml:"screenshot_with_live_type,omitempty"`
	ScreenShotWithLivePopularity bool     `json:"screenshot_with_live_popularity,omitempty" yaml:"screenshot_with_live_popularity,omitempty"`
	ScreenShotWithUITypes        []string `json:"screenshot_with_ui_types,omitempty" yaml:"screenshot_with_ui_types,omitempty"`
	ScreenShotWithClosePopups    bool     `json:"screenshot_with_close_popups,omitempty" yaml:"screenshot_with_close_popups,omitempty"`
	ScreenShotWithOCRCluster     string   `json:"screenshot_with_ocr_cluster,omitempty" yaml:"screenshot_with_ocr_cluster,omitempty"`
	ScreenShotFileName           string   `json:"screenshot_file_name,omitempty" yaml:"screenshot_file_name,omitempty"`
}

func (o *ActionOptions) Options() []ActionOption {
	options := make([]ActionOption, 0)

	if o == nil {
		return options
	}

	if o.Identifier != "" {
		options = append(options, WithIdentifier(o.Identifier))
	}

	if o.MaxRetryTimes != 0 {
		options = append(options, WithMaxRetryTimes(o.MaxRetryTimes))
	}
	if o.IgnoreNotFoundError {
		options = append(options, WithIgnoreNotFoundError(true))
	}
	if o.Interval != 0 {
		options = append(options, WithInterval(o.Interval))
	}
	if o.Duration != 0 {
		options = append(options, WithDuration(o.Duration))
	}
	if o.PressDuration != 0 {
		options = append(options, WithPressDuration(o.PressDuration))
	}
	if o.Steps != 0 {
		options = append(options, WithSteps(o.Steps))
	}

	switch v := o.Direction.(type) {
	case string:
		options = append(options, WithDirection(v))
	case []float64:
		options = append(options, WithCustomDirection(
			v[0], v[1],
			v[2], v[3],
		))
	case []interface{}:
		// loaded from json case
		// custom direction: [fromX, fromY, toX, toY]
		sx, _ := builtin.Interface2Float64(v[0])
		sy, _ := builtin.Interface2Float64(v[1])
		ex, _ := builtin.Interface2Float64(v[2])
		ey, _ := builtin.Interface2Float64(v[3])
		options = append(options, WithCustomDirection(
			sx, sy,
			ex, ey,
		))
	}

	if o.Timeout != 0 {
		options = append(options, WithTimeout(o.Timeout))
	}
	if o.Frequency != 0 {
		options = append(options, WithFrequency(o.Frequency))
	}
	if len(o.AbsScope) == 4 {
		options = append(options, WithAbsScope(
			o.AbsScope[0], o.AbsScope[1], o.AbsScope[2], o.AbsScope[3]))
	} else if len(o.Scope) == 4 {
		options = append(options, WithScope(
			o.Scope[0], o.Scope[1], o.Scope[2], o.Scope[3]))
	}
	if len(o.Offset) == 2 {
		// for tap [x,y] offset
		options = append(options, WithTapOffset(o.Offset[0], o.Offset[1]))
	} else if len(o.Offset) == 4 {
		// for swipe [fromX, fromY, toX, toY] offset
		options = append(options, WithSwipeOffset(
			o.Offset[0], o.Offset[1], o.Offset[2], o.Offset[3]))
	}
	if len(o.OffsetRandomRange) == 2 {
		options = append(options, WithOffsetRandomRange(
			o.OffsetRandomRange[0], o.OffsetRandomRange[1]))
	}

	if o.Regex {
		options = append(options, WithRegex(true))
	}
	if o.Index != 0 {
		options = append(options, WithIndex(o.Index))
	}
	if o.MatchOne {
		options = append(options, WithMatchOne(true))
	}

	// custom options
	if o.Custom != nil {
		for k, v := range o.Custom {
			options = append(options, WithCustomOption(k, v))
		}
	}

	// screenshot options
	if o.ScreenShotWithOCR {
		options = append(options, WithScreenShotOCR(true))
	}
	if o.ScreenShotWithUpload {
		options = append(options, WithScreenShotUpload(true))
	}
	if o.ScreenShotWithLiveType {
		options = append(options, WithScreenShotLiveType(true))
	}
	if o.ScreenShotWithLivePopularity {
		options = append(options, WithScreenShotLivePopularity(true))
	}
	if len(o.ScreenShotWithUITypes) > 0 {
		options = append(options, WithScreenShotUITypes(o.ScreenShotWithUITypes...))
	}
	if o.ScreenShotWithClosePopups {
		options = append(options, WithScreenShotClosePopups(true))
	}
	if o.ScreenShotWithOCRCluster != "" {
		options = append(options, WithScreenOCRCluster(o.ScreenShotWithOCRCluster))
	}
	if o.ScreenShotFileName != "" {
		options = append(options, WithScreenShotFileName(o.ScreenShotFileName))
	}

	return options
}

func (o *ActionOptions) ScreenshotActions() []string {
	actions := []string{}
	if o.ScreenShotWithUpload {
		actions = append(actions, "upload")
	}
	if o.ScreenShotWithOCR {
		actions = append(actions, "ocr")
	}
	if o.ScreenShotWithLiveType {
		actions = append(actions, "liveType")
	}
	if o.ScreenShotWithLivePopularity {
		actions = append(actions, "livePopularity")
	}
	// UI detection
	if len(o.ScreenShotWithUITypes) > 0 {
		actions = append(actions, "ui")
	}
	if o.ScreenShotWithClosePopups {
		actions = append(actions, "close")
	}
	return actions
}

func (o *ActionOptions) GetRandomOffset() float64 {
	if len(o.OffsetRandomRange) != 2 {
		// invalid offset random range, should be [min, max]
		return 0
	}

	minOffset := o.OffsetRandomRange[0]
	maxOffset := o.OffsetRandomRange[1]
	return float64(builtin.GetRandomNumber(minOffset, maxOffset)) + rand.Float64()
}

func (o *ActionOptions) UpdateData(data map[string]interface{}) {
	if o.Identifier != "" {
		data["log"] = map[string]interface{}{
			"enable": true,
			"data":   o.Identifier,
		}
	}

	if o.Steps > 0 {
		data["steps"] = o.Steps
	}
	if _, ok := data["steps"]; !ok {
		data["steps"] = 12 // default steps
	}

	if o.Duration > 0 {
		data["duration"] = o.Duration
	}
	if _, ok := data["duration"]; !ok {
		data["duration"] = 0 // default duration
	}

	if o.Frequency > 0 {
		data["frequency"] = o.Frequency
	}
	if _, ok := data["frequency"]; !ok {
		data["frequency"] = 10 // default frequency
	}

	if _, ok := data["replace"]; !ok {
		data["replace"] = true // default true
	}

	// custom options
	if o.Custom != nil {
		for k, v := range o.Custom {
			data[k] = v
		}
	}
}

func NewActionOptions(opts ...ActionOption) *ActionOptions {
	actionOptions := &ActionOptions{}
	for _, option := range opts {
		option(actionOptions)
	}
	return actionOptions
}

type ActionOption func(o *ActionOptions)

func WithCustomOption(key string, value interface{}) ActionOption {
	return func(o *ActionOptions) {
		if o.Custom == nil {
			o.Custom = make(map[string]interface{})
		}
		o.Custom[key] = value
	}
}

func WithIdentifier(identifier string) ActionOption {
	return func(o *ActionOptions) {
		o.Identifier = identifier
	}
}

func WithIndex(index int) ActionOption {
	return func(o *ActionOptions) {
		o.Index = index
	}
}

// set alias for compatibility
var WithWaitTime = WithInterval

func WithInterval(sec float64) ActionOption {
	return func(o *ActionOptions) {
		o.Interval = sec
	}
}

func WithDuration(duration float64) ActionOption {
	return func(o *ActionOptions) {
		o.Duration = duration
	}
}

func WithPressDuration(pressDuration float64) ActionOption {
	return func(o *ActionOptions) {
		o.PressDuration = pressDuration
	}
}

func WithSteps(steps int) ActionOption {
	return func(o *ActionOptions) {
		o.Steps = steps
	}
}

// WithDirection inputs direction (up, down, left, right)
func WithDirection(direction string) ActionOption {
	return func(o *ActionOptions) {
		o.Direction = direction
	}
}

// WithCustomDirection inputs sx, sy, ex, ey
func WithCustomDirection(sx, sy, ex, ey float64) ActionOption {
	return func(o *ActionOptions) {
		o.Direction = []float64{sx, sy, ex, ey}
	}
}

// WithScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all in [0, 1], which means the relative position of the screen
func WithScope(x1, y1, x2, y2 float64) ActionOption {
	return func(o *ActionOptions) {
		o.Scope = Scope{x1, y1, x2, y2}
	}
}

// WithAbsScope inputs area of [(x1,y1), (x2,y2)]
// x1, y1, x2, y2 are all absolute position of the screen
func WithAbsScope(x1, y1, x2, y2 int) ActionOption {
	return func(o *ActionOptions) {
		o.AbsScope = AbsScope{x1, y1, x2, y2}
	}
}

// Deprecated: use WithTapOffset instead
func WithOffset(offsetX, offsetY int) ActionOption {
	return func(o *ActionOptions) {
		o.Offset = []int{offsetX, offsetY}
	}
}

// tap [x, y] with offset [offsetX, offsetY]
var WithTapOffset = WithOffset

// swipe [fromX, fromY, toX, toY] with offset [offsetFromX, offsetFromY, offsetToX, offsetToY]
func WithSwipeOffset(offsetFromX, offsetFromY, offsetToX, offsetToY int) ActionOption {
	return func(o *ActionOptions) {
		o.Offset = []int{offsetFromX, offsetFromY, offsetToX, offsetToY}
	}
}

func WithOffsetRandomRange(min, max int) ActionOption {
	return func(o *ActionOptions) {
		o.OffsetRandomRange = []int{min, max}
	}
}

func WithRegex(regex bool) ActionOption {
	return func(o *ActionOptions) {
		o.Regex = regex
	}
}

func WithMatchOne(matchOne bool) ActionOption {
	return func(o *ActionOptions) {
		o.MatchOne = matchOne
	}
}

func WithFrequency(frequency int) ActionOption {
	return func(o *ActionOptions) {
		o.Frequency = frequency
	}
}

func WithMaxRetryTimes(maxRetryTimes int) ActionOption {
	return func(o *ActionOptions) {
		o.MaxRetryTimes = maxRetryTimes
	}
}

func WithTimeout(timeout int) ActionOption {
	return func(o *ActionOptions) {
		o.Timeout = timeout
	}
}

func WithIgnoreNotFoundError(ignoreError bool) ActionOption {
	return func(o *ActionOptions) {
		o.IgnoreNotFoundError = ignoreError
	}
}

func WithScreenShotOCR(ocrOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithOCR = ocrOn
	}
}

func WithScreenShotUpload(uploadOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithUpload = uploadOn
	}
}

func WithScreenShotLiveType(liveTypeOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithLiveType = liveTypeOn
	}
}

func WithScreenShotLivePopularity(livePopularityOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithLivePopularity = livePopularityOn
	}
}

func WithScreenShotUITypes(uiTypes ...string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithUITypes = uiTypes
	}
}

func WithScreenShotClosePopups(closeOn bool) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithClosePopups = closeOn
	}
}

func WithScreenOCRCluster(ocrCluster string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotWithOCRCluster = ocrCluster
	}
}

func WithScreenShotFileName(fileName string) ActionOption {
	return func(o *ActionOptions) {
		o.ScreenShotFileName = fileName
	}
}
