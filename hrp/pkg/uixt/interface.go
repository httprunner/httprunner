package uixt

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
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

func (caps Capabilities) WithAppLaunchOption(launchOpt AppLaunchOption) Capabilities {
	for k, v := range launchOpt {
		caps[k] = v
	}
	return caps
}

// WithDefaultAlertAction
func (caps Capabilities) WithDefaultAlertAction(alertAction AlertAction) Capabilities {
	caps["defaultAlertAction"] = alertAction
	return caps
}

// WithMaxTypingFrequency
//  Defaults to `60`.
func (caps Capabilities) WithMaxTypingFrequency(n int) Capabilities {
	if n <= 0 {
		n = 60
	}
	caps["maxTypingFrequency"] = n
	return caps
}

// WithWaitForIdleTimeout
//  Defaults to `10`
func (caps Capabilities) WithWaitForIdleTimeout(second float64) Capabilities {
	caps["waitForIdleTimeout"] = second
	return caps
}

// WithShouldUseTestManagerForVisibilityDetection If set to YES will ask TestManagerDaemon for element visibility
//  Defaults to  `false`
func (caps Capabilities) WithShouldUseTestManagerForVisibilityDetection(b bool) Capabilities {
	caps["shouldUseTestManagerForVisibilityDetection"] = b
	return caps
}

// WithShouldUseCompactResponses If set to YES will use compact (standards-compliant) & faster responses
//  Defaults to `true`
func (caps Capabilities) WithShouldUseCompactResponses(b bool) Capabilities {
	caps["shouldUseCompactResponses"] = b
	return caps
}

// WithElementResponseAttributes If shouldUseCompactResponses == NO,
// is the comma-separated list of fields to return with each element.
//  Defaults to `type,label`.
func (caps Capabilities) WithElementResponseAttributes(s string) Capabilities {
	caps["elementResponseAttributes"] = s
	return caps
}

// WithShouldUseSingletonTestManager
//  Defaults to `true`
func (caps Capabilities) WithShouldUseSingletonTestManager(b bool) Capabilities {
	caps["shouldUseSingletonTestManager"] = b
	return caps
}

// WithDisableAutomaticScreenshots
//  Defaults to `true`
func (caps Capabilities) WithDisableAutomaticScreenshots(b bool) Capabilities {
	caps["disableAutomaticScreenshots"] = b
	return caps
}

// WithShouldTerminateApp
//  Defaults to `true`
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
	Pid      int    `json:"pid"`
	BundleId string `json:"bundleId"`
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

// AppLaunchOption Configure app launch parameters
type AppLaunchOption map[string]interface{}

func NewAppLaunchOption() AppLaunchOption {
	return make(AppLaunchOption)
}

func (opt AppLaunchOption) WithBundleId(bundleId string) AppLaunchOption {
	opt["bundleId"] = bundleId
	return opt
}

// WithShouldWaitForQuiescence whether to wait for quiescence on application startup
//  Defaults to `true`
func (opt AppLaunchOption) WithShouldWaitForQuiescence(b bool) AppLaunchOption {
	opt["shouldWaitForQuiescence"] = b
	return opt
}

// WithArguments The optional array of application command line arguments.
// The arguments are going to be applied if the application was not running before.
func (opt AppLaunchOption) WithArguments(args []string) AppLaunchOption {
	opt["arguments"] = args
	return opt
}

// WithEnvironment The optional dictionary of environment variables for the application, which is going to be executed.
// The environment variables are going to be applied if the application was not running before.
func (opt AppLaunchOption) WithEnvironment(env map[string]string) AppLaunchOption {
	opt["environment"] = env
	return opt
}

func (opt AppLaunchOption) WithBySelector(bySelector ...BySelector) AppLaunchOption {
	opt["bySelector"] = bySelector
	return opt
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
//  only `xml` is supported.
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

const (
	// legacyWebElementIdentifier is the string constant used in the old
	// WebDriver JSON protocol that is the key for the map that contains an
	// unique element identifier.
	legacyWebElementIdentifier = "ELEMENT"

	// webElementIdentifier is the string constant defined by the W3C
	// specification that is the key for the map that contains a unique element identifier.
	webElementIdentifier = "element-6066-11e4-a52e-4f735466cecf"
)

func elementIDFromValue(val map[string]string) string {
	for _, key := range []string{webElementIdentifier, legacyWebElementIdentifier} {
		if v, ok := val[key]; ok && v != "" {
			return v
		}
	}
	return ""
}

// performance ranking: class name > accessibility id > link text > predicate > class chain > xpath
type BySelector struct {
	ClassName ElementType `json:"class name"`

	// isSearchByIdentifier
	Name            string `json:"name"`
	Id              string `json:"id"`
	AccessibilityId string `json:"accessibility id"`

	// partialSearch
	LinkText        ElementAttribute `json:"link text"`
	PartialLinkText ElementAttribute `json:"partial link text"`
	// partialSearch

	Predicate string `json:"predicate string"`

	ClassChain string `json:"class chain"`

	XPath string `json:"xpath"` // not recommended, it's slow because it is not supported by XCTest natively

	// Set the search criteria to match the given resource ResourceIdID.
	ResourceIdID string `json:"id"`
	// Set the search criteria to match the content-description property for a widget.
	ContentDescription string `json:"accessibility id"`

	UiAutomator string `json:"-android uiautomator"`
}

func (wl BySelector) getUsingAndValue() (using, value string) {
	vBy := reflect.ValueOf(wl)
	tBy := reflect.TypeOf(wl)
	for i := 0; i < vBy.NumField(); i++ {
		vi := vBy.Field(i).Interface()
		switch vi := vi.(type) {
		case ElementType:
			value = vi.String()
		case string:
			value = vi
		case ElementAttribute:
			value = vi.String()
		}
		if value != "" && value != "UNKNOWN" {
			using = tBy.Field(i).Tag.Get("json")
			return
		}
	}
	return
}

func (by BySelector) getMethodAndSelector() (method, selector string) {
	vBy := reflect.ValueOf(by)
	tBy := reflect.TypeOf(by)
	for i := 0; i < vBy.NumField(); i++ {
		vi := vBy.Field(i).Interface()
		// switch vi := vi.(type) {
		// case string:
		// 	selector = vi
		// }
		selector = vi.(string)
		if selector != "" && selector != "UNKNOWN" {
			method = tBy.Field(i).Tag.Get("json")
			return
		}
	}
	return
}

type ElementAttribute map[string]interface{}

func (ea ElementAttribute) String() string {
	for k, v := range ea {
		switch v := v.(type) {
		case bool:
			return k + "=" + strconv.FormatBool(v)
		case string:
			return k + "=" + v
		default:
			return k + "=" + fmt.Sprintf("%v", v)
		}
	}
	return "UNKNOWN"
}

func (ea ElementAttribute) getAttributeName() string {
	for k := range ea {
		return k
	}
	return "UNKNOWN"
}

func NewElementAttribute() ElementAttribute {
	return make(ElementAttribute)
}

// WithUID Element's unique identifier
func (ea ElementAttribute) WithUID(uid string) ElementAttribute {
	ea["UID"] = uid
	return ea
}

// WithAccessibilityContainer Whether element is an accessibility container
// (contains children of any depth that are accessible)
func (ea ElementAttribute) WithAccessibilityContainer(b bool) ElementAttribute {
	ea["accessibilityContainer"] = b
	return ea
}

// WithAccessible Whether element is accessible
func (ea ElementAttribute) WithAccessible(b bool) ElementAttribute {
	ea["accessible"] = b
	return ea
}

// WithEnabled Whether element is enabled
func (ea ElementAttribute) WithEnabled(b bool) ElementAttribute {
	ea["enabled"] = b
	return ea
}

// WithLabel Element's label
func (ea ElementAttribute) WithLabel(s string) ElementAttribute {
	ea["label"] = s
	return ea
}

// WithName Element's name
func (ea ElementAttribute) WithName(s string) ElementAttribute {
	ea["name"] = s
	return ea
}

// WithSelected Element's selected state
func (ea ElementAttribute) WithSelected(b bool) ElementAttribute {
	ea["selected"] = b
	return ea
}

// WithType Element's type
func (ea ElementAttribute) WithType(elemType ElementType) ElementAttribute {
	ea["type"] = elemType
	return ea
}

// WithValue Element's value
func (ea ElementAttribute) WithValue(s string) ElementAttribute {
	ea["value"] = s
	return ea
}

// WithVisible
//
// Whether element is visible
func (ea ElementAttribute) WithVisible(b bool) ElementAttribute {
	ea["visible"] = b
	return ea
}

func (et ElementType) String() string {
	vBy := reflect.ValueOf(et)
	tBy := reflect.TypeOf(et)
	for i := 0; i < vBy.NumField(); i++ {
		if vBy.Field(i).Bool() {
			return tBy.Field(i).Tag.Get("json")
		}
	}
	return "UNKNOWN"
}

// ElementType
// !!! This mapping should be updated if there are changes after each new XCTest release"`
type ElementType struct {
	Any                bool `json:"XCUIElementTypeAny"`
	Other              bool `json:"XCUIElementTypeOther"`
	Application        bool `json:"XCUIElementTypeApplication"`
	Group              bool `json:"XCUIElementTypeGroup"`
	Window             bool `json:"XCUIElementTypeWindow"`
	Sheet              bool `json:"XCUIElementTypeSheet"`
	Drawer             bool `json:"XCUIElementTypeDrawer"`
	Alert              bool `json:"XCUIElementTypeAlert"`
	Dialog             bool `json:"XCUIElementTypeDialog"`
	Button             bool `json:"XCUIElementTypeButton"`
	RadioButton        bool `json:"XCUIElementTypeRadioButton"`
	RadioGroup         bool `json:"XCUIElementTypeRadioGroup"`
	CheckBox           bool `json:"XCUIElementTypeCheckBox"`
	DisclosureTriangle bool `json:"XCUIElementTypeDisclosureTriangle"`
	PopUpButton        bool `json:"XCUIElementTypePopUpButton"`
	ComboBox           bool `json:"XCUIElementTypeComboBox"`
	MenuButton         bool `json:"XCUIElementTypeMenuButton"`
	ToolbarButton      bool `json:"XCUIElementTypeToolbarButton"`
	Popover            bool `json:"XCUIElementTypePopover"`
	Keyboard           bool `json:"XCUIElementTypeKeyboard"`
	Key                bool `json:"XCUIElementTypeKey"`
	NavigationBar      bool `json:"XCUIElementTypeNavigationBar"`
	TabBar             bool `json:"XCUIElementTypeTabBar"`
	TabGroup           bool `json:"XCUIElementTypeTabGroup"`
	Toolbar            bool `json:"XCUIElementTypeToolbar"`
	StatusBar          bool `json:"XCUIElementTypeStatusBar"`
	Table              bool `json:"XCUIElementTypeTable"`
	TableRow           bool `json:"XCUIElementTypeTableRow"`
	TableColumn        bool `json:"XCUIElementTypeTableColumn"`
	Outline            bool `json:"XCUIElementTypeOutline"`
	OutlineRow         bool `json:"XCUIElementTypeOutlineRow"`
	Browser            bool `json:"XCUIElementTypeBrowser"`
	CollectionView     bool `json:"XCUIElementTypeCollectionView"`
	Slider             bool `json:"XCUIElementTypeSlider"`
	PageIndicator      bool `json:"XCUIElementTypePageIndicator"`
	ProgressIndicator  bool `json:"XCUIElementTypeProgressIndicator"`
	ActivityIndicator  bool `json:"XCUIElementTypeActivityIndicator"`
	SegmentedControl   bool `json:"XCUIElementTypeSegmentedControl"`
	Picker             bool `json:"XCUIElementTypePicker"`
	PickerWheel        bool `json:"XCUIElementTypePickerWheel"`
	Switch             bool `json:"XCUIElementTypeSwitch"`
	Toggle             bool `json:"XCUIElementTypeToggle"`
	Link               bool `json:"XCUIElementTypeLink"`
	Image              bool `json:"XCUIElementTypeImage"`
	Icon               bool `json:"XCUIElementTypeIcon"`
	SearchField        bool `json:"XCUIElementTypeSearchField"`
	ScrollView         bool `json:"XCUIElementTypeScrollView"`
	ScrollBar          bool `json:"XCUIElementTypeScrollBar"`
	StaticText         bool `json:"XCUIElementTypeStaticText"`
	TextField          bool `json:"XCUIElementTypeTextField"`
	SecureTextField    bool `json:"XCUIElementTypeSecureTextField"`
	DatePicker         bool `json:"XCUIElementTypeDatePicker"`
	TextView           bool `json:"XCUIElementTypeTextView"`
	Menu               bool `json:"XCUIElementTypeMenu"`
	MenuItem           bool `json:"XCUIElementTypeMenuItem"`
	MenuBar            bool `json:"XCUIElementTypeMenuBar"`
	MenuBarItem        bool `json:"XCUIElementTypeMenuBarItem"`
	Map                bool `json:"XCUIElementTypeMap"`
	WebView            bool `json:"XCUIElementTypeWebView"`
	IncrementArrow     bool `json:"XCUIElementTypeIncrementArrow"`
	DecrementArrow     bool `json:"XCUIElementTypeDecrementArrow"`
	Timeline           bool `json:"XCUIElementTypeTimeline"`
	RatingIndicator    bool `json:"XCUIElementTypeRatingIndicator"`
	ValueIndicator     bool `json:"XCUIElementTypeValueIndicator"`
	SplitGroup         bool `json:"XCUIElementTypeSplitGroup"`
	Splitter           bool `json:"XCUIElementTypeSplitter"`
	RelevanceIndicator bool `json:"XCUIElementTypeRelevanceIndicator"`
	ColorWell          bool `json:"XCUIElementTypeColorWell"`
	HelpTag            bool `json:"XCUIElementTypeHelpTag"`
	Matte              bool `json:"XCUIElementTypeMatte"`
	DockItem           bool `json:"XCUIElementTypeDockItem"`
	Ruler              bool `json:"XCUIElementTypeRuler"`
	RulerMarker        bool `json:"XCUIElementTypeRulerMarker"`
	Grid               bool `json:"XCUIElementTypeGrid"`
	LevelIndicator     bool `json:"XCUIElementTypeLevelIndicator"`
	Cell               bool `json:"XCUIElementTypeCell"`
	LayoutArea         bool `json:"XCUIElementTypeLayoutArea"`
	LayoutItem         bool `json:"XCUIElementTypeLayoutItem"`
	Handle             bool `json:"XCUIElementTypeHandle"`
	Stepper            bool `json:"XCUIElementTypeStepper"`
	Tab                bool `json:"XCUIElementTypeTab"`
	TouchBar           bool `json:"XCUIElementTypeTouchBar"`
	StatusItem         bool `json:"XCUIElementTypeStatusItem"`
	EditText           bool `json:"android.widget.EditText"`
}

// ProtectedResource A system resource that requires user authorization to access.
type ProtectedResource int

// https://developer.apple.com/documentation/xctest/xcuiprotectedresource?language=objc
const (
	ProtectedResourceContacts               ProtectedResource = 1
	ProtectedResourceCalendar               ProtectedResource = 2
	ProtectedResourceReminders              ProtectedResource = 3
	ProtectedResourcePhotos                 ProtectedResource = 4
	ProtectedResourceMicrophone             ProtectedResource = 5
	ProtectedResourceCamera                 ProtectedResource = 6
	ProtectedResourceMediaLibrary           ProtectedResource = 7
	ProtectedResourceHomeKit                ProtectedResource = 8
	ProtectedResourceSystemRootDirectory    ProtectedResource = 0x40000000
	ProtectedResourceUserDesktopDirectory   ProtectedResource = 0x40000001
	ProtectedResourceUserDownloadsDirectory ProtectedResource = 0x40000002
	ProtectedResourceUserDocumentsDirectory ProtectedResource = 0x40000003
	ProtectedResourceBluetooth              ProtectedResource = -0x40000000
	ProtectedResourceKeyboardNetwork        ProtectedResource = -0x40000001
	ProtectedResourceLocation               ProtectedResource = -0x40000002
	ProtectedResourceHealth                 ProtectedResource = -0x40000003
)

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

type Rect struct {
	Point
	Size
}

type DataOptions struct {
	Data                map[string]interface{} // configurations used by ios/android driver
	Scope               []int                  // used by ocr to get text position in the scope
	Offset              []int                  // used to tap offset of point
	Index               int                    // index of the target element, should start from 1
	IgnoreNotFoundError bool                   // ignore error if target element not found
	MaxRetryTimes       int                    // max retry times if target element not found
	Interval            float64                // interval between retries in seconds
	ScreenShotFilename  string                 // turn on screenshot and specify file name
}

type DataOption func(data *DataOptions)

func WithCustomOption(key string, value interface{}) DataOption {
	return func(data *DataOptions) {
		data.Data[key] = value
	}
}

func WithDataPressDuration(duration float64) DataOption {
	return func(data *DataOptions) {
		data.Data["duration"] = duration
	}
}

func WithDataSteps(steps int) DataOption {
	return func(data *DataOptions) {
		data.Data["steps"] = steps
	}
}

func WithDataFrequency(frequency int) DataOption {
	return func(data *DataOptions) {
		data.Data["frequency"] = frequency
	}
}

func WithDataIndex(index int) DataOption {
	return func(data *DataOptions) {
		data.Index = index
	}
}

func WithDataScope(x1, x2, y1, y2 int) DataOption {
	return func(data *DataOptions) {
		data.Scope = []int{x1, x2, y1, y2}
	}
}

func WithDataOffset(offsetX, offsetY int) DataOption {
	return func(data *DataOptions) {
		data.Offset = []int{offsetX, offsetY}
	}
}

func WithDataIdentifier(identifier string) DataOption {
	if identifier == "" {
		return func(data *DataOptions) {}
	}
	return func(data *DataOptions) {
		data.Data["log"] = map[string]interface{}{
			"enable": true,
			"data":   identifier,
		}
	}
}

func WithDataIgnoreNotFoundError(ignoreError bool) DataOption {
	return func(data *DataOptions) {
		data.IgnoreNotFoundError = ignoreError
	}
}

func WithDataMaxRetryTimes(maxRetryTimes int) DataOption {
	return func(data *DataOptions) {
		data.MaxRetryTimes = maxRetryTimes
	}
}

func WithDataWaitTime(sec float64) DataOption {
	return func(data *DataOptions) {
		data.Interval = sec
	}
}

func WithScreenShot(fileName ...string) DataOption {
	return func(data *DataOptions) {
		if len(fileName) > 0 {
			data.ScreenShotFilename = fileName[0]
		} else {
			data.ScreenShotFilename = fmt.Sprintf("screenshot_%d", time.Now().Unix())
		}
	}
}

func NewDataOptions(options ...DataOption) *DataOptions {
	dataOptions := &DataOptions{
		Data: make(map[string]interface{}),
	}
	for _, option := range options {
		option(dataOptions)
	}

	if len(dataOptions.Scope) == 0 {
		dataOptions.Scope = []int{0, 0, math.MaxInt64, math.MaxInt64} // default scope
	}
	return dataOptions
}

func NewData(data map[string]interface{}, options ...DataOption) map[string]interface{} {
	dataOptions := NewDataOptions(options...)

	// merge with data options
	for k, v := range dataOptions.Data {
		data[k] = v
	}

	// handle point offset
	if len(dataOptions.Offset) == 2 {
		if x, ok := data["x"]; ok {
			xf, _ := builtin.Interface2Float64(x)
			data["x"] = xf + float64(dataOptions.Offset[0])
		}
		if y, ok := data["y"]; ok {
			yf, _ := builtin.Interface2Float64(y)
			data["y"] = yf + float64(dataOptions.Offset[1])
		}
	}

	// add default options
	if _, ok := data["steps"]; !ok {
		data["steps"] = 12 // default steps
	}

	if _, ok := data["duration"]; !ok {
		data["duration"] = 0 // default duration
	}

	if _, ok := data["frequency"]; !ok {
		data["frequency"] = 60 // default frequency
	}

	if _, ok := data["isReplace"]; !ok {
		data["isReplace"] = true // default true
	}

	return data
}

// current implemeted device: IOSDevice, AndroidDevice
type Device interface {
	UUID() string
	NewDriver(capabilities Capabilities) (driverExt *DriverExt, err error)
}

// WebDriver defines methods supported by WebDriver drivers.
type WebDriver interface {
	// NewSession starts a new session and returns the SessionInfo.
	NewSession(capabilities Capabilities) (SessionInfo, error)

	ActiveSession() (SessionInfo, error)
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
	ActiveAppInfo() (AppInfo, error)
	// ActiveAppsList Retrieves the information about the currently active apps
	ActiveAppsList() ([]AppBaseInfo, error)
	// AppState Get the state of the particular application in scope of the current session.
	// !This method is only returning reliable results since Xcode9 SDK
	AppState(bundleId string) (AppState, error)

	// IsLocked Checks if the screen is locked or not.
	IsLocked() (bool, error)
	// Unlock Forces the device under test to unlock.
	// An immediate return will happen if the device is already unlocked
	// and an error is going to be thrown if the screen has not been unlocked after the timeout.
	Unlock() error
	// Lock Forces the device under test to switch to the lock screen.
	// An immediate return will happen if the device is already locked
	// and an error is going to be thrown if the screen has not been locked after the timeout.
	Lock() error

	// Homescreen Forces the device under test to switch to the home screen
	Homescreen() error

	// AlertText Returns alert's title and description separated by new lines
	AlertText() (string, error)
	// AlertButtons Gets the labels of the buttons visible in the alert
	AlertButtons() ([]string, error)
	// AlertAccept Accepts alert, if present
	AlertAccept(label ...string) error
	// AlertDismiss Dismisses alert, if present
	AlertDismiss(label ...string) error
	// AlertSendKeys Types a text into an input inside the alert container, if it is present
	AlertSendKeys(text string) error

	// AppLaunch Launch an application with given bundle identifier in scope of current session.
	// !This method is only available since Xcode9 SDK
	AppLaunch(bundleId string, launchOpt ...AppLaunchOption) error
	// AppLaunchUnattached Launch the app with the specified bundle ID.
	AppLaunchUnattached(bundleId string) error
	// AppTerminate Terminate an application with the given bundle id.
	// Either `true` if the app has been successfully terminated or `false` if it was not running
	AppTerminate(bundleId string) (bool, error)
	// AppActivate Activate an application with given bundle identifier in scope of current session.
	// !This method is only available since Xcode9 SDK
	AppActivate(bundleId string) error
	// AppDeactivate Deactivates application for given time and then activate it again
	//  The minimum application switch wait is 3 seconds
	AppDeactivate(second float64) error

	// AppAuthReset Resets the authorization status for a protected resource. Available since Xcode 11.4
	AppAuthReset(ProtectedResource) error

	// StartCamera Starts a new camera for recording
	StartCamera() error
	// StopCamera Stops the camera for recording
	StopCamera() error

	// Tap Sends a tap event at the coordinate.
	Tap(x, y int, options ...DataOption) error
	TapFloat(x, y float64, options ...DataOption) error

	// DoubleTap Sends a double tap event at the coordinate.
	DoubleTap(x, y int) error
	DoubleTapFloat(x, y float64) error

	// TouchAndHold Initiates a long-press gesture at the coordinate, holding for the specified duration.
	//  second: The default value is 1
	TouchAndHold(x, y int, second ...float64) error
	TouchAndHoldFloat(x, y float64, second ...float64) error

	// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
	// WithPressDurationOption option can be used to set pressForDuration (default to 1 second).
	Drag(fromX, fromY, toX, toY int, options ...DataOption) error
	DragFloat(fromX, fromY, toX, toY float64, options ...DataOption) error

	// Swipe works like Drag, but `pressForDuration` value is 0
	Swipe(fromX, fromY, toX, toY int, options ...DataOption) error
	SwipeFloat(fromX, fromY, toX, toY float64, options ...DataOption) error

	ForceTouch(x, y int, pressure float64, second ...float64) error
	ForceTouchFloat(x, y, pressure float64, second ...float64) error

	// PerformW3CActions Perform complex touch action in scope of the current application.
	PerformW3CActions(actions *W3CActions) error
	PerformAppiumTouchActions(touchActs *TouchActions) error

	// SetPasteboard Sets data to the general pasteboard
	SetPasteboard(contentType PasteboardType, content string) error
	// GetPasteboard Gets the data contained in the general pasteboard.
	//  It worked when `WDA` was foreground. https://github.com/appium/WebDriverAgent/issues/330
	GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error)

	// SendKeys Types a string into active element. There must be element with keyboard focus,
	// otherwise an error is raised.
	// WithFrequency option can be used to set frequency of typing (letters per sec). The default value is 60
	SendKeys(text string, options ...DataOption) error

	// Input works like SendKeys
	Input(text string, options ...DataOption) error

	// KeyboardDismiss Tries to dismiss the on-screen keyboard
	KeyboardDismiss(keyNames ...string) error

	// PressButton Presses the corresponding hardware button on the device
	PressButton(devBtn DeviceButton) error

	// PressBack Presses the back button
	PressBack(options ...DataOption) error

	// IOHIDEvent Emulated triggering of the given low-level IOHID device event.
	//  duration: The event duration in float seconds (XCTest uses 0.005 for a single press event)
	IOHIDEvent(pageID EventPageID, usageID EventUsageID, duration ...float64) error

	// ExpectNotification Creates an expectation that is fulfilled when an expected Notification is received
	ExpectNotification(notifyName string, notifyType NotificationType, second ...int) error

	// SiriActivate Activates Siri service voice recognition with the given text to parse
	SiriActivate(text string) error
	// SiriOpenUrl Opens the particular url scheme using Siri voice recognition helpers.
	// !This will only work since XCode 8.3/iOS 10.3
	//  It doesn't actually work, right?
	SiriOpenUrl(url string) error

	Orientation() (Orientation, error)
	// SetOrientation Sets requested device interface orientation.
	SetOrientation(Orientation) error

	Rotation() (Rotation, error)
	// SetRotation Sets the devices orientation to the rotation passed.
	SetRotation(Rotation) error

	// MatchTouchID Matches or mismatches TouchID request
	MatchTouchID(isMatch bool) error

	// ActiveElement Returns the element, which currently holds the keyboard input focus or nil if there are no such elements.
	ActiveElement() (WebElement, error)
	FindElement(by BySelector) (WebElement, error)
	FindElements(by BySelector) ([]WebElement, error)

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

	// WaitWithTimeoutAndInterval waits for the condition to evaluate to true.
	WaitWithTimeoutAndInterval(condition Condition, timeout, interval time.Duration) error
	// WaitWithTimeout works like WaitWithTimeoutAndInterval, but with default polling interval.
	WaitWithTimeout(condition Condition, timeout time.Duration) error
	// Wait works like WaitWithTimeoutAndInterval, but using the default timeout and polling interval.
	Wait(condition Condition) error

	// Close inner connections properly
	Close() error

	// triggers the log capture and returns the log entries
	StartCaptureLog(identifier ...string) (err error)
	StopCaptureLog() (result interface{}, err error)
}

// WebElement defines method supported by web elements.
type WebElement interface {
	// Click Waits for element to become stable (not move) and performs sync tap on element.
	Click() error
	// SendKeys Types a text into element. It will try to activate keyboard on element,
	// if element has no keyboard focus.
	//  frequency: Frequency of typing (letters per sec). The default value is 60
	SendKeys(text string, options ...DataOption) error
	// Clear Clears text on element. It will try to activate keyboard on element,
	// if element has no keyboard focus.
	Clear() error

	// Tap Waits for element to become stable (not move) and performs sync tap on element,
	// relative to the current element position
	Tap(x, y int) error
	TapFloat(x, y float64) error

	// DoubleTap Sends a double tap event to a hittable point computed for the element.
	DoubleTap() error

	// TouchAndHold Sends a long-press gesture to a hittable point computed for the element,
	// holding for the specified duration.
	//  second: The default value is 1
	TouchAndHold(second ...float64) error
	// TwoFingerTap Sends a two finger tap event to a hittable point computed for the element.
	TwoFingerTap() error
	// TapWithNumberOfTaps Sends one or more taps with one or more touch points.
	TapWithNumberOfTaps(numberOfTaps, numberOfTouches int) error
	// ForceTouch Waits for element to become stable (not move) and performs sync force touch on element.
	//  second: The default value is 1
	ForceTouch(pressure float64, second ...float64) error
	// ForceTouchFloat works like ForceTouch, but relative to the current element position
	ForceTouchFloat(x, y, pressure float64, second ...float64) error

	// Drag Initiates a press-and-hold gesture at the coordinate, then drags to another coordinate.
	// relative to the current element position
	//  pressForDuration: The default value is 1 second.
	Drag(fromX, fromY, toX, toY int, pressForDuration ...float64) error
	DragFloat(fromX, fromY, toX, toY float64, pressForDuration ...float64) error

	// Swipe works like Drag, but `pressForDuration` value is 0.
	// relative to the current element position
	Swipe(fromX, fromY, toX, toY int) error
	SwipeFloat(fromX, fromY, toX, toY float64) error
	// SwipeDirection Performs swipe gesture on the element.
	//  velocity: swipe speed in pixels per second. Custom velocity values are only supported since Xcode SDK 11.4.
	SwipeDirection(direction Direction, velocity ...float64) error

	// Pinch Sends a pinching gesture with two touches.
	//  scale: The scale of the pinch gesture. Use a scale between 0 and 1 to "pinch close" or zoom out
	//  and a scale greater than 1 to "pinch open" or zoom in.
	//  velocity: The velocity of the pinch in scale factor per second.
	Pinch(scale, velocity float64) error
	PinchToZoomOutByW3CAction(scale ...float64) error

	// Rotate Sends a rotation gesture with two touches.
	//  rotation: The rotation of the gesture in radians.
	//  velocity: The velocity of the rotation gesture in radians per second.
	Rotate(rotation float64, velocity ...float64) error

	// PickerWheelSelect
	//  offset: The default value is 2
	PickerWheelSelect(order PickerWheelOrder, offset ...int) error

	ScrollElementByName(name string) error
	ScrollElementByPredicate(predicate string) error
	ScrollToVisible() error
	// ScrollDirection
	//  distance: The default value is 0.5
	ScrollDirection(direction Direction, distance ...float64) error

	FindElement(by BySelector) (element WebElement, err error)
	FindElements(by BySelector) (elements []WebElement, err error)
	FindVisibleCells() (elements []WebElement, err error)

	Rect() (rect Rect, err error)
	Location() (Point, error)
	Size() (Size, error)
	Text() (text string, err error)
	Type() (elemType string, err error)
	IsEnabled() (enabled bool, err error)
	IsDisplayed() (displayed bool, err error)
	IsSelected() (selected bool, err error)
	IsAccessible() (accessible bool, err error)
	IsAccessibilityContainer() (isAccessibilityContainer bool, err error)
	GetAttribute(attr ElementAttribute) (value string, err error)
	UID() (uid string)

	Screenshot() (raw *bytes.Buffer, err error)
}
