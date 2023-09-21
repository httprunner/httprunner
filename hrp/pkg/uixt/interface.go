package uixt

import (
	"bytes"
	"math"
	"strings"
	"time"

	"github.com/httprunner/funplugin"
)

var (
	DefaultWaitTimeout  = 60 * time.Second
	DefaultWaitInterval = 400 * time.Millisecond
)

type AlertAction string

const (
	AlertActionAccept  AlertAction = "accept"
	AlertActionDismiss AlertAction = "dismiss"
)

type Capabilities map[string]interface{}

func NewCapabilities() Capabilities {
	return make(Capabilities)
}

// WithDefaultAlertAction
func (caps Capabilities) WithDefaultAlertAction(alertAction AlertAction) Capabilities {
	caps["defaultAlertAction"] = alertAction
	return caps
}

// WithMaxTypingFrequency
//
//	Defaults to `60`.
func (caps Capabilities) WithMaxTypingFrequency(n int) Capabilities {
	if n <= 0 {
		n = 60
	}
	caps["maxTypingFrequency"] = n
	return caps
}

// WithWaitForIdleTimeout
//
//	Defaults to `10`
func (caps Capabilities) WithWaitForIdleTimeout(second float64) Capabilities {
	caps["waitForIdleTimeout"] = second
	return caps
}

// WithShouldUseTestManagerForVisibilityDetection If set to YES will ask TestManagerDaemon for element visibility
//
//	Defaults to  `false`
func (caps Capabilities) WithShouldUseTestManagerForVisibilityDetection(b bool) Capabilities {
	caps["shouldUseTestManagerForVisibilityDetection"] = b
	return caps
}

// WithShouldUseCompactResponses If set to YES will use compact (standards-compliant) & faster responses
//
//	Defaults to `true`
func (caps Capabilities) WithShouldUseCompactResponses(b bool) Capabilities {
	caps["shouldUseCompactResponses"] = b
	return caps
}

// WithElementResponseAttributes If shouldUseCompactResponses == NO,
// is the comma-separated list of fields to return with each element.
//
//	Defaults to `type,label`.
func (caps Capabilities) WithElementResponseAttributes(s string) Capabilities {
	caps["elementResponseAttributes"] = s
	return caps
}

// WithShouldUseSingletonTestManager
//
//	Defaults to `true`
func (caps Capabilities) WithShouldUseSingletonTestManager(b bool) Capabilities {
	caps["shouldUseSingletonTestManager"] = b
	return caps
}

// WithDisableAutomaticScreenshots
//
//	Defaults to `true`
func (caps Capabilities) WithDisableAutomaticScreenshots(b bool) Capabilities {
	caps["disableAutomaticScreenshots"] = b
	return caps
}

// WithShouldTerminateApp
//
//	Defaults to `true`
func (caps Capabilities) WithShouldTerminateApp(b bool) Capabilities {
	caps["shouldTerminateApp"] = b
	return caps
}

// WithEventloopIdleDelaySec
// Delays the invocation of '-[XCUIApplicationProcess setEventLoopHasIdled:]' by the timer interval passed.
// which is skipped on setting it to zero.
func (caps Capabilities) WithEventloopIdleDelaySec(second float64) Capabilities {
	caps["eventloopIdleDelaySec"] = second
	return caps
}

type SessionInfo struct {
	SessionId    string `json:"sessionId"`
	Capabilities struct {
		Device             string `json:"device"`
		BrowserName        string `json:"browserName"`
		SdkVersion         string `json:"sdkVersion"`
		CFBundleIdentifier string `json:"CFBundleIdentifier"`
	} `json:"capabilities"`
}

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

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Screen struct {
	StatusBarSize Size    `json:"statusBarSize"`
	Scale         float64 `json:"scale"`
}

type AppInfo struct {
	ProcessArguments struct {
		Env  interface{}   `json:"env"`
		Args []interface{} `json:"args"`
	} `json:"processArguments"`
	Name string `json:"name"`
	AppBaseInfo
}

type AppBaseInfo struct {
	Pid            int    `json:"pid,omitempty"`
	BundleId       string `json:"bundleId,omitempty"`       // ios package name
	ViewController string `json:"viewController,omitempty"` // ios view controller
	PackageName    string `json:"packageName,omitempty"`    // android package name
	Activity       string `json:"activity,omitempty"`       // android activity
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

// type KeyboardKeyLabel string
//
// const (
// 	KeyboardKeyReturn = "return"
// )

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

// EventPageID The event page identifier
type EventPageID int

const EventPageIDConsumer EventPageID = 0x0C

// EventUsageID The event usage identifier (usages are defined per-page)
type EventUsageID int

const (
	EventUsageIDCsmrVolumeUp   EventUsageID = 0xE9
	EventUsageIDCsmrVolumeDown EventUsageID = 0xEA
	EventUsageIDCsmrHome       EventUsageID = 0x40
	EventUsageIDCsmrPower      EventUsageID = 0x30
	EventUsageIDCsmrSnapshot   EventUsageID = 0x65 // Power + Home
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

// SourceOption Configure the format or attribute of the Source
type SourceOption map[string]interface{}

func NewSourceOption() SourceOption {
	return make(SourceOption)
}

// WithFormatAsJson Application elements tree in form of json string
func (opt SourceOption) WithFormatAsJson() SourceOption {
	opt["format"] = "json"
	return opt
}

// WithFormatAsXml Application elements tree in form of xml string
func (opt SourceOption) WithFormatAsXml() SourceOption {
	opt["format"] = "xml"
	return opt
}

// WithFormatAsDescription Application elements tree in form of internal XCTest debugDescription string
func (opt SourceOption) WithFormatAsDescription() SourceOption {
	opt["format"] = "description"
	return opt
}

// WithScope Allows to provide XML scope.
//
//	only `xml` is supported.
func (opt SourceOption) WithScope(scope string) SourceOption {
	if vFormat, ok := opt["format"]; ok && vFormat != "xml" {
		return opt
	}
	opt["scope"] = scope
	return opt
}

// WithExcludedAttributes Excludes the given attribute names.
// only `xml` is supported.
func (opt SourceOption) WithExcludedAttributes(attributes []string) SourceOption {
	if vFormat, ok := opt["format"]; ok && vFormat != "xml" {
		return opt
	}
	opt["excluded_attributes"] = strings.Join(attributes, ",")
	return opt
}

type Condition func(wd WebDriver) (bool, error)

type Direction string

const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

type PickerWheelOrder string

const (
	PickerWheelOrderNext     PickerWheelOrder = "next"
	PickerWheelOrderPrevious PickerWheelOrder = "previous"
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

type Rect struct {
	Point
	Size
}

type DriverOptions struct {
	capabilities Capabilities
	plugin       funplugin.IPlugin
}

type DriverOption func(*DriverOptions)

func WithDriverCapabilities(capabilities Capabilities) DriverOption {
	return func(options *DriverOptions) {
		options.capabilities = capabilities
	}
}

func WithDriverPlugin(plugin funplugin.IPlugin) DriverOption {
	return func(options *DriverOptions) {
		options.plugin = plugin
	}
}

// current implemeted device: IOSDevice, AndroidDevice
type Device interface {
	UUID() string // ios udid or android serial
	LogEnabled() bool
	NewDriver(...DriverOption) (driverExt *DriverExt, err error)

	StartPerf() error
	StopPerf() string

	StartPcap() error
	StopPcap() string
}

type ForegroundApp struct {
	PackageName string
	Activity    string
}

// WebDriver defines methods supported by WebDriver drivers.
type WebDriver interface {
	// NewSession starts a new session and returns the SessionInfo.
	NewSession(capabilities Capabilities) (SessionInfo, error)

	// DeleteSession Kills application associated with that session and removes session
	//  1) alertsMonitor disable
	//  2) testedApplicationBundleId terminate
	DeleteSession() error

	Status() (DeviceStatus, error)

	DeviceInfo() (DeviceInfo, error)

	// Location Returns device location data.
	//
	// It requires to configure location access permission by manual.
	// The response of 'latitude', 'longitude' and 'altitude' are always zero (0) without authorization.
	// 'authorizationStatus' indicates current authorization status. '3' is 'Always'.
	// https://developer.apple.com/documentation/corelocation/clauthorizationstatus
	//
	//  Settings -> Privacy -> Location Service -> WebDriverAgent-Runner -> Always
	//
	// The return value could be zero even if the permission is set to 'Always'
	// since the location service needs some time to update the location data.
	Location() (Location, error)
	BatteryInfo() (BatteryInfo, error)
	WindowSize() (Size, error)
	Screen() (Screen, error)
	Scale() (float64, error)

	// GetTimestamp returns the timestamp of the mobile device
	GetTimestamp() (timestamp int64, err error)

	// Homescreen Forces the device under test to switch to the home screen
	Homescreen() error

	// AppLaunch Launch an application with given bundle identifier in scope of current session.
	// !This method is only available since Xcode9 SDK
	AppLaunch(packageName string) error
	// AppTerminate Terminate an application with the given package name.
	// Either `true` if the app has been successfully terminated or `false` if it was not running
	AppTerminate(packageName string) (bool, error)
	// GetForegroundApp returns current foreground app package name and activity name
	GetForegroundApp() (app AppInfo, err error)
	// AssertForegroundApp returns nil if the given package and activity are in foreground
	AssertForegroundApp(packageName string, activityType ...string) error

	// StartCamera Starts a new camera for recording
	StartCamera() error
	// StopCamera Stops the camera for recording
	StopCamera() error

	// Tap Sends a tap event at the coordinate.
	Tap(x, y int, options ...ActionOption) error
	TapFloat(x, y float64, options ...ActionOption) error

	// DoubleTap Sends a double tap event at the coordinate.
	DoubleTap(x, y int) error
	DoubleTapFloat(x, y float64) error

	// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
	//  second: The default value is 1
	TouchAndHold(x, y int, second ...float64) error
	TouchAndHoldFloat(x, y float64, second ...float64) error

	// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
	// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
	Drag(fromX, fromY, toX, toY int, options ...ActionOption) error
	DragFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error

	// Swipe works like Drag, but `pressForDuration` value is 0
	Swipe(fromX, fromY, toX, toY int, options ...ActionOption) error
	SwipeFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error

	// SetPasteboard Sets data to the general pasteboard
	SetPasteboard(contentType PasteboardType, content string) error
	// GetPasteboard Gets the data contained in the general pasteboard.
	//  It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
	GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error)

	// SendKeys Types a string into active element. There must be element with keyboard focus,
	// otherwise an error is raised.
	// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
	SendKeys(text string, options ...ActionOption) error

	// Input works like SendKeys
	Input(text string, options ...ActionOption) error

	// PressButton Presses the corresponding hardware button on the device
	PressButton(devBtn DeviceButton) error

	// PressBack Presses the back button
	PressBack(options ...ActionOption) error

	Screenshot() (*bytes.Buffer, error)

	// Source Return application elements tree
	Source(srcOpt ...SourceOption) (string, error)
	// AccessibleSource Return application elements accessibility tree
	AccessibleSource() (string, error)

	// HealthCheck Health check might modify simulator state so it should only be called in-between testing sessions
	//  Checks health of XCTest by:
	//  1) Querying application for some elements,
	//  2) Triggering some device events.
	HealthCheck() error
	GetAppiumSettings() (map[string]interface{}, error)
	SetAppiumSettings(settings map[string]interface{}) (map[string]interface{}, error)

	IsHealthy() (bool, error)

	// triggers the log capture and returns the log entries
	StartCaptureLog(identifier ...string) (err error)
	StopCaptureLog() (result interface{}, err error)
}
