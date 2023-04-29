//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidVideoCrawlerTest(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("抓取抖音视频信息").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
			}).
			SetAndroid(uixt.WithSerialNumber("$device")),
		TestSteps: []hrp.IStep{
			hrp.NewStep("滑动消费 feed 至少 10 个，live 至少 3 个；滑动过程中，70% 随机间隔 0-5s，30% 随机间隔 5-10s").
				Android().
				VideoCrawler(map[string]interface{}{
					"app_package_name": "com.ss.android.ugc.aweme",
					"target_count": map[string]interface{}{
						"feed_count": 5,
						"live_count": 3,
					},
					"sleep_random": []float64{0, 5, 0.7, 5, 10, 0.3},
				}),
			hrp.NewStep("exit").
				Android().
				AppTerminate("com.ss.android.ugc.aweme").
				Validate().
				AssertAppNotInForeground("com.ss.android.ugc.aweme"),
		},
	}

	if err := testCase.Dump2JSON("demo_android_video_crawler.json"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
