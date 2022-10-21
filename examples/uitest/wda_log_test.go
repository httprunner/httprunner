//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestWDALog(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("验证 WDA 打点数据准确性").
			WithVariables(map[string]interface{}{
				"app_name": "抖音",
			}).
			SetIOS(hrp.WithLogOn(true), hrp.WithWDAPort(8700), hrp.WithWDAMjpegPort(8800)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				IOS().
				Home().
				AppTerminate("com.ss.iphone.ugc.Aweme"). // 关闭已运行的抖音
				SwipeToTapApp("$app_name", hrp.WithMaxRetryTimes(5), hrp.WithIdentifier("启动抖音")).Sleep(5).
				Validate().
				AssertOCRExists("推荐", "抖音启动失败，「推荐」不存在"),
			hrp.NewStep("处理青少年弹窗").
				IOS().
				TapByOCR("我知道了", hrp.WithIgnoreNotFoundError(true)),
			hrp.NewStep("进入购物页").
				IOS().TapByOCR("购物", hrp.WithIdentifier("点击购物")).Sleep(5),
			hrp.NewStep("进入推荐页").
				IOS().TapByOCR("推荐", hrp.WithIdentifier("点击推荐")).Sleep(5),
			hrp.NewStep("向上滑动 2 次").
				IOS().
				SwipeUp(hrp.WithIdentifier("第 1 次上划")).Sleep(2).
				SwipeUp(hrp.WithIdentifier("第 2 次上划")).Sleep(2).
				SwipeUp(hrp.WithIdentifier("第 3 次上划")).Sleep(2).
				TapXY(0.9, 0.1, hrp.WithIdentifier("点击进入搜索框")).Sleep(2).
				Input("httprunner", hrp.WithIdentifier("输入搜索关键词")),
		},
	}

	if err := testCase.Dump2JSON("wda_log_data.json"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
