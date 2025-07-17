package types

import "fmt"

// DeviceStatus example:
//
//	{
//	    "status": 0,
//	    "sessionId": "7DD3B0F7-958B-45F1-B99D-745B4EEFE178",
//	    "value": {
//	        "message": "WebDriverAgent is ready to accept commands",
//	        "state": "success",
//	        "os": {
//	            "testmanagerdVersion": 28,
//	            "name": "iOS",
//	            "sdkVersion": "16.0",
//	            "version": "15.3.1"
//	        },
//	        "ios": {
//	            "ip": "169.254.237.64"
//	        },
//	        "ready": true,
//	        "build": {
//	            "time": "Jun 21 2024 11:11:37",
//	            "productBundleIdentifier": "com.facebook.WebDriverAgentRunner"
//	        }
//	    }
//	}
type DeviceStatus struct {
	Message string `json:"message"`
	State   string `json:"state"`
	Ready   bool   `json:"ready"`
	Device  string `json:"device"`
	OS      struct {
		TestmanagerdVersion int    `json:"testmanagerdVersion"`
		Name                string `json:"name"`
		SdkVersion          string `json:"sdkVersion"`
		Version             string `json:"version"`
	} `json:"os"`
	IOS struct {
		IP               string `json:"ip"`
		SimulatorVersion string `json:"simulatorVersion"`
	} `json:"ios"`
	Build struct {
		Time                    string `json:"time"`
		ProductBundleIdentifier string `json:"productBundleIdentifier"`
		Version                 string `json:"version"`       // OpenSource WDA version
		GtfWDAVersion           string `json:"gtfWDAVersion"` // GTF WDA version
	} `json:"build"`
}

type DeviceInfo struct {
	TimeZone           string `json:"timeZone"`
	CurrentLocale      string `json:"currentLocale"`
	Model              string `json:"model"`
	UUID               string `json:"uuid"`
	UserInterfaceIdiom int    `json:"userInterfaceIdiom"`
	UserInterfaceStyle string `json:"userInterfaceStyle"`
	Name               string `json:"name"`
	IsSimulator        bool   `json:"isSimulator"`
	ThermalState       int    `json:"thermalState"`
	// ANDROID_ID A 64-bit number (as a hex string) that is uniquely generated when the user
	// first sets up the device and should remain constant for the lifetime of the user's device. The value
	// may change if a factory reset is performed on the device.
	AndroidID string `json:"androidId"`
	// Build.MANUFACTURER value
	Manufacturer string `json:"manufacturer"`
	// Build.BRAND value
	Brand string `json:"brand"`
	// Current running OS's API VERSION
	APIVersion string `json:"apiVersion"`
	// The current version string, for example "1.0" or "3.4b5"
	PlatformVersion string `json:"platformVersion"`
	// the name of the current celluar network carrier
	CarrierName string `json:"carrierName"`
	// the real size of the default display
	RealDisplaySize string `json:"realDisplaySize"`
	// The logical density of the display in Density Independent Pixel units.
	DisplayDensity int `json:"displayDensity"`
	// available networks
	Networks []networkInfo `json:"networks"`
	// current system locale
	Locale    string `json:"locale"`
	Bluetooth struct {
		State string `json:"state"`
	} `json:"bluetooth"`
}

type networkCapabilities struct {
	TransportTypes            string `json:"transportTypes"`
	NetworkCapabilities       string `json:"networkCapabilities"`
	LinkUpstreamBandwidthKbps int    `json:"linkUpstreamBandwidthKbps"`
	LinkDownBandwidthKbps     int    `json:"linkDownBandwidthKbps"`
	SignalStrength            int    `json:"signalStrength"`
	SSID                      string `json:"SSID"`
}

type networkInfo struct {
	Type          int                 `json:"type"`
	TypeName      string              `json:"typeName"`
	Subtype       int                 `json:"subtype"`
	SubtypeName   string              `json:"subtypeName"`
	IsConnected   bool                `json:"isConnected"`
	DetailedState string              `json:"detailedState"`
	State         string              `json:"state"`
	ExtraInfo     string              `json:"extraInfo"`
	IsAvailable   bool                `json:"isAvailable"`
	IsRoaming     bool                `json:"isRoaming"`
	IsFailover    bool                `json:"isFailover"`
	Capabilities  networkCapabilities `json:"capabilities"`
}

type Location struct {
	AuthorizationStatus int     `json:"authorizationStatus"`
	Longitude           float64 `json:"longitude"`
	Latitude            float64 `json:"latitude"`
	Altitude            float64 `json:"altitude"`
}

type BatteryInfo struct {
	// Battery level in range [0.0, 1.0], where 1.0 means 100% charge.
	Level float64 `json:"level"`

	// Battery state ( 1: on battery, discharging; 2: plugged in, less than 100%, 3: plugged in, at 100% )
	State BatteryState `json:"state"`

	Status BatteryStatus `json:"status"`
}

type BatteryState int

const (
	_                                  = iota
	BatteryStateUnplugged BatteryState = iota // on battery, discharging
	BatteryStateCharging                      // plugged in, less than 100%
	BatteryStateFull                          // plugged in, at 100%
)

func (v BatteryState) String() string {
	switch v {
	case BatteryStateUnplugged:
		return "On battery, discharging"
	case BatteryStateCharging:
		return "Plugged in, less than 100%"
	case BatteryStateFull:
		return "Plugged in, at 100%"
	default:
		return "UNKNOWN"
	}
}

type BatteryStatus int

const (
	_                                  = iota
	BatteryStatusUnknown BatteryStatus = iota
	BatteryStatusCharging
	BatteryStatusDischarging
	BatteryStatusNotCharging
	BatteryStatusFull
)

func (bs BatteryStatus) String() string {
	switch bs {
	case BatteryStatusUnknown:
		return "unknown"
	case BatteryStatusCharging:
		return "charging"
	case BatteryStatusDischarging:
		return "discharging"
	case BatteryStatusNotCharging:
		return "not charging"
	case BatteryStatusFull:
		return "full"
	default:
		return fmt.Sprintf("unknown status code (%d)", bs)
	}
}

// DeviceButton A physical button on a device.
type DeviceButton string

const (
	DeviceButtonHome       DeviceButton = "home"
	DeviceButtonVolumeUp   DeviceButton = "volumeUp"
	DeviceButtonVolumeDown DeviceButton = "volumeDown"
	DeviceButtonEnter      DeviceButton = "enter" // use "\n" for ios
	DeviceButtonBack       DeviceButton = "back"  // android only
)

type NotificationType string

const (
	NotificationTypePlain  NotificationType = "plain"
	NotificationTypeDarwin NotificationType = "darwin"
)

type Orientation string

const (
	// OrientationPortrait Device oriented vertically, home button on the bottom
	OrientationPortrait Orientation = "PORTRAIT"

	// OrientationPortraitUpsideDown Device oriented vertically, home button on the top
	OrientationPortraitUpsideDown Orientation = "UIA_DEVICE_ORIENTATION_PORTRAIT_UPSIDEDOWN"

	// OrientationLandscapeLeft Device oriented horizontally, home button on the right
	OrientationLandscapeLeft Orientation = "LANDSCAPE"

	// OrientationLandscapeRight Device oriented horizontally, home button on the left
	OrientationLandscapeRight Orientation = "UIA_DEVICE_ORIENTATION_LANDSCAPERIGHT"
)

type Rotation struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

type Direction string

const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

// TouchEvent represents a single touch event with all its properties
type TouchEvent struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	DeviceID  int     `json:"deviceId"`
	Pressure  float64 `json:"pressure"`
	Size      float64 `json:"size"`
	RawX      float64 `json:"rawX"`
	RawY      float64 `json:"rawY"`
	DownTime  int64   `json:"downTime"`
	EventTime int64   `json:"eventTime"`
	ToolType  int     `json:"toolType"`
	Flag      int     `json:"flag"`
	Action    int     `json:"action"`
}
