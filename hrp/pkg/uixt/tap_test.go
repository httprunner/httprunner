//go:build localtest

package uixt

import (
	"testing"
)

var iosDevice *IOSDevice

func init() {
	iosDevice, _ = NewIOSDevice()
}

func TestDriverExt_TapXY(t *testing.T) {
	driverExt, err := iosDevice.NewDriver()
	checkErr(t, err)

	err = driverExt.TapXY(0.4, 0.5)
	checkErr(t, err)
}

func TestDriverExt_TapAbsXY(t *testing.T) {
	driverExt, err := iosDevice.NewDriver()
	checkErr(t, err)

	err = driverExt.TapAbsXY(100, 300)
	checkErr(t, err)
}

func TestDriverExt_TapWithOCR(t *testing.T) {
	driverExt, err := iosDevice.NewDriver()
	checkErr(t, err)

	// 需要点击文字上方的图标
	err = driverExt.TapOffset("抖音", 0, -20)
	checkErr(t, err)
}
