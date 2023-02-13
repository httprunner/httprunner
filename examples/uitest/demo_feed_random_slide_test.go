//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidDouyinFeedTest(t *testing.T) {
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
				Sleep(10),
			hrp.NewStep("处理青少年弹窗").
				Android().
				TapByOCR("我知道了", uixt.WithIgnoreNotFoundError(true)),
			hrp.NewStep("滑动 Feed 35 次，随机间隔 0-20s").
				Loop(35).
				Android().
				SwipeUp().
				SleepRandom(0, 20),
			hrp.NewStep("滑动 Feed 15 次，随机间隔 15-50s").
				Loop(15).
				Android().
				SwipeUp().
				SleepRandom(15, 50),
		},
	}

	if err := testCase.Dump2JSON("demo_feed_random_slide.json"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
