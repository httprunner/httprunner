//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidDouYinLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("通过 feed 头像进入抖音直播间").
			SetAndroid(uixt.WithAdbLogOn(true), uixt.WithSerialNumber("2d06bf70")),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Android().
				Home().
				AppTerminate("com.ss.android.ugc.aweme"). // 关闭已运行的抖音，确保启动抖音后在「抖音」首页
				SwipeToTapApp("抖音", uixt.WithMaxRetryTimes(5)).
				Sleep(10),
			hrp.NewStep("处理青少年弹窗").
				Android().
				Tap("推荐").
				TapByOCR("我知道了", uixt.WithIgnoreNotFoundError(true)).
				Validate().
				AssertOCRExists("首页", "抖音启动失败，「首页」不存在"),
			hrp.NewStep("在推荐页上划，直到出现 feed 头像「直播」").
				Android().
				SwipeToTapText("直播", uixt.WithMaxRetryTimes(10), uixt.WithIdentifier("进入直播间")),
			hrp.NewStep("向上滑动，等待 10s").
				Android().
				SwipeUp(uixt.WithIdentifier("第一次上划")).Sleep(10).ScreenShot(). // 上划 1 次，等待 10s，截图保存
				SwipeUp(uixt.WithIdentifier("第二次上划")).Sleep(10).ScreenShot(), // 再上划 1 次，等待 10s，截图保存
		},
	}

	if err := testCase.Dump2JSON("demo_android_douyin_live.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("demo_android_douyin_live.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
