package hrpboomer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/httprunner/funplugin"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
	"github.com/httprunner/httprunner/v4/hrp/internal/sdk"
	"github.com/httprunner/httprunner/v4/hrp/pkg/boomer"
)

var pluginMap sync.Map // used for reusing plugin instance

func NewStandaloneBoomer(spawnCount int64, spawnRate float64) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewStandaloneBoomer(spawnCount, spawnRate),
		pluginsMutex: new(sync.RWMutex),
	}

	b.hrpRunner = hrp.NewRunner(nil)
	return b
}

func NewMasterBoomer(masterBindHost string, masterBindPort int) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewMasterBoomer(masterBindHost, masterBindPort),
		pluginsMutex: new(sync.RWMutex),
	}
	b.hrpRunner = hrp.NewRunner(nil)
	return b
}

func NewWorkerBoomer(masterHost string, masterPort int) *HRPBoomer {
	b := &HRPBoomer{
		Boomer:       boomer.NewWorkerBoomer(masterHost, masterPort),
		pluginsMutex: new(sync.RWMutex),
	}

	b.hrpRunner = hrp.NewRunner(nil)
	// set client transport for high concurrency load testing
	b.hrpRunner.SetClientTransport(b.GetSpawnCount(), b.GetDisableKeepAlive(), b.GetDisableCompression())
	return b
}

type HRPBoomer struct {
	*boomer.Boomer
	hrpRunner    *hrp.HRPRunner
	plugins      []funplugin.IPlugin // each task has its own plugin process
	pluginsMutex *sync.RWMutex       // avoid data race
}

func (b *HRPBoomer) InitBoomer() {
	if !b.GetProfile().DisableConsoleOutput {
		b.AddOutput(boomer.NewConsoleOutput())
	}
	if b.GetProfile().PrometheusPushgatewayURL != "" {
		b.AddOutput(boomer.NewPrometheusPusherOutput(b.GetProfile().PrometheusPushgatewayURL, "hrp", b.GetMode()))
	}
	b.SetSpawnCount(b.GetProfile().SpawnCount)
	b.SetSpawnRate(b.GetProfile().SpawnRate)
	b.SetRunTime(b.GetProfile().RunTime)
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
func (b *HRPBoomer) Run(testcases ...hrp.ITestCase) {
	startTime := time.Now()
	defer func() {
		// report boom event
		sdk.SendGA4Event("hrp_boomer_run", map[string]interface{}{
			"engagement_time_msec": time.Since(startTime).Milliseconds(),
		})

		// quit all plugins
		pluginMap.Range(func(key, value interface{}) bool {
			if plugin, ok := value.(funplugin.IPlugin); ok {
				plugin.Quit()
			}
			return true
		})
	}()

	taskSlice := b.ConvertTestCasesToBoomerTasks(testcases...)

	b.Boomer.Run(taskSlice...)
}

func (b *HRPBoomer) ConvertTestCasesToBoomerTasks(testcases ...hrp.ITestCase) (taskSlice []*boomer.Task) {
	// load all testcases
	testCases, err := hrp.LoadTestCases(testcases...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		os.Exit(code.GetErrorCode(err))
	}

	for _, testcase := range testCases {
		rendezvousList := hrp.InitRendezvous(testcase, int64(b.GetSpawnCount()))
		task := b.convertBoomerTask(testcase, rendezvousList)
		taskSlice = append(taskSlice, task)
		hrp.WaitRendezvous(rendezvousList, b)
	}
	return taskSlice
}

func (b *HRPBoomer) ParseTestCases(testCases []*hrp.TestCase) []*hrp.TestCase {
	var parsedTestCases []*hrp.TestCase
	for _, tc := range testCases {
		caseRunner, err := b.hrpRunner.NewCaseRunner(*tc)
		if err != nil {
			log.Error().Err(err).Msg("failed to create runner")
			os.Exit(code.GetErrorCode(err))
		}
		caseConfig := caseRunner.TestCase.Config.Get()
		caseConfig.Parameters = caseRunner.GetParametersIterator().Data()
		parsedTestCases = append(parsedTestCases, &hrp.TestCase{
			Config:    caseConfig,
			TestSteps: caseRunner.TestSteps,
		})
	}
	return parsedTestCases
}

func (b *HRPBoomer) TestCasesToBytes(testcases ...hrp.ITestCase) []byte {
	// load all testcases
	testCases, err := hrp.LoadTestCases(testcases...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		os.Exit(code.GetErrorCode(err))
	}
	testCasesBytes, err := json.Marshal(testCases)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal testcases")
		return nil
	}
	return testCasesBytes
}

func (b *HRPBoomer) BytesToTCases(testCasesBytes []byte) []*hrp.TestCase {
	var testcase []*hrp.TestCase
	err := json.Unmarshal(testCasesBytes, &testcase)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal testcases")
	}
	return testcase
}

func (b *HRPBoomer) Quit() {
	b.Boomer.Quit()
}

func (b *HRPBoomer) parseTCases(testCases []*hrp.TestCase) (testcases []hrp.ITestCase) {
	for _, tc := range testCases {
		// create temp dir to save testcase
		tempDir, err := os.MkdirTemp("", "hrp_testcases")
		if err != nil {
			log.Error().Err(err).Msg("failed to create hrp testcases directory")
			return
		}

		caseConfig := tc.Config.Get()
		if caseConfig.PluginSetting != nil {
			caseConfig.PluginSetting.Path = filepath.Join(tempDir, fmt.Sprintf("debugtalk.%s", caseConfig.PluginSetting.Type))
			err = builtin.Bytes2File(caseConfig.PluginSetting.Content, caseConfig.PluginSetting.Path)
			if err != nil {
				log.Error().Err(err).Msg("failed to save plugin file")
				return
			}
			caseConfig.PluginSetting.Content = nil // remove the content in testcase
		}

		if caseConfig.Environs != nil {
			envContent := ""
			for k, v := range caseConfig.Environs {
				envContent += fmt.Sprintf("%s=%s\n", k, v)
			}
			err = os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0o644)
			if err != nil {
				log.Error().Err(err).Msg("failed to dump environs")
				return
			}
		}

		caseConfig.Path = filepath.Join(tempDir, "test-case.json")
		err = builtin.Dump2JSON(tc, caseConfig.Path)
		if err != nil {
			log.Error().Err(err).Msg("failed to dump testcases")
			return
		}

		testcases = append(testcases, tc)
	}
	return testcases
}

func (b *HRPBoomer) initWorker(profile *boomer.Profile) {
	// if no IP address is specified, the default IP address is that of the master
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
}

func (b *HRPBoomer) rebalanceRunner(profile *boomer.Profile) {
	b.SetSpawnCount(profile.SpawnCount)
	b.SetSpawnRate(profile.SpawnRate)
	b.GetRebalanceChan() <- true
	log.Info().Interface("profile", profile).Msg("rebalance tasks successfully")
}

func (b *HRPBoomer) PollTasks(ctx context.Context) {
	for {
		select {
		case task := <-b.Boomer.GetTasksChan():
			// 清理过时测试用例任务
			if len(b.Boomer.GetTasksChan()) > 0 {
				continue
			}
			// Todo: 过滤掉已经传输过的task
			if task.TestCasesBytes != nil {
				// init boomer with profile
				b.initWorker(task.Profile)
				// get testcases
				testcases := b.parseTCases(b.BytesToTCases(task.TestCasesBytes))
				log.Info().Interface("testcases", testcases).Interface("profile", b.GetProfile()).Msg("starting to run tasks")
				// run testcases
				go b.Run(testcases...)
			} else {
				// rebalance runner with profile
				go b.rebalanceRunner(task.Profile)
			}

		case <-b.Boomer.GetCloseChan():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (b *HRPBoomer) PollTestCases(ctx context.Context) {
	// quit all plugins
	defer func() {
		pluginMap.Range(func(key, value interface{}) bool {
			if plugin, ok := value.(funplugin.IPlugin); ok {
				plugin.Quit()
			}
			return true
		})
	}()

	for {
		select {
		case <-b.Boomer.ParseTestCasesChan():
			var tcs []hrp.ITestCase
			for _, tc := range b.GetTestCasesPath() {
				tcp := hrp.TestCasePath(tc)
				tcs = append(tcs, &tcp)
			}
			b.TestCaseBytesChan() <- b.TestCasesToBytes(tcs...)
			log.Info().Msg("put testcase successfully")
		case <-b.Boomer.GetCloseChan():
			return
		case <-ctx.Done():
			return
		}
	}
}

func (b *HRPBoomer) convertBoomerTask(testcase *hrp.TestCase, rendezvousList []*hrp.Rendezvous) *boomer.Task {
	// init case runner for testcase
	// this runner is shared by multiple session runners
	caseRunner, err := b.hrpRunner.NewCaseRunner(*testcase)
	if err != nil {
		log.Error().Err(err).Msg("failed to create runner")
		os.Exit(code.GetErrorCode(err))
	}
	plugin := caseRunner.GetParser().Plugin
	if plugin != nil {
		b.pluginsMutex.Lock()
		b.plugins = append(b.plugins, plugin)
		b.pluginsMutex.Unlock()
	}

	// broadcast to all rendezvous at once when spawn done
	go func() {
		<-b.GetSpawnDoneChan()
		for _, rendezvous := range rendezvousList {
			rendezvous.SetSpawnDone()
		}
	}()

	// set paramters mode for load testing
	parametersIterator := caseRunner.GetParametersIterator()
	parametersIterator.SetUnlimitedMode()

	// reset start time only once
	once := sync.Once{}

	// update session variables mutex
	mutex := sync.Mutex{}

	return &boomer.Task{
		Name:   testcase.Config.Get().Name,
		Weight: testcase.Config.Get().Weight,
		Fn: func() {
			testcaseSuccess := true    // flag whole testcase result
			transactionSuccess := true // flag current transaction result

			// init session runner
			sessionRunner := caseRunner.NewSession()

			mutex.Lock()
			if parametersIterator.HasNext() {
				sessionRunner.InitWithParameters(parametersIterator.Next())
			}
			mutex.Unlock()

			defer func() {
				sessionRunner.ReleaseResources()
			}()

			startTime := time.Now()
			for _, step := range testcase.TestSteps {
				// parse step struct
				err = sessionRunner.ParseStep(step)
				if err != nil {
					log.Error().Err(err).Msg("parse step struct failed")
				}

				// reset start time only once before step
				once.Do(func() {
					b.Boomer.ResetStartTime()
				})
				stepResult, err := step.Run(sessionRunner)
				// update step result name with parsed step name
				stepResult.Name = step.Name()
				// record requests result of the step if step type is testcase
				if stepResult.StepType == hrp.StepTypeTestCase && stepResult.Data != nil {
					// record requests of testcase step
					for _, result := range stepResult.Data.([]*hrp.StepResult) {
						if result.Success {
							b.RecordSuccess(string(result.StepType), result.Name, result.Elapsed, result.ContentSize)
						} else {
							exception, _ := result.Attachments.(string)
							b.RecordFailure(string(result.StepType), result.Name, result.Elapsed, exception)
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

					log.Warn().Err(err).Msg("run step failed, continue next step")
					continue
				}

				// record step success
				if stepResult.StepType == hrp.StepTypeTransaction {
					// transaction
					// FIXME: support nested transactions
					stepTransaction := step.(*hrp.StepTransaction)
					if stepTransaction.Transaction.Type == hrp.TransactionEnd { // only record when transaction ends
						b.RecordTransaction(stepTransaction.Name(), transactionSuccess, stepResult.Elapsed, 0)
						transactionSuccess = true // reset flag for next transaction
					}
				} else if stepResult.StepType == hrp.StepTypeRendezvous {
					// rendezvous
				} else if stepResult.StepType == hrp.StepTypeThinkTime {
					// think time
					// no record required
				} else {
					// request or testcase step
					b.RecordSuccess(string(step.Type()), stepResult.Name, stepResult.Elapsed, stepResult.ContentSize)
					// update extracted variables
					for k, v := range stepResult.ExportVars {
						sessionRunner.GetSessionVariables()[k] = v
					}
				}
			}
			endTime := time.Now()

			// report duration for transaction without end
			for name, transaction := range sessionRunner.GetTransactions() {
				if len(transaction) == 1 {
					// if transaction end time not exists, use testcase end time instead
					duration := endTime.Sub(transaction[hrp.TransactionStart])
					b.RecordTransaction(name, transactionSuccess, duration.Milliseconds(), 0)
				}
			}

			// report testcase as a whole Action transaction, inspired by LoadRunner
			b.RecordTransaction("Action", testcaseSuccess, endTime.Sub(startTime).Milliseconds(), 0)
		},
	}
}
