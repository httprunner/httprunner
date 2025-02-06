package uitest

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

func TestHarmonyDouyinE2E(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_抖音_端到端时延_harmony").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
				"ups":    "${ENV(LIVEUPLIST)}",
			}).
			SetHarmony(
				options.WithConnectKey("$device"),
				options.WithLogOn(true)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Harmony().
				AppTerminate("com.ss.hm.ugc.aweme").
				SwipeToTapApp("com.ss.hm.ugc.aweme").
				Home().
				SwipeToTapApp(
					"抖音",
					uixt.WithMaxRetryTimes(5),
					uixt.WithTapOffset(0, -50),
				).
				Sleep(20).
				Validate().
				AssertOCRExists("推荐", "进入抖音失败"),
			hrp.NewStep("点击放大镜").
				Harmony().
				TapXY(0.9, 0.08).
				Sleep(5),
			hrp.NewStep("输入账号名称").
				Harmony().
				Input("$ups").
				Sleep(5),
			hrp.NewStep("点击搜索").
				Harmony().
				TapByOCR("搜索").
				Sleep(5),
			hrp.NewStep("端到端采集").Loop(5).
				Harmony().
				TapByOCR(
					"直播中",
					uixt.WithIgnoreNotFoundError(true),
					uixt.WithIndex(-1),
				).
				EndToEndDelay(uixt.WithInterval(5), uixt.WithTimeout(120)).
				TapByUITypes(uixt.WithScreenShotUITypes("close")),
		},
	}

	if err := testCase.Dump2JSON("harmony_e2e_delay_test.json"); err != nil {
		t.Fatal(err)
	}

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
