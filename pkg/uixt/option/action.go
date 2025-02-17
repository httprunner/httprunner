package option

import (
	"math/rand/v2"

	"github.com/httprunner/httprunner/v5/internal/builtin"
)

type ActionOptions struct {
	// log
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"` // used to identify the action in log

	// control related
	MaxRetryTimes int         `json:"max_retry_times,omitempty" yaml:"max_retry_times,omitempty"` // max retry times
	Interval      float64     `json:"interval,omitempty" yaml:"interval,omitempty"`               // interval between retries in seconds
	Duration      float64     `json:"duration,omitempty" yaml:"duration,omitempty"`               // used to set duration in seconds
	PressDuration float64     `json:"press_duration,omitempty" yaml:"press_duration,omitempty"`   // used to set press duration in seconds
	Steps         int         `json:"steps,omitempty" yaml:"steps,omitempty"`                     // used to set steps of action
	Direction     interface{} `json:"direction,omitempty" yaml:"direction,omitempty"`             // used by swipe to tap text or app
	Timeout       int         `json:"timeout,omitempty" yaml:"timeout,omitempty"`                 // TODO: wait timeout in seconds for mobile action
	Frequency     int         `json:"frequency,omitempty" yaml:"frequency,omitempty"`
	AbsCoordinate bool        `json:"abs,omitempty" yaml:"abs,omitempty"` // use absolute coordinate

	ScreenOptions

	// set custiom options such as textview, id, description
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty"`
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
	if o.AbsCoordinate {
		options = append(options, WithAbsoluteCoordinate(true))
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

	return options
}

func (o *ActionOptions) GetScreenOptions() []ActionOption {
	return o.ScreenOptions.Options()
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

func MergeOptions(data map[string]interface{}, opts ...ActionOption) {
	o := NewActionOptions(opts...)
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

	if o.PressDuration > 0 {
		data["pressDuration"] = o.PressDuration
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
	if actionOptions.MaxRetryTimes == 0 {
		actionOptions.MaxRetryTimes = 1
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

func WithAbsoluteCoordinate(abs bool) ActionOption {
	return func(o *ActionOptions) {
		o.AbsCoordinate = abs
	}
}

// WithCustomDirection inputs sx, sy, ex, ey
func WithCustomDirection(sx, sy, ex, ey float64) ActionOption {
	return func(o *ActionOptions) {
		o.Direction = []float64{sx, sy, ex, ey}
	}
}

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
