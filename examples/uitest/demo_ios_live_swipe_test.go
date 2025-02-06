//go:build localtest

package uitest

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func TestIOSDouyinLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("通过 feed 卡片进入抖音直播间").
			WithVariables(map[string]interface{}{
				"app_name": "抖音",
			}).
			SetIOS(
				option.WithWDALogOn(true),
				option.WithWDAPort(8700),
				option.WithWDAMjpegPort(8800),
			),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				IOS().
				Home().
				AppTerminate("com.ss.iphone.ugc.Aweme"). // 关闭已运行的抖音
				SwipeToTapApp("$app_name",
					option.WithMaxRetryTimes(5),
					option.WithIdentifier("启动抖音")).Sleep(5).
				Validate().
				AssertOCRExists("推荐", "抖音启动失败，「推荐」不存在"),
			hrp.NewStep("处理青少年弹窗").
				IOS().
				TapByOCR("我知道了", option.WithIgnoreNotFoundError(true)),
			hrp.NewStep("向上滑动 2 次").
				IOS().
				SwipeUp(option.WithIdentifier("第一次上划")).Sleep(2).ScreenShot(). // 上划 1 次，等待 2s，截图保存
				SwipeUp(option.WithIdentifier("第二次上划")).Sleep(2).ScreenShot(), // 再上划 1 次，等待 2s，截图保存
			hrp.NewStep("在推荐页上划，直到出现「点击进入直播间」").
				IOS().
				SwipeToTapText("点击进入直播间",
					option.WithMaxRetryTimes(10),
					option.WithIdentifier("进入直播间")),
		},
	}

	if err := testCase.Dump2JSON("demo_ios_live_swipe.json"); err != nil {
		t.Fatal(err)
	}

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
