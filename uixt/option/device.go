package option

// DeviceOptions unified device options for all platforms using composition
type DeviceOptions struct {
	// Common fields
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty"`

	// Embedded platform-specific options
	*AndroidDeviceOptions `json:"android,omitempty" yaml:"android,omitempty"`
	*IOSDeviceOptions     `json:"ios,omitempty" yaml:"ios,omitempty"`
	*HarmonyDeviceOptions `json:"harmony,omitempty" yaml:"harmony,omitempty"`
	*BrowserDeviceOptions `json:"browser,omitempty" yaml:"browser,omitempty"`
}

// DeviceOption unified device option function
type DeviceOption func(*DeviceOptions)

// NewDeviceOptions creates a new DeviceOptions with given options
func NewDeviceOptions(opts ...DeviceOption) *DeviceOptions {
	config := &DeviceOptions{
		AndroidDeviceOptions: &AndroidDeviceOptions{},
		IOSDeviceOptions:     &IOSDeviceOptions{},
		HarmonyDeviceOptions: &HarmonyDeviceOptions{},
		BrowserDeviceOptions: &BrowserDeviceOptions{},
	}

	for _, opt := range opts {
		opt(config)
	}

	// Apply defaults based on platform
	config.applyDefaults()

	return config
}

// Unified DeviceOption functions

// WithPlatform sets the platform
func WithPlatform(platform string) DeviceOption {
	return func(device *DeviceOptions) {
		device.Platform = platform
	}
}

// WithDeviceLogOn sets log on for any platform
func WithDeviceLogOn(logOn bool) DeviceOption {
	return func(device *DeviceOptions) {
		// Set LogOn for all platform options to avoid ambiguity
		if device.AndroidDeviceOptions != nil {
			device.AndroidDeviceOptions.LogOn = logOn
		}
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.LogOn = logOn
		}
		if device.HarmonyDeviceOptions != nil {
			device.HarmonyDeviceOptions.LogOn = logOn
		}
		if device.BrowserDeviceOptions != nil {
			device.BrowserDeviceOptions.LogOn = logOn
		}
	}
}

// Android unified options
func WithDeviceSerialNumber(serial string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.AndroidDeviceOptions != nil {
			device.AndroidDeviceOptions.SerialNumber = serial
		}
		if device.Platform == "" {
			device.Platform = "android"
		}
	}
}

func WithDeviceUIA2(uia2On bool) DeviceOption {
	return func(device *DeviceOptions) {
		if device.AndroidDeviceOptions != nil {
			device.AndroidDeviceOptions.UIA2 = uia2On
		}
		if device.Platform == "" {
			device.Platform = "android"
		}
	}
}

func WithDeviceUIA2IP(ip string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.AndroidDeviceOptions != nil {
			device.AndroidDeviceOptions.UIA2IP = ip
		}
		if device.Platform == "" {
			device.Platform = "android"
		}
	}
}

func WithDeviceUIA2Port(port int) DeviceOption {
	return func(device *DeviceOptions) {
		if device.AndroidDeviceOptions != nil {
			device.AndroidDeviceOptions.UIA2Port = port
		}
		if device.Platform == "" {
			device.Platform = "android"
		}
	}
}

// iOS unified options
func WithDeviceUDID(udid string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.UDID = udid
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceWireless(on bool) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.Wireless = on
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceWDAPort(port int) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.WDAPort = port
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceWDAMjpegPort(port int) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.WDAMjpegPort = port
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceResetHomeOnStartup(reset bool) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.ResetHomeOnStartup = reset
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceSnapshotMaxDepth(depth int) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.SnapshotMaxDepth = depth
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceAcceptAlertButtonSelector(selector string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.AcceptAlertButtonSelector = selector
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

func WithDeviceDismissAlertButtonSelector(selector string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.IOSDeviceOptions != nil {
			device.IOSDeviceOptions.DismissAlertButtonSelector = selector
		}
		if device.Platform == "" {
			device.Platform = "ios"
		}
	}
}

// Harmony unified options
func WithDeviceConnectKey(connectKey string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.HarmonyDeviceOptions != nil {
			device.HarmonyDeviceOptions.ConnectKey = connectKey
		}
		if device.Platform == "" {
			device.Platform = "harmony"
		}
	}
}

// Browser unified options
func WithDeviceBrowserID(browserID string) DeviceOption {
	return func(device *DeviceOptions) {
		if device.BrowserDeviceOptions != nil {
			device.BrowserDeviceOptions.BrowserID = browserID
		}
		if device.Platform == "" {
			device.Platform = "browser"
		}
	}
}

func WithDeviceBrowserPageSize(width, height int) DeviceOption {
	return func(device *DeviceOptions) {
		if device.BrowserDeviceOptions != nil {
			device.BrowserDeviceOptions.Width = width
			device.BrowserDeviceOptions.Height = height
		}
		if device.Platform == "" {
			device.Platform = "browser"
		}
	}
}

// setAndroidDefaults applies Android platform defaults
func (d *DeviceOptions) setAndroidDefaults() {
	if d.AndroidDeviceOptions != nil {
		// Apply defaults using existing NewAndroidDeviceOptions logic
		d.AndroidDeviceOptions = NewAndroidDeviceOptions(d.AndroidDeviceOptions.Options()...)
	}
}

// setIOSDefaults applies iOS platform defaults
func (d *DeviceOptions) setIOSDefaults() {
	if d.IOSDeviceOptions != nil {
		// Apply defaults using existing NewIOSDeviceOptions logic
		d.IOSDeviceOptions = NewIOSDeviceOptions(d.IOSDeviceOptions.Options()...)
	}
}

// setHarmonyDefaults applies Harmony platform defaults
func (d *DeviceOptions) setHarmonyDefaults() {
	if d.HarmonyDeviceOptions != nil {
		// Apply defaults using existing NewHarmonyDeviceOptions logic
		d.HarmonyDeviceOptions = NewHarmonyDeviceOptions(d.HarmonyDeviceOptions.Options()...)
	}
}

// setBrowserDefaults applies Browser platform defaults
func (d *DeviceOptions) setBrowserDefaults() {
	if d.BrowserDeviceOptions != nil {
		// Apply defaults using existing NewBrowserDeviceOptions logic
		d.BrowserDeviceOptions = NewBrowserDeviceOptions(d.BrowserDeviceOptions.Options()...)
	}
}

// applyDefaults applies platform-specific defaults based on the Platform field
func (d *DeviceOptions) applyDefaults() {
	switch d.Platform {
	case "android":
		d.setAndroidDefaults()
	case "ios":
		d.setIOSDefaults()
	case "harmony":
		d.setHarmonyDefaults()
	case "browser":
		d.setBrowserDefaults()
	}
}

// GetSerial returns the appropriate serial/identifier for the platform
func (d *DeviceOptions) GetSerial() string {
	switch d.Platform {
	case "android":
		if d.AndroidDeviceOptions != nil {
			return d.AndroidDeviceOptions.SerialNumber
		}
	case "ios":
		if d.IOSDeviceOptions != nil {
			return d.IOSDeviceOptions.UDID
		}
	case "harmony":
		if d.HarmonyDeviceOptions != nil {
			return d.HarmonyDeviceOptions.ConnectKey
		}
	case "browser":
		if d.BrowserDeviceOptions != nil {
			return d.BrowserDeviceOptions.BrowserID
		}
	}
	return "" // fallback
}

// GetPlatformOptions returns platform-specific options slice
func (d *DeviceOptions) GetPlatformOptions() interface{} {
	switch d.Platform {
	case "android":
		return d.ToAndroidOptions().Options()
	case "ios":
		return d.ToIOSOptions().Options()
	case "harmony":
		return d.ToHarmonyOptions().Options()
	case "browser":
		return d.ToBrowserOptions().Options()
	default:
		return nil
	}
}

// ToAndroidOptions converts to AndroidDeviceOptions for backward compatibility
func (d *DeviceOptions) ToAndroidOptions() *AndroidDeviceOptions {
	if d.AndroidDeviceOptions != nil {
		return d.AndroidDeviceOptions
	}
	return &AndroidDeviceOptions{}
}

// ToIOSOptions converts to IOSDeviceOptions for backward compatibility
func (d *DeviceOptions) ToIOSOptions() *IOSDeviceOptions {
	if d.IOSDeviceOptions != nil {
		return d.IOSDeviceOptions
	}
	return &IOSDeviceOptions{}
}

// ToHarmonyOptions converts to HarmonyDeviceOptions for backward compatibility
func (d *DeviceOptions) ToHarmonyOptions() *HarmonyDeviceOptions {
	if d.HarmonyDeviceOptions != nil {
		return d.HarmonyDeviceOptions
	}
	return &HarmonyDeviceOptions{}
}

// ToBrowserOptions converts to BrowserDeviceOptions for backward compatibility
func (d *DeviceOptions) ToBrowserOptions() *BrowserDeviceOptions {
	if d.BrowserDeviceOptions != nil {
		return d.BrowserDeviceOptions
	}
	return &BrowserDeviceOptions{}
}

// FromAndroidOptions creates DeviceOptions from AndroidDeviceOptions
func FromAndroidOptions(opts *AndroidDeviceOptions) *DeviceOptions {
	config := &DeviceOptions{
		Platform:             "android",
		AndroidDeviceOptions: opts,
		IOSDeviceOptions:     &IOSDeviceOptions{},
		HarmonyDeviceOptions: &HarmonyDeviceOptions{},
		BrowserDeviceOptions: &BrowserDeviceOptions{},
	}
	// Apply defaults
	config.applyDefaults()
	return config
}

// FromIOSOptions creates DeviceOptions from IOSDeviceOptions
func FromIOSOptions(opts *IOSDeviceOptions) *DeviceOptions {
	config := &DeviceOptions{
		Platform:             "ios",
		AndroidDeviceOptions: &AndroidDeviceOptions{},
		IOSDeviceOptions:     opts,
		HarmonyDeviceOptions: &HarmonyDeviceOptions{},
		BrowserDeviceOptions: &BrowserDeviceOptions{},
	}
	// Apply defaults
	config.applyDefaults()
	return config
}

// FromHarmonyOptions creates DeviceOptions from HarmonyDeviceOptions
func FromHarmonyOptions(opts *HarmonyDeviceOptions) *DeviceOptions {
	config := &DeviceOptions{
		Platform:             "harmony",
		AndroidDeviceOptions: &AndroidDeviceOptions{},
		IOSDeviceOptions:     &IOSDeviceOptions{},
		HarmonyDeviceOptions: opts,
		BrowserDeviceOptions: &BrowserDeviceOptions{},
	}
	// Apply defaults
	config.applyDefaults()
	return config
}

// FromBrowserOptions creates DeviceOptions from BrowserDeviceOptions
func FromBrowserOptions(opts *BrowserDeviceOptions) *DeviceOptions {
	config := &DeviceOptions{
		Platform:             "browser",
		AndroidDeviceOptions: &AndroidDeviceOptions{},
		IOSDeviceOptions:     &IOSDeviceOptions{},
		HarmonyDeviceOptions: &HarmonyDeviceOptions{},
		BrowserDeviceOptions: opts,
	}
	// Apply defaults
	config.applyDefaults()
	return config
}
