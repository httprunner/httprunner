package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/ai"
)

func TestNewDriverExt(t *testing.T) {
	device, _ := NewAndroidDevice()
	var driver IDriver
	var err error
	if device.UIA2 || device.LogOn {
		driver, err = NewUIA2Driver(device)
	} else if device.STUB {
		driver, err = NewStubAndroidDriver(device)
	} else {
		driver, err = NewADBDriver(device)
	}
	if err != nil {
		t.Fatal(err)
	}

	driverExt, _ := NewDriverExt(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))

	driverExt.GetDriver()
	t.Log(driverExt)
}
