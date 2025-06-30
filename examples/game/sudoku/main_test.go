package game_sudoku

import (
	"testing"

	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestGameSudoku(t *testing.T) {
	userInstruction := `每天数独是一款逻辑推理游戏，玩家需要通过推理来确定黄色方块的所在位置，以下是游戏的基本规则说明：
1、方块外面的数字代表所在那一行或一列的黄色方块数量。
2、初始状态为白色方块，选择正确后变为黄色方块，选择错误后变为红底的 X。
3、如果同一行或列有两个数字，则至少需要一个白底 X 分割它们作为间隔。
4、如果数字与格子最大数相同时，该列或行必然全都是黄色方块。
5、只能点击白色方块，不要重复点击同一个方块。

请严格按照以上游戏规则，开始游戏
`
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("每天数独小游戏自动化测试").
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
			hrp.NewStep("进入「每天数独」小游戏").
				Android().
				StartToGoal("搜索「每天数独」，进入小游戏",
					option.WithPreMarkOperation(true)).
				Validate().
				AssertAI("当前页面底部包含「开始」按钮"),
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
	err := testCase.Dump2JSON("game_sudoku.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}
