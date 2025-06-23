package parameters

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
)

// TestParametersExecutionScenarios 涵盖了参数化的核心执行场景，
// 包括纯参数驱动、参数与循环结合，以及使用随机和限制等设置。
func TestParametersExecutionScenarios(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("测试参数化核心执行场景").
			WithVariables(map[string]interface{}{"loops": 2}),
		TestSteps: []hrp.IStep{
			// 场景1: 纯参数驱动
			hrp.NewStep("API请求 - 纯参数").
				WithParameters(map[string]interface{}{
					"arg1": []int{10, 20},
					"arg2": []string{"a", "b"},
				}).
				GET("https://postman-echo.com/get").
				WithParams(map[string]interface{}{"p1": "$arg1", "p2": "$arg2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code"),

			// 场景2: 参数与 Loops 结合
			hrp.NewStep("API请求 - 参数与Loops结合").
				WithParameters(map[string]interface{}{
					"word": []string{"hello", "world"},
				}).
				Loop(3). // 每个参数执行3次
				GET("https://postman-echo.com/get").
				WithParams(map[string]interface{}{"search": "$word"}).
				Validate().
				AssertEqual("status_code", 200, "check status code"),

			// 场景3: 参数设置 (随机, 限制)
			hrp.NewStep("API请求 - 参数设置").
				WithParameters(map[string]interface{}{
					"city": []string{"chengdu", "beijing", "shanghai", "guangzhou"},
				}).
				WithParametersSetting(
					hrp.WithRandomOrder(), // 随机顺序
					hrp.WithLimit(2),      // 总共执行2次
				).
				GET("https://postman-echo.com/get").
				WithParams(map[string]interface{}{"city": "$city"}).
				Validate().
				AssertEqual("status_code", 200, "check status code"),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

// TestParametersVariableOverride 用于验证参数(parameters)如何覆盖测试用例配置(config)
// 和步骤(step)中定义的同名变量(variables)。
func TestParametersVariableOverride(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("测试参数覆盖变量的优先级").
			WithVariables(map[string]interface{}{
				"p1": "config_level", // 将被步骤级变量覆盖
				"p2": "config_level", // 将被参数覆盖
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("API请求 - 验证变量覆盖").
				WithVariables(map[string]interface{}{
					"p1": "step_level", // 不会被参数覆盖, 最终值为 "step_level"
					"p2": "step_level", // 会被参数覆盖
					"p3": "step_level", // 新增的步骤级变量
				}).
				WithParameters(map[string]interface{}{
					"p2-p4": [][]interface{}{
						{"param_level_2", "param_level_4_a"},
						{"param_level_2", "param_level_4_b"},
					},
				}).
				GET("https://postman-echo.com/get").
				WithParams(map[string]interface{}{
					"param1": "$p1", // 预期: step_level
					"param2": "$p2", // 预期: param_level_2
					"param3": "$p3", // 预期: step_level
					"param4": "$p4", // 预期: param_level_4_a/b
				}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.args.param1", "step_level", "p1 should be step_level").
				AssertEqual("body.args.param2", "param_level_2", "p2 should be param_level_2").
				AssertEqual("body.args.param3", "step_level", "p3 should be step_level"),
		},
	}
	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}

// TestParametersForMobileUI 演示了如何在移动端UI测试中使用参数化来驱动测试。
func TestParametersForMobileUI(t *testing.T) {
	testCase := &hrp.TestCase{
		Config: hrp.NewConfig("小红书UI参数化搜索").
			SetAIOptions(option.WithLLMConfig(
				option.NewLLMServiceConfig(option.DOUBAO_1_5_UI_TARS_250328),
			)),
		TestSteps: []hrp.IStep{
			hrp.NewStep("启动APP").
				Android().
				AppLaunch("com.xingin.xhs").
				Sleep(5).
				Validate().
				AssertAppInForeground("com.xingin.xhs"),
			hrp.NewStep("UI搜索 - 单参数").
				WithParameters(map[string]interface{}{
					"query": []string{"成都", "北京"},
				}).
				Android().
				StartToGoal("进入搜索框，输入「$query」，等待搜索建议出现").
				Sleep(2),
			hrp.NewStep("UI搜索 - 复合参数").
				WithParameters(map[string]interface{}{
					"query-category": [][]string{
						{"美食", "食物"},
						{"旅游", "地点"},
					},
				}).
				Android().
				StartToGoal("进入搜索框，输入「$query」，并确认其类别为「$category」，等待搜索建议出现").
				Sleep(2),
		},
	}

	err := hrp.NewRunner(t).Run(testCase)
	assert.Nil(t, err)
}
