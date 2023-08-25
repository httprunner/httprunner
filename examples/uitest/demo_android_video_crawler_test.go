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
			hrp.NewStep("启动 app").
				Android().
				ScreenShot(uixt.WithScreenShotOCR(true), uixt.WithScreenShotUpload(true)).
				AppLaunch("com.ss.android.ugc.aweme").
				Sleep(5).
				Validate().
				AssertAppInForeground("com.ss.android.ugc.aweme"),
			hrp.NewStep("滑动消费 feed 至少 10 个，live 至少 3 个；滑动过程中，70% 随机间隔 0-5s，30% 随机间隔 5-10s").
				Android().
				VideoCrawler(map[string]interface{}{
					"timeout": 600,
					"feed": map[string]interface{}{
						"target_count": 5,
						"target_labels": []map[string]interface{}{
							{"text": "^广告$", "scope": []float64{0, 0.5, 1, 1}, "regex": true, "target": 1},
							{"text": "^图文$", "scope": []float64{0, 0.5, 1, 1}, "regex": true, "target": 1},
							{"text": `^特效\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^模板\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
							{"text": `^购物\|`, "scope": []float64{0, 0.5, 1, 1}, "regex": true},
						},
						"sleep_random": []float64{0, 5, 0.7, 5, 10, 0.3},
					},
					"live": map[string]interface{}{
						"target_count": 3,
						"sleep_random": []float64{15, 20},
					},
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

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
