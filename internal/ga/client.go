package ga

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

const (
	gaAPIDebugURL = "https://www.google-analytics.com/debug/collect" // used for debug
	gaAPIURL      = "https://www.google-analytics.com/collect"
)

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
