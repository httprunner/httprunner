package hrp

import (
	"os"
	"sync"
	"time"

	"github.com/httprunner/funplugin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/boomer"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
)

func NewBoomer(spawnCount int, spawnRate float64) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		pluginsMutex: new(sync.RWMutex),
	}

	b.hrpRunner = NewRunner(nil)
	return b
}

type HRPBoomer struct {
	*boomer.Boomer
	hrpRunner    *HRPRunner
	plugins      []funplugin.IPlugin // each task has its own plugin process
	pluginsMutex *sync.RWMutex       // avoid data race
}

func (b *HRPBoomer) SetClientTransport() {
	// set client transport for high concurrency load testing
	b.hrpRunner.SetClientTransport(b.GetSpawnCount(), b.GetDisableKeepAlive(), b.GetDisableCompression())
}

// Run starts to run load test for one or multiple testcases.
func (b *HRPBoomer) Run(testcases ...ITestCase) {
	event := sdk.EventTracking{
		Category: "RunLoadTests",
		Action:   "hrp boom",
	}
	// report start event
	go sdk.SendEvent(event)
	// report execution timing event
	defer sdk.SendEvent(event.StartTiming("execution"))

	var taskSlice []*boomer.Task

	// load all testcases
	testCases, err := LoadTestCases(testcases...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		os.Exit(1)
	}

	for _, testcase := range testCases {
		rendezvousList := initRendezvous(testcase, int64(b.GetSpawnCount()))
		task := b.convertBoomerTask(testcase, rendezvousList)
		taskSlice = append(taskSlice, task)
		waitRendezvous(rendezvousList)
	}
	b.Boomer.Run(taskSlice...)
}

func (b *HRPBoomer) Quit() {
	b.pluginsMutex.Lock()
	plugins := b.plugins
	b.pluginsMutex.Unlock()
	for _, plugin := range plugins {
		plugin.Quit()
	}
	b.Boomer.Quit()
}

func (b *HRPBoomer) convertBoomerTask(testcase *TestCase, rendezvousList []*Rendezvous) *boomer.Task {
	// init runner for testcase
	// this runner is shared by multiple session runners
	caseRunner, err := b.hrpRunner.newCaseRunner(testcase)
	if err != nil {
		log.Error().Err(err).Msg("failed to create runner")
		os.Exit(1)
	}
	if caseRunner.parser.plugin != nil {
		b.pluginsMutex.Lock()
		b.plugins = append(b.plugins, caseRunner.parser.plugin)
		b.pluginsMutex.Unlock()
	}

	// broadcast to all rendezvous at once when spawn done
	go func() {
		<-b.GetSpawnDoneChan()
		for _, rendezvous := range rendezvousList {
			rendezvous.setSpawnDone()
		}
	}()

	// set paramters mode for load testing
	parametersIterator := caseRunner.parametersIterator
	parametersIterator.SetUnlimitedMode()

	// reset start time only once
	once := sync.Once{}

	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			testcaseSuccess := true    // flag whole testcase result
			transactionSuccess := true // flag current transaction result

			// init session runner
			sessionRunner := caseRunner.newSession()

			if parametersIterator.HasNext() {
				sessionRunner.updateSessionVariables(parametersIterator.Next())
			}

			startTime := time.Now()
			for _, step := range testcase.TestSteps {
				// reset start time only once before step
				once.Do(func() {
					b.Boomer.ResetStartTime()
				})
				stepResult, err := step.Run(sessionRunner)
				if err != nil {
					// step failed
					var elapsed int64
					if stepResult != nil {
						elapsed = stepResult.Elapsed
					}
					b.RecordFailure(string(step.Type()), step.Name(), elapsed, err.Error())

					// update flag
					testcaseSuccess = false
					transactionSuccess = false

					if b.hrpRunner.failfast {
						log.Error().Err(err).Msg("abort running due to failfast setting")
						break
					}
					log.Warn().Err(err).Msg("run step failed, continue next step")
					continue
				}

				// step success
				if stepResult.StepType == stepTypeTransaction {
					// transaction
					// FIXME: support nested transactions
					if step.Struct().Transaction.Type == transactionEnd { // only record when transaction ends
						b.RecordTransaction(stepResult.Name, transactionSuccess, stepResult.Elapsed, 0)
						transactionSuccess = true // reset flag for next transaction
					}
				} else if stepResult.StepType == stepTypeRendezvous {
					// rendezvous
				} else if stepResult.StepType == stepTypeThinkTime {
					// think time
					// no record required
				} else {
					// request or testcase step
					b.RecordSuccess(string(step.Type()), step.Name(), stepResult.Elapsed, stepResult.ContentSize)
					// update extracted variables
					for k, v := range stepResult.ExportVars {
						sessionRunner.sessionVariables[k] = v
					}
				}
			}
			endTime := time.Now()

			// report duration for transaction without end
			for name, transaction := range sessionRunner.transactions {
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
