package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hrp "github.com/httprunner/httprunner/v5"
)

func TestGenHTMLReport(t *testing.T) {
	summary := hrp.NewSummary()

	caseSummary1 := hrp.NewCaseSummary()
	stepResult1 := &hrp.StepResult{}
	caseSummary1.AddStepResult(stepResult1)
	summary.AddCaseSummary(caseSummary1)

	caseSummary2 := hrp.NewCaseSummary()
	stepResult2 := &hrp.StepResult{
		Name:        "Test",
		StepType:    hrp.StepTypeRequest,
		Success:     false,
		ContentSize: 0,
		Attachments: "err",
	}
	caseSummary2.AddStepResult(stepResult2)
	summary.AddCaseSummary(caseSummary2)

	_, err := summary.GenSummary()
	if err != nil {
		t.Error(err)
	}

	err = summary.GenHTMLReport()
	if err != nil {
		t.Error(err)
	}
}

func TestTestCaseSummary_AddStepResult(t *testing.T) {
	caseSummary := hrp.NewCaseSummary()
	stepResult1 := &hrp.StepResult{
		Name:        "Test1",
		StepType:    hrp.StepTypeRequest,
		Success:     true,
		ContentSize: 0,
		Attachments: "err",
	}
	caseSummary.AddStepResult(stepResult1)
	stepResult2 := &hrp.StepResult{
		Name:        "Test2",
		StepType:    hrp.StepTypeTestCase,
		Success:     false,
		ContentSize: 0,
		Attachments: "err",
		Data:        []*hrp.StepResult{stepResult1},
	}
	caseSummary.AddStepResult(stepResult2)

	if !assert.Equal(t, 2, len(caseSummary.Records)) {
		t.Fatal()
	}
	if !assert.False(t, caseSummary.Success) {
		t.Fatal()
	}
}
