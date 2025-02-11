package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
)

func TestNewDriverExt(t *testing.T) {
	device, _ := NewAndroidDevice()
	var driver IDriver
	var err error
	driver, err = NewADBDriver(device)
	if err != nil {
		t.Fatal(err)
	}

	driverExt := NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))

	texts, _ := driverExt.GetScreenTexts()
	t.Log(texts)

	// get original dirver
	driver = driverExt.IDriver.(*ADBDriver)

	// get device
	device = driver.GetDevice().(*AndroidDevice)
}
