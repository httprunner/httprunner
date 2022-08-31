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
				TapByOCR("视频号"). // 通过 OCR 识别「视频号」
				Validate().
				AssertLabelExists("视频号"),
			hrp.NewStep("处理青少年弹窗").
				IOS().
				TapByOCR("我知道了", hrp.WithIgnoreNotFoundError(false)),
			hrp.NewStep("在推荐页上划，直到出现「轻触进入直播间」").
				IOS().
				SwipeToTapText("轻触进入直播间", hrp.WithMaxRetryTimes(10)),
			hrp.NewStep("向上滑动，等待 60s").
				IOS().
				SwipeUp().Sleep(60).ScreenShot(). // 上划 1 次，等待 60s，截图保存
				SwipeUp().Times(60).ScreenShot(), // 再上划 1 次，等待 60s，截图保存
		},
	}
	fmt.Println(testCase)

	err := hrp.NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
