//go:build localtest

package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestHamonyDouyinFeedTest(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("点播_抖音_滑动场景_随机间隔_android").
			WithVariables(map[string]interface{}{
				"device": "a38c2c5c",
				"query":  "${ENV(query)}",
			}).
			SetAndroid(uixt.WithSerialNumber("$device")),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				Android().
				AppTerminate("com.ss.hm.ugc.aweme"),
		},
	}

	if err := testCase.Dump2JSON("demo_android_swipe.json"); err != nil {
		t.Fatal(err)
	}

	err := hrp.Run(t, testCase)
	if err != nil {
		t.Fatal(err)
	}
}
