//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestIOSDouyinLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("通过 feed 卡片进入抖音直播间").
			WithVariables(map[string]interface{}{
				"app_name": "抖音",
			}).
			SetIOS(
				uixt.WithWDALogOn(true),
				uixt.WithWDAPort(8700),
				uixt.WithWDAMjpegPort(8800),
				uixt.WithIOSPerfOptions(
					uixt.WithIOSPerfSystemCPU(true),
					uixt.WithIOSPerfSystemMem(true),
				),
			),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				IOS().
				Home().
				AppTerminate("com.ss.iphone.ugc.Aweme"). // 关闭已运行的抖音
				SwipeToTapApp("$app_name", uixt.WithMaxRetryTimes(5), uixt.WithIdentifier("启动抖音")).Sleep(5).
				Validate().
				AssertOCRExists("推荐", "抖音启动失败，「推荐」不存在"),
			hrp.NewStep("处理青少年弹窗").
				IOS().
				TapByOCR("我知道了", uixt.WithIgnoreNotFoundError(true)),
			hrp.NewStep("向上滑动 2 次").
				IOS().
				SwipeUp(uixt.WithIdentifier("第一次上划")).Sleep(2).ScreenShot(). // 上划 1 次，等待 2s，截图保存
				SwipeUp(uixt.WithIdentifier("第二次上划")).Sleep(2).ScreenShot(), // 再上划 1 次，等待 2s，截图保存
			hrp.NewStep("在推荐页上划，直到出现「点击进入直播间」").
				IOS().
				SwipeToTapText("点击进入直播间", uixt.WithMaxRetryTimes(10), uixt.WithIdentifier("进入直播间")),
		},
	}

	if err := testCase.Dump2JSON("demo_douyin_live.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("demo_douyin_live.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
