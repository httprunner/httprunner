//go:build localtest

package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
)

var (
	iosDevice    *IOSDevice
	iosDriverExt *XTDriver
)

func init() {
	iosDevice, _ = NewIOSDevice()
	driver, _ := iosDevice.NewDriver()
	iosDriverExt = NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
}

func TestDriverExt_TapXY(t *testing.T) {
	err := iosDriverExt.TapXY(0.4, 0.5)
	checkErr(t, err)
}

func TestDriverExt_TapAbsXY(t *testing.T) {
	err := iosDriverExt.TapAbsXY(100, 300)
	checkErr(t, err)
}
