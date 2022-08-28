package uixt

import (
	"testing"

	"github.com/electricbubble/gwda"
)

func TestDriverExtOCR(t *testing.T) {
	driver, err := gwda.NewUSBDriver(nil)
	checkErr(t, err)

	driverExt, err := Extend(driver, 0.95)
	checkErr(t, err)

	x, y, width, height, err := driverExt.FindTextByOCR("抖音")
	checkErr(t, err)

	t.Logf("x: %v, y: %v, width: %v, height: %v", x, y, width, height)
	driver.TapFloat(x, y-20)
}
