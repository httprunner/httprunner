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

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

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

var tunnelManager *tunnel.TunnelManager = nil

func RebootTunnel() (err error) {
	if tunnelManager != nil {
		_ = tunnelManager.Close()
	}
	return StartTunnel(os.TempDir(), ios.HttpApiPort(), true)
}

func NewIOSDevice(opts ...option.IOSDeviceOption) (device *IOSDevice, err error) {
	deviceOptions := option.NewIOSDeviceOptions(opts...)

	// get all attached ios devices
	devices, err := ios.ListDevices()
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}
	if len(devices.DeviceList) == 0 {
		return nil, errors.Wrapf(code.DeviceConnectionError,
			"no attached ios devices")
	}

	// filter device by udid
	var iosDevice *ios.DeviceEntry
	if deviceOptions.UDID == "" {
		if len(devices.DeviceList) > 1 {
			return nil, errors.Wrap(code.DeviceConnectionError,
				"more than one device connected, please specify the udid")
		}
		iosDevice = &devices.DeviceList[0]
		deviceOptions.UDID = iosDevice.Properties.SerialNumber
		log.Warn().Str("udid", deviceOptions.UDID).
			Msg("ios UDID is not specified, select the attached one")
	} else {
		for _, d := range devices.DeviceList {
			if d.Properties.SerialNumber == deviceOptions.UDID {
				iosDevice = &d
				break
			}
		}
		if iosDevice == nil {
			return nil, errors.Wrapf(code.DeviceConnectionError,
				"ios device %s not attached", deviceOptions.UDID)
		}
	}

	device = &IOSDevice{
		DeviceEntry: *iosDevice,
		Options:     deviceOptions,
		listeners:   make(map[int]*forward.ConnListener),
	}
	log.Info().Str("udid", device.Options.UDID).Msg("init ios device")
	return device, nil
}

type IOSDevice struct {
	ios.DeviceEntry
	Options   *option.IOSDeviceOptions
	listeners map[int]*forward.ConnListener
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

func (dev *IOSDevice) Setup() error {
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
	return dev.Options.UDID
}

func (dev *IOSDevice) LogEnabled() bool {
	return dev.Options.LogOn
}

func (dev *IOSDevice) getAppInfo(packageName string) (appInfo types.AppInfo, err error) {
	apps, err := dev.ListApps(ApplicationTypeAny)
	if err != nil {
		return types.AppInfo{}, err
	}
	for _, app := range apps {
		if app.CFBundleIdentifier == packageName {
			appInfo.BundleId = app.CFBundleIdentifier
			appInfo.AppName = app.CFBundleName
			appInfo.PackageName = app.CFBundleIdentifier
			appInfo.VersionName = app.CFBundleShortVersionString
			appInfo.VersionCode = app.CFBundleVersion
			return appInfo, err
		}
	}
	return types.AppInfo{}, fmt.Errorf("not found App by bundle id: %s", packageName)
}

func (dev *IOSDevice) NewDriver() (driverExt IDriverExt, err error) {
	driver, err := NewWDADriver(dev)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver")
	}
	settings, err := driver.SetAppiumSettings(map[string]interface{}{
		"snapshotMaxDepth":          dev.Options.SnapshotMaxDepth,
		"acceptAlertButtonSelector": dev.Options.AcceptAlertButtonSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set appium WDA settings")
	}
	log.Info().Interface("appiumWDASettings", settings).Msg("set appium WDA settings")

	if dev.Options.ResetHomeOnStartup {
		log.Info().Msg("go back to home screen")
		if err = driver.Homescreen(); err != nil {
			return nil, errors.Wrap(code.MobileUIDriverError,
				fmt.Sprintf("go back to home screen failed: %v", err))
		}
	}
	if dev.Options.LogOn {
		err = driver.StartCaptureLog("hrp_wda_log")
		if err != nil {
			return nil, err
		}
	}

	driverExt, err = NewDriverExt(driver, ai.WithCVService(ai.CVServiceTypeVEDEM))
	if err != nil {
		return nil, errors.Wrap(err, "init ios driver ext failed")
	}

	// setup driver
	if err := driverExt.GetDriver().Setup(); err != nil {
		return nil, err
	}

	return driverExt, nil
}

func (dev *IOSDevice) Install(appPath string, opts ...option.InstallOption) (err error) {
	installOpts := option.NewInstallOptions(opts...)
	for i := 0; i <= installOpts.RetryTimes; i++ {
		var conn *zipconduit.Connection
		conn, err = zipconduit.New(dev.DeviceEntry)
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

func (dev *IOSDevice) Uninstall(bundleId string) error {
	svc, err := installationproxy.New(dev.DeviceEntry)
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

func (dev *IOSDevice) Forward(localPort, remotePort int) error {
	if dev.listeners[localPort] != nil {
		log.Warn().Msg(fmt.Sprintf("local port :%d is already in use", localPort))
		_ = dev.listeners[localPort].Close()
	}
	listener, err := forward.Forward(dev.DeviceEntry, uint16(localPort), uint16(remotePort))
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("failed to forward %d to %d", localPort, remotePort))
		return err
	}
	dev.listeners[localPort] = listener
	return nil
}

func (dev *IOSDevice) GetDeviceInfo() (*DeviceDetail, error) {
	deviceInfo, err := deviceinfo.NewDeviceInfo(dev.DeviceEntry)
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
	svc, _ := installationproxy.New(dev.DeviceEntry)
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
	svc, _ := installationproxy.New(dev.DeviceEntry)
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
	conn, err := imagemounter.NewImageMounter(dev.DeviceEntry)
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
	conn, err := imagemounter.NewImageMounter(dev.DeviceEntry)
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
	imagePath, err := imagemounter.DownloadImageFor(dev.DeviceEntry, baseDir)
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
		Device:             dev.DeviceEntry,
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
	version, err = ios.GetProductVersion(dev.DeviceEntry)
	if err != nil {
		log.Error().Err(err).Msg("failed to get version")
		return nil, err
	}
	log.Info().Str("version", version.String()).Msg("get ios device version")
	return version, nil
}

func (dev *IOSDevice) ListProcess(applicationsOnly bool) (processList []instruments.ProcessInfo, err error) {
	service, err := instruments.NewDeviceInfoService(dev.DeviceEntry)
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
	err := diagnostics.Reboot(dev.DeviceEntry)
	if err != nil {
		log.Error().Err(err).Msg("failed to reboot device")
		return err
	}
	return nil
}

// NewHTTPDriver creates new remote HTTP client, this will also start a new session.
func (dev *IOSDevice) NewHTTPDriver(capabilities option.Capabilities) (driver IDriver, err error) {
	var localPort int
	localPort, err = strconv.Atoi(os.Getenv("WDA_LOCAL_PORT"))
	if err != nil {
		localPort, err = builtin.GetFreePort()
		if err != nil {
			return nil, errors.Wrap(code.DeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}
		if err = dev.Forward(localPort, dev.Options.WDAPort); err != nil {
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
		if err = dev.Forward(localMjpegPort, dev.Options.WDAMjpegPort); err != nil {
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

	wd := new(WDADriver)
	wd.Device = dev
	wd.Session.client = &http.Client{
		Timeout: time.Second * 10, // 设置超时时间为 10 秒
	}

	host := "localhost"
	if wd.Session.baseURL, err = url.Parse(fmt.Sprintf("http://%s:%d", host, localPort)); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}

	// create new session
	if err = wd.InitSession(capabilities); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}

	if wd.mjpegHTTPConn, err = net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, localMjpegPort),
	); err != nil {
		return nil, errors.Wrap(code.DeviceHTTPDriverError, err.Error())
	}
	wd.mjpegClient = NewHTTPClientWithConnection(wd.mjpegHTTPConn, 30*time.Second)
	wd.mjpegUrl = fmt.Sprintf("%s:%d", host, localMjpegPort)
	// init WDA scale
	if wd.Session.scale, err = wd.Scale(); err != nil {
		return nil, err
	}

	return wd, nil
}

func (dev *IOSDevice) GetPackageInfo(packageName string) (types.AppInfo, error) {
	svc, err := installationproxy.New(dev.DeviceEntry)
	if err != nil {
		return types.AppInfo{}, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}
	defer svc.Close()

	apps, err := svc.BrowseAllApps()
	if err != nil {
		return types.AppInfo{}, errors.Wrap(code.DeviceGetInfoError, err.Error())
	}

	for _, app := range apps {
		if app.CFBundleIdentifier != packageName {
			continue
		}

		appInfo := types.AppInfo{
			Name: app.CFBundleName,
			AppBaseInfo: types.AppBaseInfo{
				BundleId:    app.CFBundleIdentifier,
				PackageName: app.CFBundleIdentifier,
				VersionName: app.CFBundleShortVersionString,
				VersionCode: app.CFBundleVersion,
				AppName:     app.CFBundleDisplayName,
				AppPath:     app.Path,
			},
		}
		log.Info().Interface("appInfo", appInfo).Msg("get package info")
		return appInfo, nil
	}
	return types.AppInfo{}, errors.Wrap(code.DeviceAppNotInstalled,
		fmt.Sprintf("%s not found", packageName))
}
