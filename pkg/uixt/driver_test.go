package uixt

import (
	"testing"

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
