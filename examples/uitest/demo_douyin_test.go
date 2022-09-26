package uitest

import (
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestIOSDouyinLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("通过 feed 卡片进入微信直播间").
			SetIOS(hrp.WithLogOn(true), hrp.WithPort(8700), hrp.WithMjpegPort(8800)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				IOS().
				Home().
				AppTerminate("com.ss.iphone.ugc.Aweme"). // 关闭已运行的抖音
				SwipeToTapApp("抖音", hrp.WithMaxRetryTimes(5)).Sleep(5).
				Validate().
				AssertOCRExists("推荐", "抖音启动失败，「推荐」不存在"),
			// hrp.NewStep("处理青少年弹窗").
			// 	IOS().
			// 	TapByOCR("我知道了", hrp.WithIgnoreNotFoundError(true)),
			hrp.NewStep("在推荐页上划，直到出现「点击进入直播间」").
				IOS().
				SwipeToTapText("点击进入直播间", hrp.WithMaxRetryTimes(100), hrp.WithIdentifier("进入直播间")),
			hrp.NewStep("向上滑动，等待 10s").
				IOS().
				SwipeUp(hrp.WithIdentifier("第一次上划")).Sleep(2).ScreenShot(). // 上划 1 次，等待 2s，截图保存
				SwipeUp(hrp.WithIdentifier("第二次上划")).Sleep(2).ScreenShot(), // 再上划 1 次，等待 2s，截图保存
		},
	}

	if err := testCase.Dump2JSON("demo_douyin_live.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("demo_douyin_live.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t)
	sessionRunner, err := runner.NewSessionRunner(testCase)
	if err != nil {
		t.Fatal(err)
	}
	if err := sessionRunner.Start(nil); err != nil {
		t.Fatal(err)
	}
	summary := sessionRunner.GetSummary()
	fmt.Println(summary)
}
