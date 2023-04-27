//go:build localtest

package uixt

import (
	"testing"
)

func TestAndroidSwipeAction(t *testing.T) {
	setupAndroid(t)

	action := MobileAction{
		Method: ACTION_Swipe,
		Params: "up",
	}
	swipeAction := driverExt.prepareSwipeAction(action)

	err := swipeAction(driverExt)
	checkErr(t, err)

	action = MobileAction{
		Method: ACTION_Swipe,
		Params: []float64{0.5, 0.5, 0.5, 0.9},
	}
	swipeAction = driverExt.prepareSwipeAction(action)

	err = swipeAction(driverExt)
	checkErr(t, err)
}

func TestAndroidSwipeToTapApp(t *testing.T) {
	setupAndroid(t)

	err := driverExt.swipeToTapApp("抖音", MobileAction{})
	checkErr(t, err)
}

func TestAndroidSwipeToTapTexts(t *testing.T) {
	setupAndroid(t)

	err := driverExt.Driver.AppLaunch("com.ss.android.ugc.aweme")
	checkErr(t, err)

	action := MobileAction{
		Params: "up",
	}
	err = driverExt.swipeToTapTexts([]string{"点击进入直播间", "直播中"}, action)
	checkErr(t, err)
}
