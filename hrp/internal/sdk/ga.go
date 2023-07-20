package sdk

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

const (
	gaAPIDebugURL = "https://www.google-analytics.com/debug/collect" // used for debug
	gaAPIURL      = "https://www.google-analytics.com/collect"
	trackingID    = "UA-114587036-1" // Tracking ID for Google Analytics
)

var gaClient *GAClient

func init() {
	// init GA client
	gaClient = NewGAClient(trackingID, userID)
}

func SendEvent(e IEvent) error {
	if env.DISABLE_GA == "true" {
		// do not send GA events in CI environment
		return nil
	}
	return gaClient.SendEvent(e)
}

type GAClient struct {
	TrackingID string       `form:"tid"` // Tracking ID / Property ID, XX-XXXXXXX-X
	ClientID   string       `form:"cid"` // Anonymous Client ID
	Version    string       `form:"v"`   // Version
	httpClient *http.Client // http client session
}

// NewGAClient creates a new GAClient object with the trackingID and clientID.
func NewGAClient(trackingID, clientID string) *GAClient {
	return &GAClient{
		TrackingID: trackingID,
		ClientID:   clientID,
		Version:    "1", // constant v1
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// SendEvent sends one event to Google Analytics
func (g *GAClient) SendEvent(e IEvent) error {
	var data url.Values
	if event, ok := e.(UserTimingTracking); ok {
		event.duration = time.Since(event.startTime)
		data = event.ToUrlValues()
	} else {
		data = e.ToUrlValues()
	}

	// append common params
	data.Add("v", g.Version)
	data.Add("tid", g.TrackingID)
	data.Add("cid", g.ClientID)

	resp, err := g.httpClient.PostForm(gaAPIURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("response status: %d", resp.StatusCode)
	}
	return nil
}

func structToUrlValues(i interface{}) (values url.Values) {
	values = url.Values{}
	iVal := reflect.ValueOf(i)
	for i := 0; i < iVal.NumField(); i++ {
		formTagName := iVal.Type().Field(i).Tag.Get("form")
		if formTagName == "" {
			continue
		}
		if iVal.Field(i).IsZero() {
			continue
		}
		values.Set(formTagName, fmt.Sprint(iVal.Field(i)))
	}
	return
}

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
