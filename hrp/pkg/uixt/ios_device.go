package uixt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/deviceinfo"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/forward"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

const (
	defaultWDAPort          = 8100
	defaultMjpegPort        = 9100
	defaultBightInsightPort = 8000
	defaultDouyinServerPort = 32921
)

const (
	// Changes the value of maximum depth for traversing elements source tree.
	// It may help to prevent out of memory or timeout errors while getting the elements source tree,
	// but it might restrict the depth of source tree.
	// A part of elements source tree might be lost if the value was too small. Defaults to 50
	snapshotMaxDepth = 10
	// Allows to customize accept/dismiss alert button selector.
	// It helps you to handle an arbitrary element as accept button in accept alert command.
	// The selector should be a valid class chain expression, where the search root is the alert element itself.
	// The default button location algorithm is used if the provided selector is wrong or does not match any element.
	// e.g. **/XCUIElementTypeButton[`label CONTAINS[c] ‘accept’`]
	acceptAlertButtonSelector  = "**/XCUIElementTypeButton[`label IN {'允许','好','仅在使用应用期间','稍后再说'}`]"
	dismissAlertButtonSelector = "**/XCUIElementTypeButton[`label IN {'不允许','暂不'}`]"
)

var tunnelManager *tunnel.TunnelManager = nil

type IOSDeviceOption func(*IOSDevice)

func WithUDID(udid string) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.UDID = udid
	}
}

func WithWDAPort(port int) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.Port = port
	}
}

func WithWDAMjpegPort(port int) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.MjpegPort = port
	}
}

func WithWDALogOn(logOn bool) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.LogOn = logOn
	}
}

func WithIOSStub(stub bool) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.STUB = stub
	}
}

func WithResetHomeOnStartup(reset bool) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.ResetHomeOnStartup = reset
	}
}

func WithSnapshotMaxDepth(depth int) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.SnapshotMaxDepth = depth
	}
}

func WithAcceptAlertButtonSelector(selector string) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.AcceptAlertButtonSelector = selector
	}
}

func WithDismissAlertButtonSelector(selector string) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.DismissAlertButtonSelector = selector
	}
}

func GetIOSDevices(udid ...string) (deviceList []ios.DeviceEntry, err error) {
	devices, err := ios.ListDevices()
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("list ios devices failed: %v", err))
	}
	for _, d := range devices.DeviceList {
		if len(udid) > 0 {
			for _, u := range udid {
				if u != "" && u != d.Properties.SerialNumber {
					continue
				}
				// filter non-usb ios devices
				if d.Properties.ConnectionType != "USB" {
					continue
				}
				deviceList = append(deviceList, d)
			}
		} else {
			deviceList = devices.DeviceList
		}
	}

	if len(deviceList) == 0 {
		var err error
		if udid == nil || (len(udid) == 1 && udid[0] == "") {
			err = fmt.Errorf("no ios device found")
		} else {
			err = fmt.Errorf("no ios device found for udid %v", udid)
		}
		return nil, err
	}
	return deviceList, nil
}

func GetIOSDeviceOptions(dev *IOSDevice) (deviceOptions []IOSDeviceOption) {
	if dev.UDID != "" {
		deviceOptions = append(deviceOptions, WithUDID(dev.UDID))
	}
	if dev.Port != 0 {
		deviceOptions = append(deviceOptions, WithWDAPort(dev.Port))
	}
	if dev.MjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAMjpegPort(dev.MjpegPort))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithWDALogOn(true))
	}
	if dev.ResetHomeOnStartup {
		deviceOptions = append(deviceOptions, WithResetHomeOnStartup(true))
	}
	if dev.SnapshotMaxDepth != 0 {
		deviceOptions = append(deviceOptions, WithSnapshotMaxDepth(dev.SnapshotMaxDepth))
	}
	if dev.AcceptAlertButtonSelector != "" {
		deviceOptions = append(deviceOptions, WithAcceptAlertButtonSelector(dev.AcceptAlertButtonSelector))
	}
	if dev.DismissAlertButtonSelector != "" {
		deviceOptions = append(deviceOptions, WithDismissAlertButtonSelector(dev.DismissAlertButtonSelector))
	}
	return
}

func StartTunnel(recordsPath string, tunnelInfoPort int, userspaceTUN bool) (err error) {
	pm, err := tunnel.NewPairRecordManager(recordsPath)
	if err != nil {
		return err
	}
	tm := tunnel.NewTunnelManager(pm, userspaceTUN)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			err := tm.UpdateTunnels(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("failed to update tunnels")
			}
		}
	}()

	go func() {
		err := tunnel.ServeTunnelInfo(tm, tunnelInfoPort)
		if err != nil {
			log.Error().Err(err).Msg("failed to start tunnel server")
		}
	}()

	log.Info().Msg("Tunnel server started")
	return nil
}

func RebootTunnel() (err error) {
	if tunnelManager != nil {
		_ = tunnelManager.Close()
	}
	return StartTunnel(os.TempDir(), ios.HttpApiPort(), true)
}

func NewIOSDevice(options ...IOSDeviceOption) (device *IOSDevice, err error) {
	device = &IOSDevice{
		Port:                       defaultWDAPort,
		MjpegPort:                  defaultMjpegPort,
		SnapshotMaxDepth:           snapshotMaxDepth,
		AcceptAlertButtonSelector:  acceptAlertButtonSelector,
		DismissAlertButtonSelector: dismissAlertButtonSelector,
		// switch to iOS springboard before init WDA session
		// avoid getting stuck when some super app is active such as douyin or wexin
		ResetHomeOnStartup: true,
		listeners:          make(map[int]*forward.ConnListener),
	}
	for _, option := range options {
		option(device)
	}

	deviceList, err := GetIOSDevices(device.UDID)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}

	if device.UDID == "" && len(deviceList) > 1 {
		return nil, errors.Wrap(code.DeviceConnectionError, "more than one device connected, please specify the udid")
	}

	dev := deviceList[0]
	udid := dev.Properties.SerialNumber

	if device.UDID == "" {
		device.UDID = udid
		log.Warn().
			Str("udid", udid).
			Msg("ios UDID is not specified, select the first one")
	}

	device.d = dev
	log.Info().Str("udid", device.UDID).Msg("init ios device")
	err = device.Init()
	if err != nil {
		_ = device.Teardown()
		return nil, err
	}
	return device, nil
}

type IOSDevice struct {
	d         ios.DeviceEntry
	listeners map[int]*forward.ConnListener
	UDID      string `json:"udid,omitempty" yaml:"udid,omitempty"`
	Port      int    `json:"port,omitempty" yaml:"port,omitempty"`             // WDA remote port
	MjpegPort int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"` // WDA remote MJPEG port
	STUB      bool   `json:"stub,omitempty" yaml:"stub,omitempty"`             // use stub
	LogOn     bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`

	// switch to iOS springboard before init WDA session
	ResetHomeOnStartup bool `json:"reset_home_on_startup,omitempty" yaml:"reset_home_on_startup,omitempty"`

	// config appium settings
	SnapshotMaxDepth           int    `json:"snapshot_max_depth,omitempty" yaml:"snapshot_max_depth,omitempty"`
	AcceptAlertButtonSelector  string `json:"accept_alert_button_selector,omitempty" yaml:"accept_alert_button_selector,omitempty"`
	DismissAlertButtonSelector string `json:"dismiss_alert_button_selector,omitempty" yaml:"dismiss_alert_button_selector,omitempty"`
}

func (dev *IOSDevice) Options() (deviceOptions []IOSDeviceOption) {
	if dev.UDID != "" {
		deviceOptions = append(deviceOptions, WithUDID(dev.UDID))
	}
	if dev.Port != 0 {
		deviceOptions = append(deviceOptions, WithWDAPort(dev.Port))
	}
	if dev.MjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAMjpegPort(dev.MjpegPort))
	}
	if dev.STUB {
		deviceOptions = append(deviceOptions, WithIOSStub(true))
	}
	if dev.LogOn {
		deviceOptions = append(deviceOptions, WithWDALogOn(true))
	}
	if dev.ResetHomeOnStartup {
		deviceOptions = append(deviceOptions, WithResetHomeOnStartup(true))
	}
	if dev.SnapshotMaxDepth != 0 {
		deviceOptions = append(deviceOptions, WithSnapshotMaxDepth(dev.SnapshotMaxDepth))
	}
	if dev.AcceptAlertButtonSelector != "" {
		deviceOptions = append(deviceOptions, WithAcceptAlertButtonSelector(dev.AcceptAlertButtonSelector))
	}
	if dev.DismissAlertButtonSelector != "" {
		deviceOptions = append(deviceOptions, WithDismissAlertButtonSelector(dev.DismissAlertButtonSelector))
	}
	return
}

type DeviceDetail struct {
	DeviceName        string `json:"deviceName,omitempty"`
	DeviceClass       string `json:"deviceClass,omitempty"`
	ProductVersion    string `json:"productVersion,omitempty"`
	ProductType       string `json:"productType,omitempty"`
	ProductName       string `json:"productName,omitempty"`
	PasswordProtected bool   `json:"passwordProtected,omitempty"`
	ModelNumber       string `json:"modelNumber,omitempty"`
	SerialNumber      string `json:"serialNumber,omitempty"`
	SIMStatus         string `json:"simStatus,omitempty"`
	PhoneNumber       string `json:"phoneNumber,omitempty"`
	CPUArchitecture   string `json:"cpuArchitecture,omitempty"`
	ProtocolVersion   string `json:"protocolVersion,omitempty"`
	RegionInfo        string `json:"regionInfo,omitempty"`
	TimeZone          string `json:"timeZone,omitempty"`
	UniqueDeviceID    string `json:"uniqueDeviceID,omitempty"`
	WiFiAddress       string `json:"wifiAddress,omitempty"`
	BuildVersion      string `json:"buildVersion,omitempty"`
}
type ApplicationType string

const (
	ApplicationTypeSystem   ApplicationType = "System"
	ApplicationTypeUser     ApplicationType = "User"
	ApplicationTypeInternal ApplicationType = "internal"
	ApplicationTypeAny      ApplicationType = "Any"
)

func (dev *IOSDevice) Init() error {
	images, err := dev.ListImages()
	if err != nil {
		return err
	}
	version, err := dev.getVersion()
	if err != nil {
		return err
	}
	if len(images) == 0 && version.LessThan(ios.IOS17()) {
		// Notice: iOS 17.0+ does not need to mount developer image
		err = dev.AutoMountImage(os.TempDir())
		if err != nil {
			return err
		}
	}

	return nil
}

func (dev *IOSDevice) Teardown() error {
	for _, listener := range dev.listeners {
		_ = listener.Close()
	}
	return nil
}

func (dev *IOSDevice) UUID() string {
	return dev.UDID
}

func (dev *IOSDevice) LogEnabled() bool {
	return dev.LogOn
}

func (dev *IOSDevice) getAppInfo(packageName string) (appInfo AppInfo, err error) {
	apps, err := dev.ListApps(ApplicationTypeAny)
	if err != nil {
		return AppInfo{}, err
	}
	for _, app := range apps {
		if app.CFBundleIdentifier == packageName {
			appInfo.BundleId = app.CFBundleIdentifier
			appInfo.AppName = app.CFBundleName
			appInfo.VersionName = app.CFBundleShortVersionString
			appInfo.PackageName = app.CFBundleIdentifier
			versionCode, err := strconv.Atoi(app.CFBundleVersion)
			if err == nil {
				appInfo.VersionCode = versionCode
			}
			return appInfo, err
		}
	}
	return AppInfo{}, fmt.Errorf("not found App by bundle id: %s", packageName)
}

func (dev *IOSDevice) NewDriver(options ...DriverOption) (driverExt *DriverExt, err error) {
	driverOptions := NewDriverOptions()
	for _, option := range options {
		option(driverOptions)
	}

	// init WDA driver
	capabilities := driverOptions.capabilities
	if capabilities == nil {
		capabilities = NewCapabilities()
		capabilities.WithDefaultAlertAction(AlertActionAccept)
	}

	var driver IWebDriver
	if dev.STUB {
		driver, err = dev.NewStubDriver()
		if err != nil {
			return nil, errors.Wrap(err, "failed to init Stub driver")
		}
	} else {
		driver, err = dev.NewHTTPDriver(capabilities)
		if err != nil {
			return nil, errors.Wrap(err, "failed to init WDA driver")
		}
		settings, err := driver.SetAppiumSettings(map[string]interface{}{
			"snapshotMaxDepth":          dev.SnapshotMaxDepth,
			"acceptAlertButtonSelector": dev.AcceptAlertButtonSelector,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to set appium WDA settings")
		}
		log.Info().Interface("appiumWDASettings", settings).Msg("set appium WDA settings")
	}

	if dev.ResetHomeOnStartup {
		log.Info().Msg("go back to home screen")
		if err = driver.Homescreen(); err != nil {
			return nil, errors.Wrap(code.MobileUIDriverError,
				fmt.Sprintf("go back to home screen failed: %v", err))
		}
	}

	driverExt, err = newDriverExt(dev, driver, options...)
	if err != nil {
		return nil, err
	}

	settings, err := driverExt.Driver.SetAppiumSettings(map[string]interface{}{
		"snapshotMaxDepth":          dev.SnapshotMaxDepth,
		"acceptAlertButtonSelector": dev.AcceptAlertButtonSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set appium WDA settings")
	}
	log.Info().Interface("appiumWDASettings", settings).Msg("set appium WDA settings")

	if dev.LogOn {
		err = driverExt.Driver.StartCaptureLog("hrp_wda_log")
		if err != nil {
			return nil, err
		}
	}

	return driverExt, nil
}

func (dev *IOSDevice) Install(appPath string, options ...InstallOption) (err error) {
	opts := NewInstallOptions(options...)
	for i := 0; i <= opts.RetryTimes; i++ {
		var conn *zipconduit.Connection
		conn, err = zipconduit.New(dev.d)
		if err != nil {
			return err
		}
		defer conn.Close()
		err = conn.SendFile(appPath)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("failed to install app Retry time %d", i))
		}
		if err == nil {
			return nil
		}
	}
	return err
}

func (dev *IOSDevice) InstallByUrl(url string, options ...InstallOption) (err error) {
	appPath, err := builtin.DownloadFileByUrl(url)
	if err != nil {
		return err
	}
	err = dev.Install(appPath, options...)
	if err != nil {
		return err
	}
	return nil
}

func (dev *IOSDevice) Uninstall(bundleId string) error {
	svc, err := installationproxy.New(dev.d)
	if err != nil {
		return err
	}
	defer svc.Close()
	err = svc.Uninstall(bundleId)
	if err != nil {
		return err
	}
	return nil
}

func (dev *IOSDevice) forward(localPort, remotePort int) error {
	if dev.listeners[localPort] != nil {
		log.Warn().Msg(fmt.Sprintf("local port :%d is already in use", localPort))
		_ = dev.listeners[localPort].Close()
	}
	listener, err := forward.Forward(dev.d, uint16(localPort), uint16(remotePort))
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("failed to forward %d to %d", localPort, remotePort))
		return err
	}
	dev.listeners[localPort] = listener
	return nil
}

func (dev *IOSDevice) GetDeviceInfo() (*DeviceDetail, error) {
	deviceInfo, err := deviceinfo.NewDeviceInfo(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to get device info")
		return nil, err
	}
	defer deviceInfo.Close()
	info, err := deviceInfo.GetDisplayInfo()
	if err != nil {
		log.Error().Err(err).Msg("failed to get device info")
		return nil, err
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	// 将 JSON 反序列化为结构体
	detail := new(DeviceDetail)
	err = json.Unmarshal(jsonData, &detail)
	if err != nil {
		return nil, err
	}
	return detail, err
}

func (dev *IOSDevice) ListApps(appType ApplicationType) (apps []installationproxy.AppInfo, err error) {
	svc, _ := installationproxy.New(dev.d)
	defer svc.Close()
	switch appType {
	case ApplicationTypeSystem:
		apps, err = svc.BrowseSystemApps()
	case ApplicationTypeAny:
		apps, err = svc.BrowseAllApps()
	case ApplicationTypeInternal:
		apps, err = svc.BrowseFileSharingApps()
	case ApplicationTypeUser:
		apps, err = svc.BrowseUserApps()
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to list ios apps")
		return nil, err
	}
	return apps, nil
}

func (dev *IOSDevice) GetAppInfo(packageName string) (appInfo installationproxy.AppInfo, err error) {
	svc, _ := installationproxy.New(dev.d)
	defer svc.Close()
	apps, err := svc.BrowseAllApps()
	if err != nil {
		log.Error().Err(err).Msg("failed to list ios apps")
		return installationproxy.AppInfo{}, err
	}
	for _, app := range apps {
		if app.CFBundleIdentifier == packageName {
			return app, nil
		}
	}
	return installationproxy.AppInfo{}, nil
}

func (dev *IOSDevice) ListImages() (images []string, err error) {
	conn, err := imagemounter.NewImageMounter(dev.d)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}
	defer conn.Close()

	signatures, err := conn.ListImages()
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}
	for _, sig := range signatures {
		images = append(images, fmt.Sprintf("%x", sig))
	}
	return
}

func (dev *IOSDevice) MountImage(imagePath string) (err error) {
	log.Info().Str("imagePath", imagePath).Msg("mount ios developer image")
	conn, err := imagemounter.NewImageMounter(dev.d)
	if err != nil {
		return errors.Wrap(code.DeviceConnectionError, err.Error())
	}
	defer conn.Close()

	err = conn.MountImage(imagePath)
	if err != nil {
		return errors.Wrapf(code.DeviceConnectionError,
			"mount ios developer image failed: %v", err)
	}
	log.Info().Str("imagePath", imagePath).Msg("mount ios developer image success")
	return nil
}

func (dev *IOSDevice) AutoMountImage(baseDir string) (err error) {
	log.Info().Str("baseDir", baseDir).Msg("auto mount ios developer image")
	imagePath, err := imagemounter.DownloadImageFor(dev.d, baseDir)
	if err != nil {
		return errors.Wrapf(code.DeviceConnectionError,
			"download ios developer image failed: %v", err)
	}
	return dev.MountImage(imagePath)
}

func (dev *IOSDevice) RunXCTest(ctx context.Context, bundleID, testRunnerBundleID, xctestConfig string) (err error) {
	log.Info().Str("bundleID", bundleID).
		Str("testRunnerBundleID", testRunnerBundleID).
		Str("xctestConfig", xctestConfig).
		Msg("run xctest")
	listener := testmanagerd.NewTestListener(io.Discard, io.Discard, os.TempDir())
	config := testmanagerd.TestConfig{
		BundleId:           bundleID,
		TestRunnerBundleId: testRunnerBundleID,
		XctestConfigName:   xctestConfig,
		Device:             dev.d,
		Listener:           listener,
	}
	_, err = testmanagerd.RunTestWithConfig(ctx, config)
	if err != nil {
		log.Error().Err(err).Msg("run xctest failed")
		return err
	}
	return nil
}

func (dev *IOSDevice) RunXCTestDaemon(ctx context.Context, bundleID, testRunnerBundleID, xctestConfig string) {
	ctx, stopWda := context.WithCancel(ctx)
	go func() {
		err := dev.RunXCTest(ctx, bundleID, testRunnerBundleID, xctestConfig)
		if err != nil {
			log.Error().Err(err).Msg("wda ended")
		}
		stopWda()
	}()
}

func (dev *IOSDevice) getVersion() (version *semver.Version, err error) {
	version, err = ios.GetProductVersion(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to get version")
		return nil, err
	}
	log.Info().Str("version", version.String()).Msg("get ios device version")
	return version, nil
}

func (dev *IOSDevice) ListProcess(applicationsOnly bool) (processList []instruments.ProcessInfo, err error) {
	service, err := instruments.NewDeviceInfoService(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to list process")
		return
	}
	defer service.Close()
	processList, err = service.ProcessList()
	if applicationsOnly {
		var applicationProcessList []instruments.ProcessInfo
		for _, processInfo := range processList {
			if processInfo.IsApplication {
				applicationProcessList = append(applicationProcessList, processInfo)
			}
		}
		processList = applicationProcessList
	}
	return
}

func (dev *IOSDevice) Reboot() error {
	err := diagnostics.Reboot(dev.d)
	if err != nil {
		log.Error().Err(err).Msg("failed to reboot device")
		return err
	}
	return nil
}

// NewHTTPDriver creates new remote HTTP client, this will also start a new session.
func (dev *IOSDevice) NewHTTPDriver(capabilities Capabilities) (driver IWebDriver, err error) {
	var localPort int
	localPort, err = strconv.Atoi(os.Getenv("WDA_LOCAL_PORT"))
	if err != nil {
		localPort, err = builtin.GetFreePort()
		if err != nil {
			return nil, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}
		if err = dev.forward(localPort, dev.Port); err != nil {
			return nil, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("forward tcp port failed: %v", err))
		}
	} else {
		log.Info().Int("WDA_LOCAL_PORT", localPort).Msg("reuse WDA local port")
	}

	var localMjpegPort int
	localMjpegPort, err = strconv.Atoi(os.Getenv("WDA_LOCAL_MJPEG_PORT"))
	if err != nil {
		localMjpegPort, err = builtin.GetFreePort()
		if err != nil {
			return nil, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}
		if err = dev.forward(localMjpegPort, dev.MjpegPort); err != nil {
			return nil, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("forward tcp port failed: %v", err))
		}
	} else {
		log.Info().Int("WDA_LOCAL_MJPEG_PORT", localMjpegPort).
			Msg("reuse WDA local mjpeg port")
	}

	log.Info().Interface("capabilities", capabilities).
		Int("localPort", localPort).Int("localMjpegPort", localMjpegPort).
		Msg("init WDA HTTP driver")

	wd := new(wdaDriver)
	wd.device = dev
	wd.udid = dev.UDID
	wd.client = &http.Client{
		Timeout: time.Second * 10, // 设置超时时间为 10 秒
	}

	host := "localhost"
	if wd.urlPrefix, err = url.Parse(fmt.Sprintf("http://%s:%d", host, localPort)); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}

	// create new session
	var sessionInfo SessionInfo
	if sessionInfo, err = wd.NewSession(capabilities); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	wd.session.ID = sessionInfo.SessionId

	if wd.mjpegHTTPConn, err = net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, localMjpegPort),
	); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	wd.mjpegClient = convertToHTTPClient(wd.mjpegHTTPConn)
	wd.mjpegUrl = fmt.Sprintf("%s:%d", host, localMjpegPort)
	// init WDA scale
	if wd.scale, err = wd.Scale(); err != nil {
		return nil, err
	}

	return wd, nil
}

func (dev *IOSDevice) NewStubDriver() (driver IWebDriver, err error) {
	localStubPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}

	if err = dev.forward(localStubPort, defaultBightInsightPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}

	localServerPort, err := builtin.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("get free port failed: %v", err))
	}
	if err = dev.forward(localServerPort, defaultDouyinServerPort); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError,
			fmt.Sprintf("forward tcp port failed: %v", err))
	}
	host := "localhost"
	stubDriver, err := newStubIOSDriver(
		fmt.Sprintf("http://%s:%d", host, localStubPort),
		fmt.Sprintf("http://%s:%d", host, localServerPort), dev)
	if err != nil {
		return nil, err
	}
	return stubDriver, nil
}

func (dev *IOSDevice) GetCurrentWindow() (WindowInfo, error) {
	return WindowInfo{}, nil
}

func (dev *IOSDevice) GetPackageInfo(packageName string) (AppInfo, error) {
	return AppInfo{}, nil
}
