//go:build localtest

package uixt

import (
	"testing"
)

var iosDevice *IOSDevice

func init() {
	iosDevice, _ = NewIOSDevice()
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
