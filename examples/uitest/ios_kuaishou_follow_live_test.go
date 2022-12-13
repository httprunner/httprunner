//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestIOSKuaiShouLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_快手_关注天窗_ios").
			WithVariables(map[string]interface{}{
				"device": "${ENV(UDID)}",
				"ups":    "${ENV(LIVEUPLIST)}",
			}).
			SetIOS(
				uixt.WithUDID("$device"),
				uixt.WithWDALogOn(true),
				uixt.WithWDAPort(8100),
				uixt.WithWDAMjpegPort(9100)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动快手").
				IOS().
				AppTerminate("com.jiangjia.gif").
				AppLaunch("com.jiangjia.gif").
				Home().
				SwipeToTapApp("快手", uixt.WithMaxRetryTimes(5)).Sleep(10).
				Validate().
				AssertOCRExists("精选", "进入快手失败"),
			hrp.NewStep("点击首页").
				IOS().
				TapByOCR("首页", uixt.WithIndex(-1)).Sleep(10),
			hrp.NewStep("点击发现页").
				IOS().
				TapByOCR("发现", uixt.WithIndex(1)).Sleep(10),
			hrp.NewStep("点击关注页").
				IOS().
				TapByOCR("关注", uixt.WithIndex(1)).Sleep(10),
			hrp.NewStep("点击直播标签,进入直播间").
				IOS().
				SwipeToTapTexts("${split_by_comma($ups)}", uixt.WithCustomDirection(0.6, 0.2, 0.2, 0.2), uixt.WithIdentifier("click_live")).Sleep(60).
				Validate().
				AssertOCRExists("说点什么", "进入直播间失败"),
			hrp.NewStep("下滑进入下一个直播间").
				IOS().
				Swipe(0.9, 0.7, 0.9, 0.3, uixt.WithIdentifier("slide_in_live")).Sleep(60),
		},
	}

	if err := testCase.Dump2JSON("ios_kuaishou_follow_live_test.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("ios_kuaishou_follow_live_test.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
