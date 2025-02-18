//go:build localtest

package uixt

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/stretchr/testify/assert"
)

func TestNewDriver1(t *testing.T) {
	device, _ := NewAndroidDevice(option.WithUIA2(true))
	driver, _ := device.NewDriver()
	driverExt := NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
	driverExt.TapByOCR("推荐")
}

func TestNewDriver2(t *testing.T) {
	device, _ := NewAndroidDevice()
	driver, _ := NewUIA2Driver(device)
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

func TestAndroidSwipeAction(t *testing.T) {
	driver := setupDriverExt(t)

	swipeAction := prepareSwipeAction(driver, "up", option.WithDirection("down"))
	err := swipeAction(driver)
	assert.Nil(t, err)

	swipeAction = prepareSwipeAction(driver, "up", option.WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(driver)
	assert.Nil(t, err)
}

func TestAndroidSwipeToTapApp(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.SwipeToTapApp("抖音")
	assert.Nil(t, err)
}

func TestAndroidSwipeToTapTexts(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.AppLaunch("com.ss.android.ugc.aweme")
	assert.Nil(t, err)

	err = driver.swipeToTapTexts([]string{"点击进入直播间", "直播中"}, option.WithDirection("up"))
	assert.Nil(t, err)
}

func TestGetScreenShot(t *testing.T) {
	driver := setupADBDriverExt(t)

	imagePath := filepath.Join(config.ScreenShotsPath, "test_screenshot")
	_, err := driver.ScreenShot(option.WithScreenShotFileName(imagePath))
	if err != nil {
		t.Fatalf("GetScreenShot failed: %v", err)
	}

	t.Logf("screenshot saved at: %s", imagePath)
}

func TestCheckPopup(t *testing.T) {
	driver := setupADBDriverExt(t)
	popup, err := driver.CheckPopup()
	if err != nil {
		t.Logf("check popup failed, err: %v", err)
	} else if popup == nil {
		t.Log("no popup found")
	} else {
		t.Logf("found popup: %v", popup)
	}
}

func TestClosePopup(t *testing.T) {
	driver := setupADBDriverExt(t)
	if err := driver.ClosePopupsHandler(); err != nil {
		t.Fatal(err)
	}
}

func matchPopup(text string) bool {
	for _, popup := range popups {
		if regexp.MustCompile(popup[1]).MatchString(text) {
			return true
		}
	}
	return false
}

func TestMatchRegex(t *testing.T) {
	testData := []string{
		"以后再说", "我知道了", "同意", "拒绝", "稍后",
		"始终允许", "继续使用", "仅在使用中允许",
	}
	for _, text := range testData {
		if !matchPopup(text) {
			t.Fatal(text)
		}
	}
}
