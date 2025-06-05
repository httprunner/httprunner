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

func TestStartToGoal(t *testing.T) {
	userInstruction := `连连看是一款经典的益智消除类小游戏，通常以图案或图标为主要元素。以下是连连看的基本规则说明：
	1. 游戏目标: 玩家需要在规定时间内，通过连接相同的图案或图标，将它们从游戏界面中消除。
	2. 连接规则:
	- 两个相同的图案可以通过不超过三条直线连接。
	- 连接线可以水平或垂直，但不能斜线，也不能跨过其他图案。
	- 连接线的转折次数不能超过两次。
	3. 游戏界面:
	- 游戏界面通常是一个矩形区域，内含多个图案或图标，排列成行和列。
	- 图案或图标在未选中状态下背景为白色，选中状态下背景为绿色。
	4. 时间限制: 游戏通常设有时间限制，玩家需要在时间耗尽前完成所有图案的消除。
	5. 得分机制: 每成功连接并消除一对图案，玩家会获得相应的分数。完成游戏后，根据剩余时间和消除效率计算总分。
	6. 关卡设计: 游戏可能包含多个关卡，随着关卡的推进，图案的复杂度和数量会增加。

	注意事项：
	1、当连接错误时，顶部的红心会减少一个，需及时调整策略，避免红心变为0个后游戏失败
	2、不要连续 2 次点击同一个图案
	3、不要犯重复的错误

	请严格按照以上游戏规则，开始游戏
	`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("run ui action with start to goal").
			SetLLMService(option.LLMServiceTypeDoubaoVL),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音「连了又连」小游戏").
				Android().
				StartToGoal("启动抖音，搜索「连了又连」小游戏，并启动游戏").
				Validate().
				AssertAI("当前位于抖音「连了又连」小游戏页面"),
			hrp.NewStep("开始游戏").
				Android().
				StartToGoal(userInstruction, option.WithMaxRetryTimes(100)),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

func TestAIAction(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("run ui action with ai").
			SetLLMService(option.LLMServiceTypeDoubaoVL),
		TestSteps: []hrp.IStep{
			hrp.NewStep("launch settings").
				Android().AIAction("进入手机系统设置").
				Validate().
				AssertAI("当前位于手机设置页面"),
			hrp.NewStep("turn on fly mode").
				Android().AIAction("开启飞行模式").
				Validate().
				AssertAI("飞行模式已开启"),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}
