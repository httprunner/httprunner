//go:build localtest

package uixt

import (
	"testing"
)

func TestAndroidSwipeAction(t *testing.T) {
	setupAndroid(t)

	swipeAction := driverExt.prepareSwipeAction(WithDirection("up"))
	err := swipeAction(driverExt)
	checkErr(t, err)

	swipeAction = driverExt.prepareSwipeAction(WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(driverExt)
	checkErr(t, err)
}

func TestAndroidSwipeToTapApp(t *testing.T) {
	setupAndroid(t)

	err := driverExt.swipeToTapApp("抖音")
	checkErr(t, err)
}

func TestAndroidSwipeToTapTexts(t *testing.T) {
	setupAndroid(t)

	err := driverExt.Driver.AppLaunch("com.ss.android.ugc.aweme")
	checkErr(t, err)

	err = driverExt.swipeToTapTexts([]string{"点击进入直播间", "直播中"}, WithDirection("up"))
	checkErr(t, err)
}
