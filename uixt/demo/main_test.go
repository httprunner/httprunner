//go:build localtest

package demo

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestIOSDemo(t *testing.T) {
	device, err := uixt.NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithResetHomeOnStartup(false), // not reset home on startup
	)
	if err != nil {
		t.Fatal(err)
	}

	driver, err := device.NewDriver()
	if err != nil {
		t.Fatal(err)
	}
	driverExt := uixt.NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM),
	)

	// release session
	defer func() {
		driverExt.DeleteSession()
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
