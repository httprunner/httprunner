//go:build localtest

package demo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestIOSDemo(t *testing.T) {
	t.Skip("Skip demo test - requires physical mobile device")
	device, err := uixt.NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithResetHomeOnStartup(false), // not reset home on startup
	)
	assert.Nil(t, err)

	driver, err := device.NewDriver()
	assert.Nil(t, err)
	driverExt, err := uixt.NewXTDriver(driver,
		option.WithCVService(option.CVServiceTypeVEDEM),
	)
	assert.Nil(t, err)

	// release session
	defer func() {
		driverExt.DeleteSession()
	}()

	// 持续监测手机屏幕，直到出现青少年模式弹窗后，点击「我知道了」
	for {
		// take screenshot and get screen texts by OCR
		ocrTexts, err := driverExt.GetScreenTexts()
		assert.Nil(t, err)

		points, err := ocrTexts.FindTexts([]string{"青少年模式", "我知道了"})
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		point := points[1].Center()
		err = driverExt.TapAbsXY(point.X, point.Y)
		assert.Nil(t, err)
	}
}
