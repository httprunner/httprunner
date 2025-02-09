//go:build localtest

package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func TestAndroidSwipeAction(t *testing.T) {
	setupAndroidAdbDriver(t)

	dExt := driverExt.(*DriverExt)
	swipeAction := prepareSwipeAction(dExt, "up", option.WithDirection("down"))
	err := swipeAction(dExt)
	checkErr(t, err)

	swipeAction = prepareSwipeAction(dExt, "up", option.WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(dExt)
	checkErr(t, err)
}

func TestAndroidSwipeToTapApp(t *testing.T) {
	setupAndroidAdbDriver(t)

	err := driverExt.SwipeToTapApp("抖音")
	checkErr(t, err)
}

func TestAndroidSwipeToTapTexts(t *testing.T) {
	setupAndroidAdbDriver(t)

	err := driverExt.GetDriver().AppLaunch("com.ss.android.ugc.aweme")
	checkErr(t, err)

	dExt := driverExt.(*DriverExt)
	err = dExt.swipeToTapTexts([]string{"点击进入直播间", "直播中"}, option.WithDirection("up"))
	checkErr(t, err)
}
