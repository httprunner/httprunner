//go:build localtest

package tests

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
1. 游戏目标: 玩家需要通过连接相同的图案或图标，将它们从游戏界面中消除。
2. 连接规则:
- 两个相同的图案可以通过不超过三条直线连接。
- 连接线可以水平或垂直，但不能斜线，也不能跨过其他图案。
- 连接线的转折次数不能超过两次。
3. 游戏界面:
- 游戏界面是一个矩形区域，内含多个图案或图标，排列成行和列；图案或图标在未选中状态下背景为白色，选中状态下背景为绿色。
- 游戏界面下方是道具区域，共有 3 种道具，从左到右分别是：「高亮显示」、「随机打乱」、「减少种类」。
4、游戏攻略：建议多次使用道具，可以降低游戏难度
- 优先使用「减少种类」道具，可以将图案种类随机减少一种
- 遇到困难时，推荐使用「随机打乱」道具，可以获得很多新的消除机会
- 观看广告视频，待屏幕右上角出现「领取成功」后，点击其右侧的 X 即可关闭广告，继续游戏

请严格按照以上游戏规则，开始游戏
`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("run ui action with start to goal").
			SetLLMService(option.DOUBAO_1_5_THINKING_VISION_PRO_250428),
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
	err := testCase.Dump2JSON("start_llk_game.json")
	require.Nil(t, err)

	err = hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

func TestAIAction(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("run ui action with ai").
			SetLLMService(option.DOUBAO_1_5_THINKING_VISION_PRO_250428),
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
