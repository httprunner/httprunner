//go:build ocr

package uixt

import (
	"testing"
)

func TestDriverExtOCR(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	x, y, width, height, err := driverExt.FindTextByOCR("抖音")
	checkErr(t, err)

	t.Logf("x: %v, y: %v, width: %v, height: %v", x, y, width, height)
	driverExt.Driver.TapFloat(x+width*0.5, y+height*0.5-20)
}
