package uixt

import (
	"bytes"
	"fmt"
	"net"
	"reflect"

	"github.com/electricbubble/gadb"
	"github.com/pkg/errors"
)

var (
	AdbServerHost  = "localhost"
	AdbServerPort  = gadb.AdbServerPort // 5037
	UIA2ServerPort = 6790
	DeviceTempPath = "/data/local/tmp"
)

const forwardToPrefix = "forward-to-"

func InitUIAClient(device *AndroidDevice) (*DriverExt, error) {
	var deviceOptions []AndroidDeviceOption
	if device.SerialNumber != "" {
		deviceOptions = append(deviceOptions, WithSerialNumber(device.SerialNumber))
	}
	if device.IP != "" {
		deviceOptions = append(deviceOptions, WithAdbIP(device.IP))
	}
	if device.Port != 0 {
		deviceOptions = append(deviceOptions, WithAdbPort(device.Port))
	}

	// init uia device
	androidDevice, err := NewAndroidDevice(deviceOptions...)
	if err != nil {
		return nil, err
	}

	driver, err := androidDevice.NewUSBDriver(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init UIA driver")
	}
	fmt.Println(driver)

	var driverExt *DriverExt
	// TODO
	// driverExt, err = Extend(driver)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to extend UIA Driver")
	// }

	return driverExt, nil
}

type AndroidDeviceOption func(*AndroidDevice)

func WithSerialNumber(serial string) AndroidDeviceOption {
	return func(device *AndroidDevice) {
		device.SerialNumber = serial
	}
}

func WithAdbIP(ip string) AndroidDeviceOption {
	return func(device *AndroidDevice) {
		device.IP = ip
	}
}

func WithAdbPort(port int) AndroidDeviceOption {
	return func(device *AndroidDevice) {
		device.Port = port
	}
}

func WithAdbLogOn(logOn bool) AndroidDeviceOption {
	return func(device *AndroidDevice) {
		device.LogOn = logOn
	}
}

func NewAndroidDevice(options ...AndroidDeviceOption) (device *AndroidDevice, err error) {
	deviceList, err := DeviceList()
	if err != nil {
		return nil, fmt.Errorf("get attached devices failed: %v", err)
	}

	device = &AndroidDevice{
		Port: UIA2ServerPort,
		IP:   AdbServerHost,
	}
	for _, option := range options {
		option(device)
	}

	serialNumber := device.SerialNumber
	for _, dev := range deviceList {
		// find device by serial number if specified
		if serialNumber != "" && dev.Serial() != serialNumber {
			continue
		}

		device.SerialNumber = dev.Serial()
		device.d = dev
		return device, nil
	}

	return nil, fmt.Errorf("device %s not found", device.SerialNumber)
}

type AndroidDevice struct {
	d            gadb.Device
	SerialNumber string `json:"serial,omitempty" yaml:"serial,omitempty"`
	IP           string `json:"ip,omitempty" yaml:"ip,omitempty"`
	Port         int    `json:"port,omitempty" yaml:"port,omitempty"`
	LogOn        bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

func (o AndroidDevice) UUID() string {
	return o.SerialNumber
}

func DeviceList() (devices []gadb.Device, err error) {
	var adbClient gadb.Client
	if adbClient, err = gadb.NewClientWith(AdbServerHost, AdbServerPort); err != nil {
		return nil, err
	}

	return adbClient.DeviceList()
}

// NewUSBDriver creates new client via USB connected device, this will also start a new session.
// TODO: replace uiaDriver with WebDriver
func (dev *AndroidDevice) NewUSBDriver(capabilities Capabilities) (driver *uiaDriver, err error) {
	var localPort int
	if localPort, err = getFreePort(); err != nil {
		return nil, err
	}
	if err = dev.d.Forward(localPort, UIA2ServerPort); err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("http://%s%d:6790/wd/hub", forwardToPrefix, localPort)
	driver, err = NewUIADriver(capabilities, rawURL)
	if err != nil {
		_ = dev.d.ForwardKill(localPort)
		return nil, err
	}
	driver.adbDevice = dev.d
	driver.localPort = localPort

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		return nil, fmt.Errorf("adb forward: %w", err)
	}
	driver.client = convertToHTTPClient(conn)

	return driver, nil
}

// NewHTTPDriver creates new remote HTTP client, this will also start a new session.
// TODO: replace uiaDriver with WebDriver
func (dev *AndroidDevice) NewHTTPDriver(capabilities Capabilities) (driver *uiaDriver, err error) {
	rawURL := fmt.Sprintf("http://%s:%d/wd/hub", dev.IP, dev.Port)
	if driver, err = NewUIADriver(capabilities, rawURL); err != nil {
		return nil, err
	}
	driver.adbDevice = dev.d
	return driver, nil
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("free port: %w", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("free port: %w", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type UiSelectorHelper struct {
	value *bytes.Buffer
}

func NewUiSelectorHelper() UiSelectorHelper {
	return UiSelectorHelper{value: bytes.NewBufferString("new UiSelector()")}
}

func (s UiSelectorHelper) String() string {
	return s.value.String() + ";"
}

// Text Set the search criteria to match the visible text displayed
// in a widget (for example, the text label to launch an app).
//
// The text for the element must match exactly with the string in your input
// argument. Matching is case-sensitive.
func (s UiSelectorHelper) Text(text string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.text("%s")`, text))
	return s
}

// TextMatches Set the search criteria to match the visible text displayed in a layout
// element, using a regular expression.
//
// The text in the widget must match exactly with the string in your
// input argument.
func (s UiSelectorHelper) TextMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.textMatches("%s")`, regex))
	return s
}

// TextStartsWith Set the search criteria to match visible text in a widget that is
// prefixed by the text parameter.
//
// The matching is case-insensitive.
func (s UiSelectorHelper) TextStartsWith(text string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.textStartsWith("%s")`, text))
	return s
}

// TextContains Set the search criteria to match the visible text in a widget
// where the visible text must contain the string in your input argument.
//
// The matching is case-sensitive.
func (s UiSelectorHelper) TextContains(text string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.textContains("%s")`, text))
	return s
}

// ClassName Set the search criteria to match the class property
// for a widget (for example, "android.widget.Button").
func (s UiSelectorHelper) ClassName(className string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.className("%s")`, className))
	return s
}

// ClassNameMatches Set the search criteria to match the class property
// for a widget, using a regular expression.
func (s UiSelectorHelper) ClassNameMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.classNameMatches("%s")`, regex))
	return s
}

// Description Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must match exactly
// with the string in your input argument.
//
// Matching is case-sensitive.
func (s UiSelectorHelper) Description(desc string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.description("%s")`, desc))
	return s
}

// DescriptionMatches Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must match exactly
// with the string in your input argument.
func (s UiSelectorHelper) DescriptionMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.descriptionMatches("%s")`, regex))
	return s
}

// DescriptionStartsWith Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must start
// with the string in your input argument.
//
// Matching is case-insensitive.
func (s UiSelectorHelper) DescriptionStartsWith(desc string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.descriptionStartsWith("%s")`, desc))
	return s
}

// DescriptionContains Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must contain
// the string in your input argument.
//
// Matching is case-insensitive.
func (s UiSelectorHelper) DescriptionContains(desc string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.descriptionContains("%s")`, desc))
	return s
}

// ResourceId Set the search criteria to match the given resource ID.
func (s UiSelectorHelper) ResourceId(id string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.resourceId("%s")`, id))
	return s
}

// ResourceIdMatches Set the search criteria to match the resource ID
// of the widget, using a regular expression.
func (s UiSelectorHelper) ResourceIdMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.resourceIdMatches("%s")`, regex))
	return s
}

// Index Set the search criteria to match the widget by its node
// index in the layout hierarchy.
//
// The index value must be 0 or greater.
//
// Using the index can be unreliable and should only
// be used as a last resort for matching. Instead,
// consider using the `Instance(int)` method.
func (s UiSelectorHelper) Index(index int) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.index(%d)`, index))
	return s
}

// Instance Set the search criteria to match the
// widget by its instance number.
//
// The instance value must be 0 or greater, where
// the first instance is 0.
//
// For example, to simulate a user click on
// the third image that is enabled in a UI screen, you
// could specify a a search criteria where the instance is
// 2, the `className(String)` matches the image
// widget class, and `enabled(boolean)` is true.
// The code would look like this:
//  `new UiSelector().className("android.widget.ImageView")
//    .enabled(true).instance(2);`
func (s UiSelectorHelper) Instance(instance int) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.instance(%d)`, instance))
	return s
}

// Enabled Set the search criteria to match widgets that are enabled.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Enabled(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.enabled(%t)`, b))
	return s
}

// Focused Set the search criteria to match widgets that have focus.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Focused(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.focused(%t)`, b))
	return s
}

// Focusable Set the search criteria to match widgets that are focusable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Focusable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.focusable(%t)`, b))
	return s
}

// Scrollable Set the search criteria to match widgets that are scrollable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Scrollable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.scrollable(%t)`, b))
	return s
}

// Selected Set the search criteria to match widgets that
// are currently selected.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Selected(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.selected(%t)`, b))
	return s
}

// Checked Set the search criteria to match widgets that
// are currently checked (usually for checkboxes).
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Checked(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.checked(%t)`, b))
	return s
}

// Checkable Set the search criteria to match widgets that are checkable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Checkable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.checkable(%t)`, b))
	return s
}

// Clickable Set the search criteria to match widgets that are clickable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Clickable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.clickable(%t)`, b))
	return s
}

// LongClickable Set the search criteria to match widgets that are long-clickable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) LongClickable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.longClickable(%t)`, b))
	return s
}

// packageName Set the search criteria to match the package name
// of the application that contains the widget.
func (s UiSelectorHelper) packageName(name string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.packageName(%s)`, name))
	return s
}

// PackageNameMatches Set the search criteria to match the package name
// of the application that contains the widget.
func (s UiSelectorHelper) PackageNameMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.packageNameMatches(%s)`, regex))
	return s
}

// ChildSelector Adds a child UiSelector criteria to this selector.
//
// Use this selector to narrow the search scope to
// child widgets under a specific parent widget.
func (s UiSelectorHelper) ChildSelector(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.childSelector(%s)`, selector.value.String()))
	return s
}

func (s UiSelectorHelper) PatternSelector(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.patternSelector(%s)`, selector.value.String()))
	return s
}

func (s UiSelectorHelper) ContainerSelector(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.containerSelector(%s)`, selector.value.String()))
	return s
}

// FromParent Adds a child UiSelector criteria to this selector which is used to
// start search from the parent widget.
//
// Use this selector to narrow the search scope to
// sibling widgets as well all child widgets under a parent.
func (s UiSelectorHelper) FromParent(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.fromParent(%s)`, selector.value.String()))
	return s
}

type AndroidBySelector struct {
	// Set the search criteria to match the given resource ResourceIdID.
	ResourceIdID string `json:"id"`
	// Set the search criteria to match the content-description property for a widget.
	ContentDescription string `json:"accessibility id"`
	XPath              string `json:"xpath"`
	// Set the search criteria to match the class property for a widget (for example, "android.widget.Button").
	ClassName   string `json:"class name"`
	UiAutomator string `json:"-android uiautomator"`
}

func (by AndroidBySelector) getMethodAndSelector() (method, selector string) {
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
