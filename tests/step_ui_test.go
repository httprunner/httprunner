//go:build localtest

package tests

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
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
	t.Skip("Skip iOS UI test - requires physical iOS device with WDA running")
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
	t.Skip("Skip iOS UI test - requires physical iOS device with WDA running")
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
	t.Skip("Skip iOS UI test - requires physical iOS device with WDA running")
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
	t.Skip("Skip Android UI test - requires physical Android device with ADB")
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

func TestAIAction(t *testing.T) {
	t.Skip("Skip AI UI test - requires physical Android device with AI service configuration")
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
	t.Skip("Skip AI Query test - requires physical Android device with AI service configuration")
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
