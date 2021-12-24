package hrp

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/hrp/internal/boomer"
	"github.com/httprunner/hrp/internal/ga"
)

func NewBoomer(spawnCount int, spawnRate float64) *hrpBoomer {
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

func (b *hrpBoomer) convertBoomerTask(testcase *TestCase) *boomer.Task {
	config := testcase.Config.ToStruct()
	return &boomer.Task{
		Name:   config.Name,
		Weight: config.Weight,
		Fn: func() {
			runner := NewRunner(nil).SetDebug(b.debug).Reset()

			testcaseSuccess := true       // flag whole testcase result
			var transactionSuccess = true // flag current transaction result

			startTime := time.Now()
			for _, step := range testcase.TestSteps {
				stepData, err := runner.runStep(step, testcase.Config)
				if err != nil {
					// step failed
					var elapsed int64
					if stepData != nil {
						elapsed = stepData.elapsed
					}
					b.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())

					// update flag
					testcaseSuccess = false
					transactionSuccess = false

					if runner.failfast {
						log.Error().Err(err).Msg("abort running due to failfast setting")
						break
					}
					log.Warn().Err(err).Msg("run step failed, continue next step")
					continue
				}

				// step success
				if stepData.stepType == stepTypeTransaction {
					// transaction
					// FIXME: support nested transactions
					if stepData.elapsed != 0 { // only record when transaction ends
						b.RecordTransaction(stepData.name, transactionSuccess, stepData.elapsed, 0)
						transactionSuccess = true // reset flag for next transaction
					}
				} else if stepData.stepType == stepTypeRendezvous {
					// rendezvous
					// TODO: implement rendezvous in boomer
				} else {
					// request or testcase step
					b.RecordSuccess(step.Type(), step.Name(), stepData.elapsed, stepData.contentSize)
				}
			}
			endTime := time.Now()

			// report duration for transaction without end
			for name, transaction := range runner.transactions {
				if len(transaction) == 1 {
					// if transaction end time not exists, use testcase end time instead
					duration := endTime.Sub(transaction[transactionStart])
					b.RecordTransaction(name, transactionSuccess, duration.Milliseconds(), 0)
				}
			}

			// report testcase as a whole Action transaction, inspired by LoadRunner
			b.RecordTransaction("Action", testcaseSuccess, endTime.Sub(startTime).Milliseconds(), 0)
		},
	}
}
