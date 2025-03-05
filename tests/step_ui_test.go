//go:build localtest

package tests

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
)

func TestIOSSettingsAction(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("ios ui action on Settings").
			SetIOS(option.WithWDAPort(8700), option.WithWDAMjpegPort(8800)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("launch Settings").
				IOS().Home().TapByOCR("设置").
				Validate().
				AssertNameExists("飞行模式").
				AssertLabelExists("蓝牙").
				AssertOCRExists("个人热点"),
			hrp.NewStep("swipe up and down").
				IOS().SwipeUp().SwipeUp().SwipeDown(),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

func TestIOSSearchApp(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("ios ui action on Search App 资源库"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("进入 App 资源库 搜索框").
				IOS().Home().SwipeLeft().SwipeLeft().TapByCV("dewey-search-field").
				Validate().
				AssertLabelExists("取消"),
			hrp.NewStep("搜索抖音").
				IOS().Input("抖音\n"),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

func TestIOSAppLaunch(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("启动 & 关闭 App").
			SetIOS(option.WithWDAPort(8700), option.WithWDAMjpegPort(8800)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("终止今日头条").
				IOS().AppTerminate("com.ss.iphone.article.News"),
			hrp.NewStep("启动今日头条").
				IOS().AppLaunch("com.ss.iphone.article.News"),
			hrp.NewStep("终止今日头条").
				IOS().AppTerminate("com.ss.iphone.article.News"),
			hrp.NewStep("启动今日头条").
				IOS().AppLaunch("com.ss.iphone.article.News"),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

func TestAndroidAction(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("android ui action"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("launch douyin").
				Android().Serial("xxx").TapByOCR("抖音").
				Validate().
				AssertNameExists("首页", "首页 tab 不存在").
				AssertNameExists("消息", "消息 tab 不存在"),
			hrp.NewStep("swipe up and down").
				Android().Serial("xxx").SwipeUp().SwipeUp().SwipeDown(),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}
