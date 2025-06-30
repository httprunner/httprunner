package game_yanglegeyang

import (
	"testing"

	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestGameYanglegeyang(t *testing.T) {
	userInstruction := `羊了个羊是一款热门的消除类小游戏，玩法简单但具有挑战性。以下是游戏的基本规则说明：
1. 游戏目标: 玩家需要通过消除图案来完成关卡，最终目标是清空所有图案。
2. 消除规则:
- 游戏界面中会出现多个图案，玩家需要点击图案将其放入底部的槽中。
- 图案存在多层堆叠的情况，只能点击最上层的完整图案。
- 当槽中有三个相同的图案时，这三个图案会自动消除。
- 玩家需要尽量避免槽中积累过多不同的图案，以免无法继续消除。
- 严禁点击收集槽里的图案，严禁观看广告和使用道具（移出、撤回、洗牌）。
- 请持续推进游戏进程，游戏通关后继续下一关，游戏失败后重新开始。
3. 游戏界面: 图案通常以堆叠的方式呈现，玩家需要逐层消除。
4. 关卡设计: 游戏包含多个关卡，随着关卡的推进，图案的复杂度和数量会增加。
5. 策略性: 玩家需要规划消除顺序，以避免槽中积累过多无法消除的图案。

请严格按照以上游戏规则，开始游戏
`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("羊了个羊小游戏自动化测试").
			SetLLMService(option.DOUBAO_1_5_THINKING_VISION_PRO_250428).
			WithVariables(map[string]interface{}{
				"package_name": "com.ss.android.ugc.aweme",
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动抖音 app").
				Android().
				AppLaunch("$package_name").
				Sleep(5).
				Validate().
				AssertAppInForeground("$package_name"),
			hrp.NewStep("进入「羊了个羊」小游戏").
				Android().
				StartToGoal("搜索「羊了个羊星球」，进入小程序，加入羊群进入游戏",
					option.WithPreMarkOperation(true)).
				Validate().
				AssertAI("当前页面底部包含「移出」「撤回」「洗牌」按钮"),
			hrp.NewStep("开始游戏").
				Android().
				StartToGoal(userInstruction,
					option.WithPreMarkOperation(true),
					option.WithTimeout(300)), // 5 minutes
			hrp.NewStep("退出抖音 app").
				Android().
				AppTerminate("$package_name"),
		},
	}
	err := testCase.Dump2JSON("game_yanglegeyang.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}
