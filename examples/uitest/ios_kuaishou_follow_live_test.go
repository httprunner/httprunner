//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestIOSKuaiShouLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_快手_关注天窗_ios").
			WithVariables(map[string]interface{}{
				"device": "${ENV(UDID)}",
				"ups":    "${ENV(LIVEUPLIST)}",
			}).
			SetIOS(hrp.WithUDID("$device"), hrp.WithLogOn(true), hrp.WithWDAPort(8100), hrp.WithWDAMjpegPort(9100)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动快手").
				IOS().
				AppTerminate("com.jiangjia.gif").
				AppLaunch("com.jiangjia.gif").
				Home().
				SwipeToTapApp("快手", hrp.WithMaxRetryTimes(5)).Sleep(10).
				Validate().
				AssertOCRExists("精选", "进入快手失败"),
			hrp.NewStep("点击首页").
				IOS().
				TapByOCR("首页", hrp.WithIndex(-1)).Sleep(10),
			hrp.NewStep("点击发现页").
				IOS().
				TapByOCR("发现", hrp.WithIndex(1)).Sleep(10),
			hrp.NewStep("点击关注页").
				IOS().
				TapByOCR("关注", hrp.WithIndex(1)).Sleep(10),
			hrp.NewStep("点击直播标签,进入直播间").
				IOS().
				SwipeToTapTexts("${split_by_comma($ups)}", hrp.WithCustomDirection(0.6, 0.2, 0.2, 0.2), hrp.WithIdentifier("click_live")).Sleep(60).
				Validate().
				AssertOCRExists("说点什么", "进入直播间失败"),
			hrp.NewStep("下滑进入下一个直播间").
				IOS().
				Swipe(0.9, 0.7, 0.9, 0.3, hrp.WithIdentifier("slide_in_live")).Sleep(60),
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
