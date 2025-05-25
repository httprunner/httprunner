package option

import (
	"reflect"
	"strings"

	"github.com/httprunner/httprunner/v5/uixt/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type TargetDeviceRequest struct {
	Platform string `json:"platform" binding:"required" desc:"Device platform: android/ios/browser"`
	Serial   string `json:"serial" binding:"required" desc:"Device serial/udid/browser id"`
}

type TapRequest struct {
	TargetDeviceRequest
	X        float64 `json:"x" binding:"required" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y        float64 `json:"y" binding:"required" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Duration float64 `json:"duration" desc:"Tap duration in seconds (optional)"`
}

type DragRequest struct {
	TargetDeviceRequest
	FromX         float64 `json:"from_x" binding:"required" desc:"Starting X-coordinate (percentage, 0.0 to 1.0)"`
	FromY         float64 `json:"from_y" binding:"required" desc:"Starting Y-coordinate (percentage, 0.0 to 1.0)"`
	ToX           float64 `json:"to_x" binding:"required" desc:"Ending X-coordinate (percentage, 0.0 to 1.0)"`
	ToY           float64 `json:"to_y" binding:"required" desc:"Ending Y-coordinate (percentage, 0.0 to 1.0)"`
	Duration      float64 `json:"duration" desc:"Swipe duration in milliseconds (optional)"`
	PressDuration float64 `json:"press_duration" desc:"Press duration in milliseconds (optional)"`
}

type SwipeRequest struct {
	TargetDeviceRequest
	Direction     string  `json:"direction" binding:"required" desc:"The direction of the swipe. Supported directions: up, down, left, right"`
	Duration      float64 `json:"duration" desc:"Swipe duration in milliseconds (optional)"`
	PressDuration float64 `json:"press_duration" desc:"Press duration in milliseconds (optional)"`
}

type InputRequest struct {
	TargetDeviceRequest
	Text      string `json:"text" binding:"required"`
	Frequency int    `json:"frequency"` // only iOS
}

type DeleteRequest struct {
	TargetDeviceRequest
	Count int `json:"count" binding:"required"`
}

type KeycodeRequest struct {
	TargetDeviceRequest
	Keycode int `json:"keycode" binding:"required"`
}

type AppInstallRequest struct {
	TargetDeviceRequest
	AppUrl             string `json:"appUrl" binding:"required"`
	MappingUrl         string `json:"mappingUrl"`
	ResourceMappingUrl string `json:"resourceMappingUrl"`
	PackageName        string `json:"packageName"`
}

type AppInfoRequest struct {
	TargetDeviceRequest
	PackageName string `form:"packageName" binding:"required"`
}

type AppUninstallRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required"`
}

type AppClearRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required"`
}

type AppLaunchRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required" desc:"The package name of the app to launch"`
}

type AppTerminateRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required" desc:"The package name of the app to terminate"`
}

type PressButtonRequest struct {
	TargetDeviceRequest
	Button types.DeviceButton `json:"button" binding:"required" desc:"The button to press. Supported buttons: BACK (android only), HOME, VOLUME_UP, VOLUME_DOWN, ENTER."`
}

// Additional requests for missing actions
type WebLoginNoneUIRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required" desc:"Package name for the app to login"`
	PhoneNumber string `json:"phoneNumber" binding:"required" desc:"Phone number for login"`
	Captcha     string `json:"captcha" binding:"required" desc:"Captcha code"`
	Password    string `json:"password" binding:"required" desc:"Password for login"`
}

type SwipeToTapAppRequest struct {
	TargetDeviceRequest
	AppName string `json:"appName" binding:"required" desc:"App name to find and tap"`
}

type SwipeToTapTextRequest struct {
	TargetDeviceRequest
	Text string `json:"text" binding:"required" desc:"Text to find and tap"`
}

type SwipeToTapTextsRequest struct {
	TargetDeviceRequest
	Texts []string `json:"texts" binding:"required" desc:"List of texts to find and tap"`
}

type SecondaryClickRequest struct {
	TargetDeviceRequest
	X float64 `json:"x" binding:"required" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y float64 `json:"y" binding:"required" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
}

type SelectorRequest struct {
	TargetDeviceRequest
	Selector string `json:"selector" binding:"required" desc:"CSS or XPath selector"`
}

type WebCloseTabRequest struct {
	TargetDeviceRequest
	TabIndex int `json:"tabIndex" binding:"required" desc:"Index of the tab to close"`
}

type SetImeRequest struct {
	TargetDeviceRequest
	Ime string `json:"ime" binding:"required" desc:"IME package name to set"`
}

type GetSourceRequest struct {
	TargetDeviceRequest
	PackageName string `json:"packageName" binding:"required" desc:"Package name to get source from"`
}

type TapAbsXYRequest struct {
	TargetDeviceRequest
	X        float64 `json:"x" binding:"required" desc:"Absolute X coordinate in pixels"`
	Y        float64 `json:"y" binding:"required" desc:"Absolute Y coordinate in pixels"`
	Duration float64 `json:"duration" desc:"Tap duration in seconds (optional)"`
}

type TapByOCRRequest struct {
	TargetDeviceRequest
	Text string `json:"text" binding:"required" desc:"OCR text to find and tap"`
}

type TapByCVRequest struct {
	TargetDeviceRequest
	ImagePath string `json:"imagePath" desc:"Path to reference image for CV recognition"`
}

type DoubleTapXYRequest struct {
	TargetDeviceRequest
	X float64 `json:"x" binding:"required" desc:"X coordinate (0.0~1.0 for percent, or absolute pixel value)"`
	Y float64 `json:"y" binding:"required" desc:"Y coordinate (0.0~1.0 for percent, or absolute pixel value)"`
}

type SwipeAdvancedRequest struct {
	TargetDeviceRequest
	FromX         float64 `json:"fromX" binding:"required" desc:"Starting X coordinate"`
	FromY         float64 `json:"fromY" binding:"required" desc:"Starting Y coordinate"`
	ToX           float64 `json:"toX" binding:"required" desc:"Ending X coordinate"`
	ToY           float64 `json:"toY" binding:"required" desc:"Ending Y coordinate"`
	Duration      float64 `json:"duration" desc:"Swipe duration in seconds (optional)"`
	PressDuration float64 `json:"pressDuration" desc:"Press duration in seconds (optional)"`
}

type SleepMSRequest struct {
	TargetDeviceRequest
	Milliseconds int64 `json:"milliseconds" binding:"required" desc:"Sleep duration in milliseconds"`
}

type SleepRandomRequest struct {
	TargetDeviceRequest
	Params []float64 `json:"params" binding:"required" desc:"Random sleep parameters [min, max] or [min1, max1, weight1, ...]"`
}

type CallFunctionRequest struct {
	TargetDeviceRequest
	Description string `json:"description" binding:"required" desc:"Function description"`
}

type AIActionRequest struct {
	TargetDeviceRequest
	Prompt string `json:"prompt" binding:"required" desc:"AI action prompt"`
}

// NewMCPOptions generates mcp.NewTool parameters from a struct type.
// It automatically generates mcp.NewTool parameters based on the struct fields and their desc tags.
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
		switch field.Type.Kind() {
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
