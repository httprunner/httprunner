package game_yuedongxiaozi

import (
	"testing"

	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestGameZhuadaE(t *testing.T) {
	userInstruction := `跃动小子是一款开宝箱类的小游戏，以下是游戏的基本规则说明：
1、打开宝箱，按照游戏指引进行「出售」或「装备」操作。
2、请持续推进游戏进程。
3、屏幕底部的黑白按钮不要进行点击操作。

请严格按照以上游戏规则，开始游戏
`

	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("跃动小子小游戏自动化测试").
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
			hrp.NewStep("启动「跃动小子」小游戏").
				Android().
				StartToGoal("搜索「跃动小子」，启动小游戏",
					option.WithPreMarkOperation(true)).
				Validate().
				AssertAI("当前页面底部包含「领地」「试炼」按钮"),
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
	err := testCase.Dump2JSON("game_yuedongxiaozi.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}
