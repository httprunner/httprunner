package option

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

// NewMCPOptions creates MCP tool options from a struct using reflection
// This function is kept for backward compatibility with existing code
// New code should use UnifiedActionRequest.GetMCPOptions() instead
func NewMCPOptions(t interface{}) (options []mcp.ToolOption) {
	tType := reflect.TypeOf(t)

	// Handle pointer type by getting the element type
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	// Ensure we have a struct type
	if tType.Kind() != reflect.Struct {
		log.Warn().Str("type", tType.String()).Msg("NewMCPOptions expects a struct or pointer to struct")
		return options
	}

	for i := 0; i < tType.NumField(); i++ {
		field := tType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]
		binding := field.Tag.Get("binding")
		required := strings.Contains(binding, "required")
		desc := field.Tag.Get("desc")
		fieldType := field.Type
		// Handle pointer types
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
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
			// Handle slice types, especially []string and []float64
			if field.Type.Elem().Kind() == reflect.String {
				// Array of strings
				if required {
					options = append(options, mcp.WithArray(name, mcp.Required(), mcp.Description(desc)))
				} else {
					options = append(options, mcp.WithArray(name, mcp.Description(desc)))
				}
			} else if field.Type.Elem().Kind() == reflect.Float64 {
				// Array of numbers
				if required {
					options = append(options, mcp.WithArray(name, mcp.Required(), mcp.Description(desc)))
				} else {
					options = append(options, mcp.WithArray(name, mcp.Description(desc)))
				}
			}
		default:
			log.Warn().Str("field_type", field.Type.String()).Msg("Unsupported field type")
		}
	}
	return options
}

// UnifiedActionRequest represents a unified request structure that combines
// ActionOptions with specific action parameters
type UnifiedActionRequest struct {
	// Device targeting
	Platform string `json:"platform" binding:"omitempty" desc:"Device platform: android/ios/browser"`
	Serial   string `json:"serial" binding:"omitempty" desc:"Device serial/udid/browser id"`

	// Common action parameters
	X         *float64 `json:"x,omitempty" binding:"omitempty,min=0" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y         *float64 `json:"y,omitempty" binding:"omitempty,min=0" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	FromX     *float64 `json:"from_x,omitempty" binding:"omitempty,min=0" desc:"Starting X coordinate"`
	FromY     *float64 `json:"from_y,omitempty" binding:"omitempty,min=0" desc:"Starting Y coordinate"`
	ToX       *float64 `json:"to_x,omitempty" binding:"omitempty,min=0" desc:"Ending X coordinate"`
	ToY       *float64 `json:"to_y,omitempty" binding:"omitempty,min=0" desc:"Ending Y coordinate"`
	Text      string   `json:"text,omitempty" desc:"Text content for input/search operations"`
	Direction string   `json:"direction,omitempty" desc:"Direction for swipe operations: up/down/left/right"`

	// App/Package related
	PackageName        string `json:"packageName,omitempty" desc:"Package name of the app"`
	AppName            string `json:"appName,omitempty" desc:"App name to find"`
	AppUrl             string `json:"appUrl,omitempty" desc:"App URL for installation"`
	MappingUrl         string `json:"mappingUrl,omitempty" desc:"Mapping URL for app installation"`
	ResourceMappingUrl string `json:"resourceMappingUrl,omitempty" desc:"Resource mapping URL for app installation"`

	// Web/Browser related
	Selector    string `json:"selector,omitempty" desc:"CSS or XPath selector"`
	TabIndex    *int   `json:"tabIndex,omitempty" desc:"Browser tab index"`
	PhoneNumber string `json:"phoneNumber,omitempty" desc:"Phone number for login"`
	Captcha     string `json:"captcha,omitempty" desc:"Captcha code"`
	Password    string `json:"password,omitempty" desc:"Password for login"`

	// Button/Key related
	Button  types.DeviceButton `json:"button,omitempty" desc:"Device button to press"`
	Ime     string             `json:"ime,omitempty" desc:"IME package name"`
	Count   *int               `json:"count,omitempty" desc:"Count for delete operations"`
	Keycode *int               `json:"keycode,omitempty" desc:"Keycode for key press operations"`

	// Image/CV related
	ImagePath string `json:"imagePath,omitempty" desc:"Path to reference image for CV recognition"`

	// HTTP API specific fields
	FileUrl    string `json:"file_url,omitempty" desc:"File URL for upload operations"`
	FileFormat string `json:"file_format,omitempty" desc:"File format for upload operations"`
	ImageUrl   string `json:"imageUrl,omitempty" desc:"Image URL for media operations"`
	VideoUrl   string `json:"videoUrl,omitempty" desc:"Video URL for media operations"`
	Delta      *int   `json:"delta,omitempty" desc:"Delta value for scroll operations"`
	Width      *int   `json:"width,omitempty" desc:"Width for browser creation"`
	Height     *int   `json:"height,omitempty" desc:"Height for browser creation"`

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
	MaxRetryTimes *int            `json:"max_retry_times,omitempty" desc:"Maximum retry times"`
	Interval      *float64        `json:"interval,omitempty" desc:"Interval between retries in seconds"`
	Duration      *float64        `json:"duration,omitempty" desc:"Action duration in seconds"`
	PressDuration *float64        `json:"press_duration,omitempty" desc:"Press duration in seconds"`
	Steps         *int            `json:"steps,omitempty" desc:"Number of steps for action"`
	Timeout       *int            `json:"timeout,omitempty" desc:"Timeout in seconds"`
	Frequency     *int            `json:"frequency,omitempty" desc:"Action frequency"`

	// Filter options (from ScreenFilterOptions)
	Scope               []float64 `json:"scope,omitempty" desc:"Screen scope [x1,y1,x2,y2] in percentage"`
	AbsScope            []int     `json:"absScope,omitempty" desc:"Absolute screen scope [x1,y1,x2,y2] in pixels"`
	Regex               *bool     `json:"regex,omitempty" desc:"Use regex to match text"`
	TapOffset           []int     `json:"tap_offset,omitempty" desc:"Tap offset [x,y]"`
	TapRandomRect       *bool     `json:"tap_random_rect,omitempty" desc:"Tap random point in rectangle"`
	SwipeOffset         []int     `json:"swipe_offset,omitempty" desc:"Swipe offset [fromX,fromY,toX,toY]"`
	OffsetRandomRange   []int     `json:"offset_random_range,omitempty" desc:"Random offset range [min,max]"`
	Index               *int      `json:"index,omitempty" desc:"Element index when multiple matches found"`
	MatchOne            *bool     `json:"match_one,omitempty" desc:"Match only one element"`
	IgnoreNotFoundError *bool     `json:"ignore_NotFoundError,omitempty" desc:"Ignore error if element not found"`

	// Screenshot options (from ScreenShotOptions)
	ScreenShotWithOCR            *bool    `json:"screenshot_with_ocr,omitempty" desc:"Take screenshot with OCR"`
	ScreenShotWithUpload         *bool    `json:"screenshot_with_upload,omitempty" desc:"Upload screenshot"`
	ScreenShotWithLiveType       *bool    `json:"screenshot_with_live_type,omitempty" desc:"Screenshot with live type"`
	ScreenShotWithLivePopularity *bool    `json:"screenshot_with_live_popularity,omitempty" desc:"Screenshot with live popularity"`
	ScreenShotWithUITypes        []string `json:"screenshot_with_ui_types,omitempty" desc:"Screenshot with UI types"`
	ScreenShotWithClosePopups    *bool    `json:"screenshot_with_close_popups,omitempty" desc:"Close popups before screenshot"`
	ScreenShotWithOCRCluster     string   `json:"screenshot_with_ocr_cluster,omitempty" desc:"OCR cluster for screenshot"`
	ScreenShotFileName           string   `json:"screenshot_file_name,omitempty" desc:"Screenshot file name"`

	// Screen record options (from ScreenRecordOptions)
	ScreenRecordDuration   *float64 `json:"screenrecord_duration,omitempty" desc:"Screen record duration"`
	ScreenRecordWithAudio  *bool    `json:"screenrecord_with_audio,omitempty" desc:"Record with audio"`
	ScreenRecordWithScrcpy *bool    `json:"screenrecord_with_scrcpy,omitempty" desc:"Use scrcpy for recording"`
	ScreenRecordPath       string   `json:"screenrecord_path,omitempty" desc:"Screen record output path"`

	// Mark operation options (from MarkOperationOptions)
	PreMarkOperation  *bool `json:"pre_mark_operation,omitempty" desc:"Mark operation before action"`
	PostMarkOperation *bool `json:"post_mark_operation,omitempty" desc:"Mark operation after action"`

	// Custom options
	Custom map[string]interface{} `json:"custom,omitempty" desc:"Custom options"`
}

// HTTP API direct usage methods

// GetX returns the X coordinate value, handling nil pointer safely
func (r *UnifiedActionRequest) GetX() float64 {
	if r.X != nil {
		return *r.X
	}
	return 0
}

// GetY returns the Y coordinate value, handling nil pointer safely
func (r *UnifiedActionRequest) GetY() float64 {
	if r.Y != nil {
		return *r.Y
	}
	return 0
}

// GetFromX returns the FromX coordinate value, handling nil pointer safely
func (r *UnifiedActionRequest) GetFromX() float64 {
	if r.FromX != nil {
		return *r.FromX
	}
	return 0
}

// GetFromY returns the FromY coordinate value, handling nil pointer safely
func (r *UnifiedActionRequest) GetFromY() float64 {
	if r.FromY != nil {
		return *r.FromY
	}
	return 0
}

// GetToX returns the ToX coordinate value, handling nil pointer safely
func (r *UnifiedActionRequest) GetToX() float64 {
	if r.ToX != nil {
		return *r.ToX
	}
	return 0
}

// GetToY returns the ToY coordinate value, handling nil pointer safely
func (r *UnifiedActionRequest) GetToY() float64 {
	if r.ToY != nil {
		return *r.ToY
	}
	return 0
}

// GetDuration returns the duration value, handling nil pointer safely
func (r *UnifiedActionRequest) GetDuration() float64 {
	if r.Duration != nil {
		return *r.Duration
	}
	return 0
}

// GetPressDuration returns the press duration value, handling nil pointer safely
func (r *UnifiedActionRequest) GetPressDuration() float64 {
	if r.PressDuration != nil {
		return *r.PressDuration
	}
	return 0
}

// GetCount returns the count value, handling nil pointer safely
func (r *UnifiedActionRequest) GetCount() int {
	if r.Count != nil {
		return *r.Count
	}
	return 0
}

// GetKeycode returns the keycode value, handling nil pointer safely
func (r *UnifiedActionRequest) GetKeycode() int {
	if r.Keycode != nil {
		return *r.Keycode
	}
	return 0
}

// GetFrequency returns the frequency value, handling nil pointer safely
func (r *UnifiedActionRequest) GetFrequency() int {
	if r.Frequency != nil {
		return *r.Frequency
	}
	return 0
}

// GetTabIndex returns the tab index value, handling nil pointer safely
func (r *UnifiedActionRequest) GetTabIndex() int {
	if r.TabIndex != nil {
		return *r.TabIndex
	}
	return 0
}

// GetDelta returns the delta value, handling nil pointer safely
func (r *UnifiedActionRequest) GetDelta() int {
	if r.Delta != nil {
		return *r.Delta
	}
	return 0
}

// GetWidth returns the width value, handling nil pointer safely
func (r *UnifiedActionRequest) GetWidth() int {
	if r.Width != nil {
		return *r.Width
	}
	return 0
}

// GetHeight returns the height value, handling nil pointer safely
func (r *UnifiedActionRequest) GetHeight() int {
	if r.Height != nil {
		return *r.Height
	}
	return 0
}

// GetTimeout returns the timeout value, handling nil pointer safely
func (r *UnifiedActionRequest) GetTimeout() int {
	if r.Timeout != nil {
		return *r.Timeout
	}
	return 0
}

// GetMilliseconds returns the milliseconds value, handling nil pointer safely
func (r *UnifiedActionRequest) GetMilliseconds() int64 {
	if r.Milliseconds != nil {
		return *r.Milliseconds
	}
	return 0
}

// ValidateForHTTPAPI validates the request for HTTP API usage
func (r *UnifiedActionRequest) ValidateForHTTPAPI(actionType ActionMethod) error {
	// Basic validation - Platform and Serial are set from URL, so skip here
	// They will be validated by setRequestContextFromURL

	// Action-specific validation using a more efficient approach
	return r.validateActionSpecificFields(actionType)
}

// validateActionSpecificFields performs action-specific field validation
func (r *UnifiedActionRequest) validateActionSpecificFields(actionType ActionMethod) error {
	// Define validation rules for each action type using ActionMethod constants
	validationRules := map[ActionMethod]func() error{
		ACTION_Tap: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_TapXY: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_TapAbsXY: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_DoubleTap: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_DoubleTapXY: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_RightClick: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_SecondaryClick: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_Hover: func() error {
			return r.requireFields("x and y coordinates", r.X != nil && r.Y != nil)
		},
		ACTION_Drag: func() error {
			return r.requireFields("fromX, fromY, toX, toY coordinates",
				r.FromX != nil && r.FromY != nil && r.ToX != nil && r.ToY != nil)
		},
		ACTION_SwipeCoordinate: func() error {
			return r.requireFields("fromX, fromY, toX, toY coordinates",
				r.FromX != nil && r.FromY != nil && r.ToX != nil && r.ToY != nil)
		},
		ACTION_Swipe: func() error {
			return r.requireFields("direction", r.Direction != "")
		},
		ACTION_SwipeDirection: func() error {
			return r.requireFields("direction", r.Direction != "")
		},
		ACTION_Input: func() error {
			return r.requireFields("text", r.Text != "")
		},
		ACTION_Delete: func() error {
			// Count is optional, will use default if not provided
			return nil
		},
		ACTION_Backspace: func() error {
			// Count is optional, will use default if not provided
			return nil
		},
		ACTION_KeyCode: func() error {
			return r.requireFields("keycode", r.Keycode != nil)
		},
		ACTION_Scroll: func() error {
			return r.requireFields("delta", r.Delta != nil)
		},
		ACTION_AppInfo: func() error {
			return r.requireFields("packageName", r.PackageName != "")
		},
		ACTION_AppClear: func() error {
			return r.requireFields("packageName", r.PackageName != "")
		},
		ACTION_AppLaunch: func() error {
			return r.requireFields("packageName", r.PackageName != "")
		},
		ACTION_AppTerminate: func() error {
			return r.requireFields("packageName", r.PackageName != "")
		},
		ACTION_AppUninstall: func() error {
			return r.requireFields("packageName", r.PackageName != "")
		},
		ACTION_AppInstall: func() error {
			return r.requireFields("appUrl", r.AppUrl != "")
		},
		ACTION_TapByOCR: func() error {
			return r.requireFields("text", r.Text != "")
		},
		ACTION_SwipeToTapText: func() error {
			return r.requireFields("text", r.Text != "")
		},
		ACTION_TapByCV: func() error {
			return r.requireFields("imagePath", r.ImagePath != "")
		},
		ACTION_SwipeToTapApp: func() error {
			return r.requireFields("appName", r.AppName != "")
		},
		ACTION_SwipeToTapTexts: func() error {
			return r.requireFields("texts array", len(r.Texts) > 0)
		},
		ACTION_TapBySelector: func() error {
			return r.requireFields("selector", r.Selector != "")
		},
		ACTION_HoverBySelector: func() error {
			return r.requireFields("selector", r.Selector != "")
		},
		ACTION_SecondaryClickBySelector: func() error {
			return r.requireFields("selector", r.Selector != "")
		},
		ACTION_WebCloseTab: func() error {
			return r.requireFields("tabIndex", r.TabIndex != nil)
		},
		ACTION_WebLoginNoneUI: func() error {
			if r.PackageName == "" || r.PhoneNumber == "" || r.Captcha == "" || r.Password == "" {
				return fmt.Errorf("packageName, phoneNumber, captcha, and password are required for web_login_none_ui action")
			}
			return nil
		},
		ACTION_SetIme: func() error {
			return r.requireFields("ime", r.Ime != "")
		},
		ACTION_GetSource: func() error {
			return r.requireFields("packageName", r.PackageName != "")
		},
		ACTION_SleepMS: func() error {
			return r.requireFields("milliseconds", r.Milliseconds != nil)
		},
		ACTION_SleepRandom: func() error {
			return r.requireFields("params array", len(r.Params) > 0)
		},
		ACTION_AIAction: func() error {
			return r.requireFields("prompt", r.Prompt != "")
		},
		ACTION_Finished: func() error {
			return r.requireFields("content", r.Content != "")
		},
		ACTION_Upload: func() error {
			if r.X == nil || r.Y == nil || r.FileUrl == "" {
				return fmt.Errorf("x, y coordinates and fileUrl are required for upload action")
			}
			return nil
		},
		ACTION_PushMedia: func() error {
			if r.ImageUrl == "" && r.VideoUrl == "" {
				return fmt.Errorf("either imageUrl or videoUrl is required for push_media action")
			}
			return nil
		},
		ACTION_CreateBrowser: func() error {
			return r.requireFields("timeout", r.Timeout != nil)
		},
	}

	// Execute validation rule for the action type
	if validator, exists := validationRules[actionType]; exists {
		return validator()
	}

	// No specific validation needed for this action type
	return nil
}

// requireFields is a helper function to generate consistent error messages
func (r *UnifiedActionRequest) requireFields(fieldDesc string, condition bool) error {
	if !condition {
		return fmt.Errorf("%s is required for this action", fieldDesc)
	}
	return nil
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

	// Copy filter options (ScreenFilterOptions)
	opts.ScreenFilterOptions.Scope = r.Scope
	opts.ScreenFilterOptions.AbsScope = r.AbsScope
	if r.Regex != nil {
		opts.ScreenFilterOptions.Regex = *r.Regex
	}
	opts.ScreenFilterOptions.TapOffset = r.TapOffset
	if r.TapRandomRect != nil {
		opts.ScreenFilterOptions.TapRandomRect = *r.TapRandomRect
	}
	opts.ScreenFilterOptions.SwipeOffset = r.SwipeOffset
	opts.ScreenFilterOptions.OffsetRandomRange = r.OffsetRandomRange
	if r.Index != nil {
		opts.ScreenFilterOptions.Index = *r.Index
	}
	if r.MatchOne != nil {
		opts.ScreenFilterOptions.MatchOne = *r.MatchOne
	}
	if r.IgnoreNotFoundError != nil {
		opts.ScreenFilterOptions.IgnoreNotFoundError = *r.IgnoreNotFoundError
	}

	// Copy screenshot options (ScreenShotOptions)
	if r.ScreenShotWithOCR != nil {
		opts.ScreenShotOptions.ScreenShotWithOCR = *r.ScreenShotWithOCR
	}
	if r.ScreenShotWithUpload != nil {
		opts.ScreenShotOptions.ScreenShotWithUpload = *r.ScreenShotWithUpload
	}
	if r.ScreenShotWithLiveType != nil {
		opts.ScreenShotOptions.ScreenShotWithLiveType = *r.ScreenShotWithLiveType
	}
	if r.ScreenShotWithLivePopularity != nil {
		opts.ScreenShotOptions.ScreenShotWithLivePopularity = *r.ScreenShotWithLivePopularity
	}
	opts.ScreenShotOptions.ScreenShotWithUITypes = r.ScreenShotWithUITypes
	if r.ScreenShotWithClosePopups != nil {
		opts.ScreenShotOptions.ScreenShotWithClosePopups = *r.ScreenShotWithClosePopups
	}
	opts.ScreenShotOptions.ScreenShotWithOCRCluster = r.ScreenShotWithOCRCluster
	opts.ScreenShotOptions.ScreenShotFileName = r.ScreenShotFileName

	// Copy screen record options (ScreenRecordOptions)
	if r.ScreenRecordDuration != nil {
		opts.ScreenRecordOptions.ScreenRecordDuration = *r.ScreenRecordDuration
	}
	if r.ScreenRecordWithAudio != nil {
		opts.ScreenRecordOptions.ScreenRecordWithAudio = *r.ScreenRecordWithAudio
	}
	if r.ScreenRecordWithScrcpy != nil {
		opts.ScreenRecordOptions.ScreenRecordWithScrcpy = *r.ScreenRecordWithScrcpy
	}
	opts.ScreenRecordOptions.ScreenRecordPath = r.ScreenRecordPath

	// Copy mark operation options (MarkOperationOptions)
	if r.PreMarkOperation != nil {
		opts.MarkOperationOptions.PreMarkOperation = *r.PreMarkOperation
	}
	if r.PostMarkOperation != nil {
		opts.MarkOperationOptions.PostMarkOperation = *r.PostMarkOperation
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
