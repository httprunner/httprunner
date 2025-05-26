package option

import (
	"context"
	"reflect"
	"strings"

	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// UnifiedActionRequest represents a unified request structure that combines
// ActionOptions with specific action parameters
type UnifiedActionRequest struct {
	// Device targeting
	Platform string `json:"platform" binding:"required" desc:"Device platform: android/ios/browser"`
	Serial   string `json:"serial" binding:"required" desc:"Device serial/udid/browser id"`

	// Common action parameters
	X         *float64 `json:"x,omitempty" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y         *float64 `json:"y,omitempty" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	FromX     *float64 `json:"fromX,omitempty" desc:"Starting X coordinate"`
	FromY     *float64 `json:"fromY,omitempty" desc:"Starting Y coordinate"`
	ToX       *float64 `json:"toX,omitempty" desc:"Ending X coordinate"`
	ToY       *float64 `json:"toY,omitempty" desc:"Ending Y coordinate"`
	Text      string   `json:"text,omitempty" desc:"Text content for input/search operations"`
	Direction string   `json:"direction,omitempty" desc:"Direction for swipe operations: up/down/left/right"`

	// App/Package related
	PackageName string `json:"packageName,omitempty" desc:"Package name of the app"`
	AppName     string `json:"appName,omitempty" desc:"App name to find"`
	AppUrl      string `json:"appUrl,omitempty" desc:"App URL for installation"`

	// Web/Browser related
	Selector    string `json:"selector,omitempty" desc:"CSS or XPath selector"`
	TabIndex    *int   `json:"tabIndex,omitempty" desc:"Browser tab index"`
	PhoneNumber string `json:"phoneNumber,omitempty" desc:"Phone number for login"`
	Captcha     string `json:"captcha,omitempty" desc:"Captcha code"`
	Password    string `json:"password,omitempty" desc:"Password for login"`

	// Button/Key related
	Button types.DeviceButton `json:"button,omitempty" desc:"Device button to press"`
	Ime    string             `json:"ime,omitempty" desc:"IME package name"`

	// Array parameters
	Texts  []string  `json:"texts,omitempty" desc:"List of texts to search"`
	Params []float64 `json:"params,omitempty" desc:"Generic parameter array"`

	// AI related
	Prompt  string `json:"prompt,omitempty" desc:"AI action prompt"`
	Content string `json:"content,omitempty" desc:"Content for finished action"`

	// Time related
	Seconds      *float64 `json:"seconds,omitempty" desc:"Sleep duration in seconds"`
	Milliseconds *int64   `json:"milliseconds,omitempty" desc:"Sleep duration in milliseconds"`

	// Control options (from ActionOptions)
	Context       context.Context `json:"-" yaml:"-"`
	Identifier    string          `json:"identifier,omitempty" desc:"Action identifier for logging"`
	MaxRetryTimes *int            `json:"maxRetryTimes,omitempty" desc:"Maximum retry times"`
	Interval      *float64        `json:"interval,omitempty" desc:"Interval between retries in seconds"`
	Duration      *float64        `json:"duration,omitempty" desc:"Action duration in seconds"`
	PressDuration *float64        `json:"pressDuration,omitempty" desc:"Press duration in seconds"`
	Steps         *int            `json:"steps,omitempty" desc:"Number of steps for action"`
	Timeout       *int            `json:"timeout,omitempty" desc:"Timeout in seconds"`
	Frequency     *int            `json:"frequency,omitempty" desc:"Action frequency"`

	// Filter options (from ScreenFilterOptions)
	Scope               []float64 `json:"scope,omitempty" desc:"Screen scope [x1,y1,x2,y2] in percentage"`
	AbsScope            []int     `json:"absScope,omitempty" desc:"Absolute screen scope [x1,y1,x2,y2] in pixels"`
	Regex               *bool     `json:"regex,omitempty" desc:"Use regex to match text"`
	TapOffset           []int     `json:"tapOffset,omitempty" desc:"Tap offset [x,y]"`
	TapRandomRect       *bool     `json:"tapRandomRect,omitempty" desc:"Tap random point in rectangle"`
	SwipeOffset         []int     `json:"swipeOffset,omitempty" desc:"Swipe offset [fromX,fromY,toX,toY]"`
	OffsetRandomRange   []int     `json:"offsetRandomRange,omitempty" desc:"Random offset range [min,max]"`
	Index               *int      `json:"index,omitempty" desc:"Element index when multiple matches found"`
	MatchOne            *bool     `json:"matchOne,omitempty" desc:"Match only one element"`
	IgnoreNotFoundError *bool     `json:"ignoreNotFoundError,omitempty" desc:"Ignore error if element not found"`

	// Screenshot options (from ScreenShotOptions)
	ScreenShotWithOCR            *bool    `json:"screenshotWithOCR,omitempty" desc:"Take screenshot with OCR"`
	ScreenShotWithUpload         *bool    `json:"screenshotWithUpload,omitempty" desc:"Upload screenshot"`
	ScreenShotWithLiveType       *bool    `json:"screenshotWithLiveType,omitempty" desc:"Screenshot with live type"`
	ScreenShotWithLivePopularity *bool    `json:"screenshotWithLivePopularity,omitempty" desc:"Screenshot with live popularity"`
	ScreenShotWithUITypes        []string `json:"screenshotWithUITypes,omitempty" desc:"Screenshot with UI types"`
	ScreenShotWithClosePopups    *bool    `json:"screenshotWithClosePopups,omitempty" desc:"Close popups before screenshot"`
	ScreenShotWithOCRCluster     string   `json:"screenshotWithOCRCluster,omitempty" desc:"OCR cluster for screenshot"`
	ScreenShotFileName           string   `json:"screenshotFileName,omitempty" desc:"Screenshot file name"`

	// Screen record options (from ScreenRecordOptions)
	ScreenRecordDuration   *float64 `json:"screenRecordDuration,omitempty" desc:"Screen record duration"`
	ScreenRecordWithAudio  *bool    `json:"screenRecordWithAudio,omitempty" desc:"Record with audio"`
	ScreenRecordWithScrcpy *bool    `json:"screenRecordWithScrcpy,omitempty" desc:"Use scrcpy for recording"`
	ScreenRecordPath       string   `json:"screenRecordPath,omitempty" desc:"Screen record output path"`

	// Mark operation options (from MarkOperationOptions)
	PreMarkOperation  *bool `json:"preMarkOperation,omitempty" desc:"Mark operation before action"`
	PostMarkOperation *bool `json:"postMarkOperation,omitempty" desc:"Mark operation after action"`

	// Custom options
	Custom map[string]interface{} `json:"custom,omitempty" desc:"Custom options"`
}

// ToActionOptions converts UnifiedActionRequest to ActionOptions
func (r *UnifiedActionRequest) ToActionOptions() *ActionOptions {
	opts := &ActionOptions{
		Context:    r.Context,
		Identifier: r.Identifier,
		Custom:     r.Custom,
	}

	// Copy pointer values safely
	if r.MaxRetryTimes != nil {
		opts.MaxRetryTimes = *r.MaxRetryTimes
	}
	if r.Interval != nil {
		opts.Interval = *r.Interval
	}
	if r.Duration != nil {
		opts.Duration = *r.Duration
	}
	if r.PressDuration != nil {
		opts.PressDuration = *r.PressDuration
	}
	if r.Steps != nil {
		opts.Steps = *r.Steps
	}
	if r.Timeout != nil {
		opts.Timeout = *r.Timeout
	}
	if r.Frequency != nil {
		opts.Frequency = *r.Frequency
	}

	// Handle direction
	if r.Direction != "" {
		opts.Direction = r.Direction
	} else if len(r.Params) == 4 {
		opts.Direction = r.Params
	}

	// Copy filter options
	opts.Scope = r.Scope
	opts.AbsScope = r.AbsScope
	if r.Regex != nil {
		opts.Regex = *r.Regex
	}
	opts.TapOffset = r.TapOffset
	if r.TapRandomRect != nil {
		opts.TapRandomRect = *r.TapRandomRect
	}
	opts.SwipeOffset = r.SwipeOffset
	opts.OffsetRandomRange = r.OffsetRandomRange
	if r.Index != nil {
		opts.Index = *r.Index
	}
	if r.MatchOne != nil {
		opts.MatchOne = *r.MatchOne
	}
	if r.IgnoreNotFoundError != nil {
		opts.IgnoreNotFoundError = *r.IgnoreNotFoundError
	}

	// Copy screenshot options
	if r.ScreenShotWithOCR != nil {
		opts.ScreenShotWithOCR = *r.ScreenShotWithOCR
	}
	if r.ScreenShotWithUpload != nil {
		opts.ScreenShotWithUpload = *r.ScreenShotWithUpload
	}
	if r.ScreenShotWithLiveType != nil {
		opts.ScreenShotWithLiveType = *r.ScreenShotWithLiveType
	}
	if r.ScreenShotWithLivePopularity != nil {
		opts.ScreenShotWithLivePopularity = *r.ScreenShotWithLivePopularity
	}
	opts.ScreenShotWithUITypes = r.ScreenShotWithUITypes
	if r.ScreenShotWithClosePopups != nil {
		opts.ScreenShotWithClosePopups = *r.ScreenShotWithClosePopups
	}
	opts.ScreenShotWithOCRCluster = r.ScreenShotWithOCRCluster
	opts.ScreenShotFileName = r.ScreenShotFileName

	// Copy screen record options
	if r.ScreenRecordDuration != nil {
		opts.ScreenRecordDuration = *r.ScreenRecordDuration
	}
	if r.ScreenRecordWithAudio != nil {
		opts.ScreenRecordWithAudio = *r.ScreenRecordWithAudio
	}
	if r.ScreenRecordWithScrcpy != nil {
		opts.ScreenRecordWithScrcpy = *r.ScreenRecordWithScrcpy
	}
	opts.ScreenRecordPath = r.ScreenRecordPath

	// Copy mark operation options
	if r.PreMarkOperation != nil {
		opts.PreMarkOperation = *r.PreMarkOperation
	}
	if r.PostMarkOperation != nil {
		opts.PostMarkOperation = *r.PostMarkOperation
	}

	return opts
}

// GetMCPOptions generates MCP tool options for specific action types
func (r *UnifiedActionRequest) GetMCPOptions(actionType ActionMethod) []mcp.ToolOption {
	// Define field mappings for different action types
	fieldMappings := map[ActionMethod][]string{
		ACTION_TapXY:                    {"platform", "serial", "x", "y", "duration"},
		ACTION_TapAbsXY:                 {"platform", "serial", "x", "y", "duration"},
		ACTION_TapByOCR:                 {"platform", "serial", "text", "ignoreNotFoundError", "maxRetryTimes", "index", "regex", "tapRandomRect"},
		ACTION_TapByCV:                  {"platform", "serial", "ignoreNotFoundError", "maxRetryTimes", "index", "tapRandomRect"},
		ACTION_DoubleTapXY:              {"platform", "serial", "x", "y"},
		ACTION_SwipeDirection:           {"platform", "serial", "direction", "duration", "pressDuration"},
		ACTION_SwipeCoordinate:          {"platform", "serial", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_Swipe:                    {"platform", "serial", "direction", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_Drag:                     {"platform", "serial", "fromX", "fromY", "toX", "toY", "duration", "pressDuration"},
		ACTION_Input:                    {"platform", "serial", "text", "frequency"},
		ACTION_AppLaunch:                {"platform", "serial", "packageName"},
		ACTION_AppTerminate:             {"platform", "serial", "packageName"},
		ACTION_AppInstall:               {"platform", "serial", "appUrl", "packageName"},
		ACTION_AppUninstall:             {"platform", "serial", "packageName"},
		ACTION_AppClear:                 {"platform", "serial", "packageName"},
		ACTION_PressButton:              {"platform", "serial", "button"},
		ACTION_SwipeToTapApp:            {"platform", "serial", "appName", "ignoreNotFoundError", "maxRetryTimes", "index"},
		ACTION_SwipeToTapText:           {"platform", "serial", "text", "ignoreNotFoundError", "maxRetryTimes", "index", "regex"},
		ACTION_SwipeToTapTexts:          {"platform", "serial", "texts", "ignoreNotFoundError", "maxRetryTimes", "index", "regex"},
		ACTION_SecondaryClick:           {"platform", "serial", "x", "y"},
		ACTION_HoverBySelector:          {"platform", "serial", "selector"},
		ACTION_TapBySelector:            {"platform", "serial", "selector"},
		ACTION_SecondaryClickBySelector: {"platform", "serial", "selector"},
		ACTION_WebCloseTab:              {"platform", "serial", "tabIndex"},
		ACTION_WebLoginNoneUI:           {"platform", "serial", "packageName", "phoneNumber", "captcha", "password"},
		ACTION_SetIme:                   {"platform", "serial", "ime"},
		ACTION_GetSource:                {"platform", "serial", "packageName"},
		ACTION_Sleep:                    {"seconds"},
		ACTION_SleepMS:                  {"platform", "serial", "milliseconds"},
		ACTION_SleepRandom:              {"platform", "serial", "params"},
		ACTION_AIAction:                 {"platform", "serial", "prompt"},
		ACTION_Finished:                 {"content"},
		ACTION_ListAvailableDevices:     {},
		ACTION_SelectDevice:             {"platform", "serial"},
		ACTION_ScreenShot:               {"platform", "serial"},
		ACTION_GetScreenSize:            {"platform", "serial"},
		ACTION_Home:                     {"platform", "serial"},
		ACTION_Back:                     {"platform", "serial"},
		ACTION_ListPackages:             {"platform", "serial"},
		ACTION_ClosePopups:              {"platform", "serial"},
	}

	fields := fieldMappings[actionType]
	if fields == nil {
		// Fallback to all fields if not specifically mapped
		return NewMCPOptions(*r)
	}

	// Generate options only for specified fields
	return r.generateMCPOptionsForFields(fields)
}

// generateMCPOptionsForFields generates MCP options for specific fields
func (r *UnifiedActionRequest) generateMCPOptionsForFields(fields []string) []mcp.ToolOption {
	options := make([]mcp.ToolOption, 0)
	rType := reflect.TypeOf(*r)
	rValue := reflect.ValueOf(*r)

	fieldMap := make(map[string]reflect.StructField)
	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			name := strings.Split(jsonTag, ",")[0]
			fieldMap[name] = field
		}
	}

	for _, fieldName := range fields {
		field, exists := fieldMap[fieldName]
		if !exists {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		binding := field.Tag.Get("binding")
		required := strings.Contains(binding, "required")
		desc := field.Tag.Get("desc")

		// Check if field has a value
		fieldValue := rValue.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			continue
		}

		// Handle pointer types
		fieldType := field.Type
		isPointer := false
		if fieldType.Kind() == reflect.Ptr {
			isPointer = true
			fieldType = fieldType.Elem()
		}

		// Skip nil pointer fields if not required
		if isPointer && fieldValue.IsNil() && !required {
			continue
		}

		switch fieldType.Kind() {
		case reflect.Float64, reflect.Float32, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if required {
				options = append(options, mcp.WithNumber(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithNumber(name, mcp.Description(desc)))
			}
		case reflect.String:
			if required {
				options = append(options, mcp.WithString(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithString(name, mcp.Description(desc)))
			}
		case reflect.Bool:
			if required {
				options = append(options, mcp.WithBoolean(name, mcp.Required(), mcp.Description(desc)))
			} else {
				options = append(options, mcp.WithBoolean(name, mcp.Description(desc)))
			}
		case reflect.Slice:
			if fieldType.Elem().Kind() == reflect.String || fieldType.Elem().Kind() == reflect.Float64 {
				if required {
					options = append(options, mcp.WithArray(name, mcp.Required(), mcp.Description(desc)))
				} else {
					options = append(options, mcp.WithArray(name, mcp.Description(desc)))
				}
			}
		case reflect.Map, reflect.Interface:
			// Skip map and interface types for now
			continue
		default:
			log.Warn().Str("field_type", fieldType.String()).Msg("Unsupported field type")
		}
	}

	return options
}
