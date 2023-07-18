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
				IOS().AppLaunch("com.ss.iphone.article.News"),
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
				Loop(3).
				IOS().
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
