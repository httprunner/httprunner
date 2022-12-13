//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestIOSDouyinFollowLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("通过 关注天窗 进入指定主播抖音直播间").
			WithVariables(map[string]interface{}{
				"app_name": "抖音",
			}).
			SetIOS(
				uixt.WithWDALogOn(true),
				uixt.WithWDAPort(8700),
				uixt.WithWDAMjpegPort(8800),
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
			hrp.NewStep("点击首页").
				IOS().
				TapByOCR("首页", uixt.WithIndex(-1)).Sleep(10),
			hrp.NewStep("点击关注页").
				IOS().
				TapByOCR("关注", uixt.WithIndex(1)).Sleep(10),
			hrp.NewStep("向上滑动 2 次").
				IOS().SwipeToTapTexts([]string{"理肤泉", "婉宝"}, uixt.WithCustomDirection(0.6, 0.2, 0.2, 0.2), uixt.WithIdentifier("click_live")).Sleep(10).
				Swipe(0.9, 0.7, 0.9, 0.3, uixt.WithIdentifier("slide_in_live")).Sleep(10).ScreenShot(). // 上划 1 次，等待 10s，截图保存
				Swipe(0.9, 0.7, 0.9, 0.3, uixt.WithIdentifier("slide_in_live")).Sleep(10).ScreenShot(), // 再上划 1 次，等待 10s，截图保存
		},
	}

	if err := testCase.Dump2JSON("demo_douyin_follow_live.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("demo_douyin_follow_live.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
