//go:build localtest

package uitest

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestAndroidDouyinFeedTest(t *testing.T) {
	t.Skip("Skip Android UI test - requires physical Android device with ADB")
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("点播_抖音_滑动场景_随机间隔_android").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
			}).
			SetAndroid(option.WithSerialNumber("$device"),
				option.WithAdbLogOn(true)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Android().
				AppTerminate("com.ss.android.ugc.aweme").
				AppLaunch("com.ss.android.ugc.aweme").
				Sleep(10).
				Validate().
				AssertAppInForeground("com.ss.android.ugc.aweme"),
			hrp.NewStep("处理青少年弹窗").
				Android().
				TapByOCR("我知道了", option.WithIgnoreNotFoundError(true)),
			hrp.NewStep("滑动 Feed 3 次，随机间隔 0-5s").
				Loop(3).
				Android().
				SwipeUp().
				SleepRandom(0, 5),
			hrp.NewStep("滑动 Feed 1 次，随机间隔 5-10s").
				Loop(1).
				Android().
				SwipeUp().
				SleepRandom(5, 10),
			hrp.NewStep("滑动 Feed 10 次，70% 随机间隔 0-5s，30% 随机间隔 5-10s").
				Loop(10).
				Android().
				SwipeUp().
				SleepRandom(0, 5, 0.7, 5, 10, 0.3),
			hrp.NewStep("exit").
				Android().
				AppTerminate("com.ss.android.ugc.aweme").
				Validate().
				AssertAppNotInForeground("com.ss.android.ugc.aweme"),
		},
	}

	if err := testCase.Dump2JSON("demo_android_feed_swipe.json"); err != nil {
		t.Fatal(err)
	}

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
