package hrp

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/httprunner/funplugin"
	"github.com/httprunner/httprunner/v4/hrp/internal/boomer"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

func NewStandaloneBoomer(spawnCount int64, spawnRate float64) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		pluginsMutex: new(sync.RWMutex),
	}

	b.hrpRunner = NewRunner(nil)
	return b
}

func NewMasterBoomer(masterBindHost string, masterBindPort int) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewMasterBoomer(masterBindHost, masterBindPort),
		pluginsMutex: new(sync.RWMutex),
	}
	b.hrpRunner = NewRunner(nil)
	return b
}

func NewWorkerBoomer(masterHost string, masterPort int) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewWorkerBoomer(masterHost, masterPort),
		pluginsMutex: new(sync.RWMutex),
	}

	b.hrpRunner = NewRunner(nil)
	// set client transport for high concurrency load testing
	b.hrpRunner.SetClientTransport(b.GetSpawnCount(), b.GetDisableKeepAlive(), b.GetDisableCompression())
	return b
}

type HRPBoomer struct {
	*boomer.Boomer
	hrpRunner    *HRPRunner
	plugins      []funplugin.IPlugin // each task has its own plugin process
	pluginsMutex *sync.RWMutex       // avoid data race
}

func (b *HRPBoomer) InitBoomer() {
	// init output
	if !b.GetProfile().DisableConsoleOutput {
		b.AddOutput(boomer.NewConsoleOutput())
	}
	if b.GetProfile().PrometheusPushgatewayURL != "" {
		b.AddOutput(boomer.NewPrometheusPusherOutput(b.GetProfile().PrometheusPushgatewayURL, "hrp", b.GetMode()))
	}
	b.SetSpawnCount(b.GetProfile().SpawnCount)
	b.SetSpawnRate(b.GetProfile().SpawnRate)
	if b.GetProfile().LoopCount > 0 {
		b.SetLoopCount(b.GetProfile().LoopCount)
	}
	b.SetRateLimiter(b.GetProfile().MaxRPS, b.GetProfile().RequestIncreaseRate)
	b.SetDisableKeepAlive(b.GetProfile().DisableKeepalive)
	b.SetDisableCompression(b.GetProfile().DisableCompression)
	b.SetClientTransport()
	b.EnableCPUProfile(b.GetProfile().CPUProfile, b.GetProfile().CPUProfileDuration)
	b.EnableMemoryProfile(b.GetProfile().MemoryProfile, b.GetProfile().MemoryProfileDuration)
}

func (b *HRPBoomer) SetClientTransport() *HRPBoomer {
	// set client transport for high concurrency load testing
	b.hrpRunner.SetClientTransport(b.GetSpawnCount(), b.GetDisableKeepAlive(), b.GetDisableCompression())
	return b
}

// SetPython3Venv specifies python3 venv.
func (b *HRPBoomer) SetPython3Venv(venv string) *HRPBoomer {
	b.hrpRunner.SetPython3Venv(venv)
	return b
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

	// quit all plugins
	defer func() {
		if len(pluginMap) > 0 {
			for _, plugin := range pluginMap {
				plugin.Quit()
			}
		}
	}()

	taskSlice := b.ConvertTestCasesToBoomerTasks(testcases...)

	b.Boomer.Run(taskSlice...)
}

func (b *HRPBoomer) ConvertTestCasesToBoomerTasks(testcases ...ITestCase) (taskSlice []*boomer.Task) {
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
		waitRendezvous(rendezvousList, b)
	}
	return taskSlice
}

func (b *HRPBoomer) ParseTestCases(testCases []*TestCase) []*TCase {
	var parsedTestCases []*TCase
	for _, tc := range testCases {
		caseRunner, err := b.hrpRunner.newCaseRunner(tc)
		if err != nil {
			log.Error().Err(err).Msg("failed to create runner")
			os.Exit(1)
		}
		caseRunner.parsedConfig.Parameters = caseRunner.parametersIterator.outParameters()
		parsedTestCases = append(parsedTestCases, &TCase{
			Config:    caseRunner.parsedConfig,
			TestSteps: caseRunner.testCase.ToTCase().TestSteps,
		})
	}
	return parsedTestCases
}

func (b *HRPBoomer) TestCasesToBytes(testcases ...ITestCase) []byte {
	// load all testcases
	testCases, err := LoadTestCases(testcases...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		os.Exit(1)
	}
	tcs := b.ParseTestCases(testCases)
	testCasesBytes, err := json.Marshal(tcs)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal testcases")
		return nil
	}
	return testCasesBytes
}

func (b *HRPBoomer) BytesToTestCases(testCasesBytes []byte) []*TCase {
	var testcase []*TCase
	err := json.Unmarshal(testCasesBytes, &testcase)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal testcases")
	}
	return testcase
}

func (b *HRPBoomer) Quit() {
	b.Boomer.Quit()
}

func (b *HRPBoomer) runTestCases(testCases []*TCase, profile *boomer.Profile) {
	var testcases []ITestCase
	for _, tc := range testCases {
		tesecase, err := tc.toTestCase()
		if err != nil {
			log.Error().Err(err).Msg("failed to load testcases")
			return
		}
		// create temp dir to save testcase
		tempDir, err := ioutil.TempDir("", "hrp_testcases")
		if err != nil {
			log.Error().Err(err).Msg("failed to save testcases")
			return
		}

		tesecase.Config.Path = filepath.Join(tempDir, "test-case.json")
		if tesecase.Config.PluginSetting != nil {
			tesecase.Config.PluginSetting.Path = filepath.Join(tempDir, fmt.Sprintf("debugtalk.%s", tesecase.Config.PluginSetting.Type))
			err = builtin.Bytes2File(tesecase.Config.PluginSetting.Content, tesecase.Config.PluginSetting.Path)
			if err != nil {
				log.Error().Err(err).Msg("failed to save plugin file")
				return
			}
		}
		err = builtin.Dump2JSON(tesecase, tesecase.Config.Path)
		if err != nil {
			log.Error().Err(err).Msg("failed to dump testcases")
			return
		}

		testcases = append(testcases, tesecase)
	}

	if profile.PrometheusPushgatewayURL != "" {
		urlSlice := strings.Split(profile.PrometheusPushgatewayURL, ":")
		if len(urlSlice) != 2 {
			profile.PrometheusPushgatewayURL = ""
		} else {
			if urlSlice[0] == "" {
				urlSlice[0] = b.Boomer.GetMasterHost()
			}
		}
		profile.PrometheusPushgatewayURL = strings.Join(urlSlice, ":")
	}

	b.SetProfile(profile)
	b.InitBoomer()
	log.Info().Interface("testcases", testcases).Interface("profile", profile).Msg("run tasks successful")
	b.Run(testcases...)
}

func (b *HRPBoomer) rebalanceBoomer(profile *boomer.Profile) {
	b.SetProfile(profile)
	b.SetSpawnCount(b.GetProfile().SpawnCount)
	b.SetSpawnRate(b.GetProfile().SpawnRate)
	b.GetRebalanceChan() <- true
	log.Info().Interface("profile", profile).Msg("rebalance tasks successful")
}

func (b *HRPBoomer) PollTasks(ctx context.Context) {
	for {
		select {
		case task := <-b.Boomer.GetTasksChan():
			// 清理过时测试用例任务
			if len(b.Boomer.GetTasksChan()) > 0 {
				continue
			}
			//Todo: 过滤掉已经传输过的task
			if task.TestCases != nil {
				testCases := b.BytesToTestCases(task.TestCases)
				go b.runTestCases(testCases, task.Profile)
			} else {
				go b.rebalanceBoomer(task.Profile)
			}

		case <-b.Boomer.GetCloseChan():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (b *HRPBoomer) PollTestCases(ctx context.Context) {
	for {
		select {
		case <-b.Boomer.ParseTestCasesChan():
			var tcs []ITestCase
			for _, tc := range b.GetTestCasesPath() {
				tcp := TestCasePath(tc)
				tcs = append(tcs, &tcp)
			}
			b.TestCaseBytesChan() <- b.TestCasesToBytes(tcs...)
			log.Info().Msg("put testcase successful")
		case <-b.Boomer.GetCloseChan():
			return
		case <-ctx.Done():
			return
		}
	}
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

	// update session variables mutex
	mutex := sync.Mutex{}

	return &boomer.Task{
		Name:   testcase.Config.Name,
		Weight: testcase.Config.Weight,
		Fn: func() {
			testcaseSuccess := true    // flag whole testcase result
			transactionSuccess := true // flag current transaction result

			// init session runner
			sessionRunner := caseRunner.newSession()

			mutex.Lock()
			if parametersIterator.HasNext() {
				sessionRunner.updateSessionVariables(parametersIterator.Next())
			}
			mutex.Unlock()

			startTime := time.Now()
			for _, step := range testcase.TestSteps {
				// parse step name
				parsedName, err := sessionRunner.parser.ParseString(step.Name(), sessionRunner.sessionVariables)
				if err != nil {
					parsedName = step.Name()
				}
				stepName := convertString(parsedName)
				// reset start time only once before step
				once.Do(func() {
					b.Boomer.ResetStartTime()
				})
				stepResult, err := step.Run(sessionRunner)
				// update step result name with parsed step name
				stepResult.Name = stepName
				// record requests result of the step if step type is testcase
				if stepResult.StepType == stepTypeTestCase && stepResult.Data != nil {
					// record requests of testcase step
					for _, result := range stepResult.Data.([]*StepResult) {
						if result.Success {
							b.RecordSuccess(string(result.StepType), result.Name, result.Elapsed, result.ContentSize)
						} else {
							b.RecordFailure(string(result.StepType), result.Name, result.Elapsed, result.Attachment)
						}
					}
				}
				// record step failure
				if err != nil {
					// step failed
					var elapsed int64
					if stepResult != nil {
						elapsed = stepResult.Elapsed
					}
					b.RecordFailure(string(step.Type()), stepResult.Name, elapsed, err.Error())

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

				// record step success
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
					b.RecordSuccess(string(step.Type()), stepResult.Name, stepResult.Elapsed, stepResult.ContentSize)
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
