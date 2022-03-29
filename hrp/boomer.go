package hrp

import (
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin"
	"github.com/httprunner/httprunner/hrp/internal/boomer"
	"github.com/httprunner/httprunner/hrp/internal/sdk"
)

func NewBoomer(spawnCount int, spawnRate float64) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		pluginsMutex: new(sync.RWMutex),
	}
	return b
}

type HRPBoomer struct {
	*boomer.Boomer
	plugins      []funplugin.IPlugin // each task has its own plugin process
	pluginsMutex *sync.RWMutex       // avoid data race
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
	testCases, err := loadTestCases(testcases...)
	if err != nil {
		panic(err)
	}

	for _, testcase := range testCases {
		cfg := testcase.Config
		err = initParameterIterator(cfg, "boomer")
		if err != nil {
			panic(err)
		}
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
	hrpRunner := NewRunner(nil)
	// set client transport for high concurrency load testing
	hrpRunner.SetClientTransport(b.GetSpawnCount(), b.GetDisableKeepAlive(), b.GetDisableCompression())
	config := testcase.Config

	// each testcase has its own plugin process
	plugin, _ := initPlugin(config.Path, false)
	if plugin != nil {
		b.pluginsMutex.Lock()
		b.plugins = append(b.plugins, plugin)
		b.pluginsMutex.Unlock()
	}

	// broadcast to all rendezvous at once when spawn done
	go func() {
		<-b.GetSpawnDoneChan()
		for _, rendezvous := range rendezvousList {
			rendezvous.setSpawnDone()
		}
	}()

	return &boomer.Task{
		Name:   config.Name,
		Weight: config.Weight,
		Fn: func() {
			sessionRunner := hrpRunner.NewSessionRunner(testcase)
			sessionRunner.init()
			sessionRunner.parser.plugin = plugin

			testcaseSuccess := true       // flag whole testcase result
			var transactionSuccess = true // flag current transaction result

			cfg := testcase.Config
			caseConfig := &TConfig{}
			// copy config to avoid data racing
			if err := copier.Copy(caseConfig, cfg); err != nil {
				log.Error().Err(err).Msg("copy config data failed")
				return
			}
			// iterate through all parameter iterators and update case variables
			for _, it := range caseConfig.ParametersSetting.Iterators {
				if it.HasNext() {
					caseConfig.Variables = mergeVariables(it.Next(), caseConfig.Variables)
				}
			}

			if err := sessionRunner.parseConfig(caseConfig); err != nil {
				log.Error().Err(err).Msg("parse config failed")
				return
			}

			startTime := time.Now()
			for _, step := range testcase.TestSteps {
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

					if sessionRunner.hrpRunner.failfast {
						log.Error().Msg("abort running due to failfast setting")
						break
					}
					log.Warn().Err(err).Msg("run step failed, continue next step")
					continue
				}

				// step success
				if stepResult.StepType == stepTypeTransaction {
					// transaction
					// FIXME: support nested transactions
					if step.ToStruct().Transaction.Type == transactionEnd { // only record when transaction ends
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
