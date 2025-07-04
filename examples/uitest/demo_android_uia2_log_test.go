//go:build localtest

package uitest

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestUIA2Log(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("验证 UIA2 打点数据准确性").
			WithVariables(map[string]interface{}{
				"app_name": "抖音",
			}).
			SetAndroid(
				option.WithAdbLogOn(true),
			),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Android().
				Home().
				AppTerminate("com.ss.android.ugc.aweme"). // 关闭已运行的抖音
				SwipeToTapApp("$app_name",
					option.WithMaxRetryTimes(5),
					option.WithIdentifier("启动抖音"),
					option.WithPreMarkOperation(true)).Sleep(5).
				Validate().
				AssertOCRExists("推荐", "抖音启动失败，「推荐」不存在"),
			hrp.NewStep("处理青少年弹窗").
				Android().
				TapByOCR("我知道了",
					option.WithIgnoreNotFoundError(true)),
			hrp.NewStep("进入推荐页").
				Android().TapByOCR("推荐",
				option.WithIdentifier("点击推荐"),
				option.WithPreMarkOperation(true),
				option.WithTapOffset(0, -1)).Sleep(5),
			hrp.NewStep("向上滑动 2 次").
				Android().
				SwipeUp(option.WithIdentifier("第 1 次上划"), option.WithPreMarkOperation(true)).Sleep(2).
				SwipeUp(option.WithIdentifier("第 2 次上划"), option.WithPreMarkOperation(true)).Sleep(2).
				SwipeUp(option.WithIdentifier("第 3 次上划"), option.WithPreMarkOperation(true)).Sleep(2).
				TapXY(0.9, 0.1, option.WithIdentifier("点击进入搜索框"), option.WithPreMarkOperation(true)).Sleep(2).
				Input("httprunner 发版记录", option.WithIdentifier("输入搜索关键词"), option.WithPreMarkOperation(true)).
				TapByOCR("搜索", option.WithIdentifier("点击搜索")),
		},
	}

	if err := testCase.Dump2JSON("demo_android_uia2_log.json"); err != nil {
		t.Fatal(err)
	}

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
