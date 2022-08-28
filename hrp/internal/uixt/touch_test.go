package uixt

import (
	"testing"

	"github.com/electricbubble/gwda"
)

func TestDriverExt_ForceTouch(t *testing.T) {
	driver, err := gwda.NewUSBDriver(nil)
	checkErr(t, err)

	driverExt, err := Extend(driver, 0.95)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/IMG_ft.png"

	err = driverExt.ForceTouch(pathSearch, 0.5, 3)
	checkErr(t, err)

	// err = driverExt.ForceTouchOffset(pathSearch, 0.5, 0.1, 0.9)
	// checkErr(t, err)

	// err = driverExt.ForceTouchOffset(pathSearch, 0.2, 1.1, -1)
	// checkErr(t, err)
}

func TestDriverExt_TouchAndHold(t *testing.T) {
	driver, err := gwda.NewUSBDriver(nil)
	checkErr(t, err)

	driverExt, err := Extend(driver, 0.95)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/IMG_ft.png"

	// err = driverExt.TouchAndHold(pathSearch)
	// checkErr(t, err)

	// err = driverExt.TouchAndHold(pathSearch, 3)
	// checkErr(t, err)

	err = driverExt.TouchAndHoldOffset(pathSearch, 0.8, 0.1)
	checkErr(t, err)
}

func checkErr(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg, err)
		}
	}
}
