package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func TestLoadTestCases(t *testing.T) {
	// load test cases from folder path
	tc := hrp.TestCasePath("../examples/demo-with-py-plugin/testcases/")
	testCases, err := hrp.LoadTestCases(&tc)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(testCases))

	// load test cases from folder path, including sub folders
	tc = hrp.TestCasePath("../examples/demo-with-py-plugin/")
	testCases, err = hrp.LoadTestCases(&tc)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(testCases))

	// load test cases from single file path
	tc = hrp.TestCasePath(demoTestCaseWithPluginJSONPath)
	testCases, err = hrp.LoadTestCases(&tc)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(testCases))

	// load test cases from TestCase instance
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase").SetWeight(3),
	}
	testCases, err = hrp.LoadTestCases(testcase)
	assert.Nil(t, err)
	assert.Equal(t, len(testCases), 1)

	// load test cases from TestCaseJSON
	testcaseJSON := hrp.TestCaseJSON(`
	{
		"config":{"name":"TestCaseJSON"},
		"teststeps":[
			{"name": "step1", "request":{"url": "https://httpbin.org/get"}},
			{"name": "step2", "shell":{"string": "ls -l"}}
		]
	}`)
	testCases, err = hrp.LoadTestCases(&testcaseJSON)
	assert.Nil(t, err)
	assert.Equal(t, len(testCases), 1)
}

func TestLoadCase(t *testing.T) {
	tcJSON := &hrp.TestCaseDef{}
	tcYAML := &hrp.TestCaseDef{}
	err := hrp.LoadFileObject(demoTestCaseWithPluginJSONPath, tcJSON)
	assert.Nil(t, err)

	err = hrp.LoadFileObject(demoTestCaseWithPluginYAMLPath, tcYAML)
	assert.Nil(t, err)

	assert.Equal(t, tcJSON.Config.Name, tcYAML.Config.Name)
	assert.Equal(t, tcJSON.Config.BaseURL, tcYAML.Config.BaseURL)
	assert.Equal(t, tcJSON.Steps[1].StepName, tcYAML.Steps[1].StepName)
	assert.Equal(t, tcJSON.Steps[1].Request, tcJSON.Steps[1].Request)
}

func TestLoadCaseWithTapOffset(t *testing.T) {
	// Load the android_swipe_tap_loadmore.json test case
	testCasePath := "../examples/uitest/demo_ios_wda_log.json"
	tc := &hrp.TestCaseDef{}
	err := hrp.LoadFileObject(testCasePath, tc)
	assert.Nil(t, err)

	action := tc.Steps[3].IOS.Actions[0]
	assert.Equal(t, option.ACTION_TapByOCR, action.Method)
	assert.Equal(t, "推荐", action.Params)
	assert.Equal(t, "点击推荐", action.Options.Identifier)
	assert.Equal(t, []int{0, -1}, action.Options.ScreenFilterOptions.TapOffset)
}
