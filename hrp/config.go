package hrp

import (
	"reflect"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

// NewConfig returns a new constructed testcase config with specified testcase name.
func NewConfig(name string) *TConfig {
	return &TConfig{
		Name:      name,
		Environs:  make(map[string]string),
		Variables: make(map[string]interface{}),
	}
}

// TConfig represents config data structure for testcase.
// Each testcase should contain one config part.
type TConfig struct {
	Name              string                 `json:"name" yaml:"name"` // required
	Verify            bool                   `json:"verify,omitempty" yaml:"verify,omitempty"`
	BaseURL           string                 `json:"base_url,omitempty" yaml:"base_url,omitempty"`   // deprecated in v4.1, moved to env
	Headers           map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`     // public request headers
	Environs          map[string]string      `json:"environs,omitempty" yaml:"environs,omitempty"`   // environment variables
	Variables         map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"` // global variables
	Parameters        map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	ParametersSetting *TParamsConfig         `json:"parameters_setting,omitempty" yaml:"parameters_setting,omitempty"`
	ThinkTimeSetting  *ThinkTimeConfig       `json:"think_time,omitempty" yaml:"think_time,omitempty"`
	WebSocketSetting  *WebSocketConfig       `json:"websocket,omitempty" yaml:"websocket,omitempty"`
	IOS               []*uixt.IOSDevice      `json:"ios,omitempty" yaml:"ios,omitempty"`
	Android           []*uixt.AndroidDevice  `json:"android,omitempty" yaml:"android,omitempty"`
	Timeout           float64                `json:"timeout,omitempty" yaml:"timeout,omitempty"` // global timeout in seconds
	Export            []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Weight            int                    `json:"weight,omitempty" yaml:"weight,omitempty"`
	Path              string                 `json:"path,omitempty" yaml:"path,omitempty"`     // testcase file path
	PluginSetting     *PluginConfig          `json:"plugin,omitempty" yaml:"plugin,omitempty"` // plugin config
}

// WithVariables sets variables for current testcase.
func (c *TConfig) WithVariables(variables map[string]interface{}) *TConfig {
	c.Variables = variables
	return c
}

// SetBaseURL sets base URL for current testcase.
func (c *TConfig) SetBaseURL(baseURL string) *TConfig {
	c.BaseURL = baseURL
	return c
}

// SetHeaders sets global headers for current testcase.
func (c *TConfig) SetHeaders(headers map[string]string) *TConfig {
	c.Headers = headers
	return c
}

// SetVerifySSL sets whether to verify SSL for current testcase.
func (c *TConfig) SetVerifySSL(verify bool) *TConfig {
	c.Verify = verify
	return c
}

// WithParameters sets parameters for current testcase.
func (c *TConfig) WithParameters(parameters map[string]interface{}) *TConfig {
	c.Parameters = parameters
	return c
}

// SetThinkTime sets think time config for current testcase.
func (c *TConfig) SetThinkTime(strategy thinkTimeStrategy, cfg interface{}, limit float64) *TConfig {
	c.ThinkTimeSetting = &ThinkTimeConfig{strategy, cfg, limit}
	return c
}

// SetTimeout sets testcase timeout in seconds.
func (c *TConfig) SetTimeout(timeout time.Duration) *TConfig {
	c.Timeout = timeout.Seconds()
	return c
}

// ExportVars specifies variable names to export for current testcase.
func (c *TConfig) ExportVars(vars ...string) *TConfig {
	c.Export = vars
	return c
}

// SetWeight sets weight for current testcase, which is used in load testing.
func (c *TConfig) SetWeight(weight int) *TConfig {
	c.Weight = weight
	return c
}

func (c *TConfig) SetWebSocket(times, interval, timeout, size int64) *TConfig {
	c.WebSocketSetting = &WebSocketConfig{
		ReconnectionTimes:    times,
		ReconnectionInterval: interval,
		MaxMessageSize:       size,
	}
	return c
}

func (c *TConfig) SetIOS(options ...uixt.IOSDeviceOption) *TConfig {
	wdaOptions := &uixt.IOSDevice{}
	for _, option := range options {
		option(wdaOptions)
	}

	// each device can have its own settings
	if wdaOptions.UDID != "" {
		c.IOS = append(c.IOS, wdaOptions)
		return c
	}

	// device UDID is not specified, settings will be shared
	if len(c.IOS) == 0 {
		c.IOS = append(c.IOS, wdaOptions)
	} else {
		c.IOS[0] = wdaOptions
	}
	return c
}

func (c *TConfig) SetAndroid(options ...uixt.AndroidDeviceOption) *TConfig {
	uiaOptions := &uixt.AndroidDevice{}
	for _, option := range options {
		option(uiaOptions)
	}

	// each device can have its own settings
	if uiaOptions.SerialNumber != "" {
		c.Android = append(c.Android, uiaOptions)
		return c
	}

	// device UDID is not specified, settings will be shared
	if len(c.Android) == 0 {
		c.Android = append(c.Android, uiaOptions)
	} else {
		c.Android[0] = uiaOptions
	}
	return c
}

type ThinkTimeConfig struct {
	Strategy thinkTimeStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty"` // default、random、multiply、ignore
	Setting  interface{}       `json:"setting,omitempty" yaml:"setting,omitempty"`   // random(map): {"min_percentage": 0.5, "max_percentage": 1.5}; 10、multiply(float64): 1.5
	Limit    float64           `json:"limit,omitempty" yaml:"limit,omitempty"`       // limit think time no more than specific time, ignore if value <= 0
}

func (ttc *ThinkTimeConfig) checkThinkTime() {
	if ttc == nil {
		return
	}
	// unset strategy, set default strategy
	if ttc.Strategy == "" {
		ttc.Strategy = thinkTimeDefault
	}
	// check think time
	if ttc.Strategy == thinkTimeRandomPercentage {
		if ttc.Setting == nil || reflect.TypeOf(ttc.Setting).Kind() != reflect.Map {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		value, ok := ttc.Setting.(map[string]interface{})
		if !ok {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		if _, ok := value["min_percentage"]; !ok {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		if _, ok := value["max_percentage"]; !ok {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		left, err := builtin.Interface2Float64(value["min_percentage"])
		if err != nil {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		right, err := builtin.Interface2Float64(value["max_percentage"])
		if err != nil {
			ttc.Setting = thinkTimeDefaultRandom
			return
		}
		ttc.Setting = map[string]float64{"min_percentage": left, "max_percentage": right}
	} else if ttc.Strategy == thinkTimeMultiply {
		if ttc.Setting == nil {
			ttc.Setting = float64(0) // default
			return
		}
		value, err := builtin.Interface2Float64(ttc.Setting)
		if err != nil {
			ttc.Setting = float64(0) // default
			return
		}
		ttc.Setting = value
	} else if ttc.Strategy != thinkTimeIgnore {
		// unrecognized strategy, set default strategy
		ttc.Strategy = thinkTimeDefault
	}
}

type thinkTimeStrategy string

const (
	thinkTimeDefault          thinkTimeStrategy = "default"           // as recorded
	thinkTimeRandomPercentage thinkTimeStrategy = "random_percentage" // use random percentage of recorded think time
	thinkTimeMultiply         thinkTimeStrategy = "multiply"          // multiply recorded think time
	thinkTimeIgnore           thinkTimeStrategy = "ignore"            // ignore recorded think time
)

const (
	thinkTimeDefaultMultiply = 1
)

var thinkTimeDefaultRandom = map[string]float64{"min_percentage": 0.5, "max_percentage": 1.5}

type PluginConfig struct {
	Path    string
	Type    string // bin、so、py
	Content []byte
}
