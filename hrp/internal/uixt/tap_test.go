package uixt

import (
	"testing"
)

func TestDriverExt_TapWithNumber(t *testing.T) {
	driverExt, err := InitWDAClient(nil)
	checkErr(t, err)

	pathSearch := "/Users/hero/Documents/temp/2020-05/opencv/flag7.png"

	// gwda.SetDebug(true)

	err = driverExt.TapWithNumber(pathSearch, 3)
	checkErr(t, err)

	err = driverExt.TapWithNumberOffset(pathSearch, 3, 0.5, 0.75)
	checkErr(t, err)
}

func TestDriverExt_TapXY(t *testing.T) {
	driverExt, err := InitWDAClient(nil)
	checkErr(t, err)

	err = driverExt.TapXY(0.4, 0.5, "")
	checkErr(t, err)
}

func TestDriverExt_TapWithOCR(t *testing.T) {
	driverExt, err := InitWDAClient(nil)
	checkErr(t, err)

	// 需要点击文字上方的图标
	err = driverExt.TapOffset("抖音", 0.5, -1, "", false)
	checkErr(t, err)
}
