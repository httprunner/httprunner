package hrp

import (
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin"
	"github.com/httprunner/hrp/internal/boomer"
	"github.com/httprunner/hrp/internal/ga"
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
			runner := hrpRunner.newCaseRunner(testcase)
			runner.parser.plugin = plugin

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

			if err := runner.parseConfig(caseConfig); err != nil {
				log.Error().Err(err).Msg("parse config failed")
				return
			}

			startTime := time.Now()
			for index, step := range testcase.TestSteps {
				stepData, err := runner.runStep(index, caseConfig)
				if err != nil {
					// step failed
					var elapsed int64
					if stepData != nil {
						elapsed = stepData.Elapsed
					}
					b.RecordFailure(step.Type(), step.Name(), elapsed, err.Error())

					// update flag
					testcaseSuccess = false
					transactionSuccess = false

					if runner.hrpRunner.failfast {
						log.Error().Msg("abort running due to failfast setting")
						break
					}
					log.Warn().Err(err).Msg("run step failed, continue next step")
					continue
				}

				// step success
				if stepData.StepType == stepTypeTransaction {
					// transaction
					// FIXME: support nested transactions
					if step.ToStruct().Transaction.Type == transactionEnd { // only record when transaction ends
						b.RecordTransaction(stepData.Name, transactionSuccess, stepData.Elapsed, 0)
						transactionSuccess = true // reset flag for next transaction
					}
				} else if stepData.StepType == stepTypeRendezvous {
					// rendezvous
					// TODO: implement rendezvous in boomer
				} else {
					// request or testcase step
					b.RecordSuccess(step.Type(), step.Name(), stepData.Elapsed, stepData.ContentSize)
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
