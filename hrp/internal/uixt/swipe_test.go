package uixt

import (
	"testing"
)

func TestSwipeUntil(t *testing.T) {
	driverExt, err := InitWDAClient(nil)
	checkErr(t, err)

	var x, y, width, height float64
	findApp := func(d *DriverExt) error {
		var err error
		x, y, width, height, err = d.FindTextByOCR("抖音")
		return err
	}
	foundAppAction := func(d *DriverExt) error {
		// click app, launch douyin
		return d.Driver.TapFloat(x+width*0.5, y+height*0.5-20)
	}

	driverExt.Driver.Homescreen()

	// swipe to first screen
	for i := 0; i < 5; i++ {
		driverExt.SwipeRight()
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
		return d.Driver.TapFloat(x+width*0.5, y+height*0.5)
	}

	// swipe until live room found
	err = driverExt.SwipeUntil("up", findLive, foundLiveAction, 20)
	checkErr(t, err)
}
