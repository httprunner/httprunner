//go:build localtest

package uixt

import (
	"testing"
)

func TestSwipeUntil(t *testing.T) {
	driverExt, err := iosDevice.NewDriver(nil)
	checkErr(t, err)

	var point PointF
	findApp := func(d *DriverExt) error {
		var err error
		point, err = d.GetTextXY("抖音")
		return err
	}
	foundAppAction := func(d *DriverExt) error {
		// click app, launch douyin
		return d.TapAbsXY(point.X, point.Y)
	}

	driverExt.Driver.Homescreen()

	// swipe to first screen
	for i := 0; i < 5; i++ {
		driverExt.SwipeRight()
	}

	// swipe until app found
	err = driverExt.SwipeUntil("left", findApp, foundAppAction, WithDataMaxRetryTimes(10))
	checkErr(t, err)

	findLive := func(d *DriverExt) error {
		var err error
		point, err = d.GetTextXY("点击进入直播间")
		return err
	}
	foundLiveAction := func(d *DriverExt) error {
		// enter live room
		return d.TapAbsXY(point.X, point.Y)
	}

	// swipe until live room found
	err = driverExt.SwipeUntil("up", findLive, foundLiveAction, WithDataMaxRetryTimes(20))
	checkErr(t, err)
}
