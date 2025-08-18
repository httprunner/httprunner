package uitest

import (
	"testing"

	"github.com/stretchr/testify/require"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestSPHSearchPage(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("视频号搜索").
			SetLLMService(option.DOUBAO_1_5_UI_TARS_250328), // Configure LLM service for AI operations
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动视频号 app").
				Android().
				AppLaunch("com.tencent.mm").
				Sleep(5).
				Validate().
				AssertAppInForeground("com.tencent.mm"),
			hrp.NewStep("进入视频号搜索页面").
				Android().
				StartToGoal("进入「发现」页，点击进入「视频号」页面，点击搜索框").
				Validate().
				AssertAI("当前页面包含搜索框和搜索按钮"),
			hrp.NewStep("搜索短剧「$dramaName」").
				WithParameters(map[string]interface{}{
					"dramaName": []string{
						"督军，你家小福包有祖传乌鸦嘴",
						"换亲后我顺便换了江山，很合理吧",
						"穿过荆棘拥抱你",
						"认亲后，误入帮派成团宠",
						"欲念疯长",
						"花轿临门她拒嫁，只盼故人归",
						"太监武帝，功法自动大圆满",
						"容先生，你的爱意藏不住了",
						"回家给娘亲改命，心声咋还泄露了",
						"相亲遇甜妹，偷娶她闺蜜",
					},
				}).
				Android().
				StartToGoal("输入「$dramaName」，点击搜索").
				Sleep(1).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2).
				SwipeUp().SleepRandom(1, 2),
		},
	}

	err := testCase.Dump2JSON("sph_search.json")
	require.Nil(t, err)

	// err := hrp.NewRunner(t).Run(testCase)
	// assert.Nil(t, err)
}
