//go:build localtest

package tests

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GameInfo 定义游戏界面分析的输出格式
type GameInfo struct {
	Content    string   `json:"content"`     // 必须：人类可读描述
	Thought    string   `json:"thought"`     // 必须：AI推理过程
	GameType   string   `json:"game_type"`   // 游戏类型
	Rows       int      `json:"rows"`        // 行数
	Cols       int      `json:"cols"`        // 列数
	Icons      []string `json:"icons"`       // 图标类型
	TotalIcons int      `json:"total_icons"` // 图标总数
}

// UIElementInfo 定义UI元素分析的输出格式
type UIElementInfo struct {
	Content     string      `json:"content"`      // 必须：人类可读描述
	Thought     string      `json:"thought"`      // 必须：AI推理过程
	ScreenType  string      `json:"screen_type"`  // 屏幕类型
	Elements    []UIElement `json:"elements"`     // UI元素列表
	ButtonCount int         `json:"button_count"` // 按钮数量
	TextCount   int         `json:"text_count"`   // 文本数量
}

// UIElement 定义单个UI元素
type UIElement struct {
	Type        string `json:"type"`        // 元素类型 (button, text, input等)
	Text        string `json:"text"`        // 元素文本
	Clickable   bool   `json:"clickable"`   // 是否可点击
	Description string `json:"description"` // 元素描述
}

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

func TestGameLianliankan(t *testing.T) {
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
		Config: hrp.NewConfig("连连看小游戏自动化测试").
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
	err := testCase.Dump2JSON("game_llk.json")
	require.Nil(t, err)

	err = hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

func TestGame2048(t *testing.T) {
	userInstruction := `2048 是一款数字合并类的益智小游戏，以下是它的基本规则：
1、游戏目标：在一个 4x4 的网格中，通过合并相同数字的方块，最终得到一个数值为 2048 的方块。当然，若你能继续合并，也可追求更高的数字。
2、初始状态：游戏开始时，网格中会随机出现两个数字为 2 或 4 的方块。
3、移动操作：玩家可以选择上、下、左、右四个方向进行移动。每次移动时，所有方块会朝着指定方向滑动，直到碰到边界或其他方块。
4、合并规则：当两个相同数字的方块在移动过程中相遇时，它们会合并成一个新的方块，新方块的数值为原来两个方块数值之和。例如，两个 2 合并成一个 4，两个 4 合并成一个 8，依此类推。
5、新方块生成：每次移动结束后，网格中会随机出现一个新的数字为 2 或 4 的方块。
6、注意事项：若连续多次滑动无法生效，请调整策略；例如，向上无法滑动，可以尝试向下滑；向左无法滑动，可以尝试向右滑。
7、游戏结束：当网格被填满，且没有可合并的方块时，游戏结束，停止游戏。

请严格按照以上游戏规则，开始游戏
`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("2048 小游戏自动化测试").
			SetLLMService(option.DOUBAO_1_5_UI_TARS_250328),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音「2048经典」小游戏").
				Android().
				StartToGoal("启动抖音，搜索「2048经典」小游戏，并启动游戏").
				Validate().
				AssertAI("当前位于抖音「2048」小游戏页面"),
			hrp.NewStep("开始游戏").
				Android().
				StartToGoal(userInstruction, option.WithMaxRetryTimes(100)),
		},
	}
	err := testCase.Dump2JSON("game_2048.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
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

func TestAIQuery(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("AIQuery Demo with OutputSchema").
			SetLLMService(option.DOUBAO_SEED_1_6_250615), // Configure LLM service for AI operations
		TestSteps: []hrp.IStep{
			// Step 1: Take a screenshot for analysis
			hrp.NewStep("Take Screenshot").
				Android().
				ScreenShot(),

			// Step 2: Basic AIQuery without OutputSchema
			hrp.NewStep("Basic Query").
				Android().
				AIQuery("Please describe what is displayed on the screen and identify any interactive elements"),

			// Step 3: Use AIQuery to extract specific information
			hrp.NewStep("Extract App Information").
				Android().
				AIQuery("What apps are visible on the screen? List them as a comma-separated string"),

			// Step 4: Use AIQuery for UI element analysis
			hrp.NewStep("Analyze UI Elements").
				Android().
				AIQuery("Are there any buttons or clickable elements visible? Describe their locations and purposes"),

			// Step 5: Use AIQuery with validation
			hrp.NewStep("Query and Validate").
				Android().
				AIQuery("Is the home screen currently displayed?").
				Validate().
				AssertAI("The query result should indicate whether home screen is visible"),

			// Step 6: Use AIQuery with simple custom OutputSchema
			hrp.NewStep("Query with Simple Custom Schema").
				Android().
				AIQuery("Analyze the screen and provide structured information about UI elements",
					option.WithOutputSchema(struct {
						Content     string   `json:"content"`
						Thought     string   `json:"thought"`
						ElementType string   `json:"element_type"`
						ElementText []string `json:"element_text"`
						ButtonCount int      `json:"button_count"`
					}{})),

			// Step 7: Use AIQuery with GameInfo OutputSchema
			hrp.NewStep("Game Analysis with Custom Schema").
				Android().
				AIQuery("分析这个游戏界面，告诉我游戏类型、行列数和图标信息",
					option.WithOutputSchema(GameInfo{})),

			// Step 8: Use AIQuery with UIElementInfo OutputSchema
			hrp.NewStep("UI Element Analysis with Custom Schema").
				Android().
				AIQuery("分析屏幕上的UI元素，识别所有按钮、文本和可交互元素",
					option.WithOutputSchema(UIElementInfo{})),

			// Step 9: Complex analysis with nested structure
			hrp.NewStep("Complex Analysis with Nested Schema").
				Android().
				AIQuery("Provide a comprehensive analysis of this interface including all interactive elements and their properties",
					option.WithOutputSchema(struct {
						Content     string `json:"content"`
						Thought     string `json:"thought"`
						AppName     string `json:"app_name"`
						ScreenTitle string `json:"screen_title"`
						MainActions []struct {
							Name        string `json:"name"`
							Description string `json:"description"`
							Available   bool   `json:"available"`
						} `json:"main_actions"`
						NavigationElements []struct {
							Type     string `json:"type"`
							Label    string `json:"label"`
							Position string `json:"position"`
						} `json:"navigation_elements"`
						ContentSummary struct {
							HasImages bool     `json:"has_images"`
							HasText   bool     `json:"has_text"`
							HasForms  bool     `json:"has_forms"`
							Keywords  []string `json:"keywords"`
						} `json:"content_summary"`
					}{})),
		},
	}

	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}
