package hrp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

func buildHashicorpGoPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	err := BuildPlugin(tmpl("plugin/debugtalk.go"), tmpl("debugtalk.bin"))
	if err != nil {
		log.Error().Err(err).Msg("build hashicorp go plugin failed")
		os.Exit(code.GetErrorCode(err))
	}
}

func removeHashicorpGoPlugin() {
	log.Info().Msg("[teardown] remove hashicorp go plugin")
	os.Remove(tmpl("debugtalk.bin"))
	pluginPath, _ := filepath.Abs(tmpl("debugtalk.bin"))
	pluginMap.Delete(pluginPath)
}

func buildHashicorpPyPlugin() {
	log.Info().Msg("[init] prepare hashicorp python plugin")
	src, _ := os.ReadFile(tmpl("plugin/debugtalk.py"))
	err := os.WriteFile(tmpl("debugtalk.py"), src, 0o644)
	if err != nil {
		log.Error().Err(err).Msg("copy hashicorp python plugin failed")
		os.Exit(code.GetErrorCode(err))
	}
}

func removeHashicorpPyPlugin() {
	log.Info().Msg("[teardown] remove hashicorp python plugin")
	// on v4.1^, running case will generate .debugtalk_gen.py used by python plugin
	os.Remove(tmpl(PluginPySourceFile))
	os.Remove(tmpl(PluginPySourceGenFile))
}

func TestRunCaseWithGoPlugin(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	assertRunTestCases(t)
}

func TestRunCaseWithPythonPlugin(t *testing.T) {
	buildHashicorpPyPlugin()
	defer removeHashicorpPyPlugin()

	assertRunTestCases(t)
}

func assertRunTestCases(t *testing.T) {
	refCase := TestCasePath(demoTestCaseWithPluginJSONPath)
	testcase1 := &TestCase{
		Config: NewConfig("TestCase1").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{
			NewStep("testcase1-step1").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check http response Content-Type"),
			NewStep("testcase1-step2").CallRefCase(
				&TestCase{
					Config: NewConfig("testcase1-step3-ref-case").SetBaseURL("https://postman-echo.com"),
					TestSteps: []IStep{
						NewStep("ip").
							GET("/ip").
							Validate().
							AssertEqual("status_code", 200, "check status code").
							AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check http response Content-Type"),
					},
				},
			),
			NewStep("testcase1-step3").CallRefCase(&refCase),
		},
	}
	testcase2 := &TestCase{
		Config: NewConfig("TestCase2").SetWeight(3),
	}

	r := NewRunner(t)
	r.SetPluginLogOn()
	err := r.Run(testcase1, testcase2)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestRunCaseWithThinkTime(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testcases := []*TestCase{
		{
			Config: NewConfig("TestCase1"),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(2),
			},
		},
		{
			Config: NewConfig("TestCase2").
				SetThinkTime(thinkTimeIgnore, nil, 0),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(0.5),
			},
		},
		{
			Config: NewConfig("TestCase3").
				SetThinkTime(thinkTimeRandomPercentage, nil, 0),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(1),
			},
		},
		{
			Config: NewConfig("TestCase4").
				SetThinkTime(thinkTimeRandomPercentage, map[string]interface{}{"min_percentage": 2, "max_percentage": 3}, 2.5),
			TestSteps: []IStep{
				NewStep("thinkTime").SetThinkTime(1),
			},
		},
		{
			Config: NewConfig("TestCase5"),
			TestSteps: []IStep{
				// think time: 3s, random pct: {"min_percentage":1, "max_percentage":1.5}, limit: 4s
				NewStep("thinkTime").CallRefCase(&demoTestCaseWithThinkTimePath),
			},
		},
	}
	expectedMinValue := []float64{2, 0, 0.5, 2, 3}
	expectedMaxValue := []float64{2.5, 0.5, 2, 3, 10}
	for idx, testcase := range testcases {
		r := NewRunner(t)
		startTime := time.Now()
		err := r.Run(testcase)
		if err != nil {
			t.Fatalf("run testcase error: %v", err)
		}
		duration := time.Since(startTime)
		minValue := time.Duration(expectedMinValue[idx]*1000) * time.Millisecond
		maxValue := time.Duration(expectedMaxValue[idx]*1000) * time.Millisecond
		if duration < minValue || duration > maxValue {
			t.Fatalf("failed to test think time, expect value: [%v, %v], actual value: %v", minValue, maxValue, duration)
		}
	}
}

func TestRunCaseWithShell(t *testing.T) {
	testcase1 := &TestCase{
		Config: NewConfig("complex shell with env variables").
			WithVariables(map[string]interface{}{
				"SS":  "12345",
				"ABC": "$SS",
			}),
		TestSteps: []IStep{
			NewStep("shell21").Shell("echo hello world"),
			// NewStep("shell21").Shell("echo $ABC"),
			// NewStep("shell21").Shell("which hrp"),
		},
	}

	r := NewRunner(t)
	err := r.Run(testcase1)
	if err != nil {
		t.Fatal()
	}
}

func TestRunCaseWithPluginJSON(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testCase := TestCasePath(demoTestCaseWithPluginJSONPath)
	err := NewRunner(nil).Run(&testCase) // hrp.Run(testCase)
	if err != nil {
		t.Fatal()
	}
}

func TestRunCaseWithPluginYAML(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testCase := TestCasePath(demoTestCaseWithPluginYAMLPath)
	err := NewRunner(nil).Run(&testCase) // hrp.Run(testCase)
	if err != nil {
		t.Fatal()
	}
}

func TestRunCaseWithRefAPI(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testCase := TestCasePath(demoTestCaseWithRefAPIPath)
	err := NewRunner(nil).Run(&testCase)
	if err != nil {
		t.Fatal()
	}

	refAPI := APIPath(demoAPIGETPath)
	testcase := &TestCase{
		Config: NewConfig("TestCase").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []IStep{
			NewStep("run referenced api").CallRefAPI(&refAPI),
		},
	}

	r := NewRunner(t)
	err = r.Run(testcase)
	if err != nil {
		t.Fatal()
	}
}

func TestLoadTestCases(t *testing.T) {
	// load test cases from folder path
	tc := TestCasePath("../examples/demo-with-py-plugin/testcases/")
	testCases, err := LoadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, 4, len(testCases)) {
		t.Fatal()
	}

	// load test cases from folder path, including sub folders
	tc = TestCasePath("../examples/demo-with-py-plugin/")
	testCases, err = LoadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, 4, len(testCases)) {
		t.Fatal()
	}

	// load test cases from single file path
	tc = TestCasePath(demoTestCaseWithPluginJSONPath)
	testCases, err = LoadTestCases(&tc)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, 1, len(testCases)) {
		t.Fatal()
	}

	// load test cases from TestCase instance
	testcase := &TestCase{
		Config: NewConfig("TestCase").SetWeight(3),
	}
	testCases, err = LoadTestCases(testcase)
	if !assert.Nil(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, len(testCases), 1) {
		t.Fatal()
	}
}
