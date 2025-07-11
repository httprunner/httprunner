package game_romantic_restaurant

import (
	"testing"

	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestGameRomanticRestaurant(t *testing.T) {
	userInstruction := `浪漫餐厅是一款经营类游戏，以下是游戏的基本规则说明：
1、点击右下角锅铲，开始任务
2、将棋子拖拽至相同棋子，可升级生成新棋子；注意，必须是相同类别和形状的棋子才能合成，例如，长面包和圆面包不能合成，方形蛋糕和三角形蛋糕不能合成
3、拖拽相同棋子时，被部分遮挡的棋子只能作为拖拽终点，不能作为拖拽起点
4、当游戏界面中没有相同棋子时，可以点击游戏页面中央的购物袋生成新的棋子
5、若不知道如何操作，请按照游戏指引进行游玩
6、不要连续重复上一步操作，合成失败后及时更换策略

请严格按照以上游戏规则，开始游戏
`
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("浪漫餐厅小游戏自动化测试").
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
			// hrp.NewStep("进入「浪漫餐厅」小游戏").
			// 	Android().
			// 	StartToGoal("搜索「浪漫餐厅」，点击进入「游戏」tab，进入小游戏",
			// 		option.WithPreMarkOperation(true)).
			// 	Validate().
			// 	AssertAI("当前位于游戏界面"),
			hrp.NewStep("进入「浪漫餐厅」小游戏").
				Android().
				Home().
				StartToGoal("在手机桌面点击「浪漫餐厅」启动小游戏，等待游戏加载完成",
					option.WithPreMarkOperation(true)).
				Validate().
				AssertAI("当前位于游戏界面"),
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
	err := testCase.Dump2JSON("game_romantic_restaurant.json")
	require.Nil(t, err)

	// err = hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}
