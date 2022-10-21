//go:build localtest

package uixt

import (
	"testing"
)

var iosDevice *IOSDevice

func init() {
	iosDevice, _ = NewIOSDevice()
}

func TestDriverExt_TapWithNumber(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/flag7.png"

	err = driverExt.TapWithNumber(pathSearch, 3)
	checkErr(t, err)

	err = driverExt.TapWithNumberOffset(pathSearch, 3, 0.5, 0.75)
	checkErr(t, err)
}

func TestDriverExt_TapXY(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	err = driverExt.TapXY(0.4, 0.5)
	checkErr(t, err)
}

func TestDriverExt_TapAbsXY(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	err = driverExt.TapAbsXY(100, 300)
	checkErr(t, err)
}

func TestDriverExt_TapWithOCR(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	// 需要点击文字上方的图标
	err = driverExt.TapOffset("抖音", 0.5, -1)
	checkErr(t, err)
}
