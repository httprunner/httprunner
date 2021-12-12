package hrp

import (
	"time"

	"github.com/myzhan/boomer"

	"github.com/httprunner/hrp/internal/ga"
)

func NewStandaloneBoomer(spawnCount int, spawnRate float64) *hrpBoomer {
	b := &hrpBoomer{
		Boomer: boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		debug:  false,
	}
	return b
}

type hrpBoomer struct {
	*boomer.Boomer
	debug bool
}

// SetDebug configures whether to log HTTP request and response content.
func (b *hrpBoomer) SetDebug(debug bool) *hrpBoomer {
	b.debug = debug
	return b
}

// Run starts to run load test for one or multiple testcases.
func (b *hrpBoomer) Run(testcases ...ITestCase) {
	event := ga.EventTracking{
		Category: "RunLoadTests",
		Action:   "hrp boom",
	}
	// report start event
	go ga.SendEvent(event)
	// report execution timing event
	defer ga.SendEvent(event.StartTiming("execution"))

	var taskSlice []*boomer.Task
	for _, iTestCase := range testcases {
		testcase, err := iTestCase.ToTestCase()
		if err != nil {
			panic(err)
		}
		task := b.convertBoomerTask(testcase)
		taskSlice = append(taskSlice, task)
	}
	b.Boomer.Run(taskSlice...)
}

// Quit stops running load test.
func (b *hrpBoomer) Quit() {
	b.Boomer.Quit()
}

func (b *hrpBoomer) convertBoomerTask(testcase *TestCase) *boomer.Task {
	config := testcase.Config.ToStruct()
	return &boomer.Task{
		Name:   config.Name,
		Weight: config.Weight,
		Fn: func() {
			runner := NewRunner(nil).SetDebug(b.debug).Reset()
			for _, step := range testcase.TestSteps {
				stepData, err := runner.runStep(step, testcase.Config)

				if stepData.stepType == stepTypeRendezvous {
					// TODO: implement rendezvous in boomer
					continue
				}

				// record transaction
				if stepData.stepType == stepTypeTransaction {
					// TODO: implement recording transaction in boomer
					if stepData.elapsed != 0 {
						b.RecordSuccess(string(stepTypeTransaction), stepData.name, stepData.elapsed, 0)
					}
					continue
				}

				if err == nil {
					b.RecordSuccess(step.Type(), step.Name(), stepData.elapsed, stepData.responseLength)
				} else {
					var elapsed int64
					if stepData != nil {
						elapsed = stepData.elapsed
					}
					b.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())
				}
			}
			endTime := time.Now()

			// report duration for transaction without end
			for name, transaction := range runner.transactions {
				if len(transaction) == 1 {
					// if transaction end time not exists, use testcase end time instead
					duration := endTime.Sub(transaction[TransactionStart])
					b.RecordSuccess(string(stepTypeTransaction), name, duration.Milliseconds(), 0)
				}
			}
		},
	}
}
