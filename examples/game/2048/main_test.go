package game_2048

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/require"
)

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
