//go:build localtest

package main

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestMain(t *testing.T) {
	main()
}

func TestRunCaseWithShell(t *testing.T) {
	testcase1 := &hrp.TestCase{
		Config: hrp.NewConfig("run ui test on bili android").
			WithVariables(map[string]interface{}{
				"SerialNumber": "${ENV(SerialNumber)}",
				"RunTimes":     3,
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("run bili android").
				Shell("bili_android"),
		},
	}

	testcase1.Dump2JSON("bili_android.json")

	r := hrp.NewRunner(t)
	err := r.Run(testcase1)
	if err != nil {
		t.Fatal()
	}
}
