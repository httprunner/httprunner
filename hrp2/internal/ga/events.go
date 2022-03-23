package ga

import (
	"fmt"
	"net/url"
	"time"

	"github.com/httprunner/hrp/internal/version"
)

type IEvent interface {
	ToUrlValues() url.Values
}

type EventTracking struct {
	HitType  string `form:"t"`  // Event hit type = event
	Category string `form:"ec"` // Required. Event Category.
	Action   string `form:"ea"` // Required. Event Action.
	Label    string `form:"el"` // Optional. Event label, used as version.
	Value    int    `form:"ev"` // Optional. Event value, must be non-negative integer
}

func (e EventTracking) StartTiming(variable string) UserTimingTracking {
	return UserTimingTracking{
		HitType:   "timing",
		Category:  e.Category,
		Variable:  variable,
		Label:     e.Label,
		startTime: time.Now(), // starts the timer
	}
}

func (e EventTracking) ToUrlValues() url.Values {
	e.HitType = "event"
	e.Label = version.VERSION
	return structToUrlValues(e)
}

type UserTimingTracking struct {
	HitType   string `form:"t"`   // Timing hit type
	Category  string `form:"utc"` // Required. user timing category. e.g. jsonLoader
	Variable  string `form:"utv"` // Required. timing variable. e.g. load
	Duration  string `form:"utt"` // Required. time took duration.
	Label     string `form:"utl"` // Optional. user timing label. e.g jQuery
	startTime time.Time
	duration  time.Duration // time took duration
}

func (e UserTimingTracking) ToUrlValues() url.Values {
	e.HitType = "timing"
	e.Label = version.VERSION
	e.Duration = fmt.Sprintf("%d", int64(e.duration.Seconds()*1000))
	return structToUrlValues(e)
}

type Exception struct {
	HitType     string `form:"t"`   // Hit Type = exception
	Description string `form:"exd"` // exception description. i.e. IOException
	IsFatal     string `form:"exf"` // if the exception was fatal
	isFatal     bool
}

func (e Exception) ToUrlValues() url.Values {
	e.HitType = "exception"
	if e.isFatal {
		e.IsFatal = "1"
	} else {
		e.IsFatal = "0"
	}
	return structToUrlValues(e)
}
