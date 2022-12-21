//go:build localtest

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
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

func TestIOSDouyinWorldCupLive(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("直播_抖音_世界杯_ios").
			WithVariables(map[string]interface{}{
				"appBundleID": "com.ss.iphone.ugc.Aweme",
			}).
			SetIOS(
				uixt.WithUDID(uuid),
				uixt.WithWDALogOn(true),
				uixt.WithWDAPort(8700),
				uixt.WithWDAMjpegPort(8800),
				uixt.WithXCTest("com.gtf.wda.runner.xctrunner"),
				uixt.WithIOSPerfOptions(
					uixt.WithIOSPerfNetwork(true),
					// uixt.WithIOSPerfBundleID("com.ss.iphone.ugc.Aweme"),
				),
				uixt.WithIOSPcapOptions(
					// uixt.WithIOSPcapAll(true),
					// uixt.WithIOSPcapPID(1234),
					// uixt.WithIOSPcapProcName("Awe"),
					uixt.WithIOSPcapBundleID("com.ss.iphone.ugc.Aweme"),
				),
			),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音").
				IOS().
				Home().
				AppTerminate("$appBundleID"). // 关闭已运行的抖音
				AppLaunch("$appBundleID").
				TapByOCR("我知道了", uixt.WithIgnoreNotFoundError(true)). // 处理青少年弹窗
				Validate().
				AssertOCRExists("首页", "抖音启动失败，「首页」不存在"),
			hrp.NewStep("点击首页").
				IOS().
				TapByOCR("首页", uixt.WithIndex(-1)).Sleep(2),
			hrp.NewStep("点击世界杯页").
				IOS().
				SwipeToTapText("世界杯",
					uixt.WithMaxRetryTimes(5),
					uixt.WithCustomDirection(0.4, 0.07, 0.6, 0.07), // 滑动 tab，从左到右，解决「世界杯」被遮挡的问题
					uixt.WithScope(0, 0, 1, 0.15),                  // 限定 tab 区域
					uixt.WithWaitTime(1),
				),
			hrp.NewStep("点击进入赛程晋级").
				Loop(5). // 重复执行 5 次
				IOS().
				TapByOCR("赛程晋级", uixt.WithIdentifier("click_live"), uixt.WithIndex(-1)).
				Sleep(3).Back().Sleep(3),
			hrp.NewStep("关闭抖音").
				IOS().
				AppTerminate("$appBundleID"),
		},
	}

	runner := hrp.NewRunner(t).SetSaveTests(true)
	err := runner.Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
