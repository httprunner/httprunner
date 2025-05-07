package hrp

import (
	"reflect"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type IConfig interface {
	Get() *TConfig
}

// NewConfig returns a new constructed testcase config with specified testcase name.
func NewConfig(name string) *TConfig {
	return &TConfig{
		Name:      name,
		Environs:  make(map[string]string),
		Variables: make(map[string]interface{}),
	}
}

// define struct for testcase config
type TConfig struct {
	Name              string                         `json:"name" yaml:"name"` // required
	Verify            bool                           `json:"verify,omitempty" yaml:"verify,omitempty"`
	BaseURL           string                         `json:"base_url,omitempty" yaml:"base_url,omitempty"`   // deprecated in v4.1, moved to env
	Headers           map[string]string              `json:"headers,omitempty" yaml:"headers,omitempty"`     // public request headers
	Environs          map[string]string              `json:"environs,omitempty" yaml:"environs,omitempty"`   // environment variables
	Variables         map[string]interface{}         `json:"variables,omitempty" yaml:"variables,omitempty"` // global variables
	Parameters        map[string]interface{}         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	ParametersSetting *TParamsConfig                 `json:"parameters_setting,omitempty" yaml:"parameters_setting,omitempty"`
	ThinkTimeSetting  *ThinkTimeConfig               `json:"think_time,omitempty" yaml:"think_time,omitempty"`
	WebSocketSetting  *WebSocketConfig               `json:"websocket,omitempty" yaml:"websocket,omitempty"`
	IOS               []*option.IOSDeviceOptions     `json:"ios,omitempty" yaml:"ios,omitempty"`
	Android           []*option.AndroidDeviceOptions `json:"android,omitempty" yaml:"android,omitempty"`
	Harmony           []*option.HarmonyDeviceOptions `json:"harmony,omitempty" yaml:"harmony,omitempty"`
	RequestTimeout    float32                        `json:"request_timeout,omitempty" yaml:"request_timeout,omitempty"` // request timeout in seconds
	CaseTimeout       float32                        `json:"case_timeout,omitempty" yaml:"case_timeout,omitempty"`       // testcase timeout in seconds
	Export            []string                       `json:"export,omitempty" yaml:"export,omitempty"`
	Weight            int                            `json:"weight,omitempty" yaml:"weight,omitempty"`
	Path              string                         `json:"path,omitempty" yaml:"path,omitempty"`     // testcase file path
	PluginSetting     *PluginConfig                  `json:"plugin,omitempty" yaml:"plugin,omitempty"` // plugin config
	IgnorePopup       bool                           `json:"ignore_popup,omitempty" yaml:"ignore_popup,omitempty"`
	LLMService        option.LLMServiceType          `json:"llm_service,omitempty" yaml:"llm_service,omitempty"`
	CVService         option.CVServiceType           `json:"cv_service,omitempty" yaml:"cv_service,omitempty"`
}

func (c *TConfig) Get() *TConfig {
	return c
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
func (c *TConfig) SetThinkTime(strategy ThinkTimeStrategy, cfg interface{}, limit float64) *TConfig {
	c.ThinkTimeSetting = &ThinkTimeConfig{strategy, cfg, limit}
	return c
}

// SetRequestTimeout sets request timeout in seconds.
func (c *TConfig) SetRequestTimeout(seconds float32) *TConfig {
	c.RequestTimeout = seconds
	return c
}

// SetCaseTimeout sets testcase timeout in seconds.
func (c *TConfig) SetCaseTimeout(seconds float32) *TConfig {
	c.CaseTimeout = seconds
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

// SetLLMService sets LLM service for current testcase.
func (c *TConfig) SetLLMService(llmService option.LLMServiceType) *TConfig {
	c.LLMService = llmService
	return c
}

// SetCVService sets CV service for current testcase.
func (c *TConfig) SetCVService(cvService option.CVServiceType) *TConfig {
	c.CVService = cvService
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

func (c *TConfig) SetIOS(opts ...option.IOSDeviceOption) *TConfig {
	iosOptions := option.NewIOSDeviceOptions(opts...)

	// each device can have its own settings
	if iosOptions.UDID != "" {
		c.IOS = append(c.IOS, iosOptions)
		return c
	}

	// device UDID is not specified, settings will be shared
	if len(c.IOS) == 0 {
		c.IOS = append(c.IOS, iosOptions)
	} else {
		c.IOS[0] = iosOptions
	}
	return c
}

func (c *TConfig) SetHarmony(opts ...option.HarmonyDeviceOption) *TConfig {
	harmonyOptions := option.NewHarmonyDeviceOptions(opts...)

	// each device can have its own settings
	if harmonyOptions.ConnectKey != "" {
		c.Harmony = append(c.Harmony, harmonyOptions)
		return c
	}

	// device UDID is not specified, settings will be shared
	if len(c.Harmony) == 0 {
		c.Harmony = append(c.Harmony, harmonyOptions)
	} else {
		c.Harmony[0] = harmonyOptions
	}
	return c
}

func (c *TConfig) SetAndroid(opts ...option.AndroidDeviceOption) *TConfig {
	uiaOptions := option.NewAndroidDeviceOptions(opts...)

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

// EnablePlugin enables plugin for current testcase.
// default to disable plugin
func (c *TConfig) EnablePlugin() *TConfig {
	c.PluginSetting = &PluginConfig{}
	return c
}

func (c *TConfig) DisableAutoPopupHandler() *TConfig {
	c.IgnorePopup = true
	return c
}

type ThinkTimeConfig struct {
	Strategy ThinkTimeStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty"` // default、random、multiply、ignore
	Setting  interface{}       `json:"setting,omitempty" yaml:"setting,omitempty"`   // random(map): {"min_percentage": 0.5, "max_percentage": 1.5}; 10、multiply(float64): 1.5
	Limit    float64           `json:"limit,omitempty" yaml:"limit,omitempty"`       // limit think time no more than specific time, ignore if value <= 0
}

func (ttc *ThinkTimeConfig) checkThinkTime() {
	if ttc == nil {
		return
	}
	// unset strategy, set default strategy
	if ttc.Strategy == "" {
		ttc.Strategy = ThinkTimeDefault
	}
	// check think time
	if ttc.Strategy == ThinkTimeRandomPercentage {
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
	} else if ttc.Strategy == ThinkTimeMultiply {
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
	} else if ttc.Strategy != ThinkTimeIgnore {
		// unrecognized strategy, set default strategy
		ttc.Strategy = ThinkTimeDefault
	}
}

type ThinkTimeStrategy string

const (
	ThinkTimeDefault          ThinkTimeStrategy = "default"           // as recorded
	ThinkTimeRandomPercentage ThinkTimeStrategy = "random_percentage" // use random percentage of recorded think time
	ThinkTimeMultiply         ThinkTimeStrategy = "multiply"          // multiply recorded think time
	ThinkTimeIgnore           ThinkTimeStrategy = "ignore"            // ignore recorded think time
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
