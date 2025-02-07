package uixt

import (
	"fmt"
	"math"
)

type DeviceStatus struct {
	Message string `json:"message"`
	State   string `json:"state"`
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
	Ready bool `json:"ready"`
	Build struct {
		Time                    string `json:"time"`
		ProductBundleIdentifier string `json:"productBundleIdentifier"`
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

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (s Size) IsNil() bool {
	return s.Width == 0 && s.Height == 0
}

type Screen struct {
	StatusBarSize Size    `json:"statusBarSize"`
	Scale         float64 `json:"scale"`
}

type AppInfo struct {
	Name string `json:"name,omitempty"`
	AppBaseInfo
}

type WindowInfo struct {
	PackageName string `json:"packageName,omitempty"`
	Activity    string `json:"activity,omitempty"`
}

type AppBaseInfo struct {
	Pid            int         `json:"pid,omitempty"`
	BundleId       string      `json:"bundleId,omitempty"`       // ios package name
	ViewController string      `json:"viewController,omitempty"` // ios view controller
	PackageName    string      `json:"packageName,omitempty"`    // android package name
	Activity       string      `json:"activity,omitempty"`       // android activity
	VersionName    string      `json:"versionName,omitempty"`
	VersionCode    interface{} `json:"versionCode,omitempty"` // int or string
	AppName        string      `json:"appName,omitempty"`
	AppPath        string      `json:"appPath,omitempty"`
	AppMD5         string      `json:"appMD5,omitempty"`
	// AppIcon        string `json:"appIcon,omitempty"`
}

type AppState int

const (
	AppStateNotRunning AppState = 1 << iota
	AppStateRunningBack
	AppStateRunningFront
)

func (v AppState) String() string {
	switch v {
	case AppStateNotRunning:
		return "Not Running"
	case AppStateRunningBack:
		return "Running (Back)"
	case AppStateRunningFront:
		return "Running (Front)"
	default:
		return "UNKNOWN"
	}
}

// PasteboardType The type of the item on the pasteboard.
type PasteboardType string

const (
	PasteboardTypePlaintext PasteboardType = "plaintext"
	PasteboardTypeImage     PasteboardType = "image"
	PasteboardTypeUrl       PasteboardType = "url"
)

const (
	TextBackspace string = "\u0008"
	TextDelete    string = "\u007F"
)

// DeviceButton A physical button on an iOS device.
type DeviceButton string

const (
	DeviceButtonHome       DeviceButton = "home"
	DeviceButtonVolumeUp   DeviceButton = "volumeUp"
	DeviceButtonVolumeDown DeviceButton = "volumeDown"
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

type Condition func(wd IDriver) (bool, error)

type Direction string

const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

type Point struct {
	X int `json:"x"` // upper left X coordinate of selected element
	Y int `json:"y"` // upper left Y coordinate of selected element
}

type PointF struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p PointF) IsIdentical(p2 PointF) bool {
	// set the coordinate precision to 1 pixel
	return math.Abs(p.X-p2.X) < 1 && math.Abs(p.Y-p2.Y) < 1
}
