package uixt

import (
	"testing"

	"github.com/electricbubble/gwda"
)

func TestDriverExt_Drag(t *testing.T) {
	driver, err := gwda.NewUSBDriver(nil)
	checkErr(t, err)

	driverExt, err := Extend(driver, 0.95)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/IMG_map.png"

	// err = driverExt.Drag(pathSearch, 300, 500, 2)
	// checkErr(t, err)

	err = driverExt.DragOffset(pathSearch, 300, 500, 2.1, 0.5, 2)
	checkErr(t, err)
}
