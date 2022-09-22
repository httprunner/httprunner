package uixt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/electricbubble/gwda"
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

type WDAOptions struct {
	UDID      string `json:"udid,omitempty" yaml:"udid,omitempty"`
	Port      int    `json:"port,omitempty" yaml:"port,omitempty"`
	MjpegPort int    `json:"mjpeg_port,omitempty" yaml:"mjpeg_port,omitempty"`
	LogOn     bool   `json:"log_on,omitempty" yaml:"log_on,omitempty"`
}

type WDAOption func(*WDAOptions)

func WithUDID(udid string) WDAOption {
	return func(device *WDAOptions) {
		device.UDID = udid
	}
}

func WithPort(port int) WDAOption {
	return func(device *WDAOptions) {
		device.Port = port
	}
}

func WithMjpegPort(port int) WDAOption {
	return func(device *WDAOptions) {
		device.MjpegPort = port
	}
}

func WithLogOn(logOn bool) WDAOption {
	return func(device *WDAOptions) {
		device.LogOn = logOn
	}
}

func InitWDAClient(options *WDAOptions) (*DriverExt, error) {
	var deviceOptions []gwda.DeviceOption
	if options.UDID != "" {
		deviceOptions = append(deviceOptions, gwda.WithSerialNumber(options.UDID))
	}
	if options.Port != 0 {
		deviceOptions = append(deviceOptions, gwda.WithPort(options.Port))
	}
	if options.MjpegPort != 0 {
		deviceOptions = append(deviceOptions, gwda.WithMjpegPort(options.MjpegPort))
	}

	// init wda device
	targetDevice, err := gwda.NewDevice(deviceOptions...)
	if err != nil {
		return nil, err
	}

	// switch to iOS springboard before init WDA session
	// aviod getting stuck when some super app is activate such as douyin or wexin
	log.Info().Msg("switch to iOS springboard")
	bundleID := "com.apple.springboard"
	_, err = targetDevice.GIDevice().AppLaunch(bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "launch springboard failed")
	}

	// init WDA driver
	gwda.SetDebug(true)
	capabilities := gwda.NewCapabilities()
	capabilities.WithDefaultAlertAction(gwda.AlertActionAccept)
	driver, err := gwda.NewUSBDriver(capabilities, *targetDevice)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init WDA driver")
	}
	driverExt, err := Extend(driver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extend gwda.WebDriver")
	}
	settings, err := driverExt.SetAppiumSettings(map[string]interface{}{
		"snapshotMaxDepth":          snapshotMaxDepth,
		"acceptAlertButtonSelector": acceptAlertButtonSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set appium WDA settings")
	}
	log.Info().Interface("appiumWDASettings", settings).Msg("set appium WDA settings")

	driverExt.host = fmt.Sprintf("http://127.0.0.1:%d", targetDevice.Port)
	if options.LogOn {
		err = driverExt.StartWDALog("hrp_wda_log")
		if err != nil {
			return nil, err
		}
	}

	return driverExt, nil
}

type wdaResponse struct {
	Value     string `json:"value"`
	SessionID string `json:"sessionId"`
}

func (dExt *DriverExt) StartWDALog(identifier string) error {
	log.Info().Msg("start WDA log recording")
	data := map[string]interface{}{"action": "start", "type": 2, "identifier": identifier}
	_, err := dExt.triggerWDALog(data)
	if err != nil {
		return errors.Wrap(err, "failed to start WDA log recording")
	}

	return nil
}

func (dExt *DriverExt) GetWDALog() (string, error) {
	log.Info().Msg("stop WDA log recording")
	data := map[string]interface{}{"action": "stop"}
	reply, err := dExt.triggerWDALog(data)
	if err != nil {
		return "", errors.Wrap(err, "failed to get WDA logs")
	}

	return reply.Value, nil
}

func (dExt *DriverExt) triggerWDALog(data map[string]interface{}) (*wdaResponse, error) {
	// [[FBRoute POST:@"/gtf/automation/log"].withoutSession respondWithTarget:self action:@selector(handleAutomationLog:)]
	postJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/gtf/automation/log", dExt.host)
	log.Info().Str("url", url).Interface("data", data).Msg("trigger WDA log")
	resp, err := http.DefaultClient.Post(url, "application/json", bytes.NewBuffer(postJSON))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to trigger wda log, response status code: %d", resp.StatusCode)
	}

	rawResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reply := new(wdaResponse)
	if err = json.Unmarshal(rawResp, reply); err != nil {
		return nil, err
	}

	return reply, nil
}
