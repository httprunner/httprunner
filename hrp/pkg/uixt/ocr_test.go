//go:build ocr

package uixt

import (
	"testing"
)

func TestDriverExtOCR(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	point, err := driverExt.FindScreenText("抖音")
	checkErr(t, err)

	t.Logf("point.X: %v, point.Y: %v", point.X, point.Y)
	driverExt.Driver.TapFloat(point.X, point.Y-20)
}
