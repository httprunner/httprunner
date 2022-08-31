package uixt

import (
	"testing"
)

func TestDriverExt_Swipe(t *testing.T) {
	driverExt, err := InitWDAClient()
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

func TestSwipeUntil(t *testing.T) {
	driverExt, err := InitWDAClient()
	checkErr(t, err)

	var x, y, width, height float64
	findApp := func(d *DriverExt) error {
		var err error
		x, y, width, height, err = d.FindTextByOCR("抖音")
		return err
	}
	foundAppAction := func(d *DriverExt) error {
		// click app, launch douyin
		return d.TapFloat(x+width*0.5, y+height*0.5-20)
	}

	driverExt.Homescreen()

	// swipe to first screen
	for i := 0; i < 5; i++ {
		driverExt.SwipeTo("right")
	}

	// swipe until app found
	err = driverExt.SwipeUntil("left", findApp, foundAppAction, 10)
	checkErr(t, err)

	findLive := func(d *DriverExt) error {
		var err error
		x, y, width, height, err = d.FindTextByOCR("点击进入直播间")
		return err
	}
	foundLiveAction := func(d *DriverExt) error {
		// enter live room
		return d.TapFloat(x+width*0.5, y+height*0.5)
	}

	// swipe until live room found
	err = driverExt.SwipeUntil("up", findLive, foundLiveAction, 20)
	checkErr(t, err)
}
