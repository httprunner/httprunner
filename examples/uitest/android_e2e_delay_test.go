package uitest

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
)

func TestAndroidDouyinE2E(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_抖音_端到端时延_android").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
				"ups":    "${ENV(LIVEUPLIST)}",
			}).
			SetAndroid(
				options.WithSerialNumber("$device"),
				options.WithAdbLogOn(true)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Android().
				AppTerminate("com.ss.android.ugc.aweme").
				AppLaunch("com.ss.android.ugc.aweme").
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
				Android().
				TapXY(0.9, 0.08).
				Sleep(5),
			hrp.NewStep("输入账号名称").
				Android().
				Input("$ups").
				Sleep(5),
			hrp.NewStep("点击搜索").
				Android().
				TapByOCR("搜索").
				Sleep(5),
			hrp.NewStep("端到端采集").Loop(5).
				Android().
				TapByOCR(
					"直播中",
					uixt.WithIgnoreNotFoundError(true),
					uixt.WithIndex(-1),
				).
				EndToEndDelay(uixt.WithInterval(5), uixt.WithTimeout(120)).
				TapByUITypes(uixt.WithScreenShotUITypes("close")),
		},
	}

	if err := testCase.Dump2JSON("android_e2e_delay_test.json"); err != nil {
		t.Fatal(err)
	}
}
