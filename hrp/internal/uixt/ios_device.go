package uixt

import (
	"bytes"
	"encoding/base64"
	builtinJSON "encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/json"
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

const (
	defaultWDAPort   = 8100
	defaultMjpegPort = 9100
)

func InitWDAClient(device *IOSDevice) (*DriverExt, error) {
	// init wda device
	iosDevice, err := NewIOSDevice(device.opitons()...)
	if err != nil {
		return nil, err
	}

	// init WDA driver
	capabilities := NewCapabilities()
	capabilities.WithDefaultAlertAction(AlertActionAccept)
	var driver WebDriver
	if iosDevice.LocalPort != 0 && iosDevice.LocalMjpegPort != 0 {
		driver, err = iosDevice.NewHTTPDriver(capabilities)
	} else {
		driver, err = iosDevice.NewUSBDriver(capabilities)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver")
	}

	// switch to iOS springboard before init WDA session
	// aviod getting stuck when some super app is activate such as douyin or wexin
	log.Info().Msg("go back to home screen")
	if err = driver.Homescreen(); err != nil {
		return nil, errors.Wrap(err, "failed to go back to home screen")
	}

	driverExt, err := Extend(driver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extend WebDriver")
	}
	settings, err := driverExt.Driver.SetAppiumSettings(map[string]interface{}{
		"snapshotMaxDepth":          snapshotMaxDepth,
		"acceptAlertButtonSelector": acceptAlertButtonSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set appium WDA settings")
	}
	log.Info().Interface("appiumWDASettings", settings).Msg("set appium WDA settings")

	if device.LogOn {
		err = driverExt.Driver.StartCaptureLog("hrp_wda_log")
		if err != nil {
			return nil, err
		}
	}

	driverExt.UUID = iosDevice.UUID()
	return driverExt, nil
}

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

func WithWDALocalPort(port int) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.LocalPort = port
	}
}

func WithWDALocalMjpegPort(port int) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.LocalMjpegPort = port
	}
}

func WithLogOn(logOn bool) IOSDeviceOption {
	return func(device *IOSDevice) {
		device.LogOn = logOn
	}
}

func NewIOSDevice(options ...IOSDeviceOption) (device *IOSDevice, err error) {
	var usbmux giDevice.Usbmux
	if usbmux, err = giDevice.NewUsbmux(); err != nil {
		return nil, fmt.Errorf("init usbmux failed: %v", err)
	}

	var deviceList []giDevice.Device
	if deviceList, err = usbmux.Devices(); err != nil {
		return nil, fmt.Errorf("get attached devices failed: %v", err)
	}

	device = &IOSDevice{
		Port:      defaultWDAPort,
		MjpegPort: defaultMjpegPort,
	}
	for _, option := range options {
		option(device)
	}

	serialNumber := device.UDID
	for _, dev := range deviceList {
		// find device by serial number if specified
		if serialNumber != "" && dev.Properties().SerialNumber != serialNumber {
			continue
		}

		device.UDID = dev.Properties().SerialNumber
		device.d = dev
		return device, nil
	}

	return nil, fmt.Errorf("device %s not found", device.UDID)
}

type IOSDevice struct {
	d              giDevice.Device
	UDID           string `json:"udid,omitempty" yaml:"udid,omitempty"`
	Port           int    `json:"port,omitempty" yaml:"port,omitempty"`                         // WDA remote port
	MjpegPort      int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"`             // WDA remote MJPEG port
	LocalPort      int    `json:"local_port,omitempty" yaml:"local_port,omitempty"`             // WDA local port
	LocalMjpegPort int    `json:"local_mjpeg_port,omitempty" yaml:"local_mjpeg_port,omitempty"` // WDA local MJPEG port
	LogOn          bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

func (dev *IOSDevice) UUID() string {
	return dev.UDID
}

func (dev *IOSDevice) opitons() (deviceOptions []IOSDeviceOption) {
	if dev.UDID != "" {
		deviceOptions = append(deviceOptions, WithUDID(dev.UDID))
	}
	if dev.Port != 0 {
		deviceOptions = append(deviceOptions, WithWDAPort(dev.Port))
	}
	if dev.MjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDAMjpegPort(dev.MjpegPort))
	}

	if wda_port := os.Getenv("WDA_LOCAL_PORT"); wda_port != "" {
		if port, err := strconv.Atoi(wda_port); err == nil {
			log.Info().Int("WDA_LOCAL_PORT", port).
				Msg("override with environment variable")
			dev.LocalPort = port
		} else {
			log.Error().Err(err).Str("WDA_LOCAL_PORT", wda_port).
				Msg("invalid WDA_LOCAL_PORT, ignored")
		}
	}
	if wda_mjpeg_port := os.Getenv("WDA_LOCAL_MJPEG_PORT"); wda_mjpeg_port != "" {
		if mjpeg_port, err := strconv.Atoi(wda_mjpeg_port); err == nil {
			log.Info().Int("WDA_LOCAL_MJPEG_PORT", mjpeg_port).
				Msg("override with environment variable")
			dev.LocalMjpegPort = mjpeg_port
		} else {
			log.Error().Err(err).Str("WDA_LOCAL_MJPEG_PORT", wda_mjpeg_port).
				Msg("invalid WDA_LOCAL_MJPEG_PORT, ignored")
		}
	}
	if dev.LocalPort != 0 {
		deviceOptions = append(deviceOptions, WithWDALocalPort(dev.LocalPort))
	}
	if dev.LocalMjpegPort != 0 {
		deviceOptions = append(deviceOptions, WithWDALocalMjpegPort(dev.LocalMjpegPort))
	}

	return
}

// NewHTTPDriver creates new remote HTTP client, this will also start a new session.
// WDA port and mjpeg port must be proxied to local ports:
// iproxy -u UDID WDA_LOCAL_PORT WDA_PORT
// iproxy -u UDID WDA_LOCAL_MJPEG_PORT WDA_MJPEG_PORT
func (dev *IOSDevice) NewHTTPDriver(capabilities Capabilities) (driver WebDriver, err error) {
	host := "127.0.0.1"
	log.Info().Interface("capabilities", capabilities).
		Str("host", host).Msg("init WDA HTTP driver")
	wd := new(wdaDriver)
	wd.client = http.DefaultClient

	if wd.urlPrefix, err = url.Parse(fmt.Sprintf("http://%s:%d", host, dev.LocalPort)); err != nil {
		return nil, err
	}
	var sessionInfo SessionInfo
	if sessionInfo, err = wd.NewSession(capabilities); err != nil {
		return nil, err
	}
	wd.sessionId = sessionInfo.SessionId

	if wd.mjpegHTTPConn, err = net.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", host, dev.LocalMjpegPort),
	); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("connect port %d failed: %w",
			dev.Port, err)
	}
	wd.client = convertToHTTPClient(wd.defaultConn.RawConn())

	if wd.mjpegUSBConn, err = dev.d.NewConnect(dev.MjpegPort, 0); err != nil {
		return nil, fmt.Errorf("connect MJPEG port %d failed: %w",
			dev.MjpegPort, err)
	}
	wd.mjpegClient = convertToHTTPClient(wd.mjpegUSBConn.RawConn())

	if wd.urlPrefix, err = url.Parse("http://" + dev.UDID); err != nil {
		return nil, err
	}
	_, err = wd.NewSession(capabilities)

	return wd, err
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
