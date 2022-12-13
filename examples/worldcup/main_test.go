//go:build localtest

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestConvertTimeToSeconds(t *testing.T) {
	testData := []struct {
		timeStr string
		seconds int
	}{
		{"00:00", 0},
		{"00:01", 1},
		{"01:00", 60},
		{"01:01", 61},
		{"00:01:02", 62},
		{"01:02:03", 3723},
	}

	for _, td := range testData {
		seconds, err := convertTimeToSeconds(td.timeStr)
		assert.Nil(t, err)
		assert.Equal(t, td.seconds, seconds)
	}
}

func TestMainIOS(t *testing.T) {
	uuid := "00008030-00194DA421C1802E"
	device := initIOSDevice(uuid)
	bundleID := "com.ss.iphone.ugc.Aweme"
	wc := NewWorldCupLive(device, "", bundleID, 30, 10)
	wc.EnterLive(bundleID)
	wc.Start()
}

func TestMainAndroid(t *testing.T) {
	device := initAndroidDevice(uuid)
	bundleID := "com.ss.android.ugc.aweme"
	wc := NewWorldCupLive(device, "", bundleID, 30, 10)
	wc.EnterLive(bundleID)
	wc.Start()
}

func init() {
	os.Setenv("UDID", "00008030-00194DA421C1802E")
}

func TestIOSDouyinWorldCupLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_抖音_世界杯_ios").
			WithVariables(map[string]interface{}{
				"device": "${ENV(UDID)}",
			}).
			SetIOS(
				hrp.WithUDID("$device"),
				hrp.WithLogOn(true),
				hrp.WithWDAPort(8700),
				hrp.WithWDAMjpegPort(8800),
				hrp.WithXCTest("com.gtf.wda.runner.xctrunner"),
			),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				IOS().
				Home().
				AppTerminate("com.ss.iphone.ugc.Aweme"). // 关闭已运行的抖音
				AppLaunch("com.ss.iphone.ugc.Aweme").
				Validate().
				AssertOCRExists("首页", "抖音启动失败，「首页」不存在"),
			hrp.NewStep("处理青少年弹窗").
				IOS().
				TapByOCR("我知道了", hrp.WithIgnoreNotFoundError(true)),
			hrp.NewStep("点击首页").
				IOS().
				TapByOCR("首页", hrp.WithIndex(-1)).Sleep(5),
			hrp.NewStep("点击世界杯页").
				IOS().
				SwipeToTapText("世界杯",
					hrp.WithMaxRetryTimes(5),
					hrp.WithCustomDirection(0.4, 0.07, 0.6, 0.07), // 滑动 tab，从左到右，解决「世界杯」被遮挡的问题
					hrp.WithScope(0, 0, 1, 0.15),                  // 限定 tab 区域
					hrp.WithWaitTime(1),
				).
				Swipe(0.5, 0.3, 0.5, 0.2), // 少量上划，解决「直播中」未展示的问题
			hrp.NewStep("点击进入直播间").
				IOS().
				LoopTimes(30). // 重复执行 30 次
				TapByOCR("直播中", hrp.WithIdentifier("click_live"), hrp.WithIndex(-1)).
				Sleep(30).Back().Sleep(30),
			hrp.NewStep("关闭抖音").
				IOS().
				AppTerminate("com.ss.iphone.ugc.Aweme"),
			hrp.NewStep("返回主界面，并打开本地时间戳").
				IOS().
				Home().SwipeToTapApp("local", hrp.WithMaxRetryTimes(5)).Sleep(10).
				Validate().
				AssertOCRExists("16", "打开本地时间戳失败"),
		},
	}

	if err := testCase.Dump2JSON("ios_worldcup_live_douyin_test.json"); err != nil {
		t.Fatal(err)
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
