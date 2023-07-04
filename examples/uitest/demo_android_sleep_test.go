package uitest

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestAndroidNaiveSleep(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("浏览器时间戳截图").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
			}).
			SetAndroid(uixt.WithSerialNumber("$device")),
		TestSteps: []hrp.IStep{
			hrp.NewStep("打开浏览器local-time时间戳界面").
				Android().
				Home().
				SwipeToTapApp("local", uixt.WithMaxRetryTimes(5)).
				Sleep(7),
			hrp.NewStep("循环2次，每次截图四张，截图间隔为1s/2s/3s").
				Loop(2).
				Android().
				ScreenShot().Sleep(1).
				ScreenShot().Sleep(2).
				ScreenShot().Sleep(3).
				ScreenShot(),
		},
	}

	if err := testCase.Dump2JSON("demo_android_naive_sleep.json"); err != nil {
		t.Fatal(err)
	}

	// err := hrp.Run(t, testCase)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func TestAndroidStrictSleep(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("浏览器时间戳截图").
			WithVariables(map[string]interface{}{
				"device": "${ENV(SerialNumber)}",
			}).
			SetAndroid(uixt.WithSerialNumber("$device")),
		TestSteps: []hrp.IStep{
			hrp.NewStep("打开浏览器local-time时间戳界面").
				Android().
				Home().
				SwipeToTapApp("local", uixt.WithMaxRetryTimes(5)).
				Sleep(7),
			hrp.NewStep("循环2次，每次截图四张，截图间隔为1s/2s/3s").
				Loop(2).
				Android().
				ScreenShot().SleepStrict(1).
				ScreenShot().SleepStrict(2).
				ScreenShot().SleepStrict(3).
				ScreenShot(),
		},
	}

	if err := testCase.Dump2JSON("demo_android_strict_sleep.json"); err != nil {
		t.Fatal(err)
	}

	// err := hrp.Run(t, testCase)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}
