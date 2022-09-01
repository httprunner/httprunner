package uitest

import (
	"fmt"
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestIOSWeixinLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("通过 feed 卡片进入微信直播间"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动微信").
				IOS().
				Home().
				AppTerminate("com.tencent.xin"). // 关闭已运行的微信，确保启动微信后在「微信」首页
				SwipeToTapApp("微信", hrp.WithMaxRetryTimes(5)).
				Validate().
				AssertLabelExists("通讯录", "微信启动失败，「通讯录」不存在"),
			hrp.NewStep("进入直播页").
				IOS().
				Tap("发现").       // 进入「发现页」
				TapByOCR("视频号"), // 通过 OCR 识别「视频号」
			hrp.NewStep("处理青少年弹窗").
				IOS().
				TapByOCR("我知道了", hrp.WithIgnoreNotFoundError(true)),
			hrp.NewStep("在推荐页上划，直到出现「轻触进入直播间」").
				IOS().
				SwipeToTapText("轻触进入直播间", hrp.WithMaxRetryTimes(10)),
			hrp.NewStep("向上滑动，等待 10s").
				IOS().
				SwipeUp().Sleep(10).ScreenShot(). // 上划 1 次，等待 10s，截图保存
				SwipeUp().Sleep(10).ScreenShot(), // 再上划 1 次，等待 10s，截图保存
		},
	}
	fmt.Println(testCase)
	if err := testCase.Dump2JSON("demo_weixin_live.json"); err != nil {
		t.Fatal(err)
	}
	if err := testCase.Dump2YAML("demo_weixin_live.yaml"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t)
	sessionRunner, _ := runner.NewSessionRunner(testCase)
	if err := sessionRunner.Start(nil); err != nil {
		t.Fatal(err)
	}
	summary := sessionRunner.GetSummary()
	fmt.Println(summary)
}
