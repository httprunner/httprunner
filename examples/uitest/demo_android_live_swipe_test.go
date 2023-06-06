//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidLiveSwipeTest(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("点播_抖音_滑动场景_随机间隔_android").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
			}).
			SetAndroid(uixt.WithSerialNumber("$device")),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Android().
				AppTerminate("com.ss.android.ugc.aweme").
				AppLaunch("com.ss.android.ugc.aweme").
				Sleep(5).
				Validate().
				AssertAppInForeground("com.ss.android.ugc.aweme"),
			hrp.NewStep("处理青少年弹窗").
				Android().
				TapByOCR("我知道了", uixt.WithIgnoreNotFoundError(true)),
			hrp.NewStep("在推荐页上划，直到出现「点击进入直播间」").
				Android().
				SwipeToTapText("点击进入直播间", uixt.WithMaxRetryTimes(10), uixt.WithIdentifier("进入直播间")),
			hrp.NewStep("滑动 Feed 5 次，60% 随机间隔 0-5s，40% 随机间隔 5-10s").
				Loop(5).
				Android().
				SwipeUp().
				SleepRandom(0, 5, 0.6, 5, 10, 0.4),
			hrp.NewStep("向上滑动，等待 10s").
				Android().
				SwipeUp(uixt.WithIdentifier("第一次上划")).Sleep(5).ScreenShot(). // 上划 1 次，等待 5s，截图保存
				SwipeUp(uixt.WithIdentifier("第二次上划")).Sleep(5).ScreenShot(), // 再上划 1 次，等待 5s，截图保存
			hrp.NewStep("exit").
				Android().
				AppTerminate("com.ss.android.ugc.aweme").
				Validate().
				AssertAppNotInForeground("com.ss.android.ugc.aweme"),
		},
	}

	if err := testCase.Dump2JSON("demo_android_live_swipe.json"); err != nil {
		t.Fatal(err)
	}

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
