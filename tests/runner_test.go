package tests

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
	"github.com/httprunner/httprunner/v5/code"
)

func buildHashicorpGoPlugin() {
	log.Info().Msg("[init] build hashicorp go plugin")
	err := hrp.BuildPlugin(tmpl("plugin/debugtalk.go"), tmpl("debugtalk.bin"))
	if err != nil {
		log.Error().Err(err).Msg("build hashicorp go plugin failed")
		os.Exit(code.GetErrorCode(err))
	}
}

func removeHashicorpGoPlugin() {
	log.Info().Msg("[teardown] remove hashicorp go plugin")
	os.Remove(tmpl("debugtalk.bin"))
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
	os.Remove(tmpl(hrp.PluginPySourceFile))
	os.Remove(tmpl(hrp.PluginPySourceGenFile))
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
	refCase := hrp.TestCasePath(demoTestCaseWithPluginJSONPath)
	testcase1 := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase1").
			SetBaseURL("https://postman-echo.com").
			EnablePlugin(), // TODO: FIXME
		TestSteps: []hrp.IStep{
			hrp.NewStep("testcase1-step1").
				GET("/headers").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check http response Content-Type"),
			hrp.NewStep("testcase1-step2").CallRefCase(
				&hrp.TestCase{
					Config: hrp.NewConfig("testcase1-step3-ref-case").SetBaseURL("https://postman-echo.com"),
					TestSteps: []hrp.IStep{
						hrp.NewStep("ip").
							GET("/ip").
							Validate().
							AssertEqual("status_code", 200, "check status code").
							AssertEqual("headers.\"Content-Type\"", "application/json; charset=utf-8", "check http response Content-Type"),
					},
				},
			),
			hrp.NewStep("testcase1-step3").CallRefCase(&refCase),
		},
	}
	testcase2 := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase2").SetWeight(3),
	}

	r := hrp.NewRunner(t)
	r.SetPluginLogOn()
	err := r.Run(testcase1, testcase2)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestRunCaseWithThinkTime(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testcases := []*hrp.TestCase{
		{
			Config: hrp.NewConfig("TestCase1"),
			TestSteps: []hrp.IStep{
				hrp.NewStep("thinkTime").SetThinkTime(2),
			},
		},
		{
			Config: hrp.NewConfig("TestCase2").
				SetThinkTime(hrp.ThinkTimeIgnore, nil, 0),
			TestSteps: []hrp.IStep{
				hrp.NewStep("thinkTime").SetThinkTime(0.5),
			},
		},
		{
			Config: hrp.NewConfig("TestCase3").
				SetThinkTime(hrp.ThinkTimeRandomPercentage, nil, 0),
			TestSteps: []hrp.IStep{
				hrp.NewStep("thinkTime").SetThinkTime(1),
			},
		},
		{
			Config: hrp.NewConfig("TestCase4").
				SetThinkTime(hrp.ThinkTimeRandomPercentage, map[string]interface{}{"min_percentage": 2, "max_percentage": 3}, 2.5),
			TestSteps: []hrp.IStep{
				hrp.NewStep("thinkTime").SetThinkTime(1),
			},
		},
		{
			Config: hrp.NewConfig("TestCase5"),
			TestSteps: []hrp.IStep{
				// think time: 3s, random pct: {"min_percentage":1, "max_percentage":1.5}, limit: 4s
				hrp.NewStep("thinkTime").CallRefCase(&demoTestCaseWithThinkTimePath),
			},
		},
	}
	expectedMinValue := []float64{2, 0, 0.5, 2, 3}
	expectedMaxValue := []float64{2.5, 0.5, 2, 3, 10}
	for idx, testcase := range testcases {
		r := hrp.NewRunner(t)
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
	testcase1 := &hrp.TestCase{
		Config: hrp.NewConfig("complex shell with env variables").
			WithVariables(map[string]interface{}{
				"SS":  "12345",
				"ABC": "$SS",
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("shell21").Shell("echo hello world"),
			// NewStep("shell21").Shell("echo $ABC"),
			// NewStep("shell21").Shell("which hrp"),
		},
	}

	r := hrp.NewRunner(t)
	err := r.Run(testcase1)
	if err != nil {
		t.Fatal()
	}
}

func TestRunCaseWithFunction(t *testing.T) {
	fn1 := func() {
		fmt.Println("call function1 without return")
	}
	num := 0
	fn2 := func() {
		num++
		fmt.Println("call function2 with return value")
	}
	var err3 error
	fn3 := func() {
		num++
		err3 = errors.New("func3 error")
		fmt.Println("call function3 with return value and error")
	}
	testcase1 := &hrp.TestCase{
		Config: hrp.NewConfig("call function"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("fn1").Function(fn1),
			hrp.NewStep("fn2").Function(fn2),
			hrp.NewStep("fn3").Function(fn3),
		},
	}

	r := hrp.NewRunner(t)
	err := r.Run(testcase1)
	if err != nil {
		t.Fatal()
	}

	if !assert.Equal(t, num, 2) {
		t.Fatal()
	}
	if !assert.NotNil(t, err3) {
		t.Fatal()
	}
}

func TestRunCaseWithPluginJSON(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testCase := hrp.TestCasePath(demoTestCaseWithPluginJSONPath)
	// TODO: FIXME, enable plugin
	err := hrp.NewRunner(nil).Run(&testCase) // hrp.Run(testCase)
	if err != nil {
		t.Fatal()
	}
}

// TODO: FIXME
// func TestRunCaseWithPluginYAML(t *testing.T) {
// 	buildHashicorpGoPlugin()
// 	defer removeHashicorpGoPlugin()

// 	testCase := TestCasePath(demoTestCaseWithPluginYAMLPath)
// 	err := NewRunner(nil).Run(&testCase) // hrp.Run(testCase)
// 	if err != nil {
// 		t.Fatal()
// 	}
// }

func TestRunCaseWithRefAPI(t *testing.T) {
	buildHashicorpGoPlugin()
	defer removeHashicorpGoPlugin()

	testCase := hrp.TestCasePath(demoTestCaseWithRefAPIPath)
	err := hrp.NewRunner(nil).Run(&testCase)
	if err != nil {
		t.Fatal()
	}

	refAPI := hrp.APIPath(demoAPIGETPath)
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("TestCase").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("run referenced api").CallRefAPI(&refAPI),
		},
	}

	r := hrp.NewRunner(t)
	err = r.Run(testcase)
	if err != nil {
		t.Fatal()
	}
}

func TestSessionRunner(t *testing.T) {
	testcase := hrp.TestCase{
		Config: hrp.NewConfig("TestCase").
			WithVariables(map[string]interface{}{
				"a":      12.3,
				"b":      3.45,
				"varFoo": "${max($a, $b)}",
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("check variables").
				WithVariables(map[string]interface{}{
					"a":      12.3,
					"b":      34.5,
					"varFoo": "${max($a, $b)}",
				}).
				GET("/hello").
				Validate().
				AssertEqual("status_code", 200, "check status code"),
			// AssertEqual("$varFoo", "$b", "check varFoo value"),
		},
	}

	caseRunner, _ := hrp.NewRunner(t).NewCaseRunner(testcase)
	sessionRunner := caseRunner.NewSession()
	step := testcase.TestSteps[0]
	if !assert.Equal(t, step.Config().Variables["varFoo"], "${max($a, $b)}") {
		t.Fatal()
	}

	err := sessionRunner.ParseStep(step)
	if err != nil {
		t.Fatal()
	}
	if !assert.Equal(t, step.Config().Variables["varFoo"], 34.5) {
		t.Fatal()
	}
}
