//go:build localtest

package uixt

import (
	"testing"
)

func TestDriverExt_Drag(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/IMG_map.png"

	// err = driverExt.Drag(pathSearch, 300, 500, 2)
	// checkErr(t, err)

	err = driverExt.DragOffset(pathSearch, 300, 500, 2.1, 0.5, 2)
	checkErr(t, err)
}
