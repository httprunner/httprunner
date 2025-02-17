//go:build localtest

package uixt

import (
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
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

func setupDriverExt(t *testing.T, driverType ...string) *XTDriver {
	var dType string
	if len(driverType) > 0 {
		dType = driverType[0]
	}
	switch dType {
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

func TestDriverExt_TapXY(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.TapXY(0.4, 0.5)
	checkErr(t, err)
}

func TestDriverExt_TapAbsXY(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.TapAbsXY(100, 300)
	checkErr(t, err)
}

func TestAndroidSwipeAction(t *testing.T) {
	driver := setupDriverExt(t)

	swipeAction := prepareSwipeAction(driver, "up", option.WithDirection("down"))
	err := swipeAction(driver)
	checkErr(t, err)

	swipeAction = prepareSwipeAction(driver, "up", option.WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(driver)
	checkErr(t, err)
}

func TestAndroidSwipeToTapApp(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.SwipeToTapApp("抖音")
	checkErr(t, err)
}

func TestAndroidSwipeToTapTexts(t *testing.T) {
	driver := setupDriverExt(t)
	err := driver.AppLaunch("com.ss.android.ugc.aweme")
	checkErr(t, err)

	err = driver.swipeToTapTexts([]string{"点击进入直播间", "直播中"}, option.WithDirection("up"))
	checkErr(t, err)
}

func checkErr(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg, err)
		}
	}
}

func TestGetSimulationDuration(t *testing.T) {
	params := []float64{1.23}
	duration := getSimulationDuration(params)
	if duration != 1230 {
		t.Fatal("getSimulationDuration failed")
	}

	params = []float64{1, 2}
	duration = getSimulationDuration(params)
	if duration < 1000 || duration > 2000 {
		t.Fatal("getSimulationDuration failed")
	}

	params = []float64{1, 5, 0.7, 5, 10, 0.3}
	duration = getSimulationDuration(params)
	if duration < 1000 || duration > 10000 {
		t.Fatal("getSimulationDuration failed")
	}
}

func TestSleepStrict(t *testing.T) {
	startTime := time.Now()
	sleepStrict(startTime, 1230)
	dur := time.Since(startTime).Milliseconds()
	t.Log(dur)
	if dur < 1230 || dur > 1300 {
		t.Fatalf("sleepRandom failed, dur: %d", dur)
	}
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
