package hrp

import "testing"

func TestGenHTMLReport(t *testing.T) {
	summary := newOutSummary()
	caseSummary1 := newSummary()
	caseSummary2 := newSummary()
	stepResult1 := &StepResult{}
	stepResult2 := &StepResult{
		Name:        "Test",
		StepType:    stepTypeRequest,
		Success:     false,
		ContentSize: 0,
		Attachments: "err",
	}
	caseSummary1.Records = []*StepResult{stepResult1, stepResult2, nil}
	summary.appendCaseSummary(caseSummary1)
	summary.appendCaseSummary(caseSummary2)
	err := summary.genHTMLReport()
	if err != nil {
		t.Error(err)
	}
}
