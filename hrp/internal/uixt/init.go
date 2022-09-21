package uixt

import (
	"github.com/electricbubble/gwda"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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

func InitWDAClient(options ...gwda.DeviceOption) (*DriverExt, error) {
	// init wda device
	targetDevice, err := gwda.NewDevice(options...)
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

	return driverExt, nil
}
