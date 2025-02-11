//go:build localtest

package uixt

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func TestAndroidSwipeAction(t *testing.T) {
	setupAndroidAdbDriver(t)

	swipeAction := prepareSwipeAction(driverExt, "up", option.WithDirection("down"))
	err := swipeAction(driverExt)
	checkErr(t, err)

	swipeAction = prepareSwipeAction(driverExt, "up", option.WithCustomDirection(0.5, 0.5, 0.5, 0.9))
	err = swipeAction(driverExt)
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

	err = driverExt.swipeToTapTexts([]string{"点击进入直播间", "直播中"}, option.WithDirection("up"))
	checkErr(t, err)
}
