//go:build localtest

package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriverExt_NewMethod1(t *testing.T) {
	device, err := NewAndroidDevice(option.WithUIA2(true))
	require.Nil(t, err)
	driver, err := device.NewDriver()
	require.Nil(t, err)
	driverExt := NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
	driverExt.TapByOCR("推荐")
}

func TestDriverExt_NewMethod2(t *testing.T) {
	device, err := NewAndroidDevice()
	require.Nil(t, err)
	driver, err := NewUIA2Driver(device)
	require.Nil(t, err)
	driverExt := NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
	driverExt.TapByOCR("推荐")
}

func TestDriverExt(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, _ := NewADBDriver(device)
	driverExt := NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))

	// call IDriver methods
	driverExt.TapXY(0.2, 0.5)
	driverExt.Swipe(0.2, 0.5, 0.8, 0.5)
	driverExt.AppLaunch("com.ss.android.ugc.aweme")

	// call AI extended methods
	driverExt.TapByOCR("推荐")
	texts, _ := driverExt.GetScreenTexts()
	t.Log(texts)
	point, _ := driverExt.FindScreenText("hello")
	t.Log(point)

	// call IDriver methods
	driverExt.GetDevice().Install("/path/to/app")
	driverExt.GetDevice().GetPackageInfo("com.ss.android.ugc.aweme")

	// get original driver and call its methods
	adbDriver := driverExt.IDriver.(*ADBDriver)
	adbDriver.TapByHierarchy("hello")
	wdaDriver := driverExt.IDriver.(*WDADriver)
	wdaDriver.GetMjpegClient()
	wdaDriver.Scale()

	// get original device and call its methods
	androidDevice := driver.GetDevice().(*AndroidDevice)
	androidDevice.InstallAPK("/path/to/app.apk")
}

var driverType = "ADB"

func setupDriverExt(t *testing.T) *XTDriver {
	switch driverType {
	case "ADB":
		return setupADBDriverExt(t)
	case "UIA2":
		return setupUIA2DriverExt(t)
	case "WDA":
		return setupWDADriverExt(t)
	case "HDC":
		return setupHDCDriverExt(t)
	default:
		return setupADBDriverExt(t)
	}
}

func TestDriverExt_TapByOCR(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.TapByOCR("天气")
	assert.Nil(t, err)
}

func TestDriverExt_prepareSwipeAction(t *testing.T) {
	driver := setupDriverExt(t)

	swipeAction := prepareSwipeAction(driver, "up", option.WithDirection("down"))
	err := swipeAction(driver)
	assert.Nil(t, err)

	swipeAction = prepareSwipeAction(driver, "up", option.WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(driver)
	assert.Nil(t, err)
}

func TestDriverExt_SwipeToTapApp(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.SwipeToTapApp("抖音")
	assert.Nil(t, err)
}

func TestDriverExt_SwipeToTapTexts(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.AppLaunch("com.ss.android.ugc.aweme")
	assert.Nil(t, err)

	err = driver.SwipeToTapTexts(
		[]string{"点击进入直播间", "直播中"},
		option.WithDirection("up"),
		option.WithMaxRetryTimes(10))
	assert.Nil(t, err)
}

func TestDriverExt_CheckPopup(t *testing.T) {
	driver := setupADBDriverExt(t)
	popup, err := driver.CheckPopup()
	require.Nil(t, err)
	if popup == nil {
		t.Log("no popup found")
	} else {
		t.Logf("found popup: %v", popup)
	}
}

func TestDriverExt_ClosePopupsHandler(t *testing.T) {
	driver := setupADBDriverExt(t)
	err := driver.ClosePopupsHandler()
	assert.Nil(t, err)
}
