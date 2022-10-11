package demo

import (
	"testing"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestIOSDemo(t *testing.T) {
	device, err := uixt.NewIOSDevice(
		uixt.WithWDAPort(8700), uixt.WithWDAMjpegPort(8800),
		uixt.WithResetHomeOnStartup(false), // not reset home on startup
	)
	if err != nil {
		t.Fatal(err)
	}

	capabilities := uixt.NewCapabilities()
	capabilities.WithDefaultAlertAction(uixt.AlertActionAccept) // or uixt.AlertActionDismiss
	driverExt, err := device.NewDriver(capabilities)
	if err != nil {
		t.Fatal(err)
	}

	// 持续监测手机屏幕，直到出现青少年模式弹窗后，点击「我知道了」
	for {
		_, err1 := driverExt.GetTextXY("青少年模式")
		point, err2 := driverExt.GetTextXY("我知道了")
		if err1 != nil || err2 != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		err := driverExt.TapAbsXY(point.X, point.Y, "")
		if err != nil {
			t.Fatal(err)
		}
	}
}
