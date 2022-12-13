package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidKuaiShouFeedCardLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_快手_Feed卡片_android").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
			}).
			SetAndroid(
				uixt.WithSerialNumber("$device"),
				uixt.WithAdbLogOn(true)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动快手").
				Android().
				AppTerminate("com.smile.gifmaker").
				AppLaunch("com.smile.gifmaker").
				Home().
				SwipeToTapApp("快手", uixt.WithMaxRetryTimes(5)).Sleep(10),
			hrp.NewStep("处理青少年弹窗").
				Android().
				TapByOCR("我知道了", uixt.WithIgnoreNotFoundError(true)).
				Validate().
				AssertOCRExists("精选", "进入快手失败"),
			hrp.NewStep("点击精选").
				Android().
				TapByOCR("精选", uixt.WithIndex(-1), uixt.WithOffset(0, -50)).Sleep(10),
			hrp.NewStep("点击直播标签,进入直播间").
				Android().
				SwipeToTapText("点击进入直播间",
					uixt.WithCustomDirection(0.9, 0.7, 0.9, 0.3),
					uixt.WithScope(0.2, 0.5, 0.8, 0.8),
					uixt.WithMaxRetryTimes(20),
					uixt.WithWaitTime(60),
					uixt.WithIdentifier("click_live"),
				),
			hrp.NewStep("等待1分钟").
				Android().
				Sleep(60),
			hrp.NewStep("上滑进入下一个直播间").
				Android().
				Swipe(0.9, 0.7, 0.9, 0.3, uixt.WithIdentifier("slide_in_live")).Sleep(60),
			hrp.NewStep("返回主界面，并打开本地时间戳").
				Android().
				Home().SwipeToTapApp("local", uixt.WithMaxRetryTimes(5)).Sleep(10).
				Validate().
				AssertOCRExists("16", "打开本地时间戳失败"),
		},
	}

	if err := testCase.Dump2JSON("android_feed_card_live_test.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("android_feed_card_live_test.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
