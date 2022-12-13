//go:build localtest

package hrp

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func TestIOSSettingsAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on Settings").
			SetIOS(uixt.WithWDAPort(8700), uixt.WithWDAMjpegPort(8800)),
		TestSteps: []IStep{
			NewStep("launch Settings").
				IOS().Home().Tap("设置").
				Validate().
				AssertNameExists("飞行模式").
				AssertLabelExists("蓝牙").
				AssertOCRExists("个人热点"),
			NewStep("swipe up and down").
				IOS().SwipeUp().SwipeUp().SwipeDown(),
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSSearchApp(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on Search App 资源库"),
		TestSteps: []IStep{
			NewStep("进入 App 资源库 搜索框").
				IOS().Home().SwipeLeft().SwipeLeft().Tap("dewey-search-field").
				Validate().
				AssertLabelExists("取消"),
			NewStep("搜索抖音").
				IOS().Input("抖音\n"),
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSAppLaunch(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("启动 & 关闭 App").
			SetIOS(uixt.WithWDAPort(8700), uixt.WithWDAMjpegPort(8800)),
		TestSteps: []IStep{
			NewStep("终止今日头条").
				IOS().AppTerminate("com.ss.iphone.article.News"),
			NewStep("启动今日头条").
				IOS().AppLaunch("com.ss.iphone.article.News"),
			NewStep("终止今日头条").
				IOS().AppTerminate("com.ss.iphone.article.News"),
			NewStep("启动今日头条").
				IOS().AppLaunchUnattached("com.ss.iphone.article.News"),
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSWeixinLive(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on 微信直播").
			SetIOS(uixt.WithWDALogOn(true), uixt.WithWDAPort(8100), uixt.WithWDAMjpegPort(9100)),
		TestSteps: []IStep{
			NewStep("启动微信").
				IOS().
				Home().
				AppTerminate("com.tencent.xin"). // 关闭已运行的微信，确保启动微信后在「微信」首页
				Tap("微信").
				Validate().
				AssertLabelExists("通讯录", "微信启动失败，「通讯录」不存在"),
			NewStep("进入直播页").
				IOS().
				Tap("发现").Sleep(5). // 进入「发现页」；等待 5 秒确保加载完成
				TapByOCR("直播").     // 通过 OCR 识别「直播」
				Validate().
				AssertLabelExists("直播"),
			NewStep("向上滑动 3 次，截图保存").
				IOS().
				Loop(3).                          // 整体循环 3 次
				SwipeUp().SwipeUp().ScreenShot(), // 上划 2 次，截图保存
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSCameraPhotoCapture(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios camera photo capture"),
		TestSteps: []IStep{
			NewStep("launch camera").
				IOS().Home().
				StopCamera().
				StartCamera().
				Validate().
				AssertLabelExists("PhotoCapture", "拍照按钮不存在"),
			NewStep("start recording").
				IOS().Tap("PhotoCapture"),
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSCameraVideoCapture(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios camera video capture"),
		TestSteps: []IStep{
			NewStep("launch camera").
				IOS().Home().
				StopCamera().
				StartCamera().
				Validate().
				AssertLabelExists("PhotoCapture", "录像按钮不存在"),
			NewStep("switch to video capture").
				IOS().
				SwipeRight().
				Validate().
				AssertLabelExists("VideoCapture", "拍摄按钮不存在"),
			NewStep("start recording").
				IOS().
				Tap("VideoCapture"). // 开始录像
				Sleep(5).
				Tap("VideoCapture"), // 停止录像
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIOSDouyinAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on 抖音"),
		TestSteps: []IStep{
			NewStep("launch douyin").
				IOS().Home().Tap("//*[@label='抖音']").
				Validate().
				AssertLabelExists("首页", "首页 tab 不存在").
				AssertLabelExists("消息", "消息 tab 不存在"),
			NewStep("swipe up and down").
				IOS().
				Loop(3).
				SwipeUp().SwipeDown(),
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAndroidAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("android ui action"),
		TestSteps: []IStep{
			NewStep("launch douyin").
				Android().Serial("xxx").Tap("抖音").
				Validate().
				AssertNameExists("首页", "首页 tab 不存在").
				AssertNameExists("消息", "消息 tab 不存在"),
			NewStep("swipe up and down").
				Android().Serial("xxx").SwipeUp().SwipeUp().SwipeDown(),
		},
	}
	err := NewRunner(t).Run(testCase)
	if err != nil {
		t.Fatal(err)
	}
}
