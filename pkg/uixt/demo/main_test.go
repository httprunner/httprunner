//go:build localtest

package demo

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

func TestIOSDemo(t *testing.T) {
	device, err := uixt.NewIOSDevice(
		options.WithWDAPort(8700),
		options.WithWDAMjpegPort(8800),
		options.WithResetHomeOnStartup(false), // not reset home on startup
	)
	if err != nil {
		t.Fatal(err)
	}

	capabilities := uixt.NewCapabilities()
	capabilities.WithDefaultAlertAction(uixt.AlertActionAccept) // or uixt.AlertActionDismiss
	driverExt, err := device.NewDriver(uixt.WithDriverCapabilities(capabilities))
	if err != nil {
		t.Fatal(err)
	}

	// release session
	defer func() {
		driverExt.Driver.DeleteSession()
	}()

	// 持续监测手机屏幕，直到出现青少年模式弹窗后，点击「我知道了」
	for {
		// take screenshot and get screen texts by OCR
		ocrTexts, err := driverExt.GetScreenTexts()
		if err != nil {
			log.Error().Err(err).Msg("OCR GetTexts failed")
			t.Fatal(err)
		}

		points, err := ocrTexts.FindTexts([]string{"青少年模式", "我知道了"})
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		point := points[1].Center()
		err = driverExt.TapAbsXY(point.X, point.Y)
		if err != nil {
			t.Fatal(err)
		}
	}
}
