package game_zhuadae

import (
	"testing"

	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestGameZhuadaE(t *testing.T) {
	userInstruction := `抓大鹅是一款抓取类小游戏，以下是游戏的基本规则说明：
1. 游戏目标: 玩家需要通过抓取图案来完成关卡，最终目标是清空所有图案。
2. 抓取规则:
- 游戏界面中会出现多个图案，图案存在多层堆叠的情况，玩家需要点击图案将其抓取放入到槽中。
- 当抓取到三个相同的图案放入抓取槽时，这三个图案会成功消除。
- 需要尽量避免抓取槽满的情况，抓取槽满时游戏失败。
- 游戏通关后继续进入下一关，游戏失败后重新开始游戏。

请严格按照以上游戏规则，开始游戏
`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("抓大鹅小游戏自动化测试").
			SetLLMService(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithVariables(map[string]interface{}{
				"package_name": "com.ss.android.ugc.aweme",
			}),
		TestSteps: []hrp.IStep{
			// hrp.NewStep("启动抖音 app").
			// 	Android().
			// 	AppLaunch("$package_name").
			// 	Sleep(5).
			// 	Validate().
			// 	AssertAppInForeground("$package_name"),
			// hrp.NewStep("启动「抓大鹅」小游戏").
			// 	Android().
			// 	StartToGoal("搜索「抓大鹅」，启动小游戏",
			// 		option.WithPreMarkOperation(true)).
			// 	Sleep(10).
			// 	Validate().
			// 	AssertAI("当前页面底部包含「抓大鹅」"),
			hrp.NewStep("启动「抓大鹅」小游戏").
				Android().
				Home().
				StartToGoal("在手机桌面点击「抓大鹅」启动小游戏，处理弹窗，等待游戏加载完成",
					option.WithPreMarkOperation(true)).
				Sleep(10).
				Validate().
				AssertAI("当前页面底部包含「抓大鹅」"),
			hrp.NewStep("进入「抓大鹅」小游戏").
				Android().
				StartToGoal("点击「抓大鹅」，进入小游戏",
					option.WithPreMarkOperation(true)).
				Sleep(10).
				Validate().
				AssertAI("当前页面底部包含「移出」「凑齐」「打乱」按钮"),
			hrp.NewStep("开始游戏").
				Android().
				StartToGoal(userInstruction,
					option.WithPreMarkOperation(true),
					option.WithTimeLimit(300)), // 5 minutes
			hrp.NewStep("退出抖音 app").
				Android().
				AppTerminate("$package_name"),
		},
	}
	err := testCase.Dump2JSON("game_zhuadae.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}
