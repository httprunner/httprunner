package hrp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenHTMLReport(t *testing.T) {
	summary := NewSummary()

	caseSummary1 := NewCaseSummary()
	stepResult1 := &StepResult{}
	caseSummary1.AddStepResult(stepResult1)
	summary.AddCaseSummary(caseSummary1)

	caseSummary2 := NewCaseSummary()
	stepResult2 := &StepResult{
		Name:        "Test",
		StepType:    stepTypeRequest,
		Success:     false,
		ContentSize: 0,
		Attachments: "err",
	}
	caseSummary2.AddStepResult(stepResult2)
	summary.AddCaseSummary(caseSummary2)

	err := summary.GenHTMLReport()
	if err != nil {
		t.Error(err)
	}
}

func TestTestCaseSummary_AddStepResult(t *testing.T) {
	caseSummary := NewCaseSummary()
	stepResult1 := &StepResult{
		Name:        "Test1",
		StepType:    stepTypeRequest,
		Success:     true,
		ContentSize: 0,
		Attachments: "err",
	}
	caseSummary.AddStepResult(stepResult1)
	stepResult2 := &StepResult{
		Name:        "Test2",
		StepType:    stepTypeTestCase,
		Success:     false,
		ContentSize: 0,
		Attachments: "err",
		Data:        []*StepResult{stepResult1},
	}
	caseSummary.AddStepResult(stepResult2)

	if !assert.Equal(t, 2, len(caseSummary.Records)) {
		t.Fatal()
	}
	if !assert.False(t, caseSummary.Success) {
		t.Fatal()
	}
}
