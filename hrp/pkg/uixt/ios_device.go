package uixt

import (
	"bytes"
	"context"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"io"
	builtinLog "log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

const (
	defaultWDAPort   = 8100
	defaultMjpegPort = 9100
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

type IOSPerfOption = gidevice.PerfOption

var (
	WithIOSPerfSystemCPU         = gidevice.WithPerfSystemCPU
	WithIOSPerfSystemMem         = gidevice.WithPerfSystemMem
	WithIOSPerfSystemDisk        = gidevice.WithPerfSystemDisk
	WithIOSPerfSystemNetwork     = gidevice.WithPerfSystemNetwork
	WithIOSPerfGPU               = gidevice.WithPerfGPU
	WithIOSPerfFPS               = gidevice.WithPerfFPS
	WithIOSPerfNetwork           = gidevice.WithPerfNetwork
	WithIOSPerfBundleID          = gidevice.WithPerfBundleID
	WithIOSPerfPID               = gidevice.WithPerfPID
	WithIOSPerfOutputInterval    = gidevice.WithPerfOutputInterval
	WithIOSPerfProcessAttributes = gidevice.WithPerfProcessAttributes
	WithIOSPerfSystemAttributes  = gidevice.WithPerfSystemAttributes
)

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

func WithXCTest(bundleID string) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.XCTestBundleID = bundleID
	}
}

func WithIOSPerfOptions(options ...gidevice.PerfOption) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.PerfOptions = &gidevice.PerfOptions{}
		for _, option := range options {
			option(device.PerfOptions)
		}
	}
}

func IOSDevices(udid ...string) (devices []gidevice.Device, err error) {
	var usbmux gidevice.Usbmux
	if usbmux, err = gidevice.NewUsbmux(); err != nil {
		return nil, errors.Wrap(code.IOSDeviceConnectionError,
			fmt.Sprintf("init usbmux failed: %v", err))
	}

	if devices, err = usbmux.Devices(); err != nil {
		return nil, errors.Wrap(code.IOSDeviceConnectionError,
			fmt.Sprintf("list ios devices failed: %v", err))
	}

	// filter by udid
	var deviceList []gidevice.Device
	for _, d := range devices {
		for _, u := range udid {
			if u != "" && u != d.Properties().SerialNumber {
				continue
			}
			// filter non-usb ios devices
			if d.Properties().ConnectionType != "USB" {
				continue
			}
			deviceList = append(deviceList, d)
		}
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
	if dev.PerfOptions != nil {
		deviceOptions = append(deviceOptions, WithIOSPerfOptions(dev.perfOpitons()...))
	}
	if dev.XCTestBundleID != "" {
		deviceOptions = append(deviceOptions, WithXCTest(dev.XCTestBundleID))
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
	}
	for _, option := range options {
		option(device)
	}

	deviceList, err := IOSDevices(device.UDID)
	if err != nil {
		return nil, err
	}

	for _, dev := range deviceList {
		udid := dev.Properties().SerialNumber
		device.UDID = udid
		device.d = dev

		// run xctest if XCTestBundleID is set
		if device.XCTestBundleID != "" {
			_, err = device.RunXCTest(device.XCTestBundleID)
			if err != nil {
				log.Error().Err(err).Str("udid", udid).Msg("failed to init XCTest")
				continue
			}
		}

		log.Info().Str("udid", device.UDID).Msg("select device")
		return device, nil
	}

	return nil, errors.Wrap(code.IOSDeviceConnectionError,
		fmt.Sprintf("device %s not found", device.UDID))
}

type IOSDevice struct {
	d              gidevice.Device
	PerfOptions    *gidevice.PerfOptions `json:"perf_options,omitempty" yaml:"perf_options,omitempty"`
	UDID           string                `json:"udid,omitempty" yaml:"udid,omitempty"`
	Port           int                   `json:"port,omitempty" yaml:"port,omitempty"`             // WDA remote port
	MjpegPort      int                   `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"` // WDA remote MJPEG port
	LogOn          bool                  `json:"log_on,omitempty" yaml:"log_on,omitempty"`
	XCTestBundleID string                `json:"xctest_bundle_id,omitempty" yaml:"xctest_bundle_id,omitempty"`

	// switch to iOS springboard before init WDA session
	ResetHomeOnStartup bool `json:"reset_home_on_startup,omitempty" yaml:"reset_home_on_startup,omitempty"`

	// config appium settings
	SnapshotMaxDepth           int    `json:"snapshot_max_depth,omitempty" yaml:"snapshot_max_depth,omitempty"`
	AcceptAlertButtonSelector  string `json:"accept_alert_button_selector,omitempty" yaml:"accept_alert_button_selector,omitempty"`
	DismissAlertButtonSelector string `json:"dismiss_alert_button_selector,omitempty" yaml:"dismiss_alert_button_selector,omitempty"`
}

func (dev *IOSDevice) UUID() string {
	return dev.UDID
}

func (dev *IOSDevice) NewDriver(capabilities Capabilities) (driverExt *DriverExt, err error) {
	// init WDA driver
	if capabilities == nil {
		capabilities = NewCapabilities()
		capabilities.WithDefaultAlertAction(AlertActionAccept)
	}

	var driver WebDriver
	if env.WDA_USB_DRIVER == "" {
		// default use http driver
		driver, err = dev.NewHTTPDriver(capabilities)
	} else {
		driver, err = dev.NewUSBDriver(capabilities)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver")
	}

	if dev.ResetHomeOnStartup {
		log.Info().Msg("go back to home screen")
		if err = driver.Homescreen(); err != nil {
			return nil, errors.Wrap(code.MobileUIDriverError,
				fmt.Sprintf("go back to home screen failed: %v", err))
		}
	}

	driverExt, err = Extend(driver)
	if err != nil {
		return nil, errors.Wrap(code.MobileUIDriverError,
			fmt.Sprintf("extend WebDriver failed: %v", err))
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

	if dev.PerfOptions != nil {
		data, err := dev.d.PerfStart(dev.perfOpitons()...)
		if err != nil {
			return nil, err
		}

		driverExt.perfStop = make(chan struct{})
		// start performance monitor
		go func() {
			for {
				select {
				case <-driverExt.perfStop:
					dev.d.PerfStop()
					return
				case d := <-data:
					fmt.Println(string(d))
					driverExt.perfData = append(driverExt.perfData, string(d))
				}
			}
		}()
	}

	driverExt.UUID = dev.UUID()
	return driverExt, nil
}

func (dev *IOSDevice) forward(localPort, remotePort int) error {
	log.Info().Int("localPort", localPort).Int("remotePort", remotePort).
		Str("udid", dev.UDID).Msg("forward tcp port")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		log.Error().Err(err).Msg("listen tcp error")
		return err
	}

	go func(listener net.Listener, device gidevice.Device) {
		for {
			accept, err := listener.Accept()
			if err != nil {
				log.Error().Err(err).Msg("accept error")
				continue
			}

			rInnerConn, err := device.NewConnect(remotePort)
			if err != nil {
				log.Error().Err(err).Msg("connect to ios device failed")
				os.Exit(code.GetErrorCode(code.IOSDeviceConnectionError))
			}

			rConn := rInnerConn.RawConn()
			_ = rConn.SetDeadline(time.Time{})

			go func(lConn net.Conn) {
				go func(lConn, rConn net.Conn) {
					if _, err := io.Copy(lConn, rConn); err != nil {
						log.Error().Err(err).Msg("copy local -> remote")
					}
				}(lConn, rConn)
				go func(lConn, rConn net.Conn) {
					if _, err := io.Copy(rConn, lConn); err != nil {
						log.Error().Err(err).Msg("copy local <- remote")
					}
				}(lConn, rConn)
			}(accept)
		}
	}(listener, dev.d)

	return nil
}

func (dev *IOSDevice) perfOpitons() (perfOptions []gidevice.PerfOption) {
	if dev.PerfOptions == nil {
		return
	}

	// system
	if dev.PerfOptions.SysCPU {
		perfOptions = append(perfOptions, gidevice.WithPerfSystemCPU(true))
	}
	if dev.PerfOptions.SysMem {
		perfOptions = append(perfOptions, gidevice.WithPerfSystemMem(true))
	}
	if dev.PerfOptions.SysDisk {
		perfOptions = append(perfOptions, gidevice.WithPerfSystemDisk(true))
	}
	if dev.PerfOptions.SysNetwork {
		perfOptions = append(perfOptions, gidevice.WithPerfSystemNetwork(true))
	}
	if dev.PerfOptions.FPS {
		perfOptions = append(perfOptions, gidevice.WithPerfFPS(true))
	}
	if dev.PerfOptions.Network {
		perfOptions = append(perfOptions, gidevice.WithPerfNetwork(true))
	}

	// process
	if dev.PerfOptions.BundleID != "" {
		perfOptions = append(perfOptions,
			gidevice.WithPerfBundleID(dev.PerfOptions.BundleID))
	}
	if dev.PerfOptions.Pid != 0 {
		perfOptions = append(perfOptions,
			gidevice.WithPerfPID(dev.PerfOptions.Pid))
	}

	// config
	if dev.PerfOptions.OutputInterval != 0 {
		perfOptions = append(perfOptions,
			gidevice.WithPerfOutputInterval(dev.PerfOptions.OutputInterval))
	}
	if dev.PerfOptions.SystemAttributes != nil {
		perfOptions = append(perfOptions,
			gidevice.WithPerfSystemAttributes(dev.PerfOptions.SystemAttributes...))
	}
	if dev.PerfOptions.ProcessAttributes != nil {
		perfOptions = append(perfOptions,
			gidevice.WithPerfProcessAttributes(dev.PerfOptions.ProcessAttributes...))
	}
	return
}

// NewHTTPDriver creates new remote HTTP client, this will also start a new session.
func (dev *IOSDevice) NewHTTPDriver(capabilities Capabilities) (driver WebDriver, err error) {
	var localPort int
	localPort, err = strconv.Atoi(env.WDA_LOCAL_PORT)
	if err != nil {
		localPort, err = getFreePort()
		if err != nil {
			return nil, errors.Wrap(code.IOSDeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}

		if err = dev.forward(localPort, dev.Port); err != nil {
			return nil, errors.Wrap(code.IOSDeviceHTTPDriverError,
				fmt.Sprintf("forward tcp port failed: %v", err))
		}
	} else {
		log.Info().Int("WDA_LOCAL_PORT", localPort).Msg("reuse WDA local port")
	}

	var localMjpegPort int
	localMjpegPort, err = strconv.Atoi(env.WDA_LOCAL_MJPEG_PORT)
	if err != nil {
		localMjpegPort, err = getFreePort()
		if err != nil {
			return nil, errors.Wrap(code.IOSDeviceHTTPDriverError,
				fmt.Sprintf("get free port failed: %v", err))
		}
		if err = dev.forward(localMjpegPort, dev.MjpegPort); err != nil {
			return nil, errors.Wrap(code.IOSDeviceHTTPDriverError,
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
	wd.client = http.DefaultClient

	host := "127.0.0.1"
	if wd.urlPrefix, err = url.Parse(fmt.Sprintf("http://%s:%d", host, localPort)); err != nil {
		return nil, errors.Wrap(code.IOSDeviceHTTPDriverError, err.Error())
	}
	var sessionInfo SessionInfo
	if sessionInfo, err = wd.NewSession(capabilities); err != nil {
		return nil, errors.Wrap(code.IOSDeviceHTTPDriverError, err.Error())
	}
	wd.sessionId = sessionInfo.SessionId

	if wd.mjpegHTTPConn, err = net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, localMjpegPort),
	); err != nil {
		return nil, errors.Wrap(code.IOSDeviceHTTPDriverError, err.Error())
	}
	wd.mjpegClient = convertToHTTPClient(wd.mjpegHTTPConn)

	return wd, nil
}

// NewUSBDriver creates new client via USB connected device, this will also start a new session.
func (dev *IOSDevice) NewUSBDriver(capabilities Capabilities) (driver WebDriver, err error) {
	log.Info().Interface("capabilities", capabilities).
		Str("udid", dev.UDID).Msg("init WDA USB driver")

	wd := new(wdaDriver)
	if wd.defaultConn, err = dev.d.NewConnect(dev.Port, 0); err != nil {
		return nil, errors.Wrap(code.IOSDeviceUSBDriverError,
			fmt.Sprintf("connect port %d failed: %v", dev.Port, err))
	}
	wd.client = convertToHTTPClient(wd.defaultConn.RawConn())

	if wd.mjpegUSBConn, err = dev.d.NewConnect(dev.MjpegPort, 0); err != nil {
		return nil, errors.Wrap(code.IOSDeviceUSBDriverError,
			fmt.Sprintf("connect MJPEG port %d failed: %v", dev.MjpegPort, err))
	}
	wd.mjpegClient = convertToHTTPClient(wd.mjpegUSBConn.RawConn())

	if wd.urlPrefix, err = url.Parse("http://" + dev.UDID); err != nil {
		return nil, errors.Wrap(code.IOSDeviceUSBDriverError, err.Error())
	}
	if _, err = wd.NewSession(capabilities); err != nil {
		return nil, errors.Wrap(code.IOSDeviceUSBDriverError, err.Error())
	}

	return wd, nil
}

func (dev *IOSDevice) RunXCTest(bundleID string) (cancel context.CancelFunc, err error) {
	log.Info().Str("bundleID", bundleID).Msg("run xctest")
	out, cancel, err := dev.d.XCTest(bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "run xctest failed")
	}
	// wait for xctest to start
	time.Sleep(5 * time.Second)

	f, err := os.OpenFile(fmt.Sprintf("xctest_%s.log", dev.UDID),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		return nil, err
	}
	defer builtinLog.SetOutput(f)

	// print xctest running logs
	go func() {
		for s := range out {
			builtinLog.Print(s)
		}
		f.Close()
	}()

	return cancel, nil
}

func (dExt *DriverExt) ConnectMjpegStream(httpClient *http.Client) (err error) {
	if httpClient == nil {
		return errors.New(`'httpClient' can't be nil`)
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, "http://*", nil); err != nil {
		return err
	}

	var resp *http.Response
	if resp, err = httpClient.Do(req); err != nil {
		return err
	}
	// defer func() { _ = resp.Body.Close() }()

	var boundary string
	if _, param, err := mime.ParseMediaType(resp.Header.Get("Content-Type")); err != nil {
		return err
	} else {
		boundary = strings.Trim(param["boundary"], "-")
	}

	mjpegReader := multipart.NewReader(resp.Body, boundary)

	go func() {
		for {
			select {
			case <-dExt.doneMjpegStream:
				_ = resp.Body.Close()
				return
			default:
				var part *multipart.Part
				if part, err = mjpegReader.NextPart(); err != nil {
					dExt.frame = nil
					continue
				}

				raw := new(bytes.Buffer)
				if _, err = raw.ReadFrom(part); err != nil {
					dExt.frame = nil
					continue
				}
				dExt.frame = raw
			}
		}
	}()

	return
}

func (dExt *DriverExt) CloseMjpegStream() {
	dExt.doneMjpegStream <- true
}

type rawResponse []byte

func (r rawResponse) checkErr() (err error) {
	reply := new(struct {
		Value struct {
			Err        string `json:"error"`
			Message    string `json:"message"`
			Traceback  string `json:"traceback"`  // wda
			Stacktrace string `json:"stacktrace"` // uia
		}
	})
	if err = json.Unmarshal(r, reply); err != nil {
		return err
	}
	if reply.Value.Err != "" {
		errText := reply.Value.Message
		re := regexp.MustCompile(`{.+?=(.+?)}`)
		if re.MatchString(reply.Value.Message) {
			subMatch := re.FindStringSubmatch(reply.Value.Message)
			errText = subMatch[len(subMatch)-1]
		}
		return fmt.Errorf("%s: %s", reply.Value.Err, errText)
	}
	return
}

func (r rawResponse) valueConvertToString() (s string, err error) {
	reply := new(struct{ Value string })
	if err = json.Unmarshal(r, reply); err != nil {
		return "", errors.Wrapf(err, "json.Unmarshal failed, rawResponse: %s", string(r))
	}
	s = reply.Value
	return
}

func (r rawResponse) valueConvertToBool() (b bool, err error) {
	reply := new(struct{ Value bool })
	if err = json.Unmarshal(r, reply); err != nil {
		return false, err
	}
	b = reply.Value
	return
}

func (r rawResponse) valueConvertToSessionInfo() (sessionInfo SessionInfo, err error) {
	reply := new(struct{ Value struct{ SessionInfo } })
	if err = json.Unmarshal(r, reply); err != nil {
		return SessionInfo{}, err
	}
	sessionInfo = reply.Value.SessionInfo
	return
}

func (r rawResponse) valueConvertToJsonRawMessage() (raw builtinJSON.RawMessage, err error) {
	reply := new(struct{ Value builtinJSON.RawMessage })
	if err = json.Unmarshal(r, reply); err != nil {
		return nil, err
	}
	raw = reply.Value
	return
}

func (r rawResponse) valueDecodeAsBase64() (raw *bytes.Buffer, err error) {
	str, err := r.valueConvertToString()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert value to string")
	}
	decodeString, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode base64 string")
	}
	raw = bytes.NewBuffer(decodeString)
	return
}

var errNoSuchElement = errors.New("no such element")

func (r rawResponse) valueConvertToElementID() (id string, err error) {
	reply := new(struct{ Value map[string]string })
	if err = json.Unmarshal(r, reply); err != nil {
		return "", err
	}
	if len(reply.Value) == 0 {
		return "", errNoSuchElement
	}
	if id = elementIDFromValue(reply.Value); id == "" {
		return "", fmt.Errorf("invalid element returned: %+v", reply)
	}
	return
}

func (r rawResponse) valueConvertToElementIDs() (IDs []string, err error) {
	reply := new(struct{ Value []map[string]string })
	if err = json.Unmarshal(r, reply); err != nil {
		return nil, err
	}
	if len(reply.Value) == 0 {
		return nil, errNoSuchElement
	}
	IDs = make([]string, len(reply.Value))
	for i, elem := range reply.Value {
		var id string
		if id = elementIDFromValue(elem); id == "" {
			return nil, fmt.Errorf("invalid element returned: %+v", reply)
		}
		IDs[i] = id
	}
	return
}
