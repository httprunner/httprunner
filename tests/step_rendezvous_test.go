package tests

import (
	"math"
	"testing"

	hrp "github.com/httprunner/httprunner/v5"
)

func TestRunCaseWithRendezvous(t *testing.T) {
	rendezvousBoundaryTestcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with functions").
			SetBaseURL("https://postman-echo.com").
			WithVariables(map[string]interface{}{
				"n": 5,
				"a": 12.3,
				"b": 3.45,
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("test negative number").
				SetRendezvous("test negative number").
				WithUserNumber(-1),
			hrp.NewStep("test overflow number").
				SetRendezvous("test overflow number").
				WithUserNumber(1000000),
			hrp.NewStep("test negative percent").
				SetRendezvous("test very low percent").
				WithUserPercent(-0.5),
			hrp.NewStep("test very low percent").
				SetRendezvous("test very low percent").
				WithUserPercent(0.00001),
			hrp.NewStep("test overflow percent").
				SetRendezvous("test overflow percent").
				WithUserPercent(1.5),
			hrp.NewStep("test conflict params").
				SetRendezvous("test conflict params").
				WithUserNumber(1).
				WithUserPercent(0.123),
			hrp.NewStep("test negative timeout").
				SetRendezvous("test negative timeout").
				WithTimeout(-1000),
		},
	}

	type rendezvousParam struct {
		number  int64
		percent float32
		timeout int64
	}
	expectedRendezvousParams := []rendezvousParam{
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 0, percent: 0.00001, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
		{number: 100, percent: 1, timeout: 5000},
	}

	rendezvousList := hrp.InitRendezvous(rendezvousBoundaryTestcase, 100)

	for i, r := range rendezvousList {
		if r.Number != expectedRendezvousParams[i].number {
			t.Fatalf("run rendezvous %v error: expected number: %v, real number: %v", r.Name, expectedRendezvousParams[i].number, r.Number)
		}
		if math.Abs(float64(r.Percent-expectedRendezvousParams[i].percent)) > 0.001 {
			t.Fatalf("run rendezvous %v error: expected percent: %v, real percent: %v", r.Name, expectedRendezvousParams[i].percent, r.Percent)
		}
		if r.Timeout != expectedRendezvousParams[i].timeout {
			t.Fatalf("run rendezvous %v error: expected timeout: %v, real timeout: %v", r.Name, expectedRendezvousParams[i].timeout, r.Timeout)
		}
	}
}
