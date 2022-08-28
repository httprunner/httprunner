package uixt

import (
	"testing"

	"github.com/electricbubble/gwda"
)

func TestDriverExt_Swipe(t *testing.T) {
	driver, err := gwda.NewUSBDriver(nil)
	checkErr(t, err)

	driverExt, err := Extend(driver, 0.95)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/flag7.png"

	// gwda.SetDebug(true)

	err = driverExt.Swipe(pathSearch, 300, 500)
	checkErr(t, err)

	err = driverExt.SwipeFloat(pathSearch, 300.9, 500)
	checkErr(t, err)

	err = driverExt.SwipeOffset(pathSearch, 300, 500, 0.2, 0.5)
	checkErr(t, err)

	driverExt.Debug(DmNotMatch)

	err = driverExt.OnlyOnceThreshold(0.92).SwipeOffsetFloat(pathSearch, 300.9, 499.1, 0.2, 0.5)
	checkErr(t, err)
}
