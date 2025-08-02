//go:build localtest

package main

import (
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
)

func TestMain(t *testing.T) {
	t.Skip("Skip Bilibili test - requires physical Android device")
	main()
}

func TestRunCaseWithShell(t *testing.T) {
	t.Skip("Skip Bilibili test - requires physical Android device")
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
