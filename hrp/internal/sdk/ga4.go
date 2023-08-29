package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"

	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

// Measurement Protocol (Google Analytics 4) docs reference:
// https://developers.google.com/analytics/devguides/collection/protocol/ga4
// debugging tools: https://ga-dev-tools.google/ga4/event-builder/
const (
	ga4APISecret     = "w7lKNQIrQsKNS4ikgMPp0Q"
	ga4MeasurementID = "G-9KHR3VC2LN"
)

var (
	ga4Client *GA4Client
	userID    string
)

func init() {
	var err error
	userID, err = machineid.ProtectedID("hrp")
	if err != nil {
		userID = uuid.NewV1().String()
	}

	// init GA4 client
	ga4Client = NewGA4Client(ga4MeasurementID, ga4APISecret, false)
}

type GA4Client struct {
	apiSecret     string       // Measurement Protocol API secret value
	measurementID string       // MEASUREMENT ID, G-XXXXXXXXXX
	userID        string       // A unique identifier for a user
	httpClient    *http.Client // http client session
	debug         bool         // send events for validation, used for debug
}

// NewGA4Client creates a new GA4Client object with the measurementID and apiSecret.
func NewGA4Client(measurementID, apiSecret string, debug ...bool) *GA4Client {
	dbg := false
	if len(debug) > 0 {
		dbg = debug[0]
	}

	return &GA4Client{
		measurementID: measurementID,
		apiSecret:     apiSecret,
		userID:        userID,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		debug: dbg,
	}
}

type Event struct {
	// Required. The name for the event.
	Name string `json:"name"`
	// Optional. The parameters for the event.
	// engagement_time_msec/session_id
	Params map[string]interface{} `json:"params,omitempty"`
}

// payload docs reference:
// https://developers.google.com/analytics/devguides/collection/protocol/ga4/reference?client_type=gtag
type Payload struct {
	// Required. Uniquely identifies a user instance of a web client
	ClientID string `json:"client_id"`
	// Optional. A unique identifier for a user
	UserID string `json:"user_id,omitempty"`
	// Optional. A Unix timestamp (in microseconds) for the time to associate with the event.
	// This should only be set to record events that happened in the past.
	// This value can be overridden via user_property or event timestamps.
	// Events can be backdated up to 3 calendar days based on the property's timezone.
	TimestampMicros int64 `json:"timestamp_micros,omitempty"`
	// Optional. The user properties for the measurement.
	UserProperties map[string]string `json:"user_properties,omitempty"`
	// Optional. Set to true to indicate these events should not be used for personalized ads.
	NonPersonalizedAds bool `json:"non_personalized_ads,omitempty"`
	// Required. An array of event items. Up to 25 events can be sent per request.
	Events []Event `json:"events"`
}

// validation docs reference:
// https://developers.google.com/analytics/devguides/collection/protocol/ga4/validating-events?client_type=gtag
type ValidationResponse struct {
	ValidationMessages []ValidationMessage `json:"validationMessages"` // An array of validation messages.
}

type ValidationMessage struct {
	FieldPath      string         `json:"fieldPath"`      // The path to the field that was invalid.
	Description    string         `json:"description"`    // A description of the error.
	ValidationCode ValidationCode `json:"validationCode"` // A ValidationCode that corresponds to the error.
}

type ValidationCode string

const (
	VALUE_INVALID         ValidationCode = "VALUE_INVALID"         // The value provided for a fieldPath was invalid.
	VALUE_REQUIRED        ValidationCode = "VALUE_REQUIRED"        // A required value for a fieldPath was not provided.
	NAME_INVALID          ValidationCode = "NAME_INVALID"          // The name provided was invalid.
	NAME_RESERVED         ValidationCode = "NAME_RESERVED"         // The name provided was one of the reserved names.
	VALUE_OUT_OF_BOUNDS   ValidationCode = "VALUE_OUT_OF_BOUNDS"   // The value provided was too large.
	EXCEEDED_MAX_ENTITIES ValidationCode = "EXCEEDED_MAX_ENTITIES" // There were too many parameters in the request.
	NAME_DUPLICATED       ValidationCode = "NAME_DUPLICATED"       // The same name was provided more than once in the request.
)

// SendEvent sends one event to Google Analytics
func (g *GA4Client) SendEvent(event Event) error {
	query := url.Values{}
	query.Add("api_secret", g.apiSecret)
	query.Add("measurement_id", g.measurementID)

	var uri string
	if g.debug {
		uri = fmt.Sprintf("https://www.google-analytics.com/debug/mp/collect?%s", query.Encode())
	} else {
		uri = fmt.Sprintf("https://www.google-analytics.com/mp/collect?%s", query.Encode())
	}

	// append event params
	if event.Params == nil {
		event.Params = map[string]interface{}{}
	}
	event.Params["os"] = runtime.GOOS
	event.Params["arch"] = runtime.GOARCH
	event.Params["go_version"] = runtime.Version()
	event.Params["hrp_version"] = version.VERSION

	payload := Payload{
		ClientID:        fmt.Sprintf("%d.%d", rand.Int31(), time.Now().Unix()),
		UserID:          g.userID,
		TimestampMicros: time.Now().UnixMicro(),
		Events:          []Event{event},
	}

	bs, err := json.Marshal(payload)
	if g.debug {
		log.Debug().
			Str("uri", uri).
			Interface("payload", payload).
			Msg("send GA4 event")
	}
	if err != nil {
		return errors.Wrap(err, "marshal GA4 request payload failed")
	}

	body := bytes.NewReader(bs)
	res, err := g.httpClient.Post(uri, "application/json", body)
	if err != nil {
		return errors.Wrap(err, "request GA4 failed")
	}

	if res.StatusCode >= 300 {
		return fmt.Errorf("validation response got unexpected status %d", res.StatusCode)
	}

	if !g.debug {
		return nil
	}

	bs, err = io.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "read GA4 response body failed")
	}

	validationResponse := ValidationResponse{}
	err = json.Unmarshal(bs, &validationResponse)
	if err != nil {
		return errors.Wrap(err, "unmarshal GA4 response body failed")
	}

	log.Debug().
		Int("statusCode", res.StatusCode).
		Interface("validationResponse", validationResponse).
		Msg("get GA4 validation response")
	return nil
}

func SendGA4Event(name string, params map[string]interface{}) {
	if env.DISABLE_GA == "true" {
		// do not send GA4 events in CI environment
		return
	}

	event := Event{
		Name:   name,
		Params: params,
	}
	err := ga4Client.SendEvent(event)
	if err != nil {
		log.Error().Err(err).Msg("send GA4 event failed")
	}
}
