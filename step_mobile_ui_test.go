//go:build localtest

package hrp

import (
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func TestIOSSettingsAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("ios ui action on Settings").
			SetIOS(option.WithWDAPort(8700), option.WithWDAMjpegPort(8800)),
		TestSteps: []IStep{
			NewStep("launch Settings").
				IOS().Home().TapByOCR("设置").
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
				IOS().Home().SwipeLeft().SwipeLeft().TapByCV("dewey-search-field").
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
			SetIOS(option.WithWDAPort(8700), option.WithWDAMjpegPort(8800)),
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

func TestAndroidAction(t *testing.T) {
	testCase := &TestCase{
		Config: NewConfig("android ui action"),
		TestSteps: []IStep{
			NewStep("launch douyin").
				Android().Serial("xxx").TapByOCR("抖音").
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
