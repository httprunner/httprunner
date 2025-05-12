package hrp

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"

	"github.com/httprunner/funplugin"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// Run starts to run testcase with default configs.
func Run(t *testing.T, testcases ...ITestCase) error {
	err := NewRunner(t).SetSaveTests(true).Run(testcases...)
	code.GetErrorCode(err)
	return err
}

// NewRunner constructs a new runner instance.
func NewRunner(t *testing.T) *HRPRunner {
	if t == nil {
		t = &testing.T{}
	}
	jar, _ := cookiejar.New(nil)
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGTERM, syscall.SIGINT)
	return &HRPRunner{
		t:             t,
		failfast:      true, // default to failfast
		genHTMLReport: false,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar:     jar, // insert response cookies into request
			Timeout: 120 * time.Second,
		},
		http2Client: &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar:     jar, // insert response cookies into request
			Timeout: 120 * time.Second,
		},
		// use default handshake timeout (no timeout limit) here, enable timeout at step level
		wsDialer: &websocket.Dialer{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		caseTimeoutTimer: time.NewTimer(time.Hour * 2), // default case timeout to 2 hour
		interruptSignal:  interruptSignal,
	}
}

type HRPRunner struct {
	t                *testing.T
	failfast         bool
	httpStatOn       bool
	requestsLogOn    bool
	pluginLogOn      bool
	venv             string
	saveTests        bool
	genHTMLReport    bool
	httpClient       *http.Client
	http2Client      *http.Client
	wsDialer         *websocket.Dialer
	caseTimeoutTimer *time.Timer    // case timeout timer
	interruptSignal  chan os.Signal // interrupt signal channel
}

// SetClientTransport configures transport of http client for high concurrency load testing
func (r *HRPRunner) SetClientTransport(maxConns int, disableKeepAlive bool, disableCompression bool) *HRPRunner {
	log.Info().
		Int("maxConns", maxConns).
		Bool("disableKeepAlive", disableKeepAlive).
		Bool("disableCompression", disableCompression).
		Msg("[init] SetClientTransport")
	r.httpClient.Transport = &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DialContext:         (&net.Dialer{}).DialContext,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: maxConns,
		DisableKeepAlives:   disableKeepAlive,
		DisableCompression:  disableCompression,
	}
	r.http2Client.Transport = &http2.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: disableCompression,
	}
	r.wsDialer.EnableCompression = !disableCompression
	return r
}

// SetFailfast configures whether to stop running when one step fails.
func (r *HRPRunner) SetFailfast(failfast bool) *HRPRunner {
	log.Info().Bool("failfast", failfast).Msg("[init] SetFailfast")
	r.failfast = failfast
	return r
}

// SetRequestsLogOn turns on request & response details logging.
func (r *HRPRunner) SetRequestsLogOn() *HRPRunner {
	log.Info().Msg("[init] SetRequestsLogOn")
	r.requestsLogOn = true
	return r
}

// SetHTTPStatOn turns on HTTP latency stat.
func (r *HRPRunner) SetHTTPStatOn() *HRPRunner {
	log.Info().Msg("[init] SetHTTPStatOn")
	r.httpStatOn = true
	return r
}

// SetPluginLogOn turns on plugin logging.
func (r *HRPRunner) SetPluginLogOn() *HRPRunner {
	log.Info().Msg("[init] SetPluginLogOn")
	r.pluginLogOn = true
	return r
}

// SetPython3Venv specifies python3 venv.
func (r *HRPRunner) SetPython3Venv(venv string) *HRPRunner {
	log.Info().Str("venv", venv).Msg("[init] SetPython3Venv")
	r.venv = venv
	return r
}

// SetProxyUrl configures the proxy URL, which is usually used to capture HTTP packets for debugging.
func (r *HRPRunner) SetProxyUrl(proxyUrl string) *HRPRunner {
	log.Info().Str("proxyUrl", proxyUrl).Msg("[init] SetProxyUrl")
	p, err := url.Parse(proxyUrl)
	if err != nil {
		log.Error().Err(err).Str("proxyUrl", proxyUrl).Msg("[init] invalid proxyUrl")
		return r
	}
	r.httpClient.Transport = &http.Transport{
		Proxy:           http.ProxyURL(p),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	r.wsDialer.Proxy = http.ProxyURL(p)
	return r
}

// SetRequestTimeout configures global request timeout in seconds.
func (r *HRPRunner) SetRequestTimeout(seconds float32) *HRPRunner {
	log.Info().Float32("timeout_seconds", seconds).Msg("[init] SetRequestTimeout")
	r.httpClient.Timeout = time.Duration(seconds*1000) * time.Millisecond
	return r
}

// SetCaseTimeout configures global testcase timeout in seconds.
func (r *HRPRunner) SetCaseTimeout(seconds float32) *HRPRunner {
	log.Info().Float32("timeout_seconds", seconds).Msg("[init] SetCaseTimeout")
	r.caseTimeoutTimer = time.NewTimer(time.Duration(seconds*1000) * time.Millisecond)
	return r
}

// SetSaveTests configures whether to save summary of tests.
func (r *HRPRunner) SetSaveTests(saveTests bool) *HRPRunner {
	log.Info().Bool("saveTests", saveTests).Msg("[init] SetSaveTests")
	r.saveTests = saveTests
	return r
}

// GenHTMLReport configures whether to gen html report of api tests.
func (r *HRPRunner) GenHTMLReport() *HRPRunner {
	log.Info().Bool("genHTMLReport", true).Msg("[init] SetgenHTMLReport")
	r.genHTMLReport = true
	return r
}

// Run starts to execute one or multiple testcases.
func (r *HRPRunner) Run(testcases ...ITestCase) (err error) {
	log.Info().Str("hrp_version", version.VERSION).Msg("start running")

	startTime := time.Now()
	defer func() {
		// report run event
		sdk.SendGA4Event("hrp_run", map[string]interface{}{
			"success":              err == nil,
			"engagement_time_msec": time.Since(startTime).Milliseconds(),
		})
	}()

	// record execution data to summary
	s := NewSummary()

	// load all testcases
	testCases, err := LoadTestCases(testcases...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		return err
	}

	// quit all plugins
	defer func() {
		pluginMap.Range(func(key, value interface{}) bool {
			if plugin, ok := value.(funplugin.IPlugin); ok {
				plugin.Quit()
			}
			return true
		})
	}()

	var runErr error
	// run testcase one by one
	for _, testcase := range testCases {
		// each testcase has its own case runner
		caseRunner, err := NewCaseRunner(*testcase, r)
		if err != nil {
			log.Error().Err(err).Msg("[Run] init case runner failed")
			return err
		}

		// release UI driver session
		defer func() {
			for _, client := range caseRunner.uixtDrivers {
				client.DeleteSession()
			}
		}()

		for it := caseRunner.parametersIterator; it.HasNext(); {
			// case runner can run multiple times with different parameters
			// each run has its own session runner
			sessionRunner := caseRunner.NewSession()
			caseSummary, err := sessionRunner.Start(it.Next())
			s.AddCaseSummary(caseSummary)
			if err != nil {
				log.Error().Err(err).Msg("[Run] run testcase failed")
				runErr = err
			}

			if runErr != nil && r.failfast {
				break
			}
		}
	}
	s.Time.Duration = time.Since(s.Time.StartAt).Seconds()

	// save summary
	if r.saveTests {
		if _, err := s.GenSummary(); err != nil {
			return err
		}
	}

	// generate HTML report
	if r.genHTMLReport {
		if err := s.GenHTMLReport(); err != nil {
			return err
		}
	}

	return runErr
}

// NewCaseRunner creates a new case runner for testcase.
// each testcase has its own case runner
// If the provided hrpRunner is nil, a default HRPRunner will be created and used.
func NewCaseRunner(testcase TestCase, hrpRunner *HRPRunner) (*CaseRunner, error) {
	if hrpRunner == nil {
		hrpRunner = NewRunner(nil)
	}
	caseRunner := &CaseRunner{
		TestCase:    testcase,
		hrpRunner:   hrpRunner,
		parser:      NewParser(),
		uixtDrivers: make(map[string]*uixt.XTDriver),
	}
	config := testcase.Config.Get()

	// init parser plugin
	if config.PluginSetting != nil {
		plugin, err := initPlugin(config.Path, hrpRunner.venv, hrpRunner.pluginLogOn)
		if err != nil {
			return nil, errors.Wrap(err, "init plugin failed")
		}
		caseRunner.parser.Plugin = plugin

		// load plugin info to testcase config
		pluginPath := plugin.Path()
		pluginContent, err := builtin.LoadFile(pluginPath)
		if err != nil {
			return nil, err
		}
		config.PluginSetting.Path = pluginPath
		config.PluginSetting.Content = pluginContent
		tp := strings.Split(pluginPath, ".")
		config.PluginSetting.Type = tp[len(tp)-1]
		log.Info().Str("pluginPath", pluginPath).
			Str("pluginType", config.PluginSetting.Type).
			Msg("plugin info loaded")
	}

	// parse testcase config
	parsedConfig, err := caseRunner.parseConfig()
	if err != nil {
		return nil, errors.Wrap(err, "parse testcase config failed")
	}

	// set request timeout in seconds
	if parsedConfig.RequestTimeout != 0 {
		hrpRunner.SetRequestTimeout(parsedConfig.RequestTimeout)
	}
	// set testcase timeout in seconds
	if parsedConfig.CaseTimeout != 0 {
		hrpRunner.SetCaseTimeout(parsedConfig.CaseTimeout)
	}

	caseRunner.TestCase.Config = parsedConfig
	return caseRunner, nil
}

type CaseRunner struct {
	TestCase // each testcase init its own CaseRunner

	hrpRunner *HRPRunner // all case runners share one HRPRunner
	parser    *Parser    // each CaseRunner init its own Parser

	parametersIterator *ParametersIterator

	// UI automation clients for iOS and Android, key is udid/serial
	uixtDrivers map[string]*uixt.XTDriver
}

func (r *CaseRunner) GetParametersIterator() *ParametersIterator {
	return r.parametersIterator
}

func (r *CaseRunner) GetParser() *Parser {
	return r.parser
}

// parseConfig parses testcase config, stores to parsedConfig.
func (r *CaseRunner) parseConfig() (parsedConfig *TConfig, err error) {
	cfg := r.TestCase.Config.Get()

	parsedConfig = &TConfig{}
	// deep copy config to avoid data racing
	if err := copier.Copy(parsedConfig, cfg); err != nil {
		log.Error().Err(err).Msg("copy testcase config failed")
		return nil, err
	}

	// parse config variables
	parsedVariables, err := r.parser.ParseVariables(cfg.Variables)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return nil, err
	}
	parsedConfig.Variables = parsedVariables

	// parse config name
	parsedName, err := r.parser.ParseString(cfg.Name, parsedVariables)
	if err != nil {
		return nil, errors.Wrap(err, "parse config name failed")
	}
	parsedConfig.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := r.parser.ParseString(cfg.BaseURL, parsedVariables)
	if err != nil {
		return nil, errors.Wrap(err, "parse config base url failed")
	}
	parsedConfig.BaseURL = convertString(parsedBaseURL)

	// merge config environment variables with base_url
	// priority: env base_url > base_url
	if cfg.Environs != nil {
		parsedConfig.Environs = cfg.Environs
	} else {
		parsedConfig.Environs = make(map[string]string)
	}
	if value, ok := parsedConfig.Environs["base_url"]; !ok || value == "" {
		if parsedConfig.BaseURL != "" {
			parsedConfig.Environs["base_url"] = parsedConfig.BaseURL
		}
	}

	// merge config variables with environment variables
	// priority: env > config variables
	for k, v := range parsedConfig.Environs {
		parsedConfig.Variables[k] = v
	}

	// ensure correction of think time config
	parsedConfig.ThinkTimeSetting.checkThinkTime()

	// ensure correction of websocket config
	parsedConfig.WebSocketSetting.checkWebSocket()

	// parse testcase config parameters
	parametersIterator, err := r.parser.InitParametersIterator(parsedConfig)
	if err != nil {
		log.Error().Err(err).
			Interface("parameters", parsedConfig.Parameters).
			Interface("parametersSetting", parsedConfig.ParametersSetting).
			Msg("parse config parameters failed")
		return nil, errors.Wrap(err, "parse testcase config parameters failed")
	}
	r.parametersIterator = parametersIterator

	// ai options
	aiOpts := []option.AIServiceOption{}
	if parsedConfig.LLMService != "" {
		aiOpts = append(aiOpts, option.WithLLMService(option.LLMServiceType(parsedConfig.LLMService)))
	}
	if parsedConfig.CVService == "" {
		// default to vedem
		parsedConfig.CVService = option.CVServiceTypeVEDEM
	}
	aiOpts = append(aiOpts, option.WithCVService(parsedConfig.CVService))

	// parse android devices config
	for _, androidDeviceOptions := range parsedConfig.Android {
		err := r.parseDeviceConfig(androidDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse android config failed: %v", err))
		}

		device, err := uixt.NewAndroidDevice(androidDeviceOptions.Options()...)
		if err != nil {
			return nil, errors.Wrap(err, "init android device failed")
		}
		driver, err := device.NewDriver()
		if err != nil {
			return nil, errors.Wrap(err, "init android driver failed")
		}

		driverExt, err := uixt.NewXTDriver(driver, aiOpts...)
		if err != nil {
			return nil, errors.Wrap(err, "init android XTDriver failed")
		}
		r.RegisterUIXTDriver(androidDeviceOptions.SerialNumber, driverExt)
	}
	// parse iOS devices config
	for _, iosDeviceOptions := range parsedConfig.IOS {
		err := r.parseDeviceConfig(iosDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse ios config failed: %v", err))
		}

		device, err := uixt.NewIOSDevice(iosDeviceOptions.Options()...)
		if err != nil {
			return nil, errors.Wrap(err, "init ios device failed")
		}
		driver, err := device.NewDriver()
		if err != nil {
			return nil, errors.Wrap(err, "init ios driver failed")
		}

		driverExt, err := uixt.NewXTDriver(driver, aiOpts...)
		if err != nil {
			return nil, errors.Wrap(err, "init ios XTDriver failed")
		}
		r.RegisterUIXTDriver(iosDeviceOptions.UDID, driverExt)
	}
	// parse harmony devices config
	for _, harmonyDeviceOptions := range parsedConfig.Harmony {
		err := r.parseDeviceConfig(harmonyDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse harmony config failed: %v", err))
		}

		device, err := uixt.NewHarmonyDevice(harmonyDeviceOptions.Options()...)
		if err != nil {
			return nil, errors.Wrap(err, "init harmony device failed")
		}
		driver, err := device.NewDriver()
		if err != nil {
			return nil, errors.Wrap(err, "init harmony driver failed")
		}

		driverExt, err := uixt.NewXTDriver(driver, aiOpts...)
		if err != nil {
			return nil, errors.Wrap(err, "init harmony XTDriver failed")
		}
		r.RegisterUIXTDriver(harmonyDeviceOptions.ConnectKey, driverExt)
	}
	// parse browser devices config
	for _, browserDeviceOptions := range parsedConfig.Browser {
		err := r.parseDeviceConfig(browserDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse browser config failed: %v", err))
		}
		device, err := uixt.NewBrowserDevice(browserDeviceOptions.Options()...)
		if err != nil {
			return nil, errors.Wrap(err, "init browser device failed")
		}
		driver, err := device.NewDriver()
		if err != nil {
			return nil, errors.Wrap(err, "init browser driver failed")
		}
		driverExt, err := uixt.NewXTDriver(driver, aiOpts...)
		if err != nil {
			return nil, errors.Wrap(err, "init browser XTDriver failed")
		}
		r.RegisterUIXTDriver(browserDeviceOptions.BrowserID, driverExt)
	}

	return parsedConfig, nil
}

func (r *CaseRunner) RegisterUIXTDriver(serial string, driver *uixt.XTDriver) {
	r.uixtDrivers[serial] = driver
}

func (r *CaseRunner) parseDeviceConfig(device interface{}, configVariables map[string]interface{}) error {
	deviceValue := reflect.ValueOf(device).Elem()
	deviceType := deviceValue.Type()
	for i := 0; i < deviceType.NumField(); i++ {
		field := deviceType.Field(i)
		fieldValue := deviceValue.Field(i)
		if fieldValue.Kind() != reflect.String {
			continue
		}

		// skip if field cannot be set
		if !fieldValue.CanSet() {
			log.Warn().Str("field", field.Name).Msg("field cannot be set, skip")
			continue
		}

		parsedValue, err := r.parser.ParseString(
			fieldValue.String(), configVariables)
		if err != nil {
			log.Error().Err(err).Msgf(
				"parse config device variable %s failed", field.Name)
			return err
		}

		parsedValueReflect := reflect.ValueOf(parsedValue)
		if parsedValueReflect.Type().ConvertibleTo(fieldValue.Type()) {
			convertedValue := parsedValueReflect.Convert(fieldValue.Type())
			fieldValue.Set(convertedValue)
		} else {
			log.Error().Msgf("update config device variable %s failed", field.Name)
			return err
		}
	}
	return nil
}

func (r *CaseRunner) GetUIXTDriver(serial string) (driver *uixt.XTDriver, err error) {
	for key, driver := range r.uixtDrivers {
		// return the driver with the same serial
		if key == serial {
			return driver, nil
		}
		// or return the first driver if serial is empty
		if serial == "" {
			r.uixtDrivers[serial] = driver
			return driver, nil
		}
	}
	return nil, errors.New("no driver found")
}

// each boomer task initiates a new session
// in order to avoid data racing
func (r *CaseRunner) NewSession() *SessionRunner {
	log.Info().Msg("create new session runner")
	sessionRunner := &SessionRunner{
		caseRunner:       r,
		sessionVariables: make(map[string]interface{}),
		summary:          NewCaseSummary(),

		transactions: make(map[string]map[TransactionType]time.Time),
		ws:           newWSSession(),
	}
	return sessionRunner
}

// SessionRunner is used to run testcase and its steps.
// each testcase has its own SessionRunner instance and share session variables.
type SessionRunner struct {
	caseRunner *CaseRunner // all session runners share one CaseRunner

	sessionVariables map[string]interface{} // testcase execution session variables
	summary          *TestCaseSummary       // record test case summary

	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[TransactionType]time.Time

	// websocket session
	ws *wsSession
}

// Start runs the test steps in sequential order.
// givenVars is used for data driven
func (r *SessionRunner) Start(givenVars map[string]interface{}) (summary *TestCaseSummary, err error) {
	// report GA event
	sdk.SendGA4Event("hrp_session_runner_start", nil)

	config := r.caseRunner.TestCase.Config.Get()
	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	// update config variables with given variables
	r.InitWithParameters(givenVars)

	defer func() {
		// release session resources
		r.ReleaseResources()

		summary = r.summary
		summary.Name = config.Name
		summary.Time.Duration = time.Since(summary.Time.StartAt).Seconds()
		exportVars := make(map[string]interface{})
		for _, value := range config.Export {
			exportVars[value] = r.sessionVariables[value]
		}
		summary.InOut.ExportVars = exportVars
		summary.InOut.ConfigVars = config.Variables

		// TODO: move to mobile ui step
		for uuid, client := range r.caseRunner.uixtDrivers {
			// add WDA/UIA logs to summary
			logs := map[string]interface{}{
				"uuid": uuid,
			}

			if client.GetDevice().LogEnabled() {
				log, err1 := client.StopCaptureLog()
				if err1 != nil {
					if err == nil {
						err = errors.Wrap(err1, "stop capture log failed")
					} else {
						err = errors.Wrap(err, "stop capture log failed")
					}
					return
				}
				logs["content"] = log
			}

			summary.Logs = append(summary.Logs, logs)
		}
	}()

	// run step in sequential order
	for _, step := range r.caseRunner.TestSteps {
		select {
		case <-r.caseRunner.hrpRunner.caseTimeoutTimer.C:
			log.Warn().Msg("timeout in session runner")
			return summary, errors.Wrap(code.TimeoutError, "session runner timeout")
		case <-r.caseRunner.hrpRunner.interruptSignal:
			log.Warn().Msg("interrupted in session runner")
			return summary, errors.Wrap(code.InterruptError, "session runner interrupted")
		default:
			_, err := r.RunStep(step)
			if err == nil {
				continue
			}
			// interrupted or timeout, abort running
			if errors.Is(err, code.InterruptError) || errors.Is(err, code.TimeoutError) {
				return summary, err
			}

			// check if failfast
			if r.caseRunner.hrpRunner.failfast {
				return summary, errors.Wrap(err, "abort running due to failfast setting")
			}
		}
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return summary, nil
}

func (r *SessionRunner) RunStep(step IStep) (stepResult *StepResult, err error) {
	// parse step struct
	if err = r.ParseStep(step); err != nil {
		log.Error().Err(err).Msg("parse step struct failed")
		if r.caseRunner.hrpRunner.failfast {
			return nil, errors.Wrap(err, "parse step struct failed")
		}
	}

	stepName := step.Name()
	stepType := string(step.Type())
	log.Info().Str("step", stepName).Str("type", stepType).Msg("run step start")

	// run times of step
	loopTimes := step.Config().Loops
	if loopTimes < 0 {
		log.Warn().Int("loops", loopTimes).Msg("loop times should be positive, set to 1")
		loopTimes = 1
	} else if loopTimes == 0 {
		loopTimes = 1
	} else if loopTimes > 1 {
		log.Info().Int("loops", loopTimes).Msg("run step with specified loop times")
	}

	// run step with specified loop times
	for i := 1; i <= loopTimes; i++ {
		var loopIndex string
		if loopTimes > 1 {
			log.Info().Int("index", i).Msg("start running step in loop")
			loopIndex = fmt.Sprintf("_loop_%d", i)
		}

		// run step
		stepResult, err = step.Run(r)
		stepResult.Name = stepName + loopIndex

		// add step result to summary
		r.summary.AddStepResult(stepResult)

		// update extracted variables
		for k, v := range stepResult.ExportVars {
			r.sessionVariables[k] = v
		}

		// run step success
		if err == nil {
			log.Info().Str("step", stepName).
				Str("type", stepType).
				Bool("success", true).
				Int64("elapsed(ms)", stepResult.Elapsed).
				Interface("exportVars", stepResult.ExportVars).
				Msg("run step end")
			continue
		}
		// run step failed
		log.Error().Err(err).Str("step", stepName).
			Str("type", stepType).
			Bool("success", false).
			Int64("elapsed(ms)", stepResult.Elapsed).
			Msg("run step end")
		return stepResult, err
	}

	return stepResult, nil
}

func (r *SessionRunner) GetSummary() *TestCaseSummary {
	r.summary.Time.Duration = time.Since(r.summary.Time.StartAt).Seconds()
	return r.summary
}

func (r *SessionRunner) ParseStep(step IStep) error {
	caseConfig := r.caseRunner.TestCase.Config.Get()
	stepConfig := step.Config()

	// update step variables: merges step variables with config variables and session variables
	// variables priority: step variables > session variables (extracted variables from previous steps)
	overrideVars := mergeVariables(stepConfig.Variables, r.sessionVariables)
	// step variables > testcase config variables
	overrideVars = mergeVariables(overrideVars, caseConfig.Variables)

	// parse step variables
	parsedVariables, err := r.caseRunner.parser.ParseVariables(overrideVars)
	if err != nil {
		log.Error().Interface("variables", caseConfig.Variables).
			Err(err).Msg("parse step variables failed")
		return errors.Wrap(err, "parse step variables failed")
	}
	stepConfig.Variables = parsedVariables

	// parse step name
	parsedName, err := r.caseRunner.parser.ParseString(
		stepConfig.StepName, stepConfig.Variables)
	if err != nil {
		parsedName = step.Name()
	}
	stepConfig.StepName = convertString(parsedName)

	// parse step validators
	var parsedValidators []interface{}
	for _, iValidator := range stepConfig.Validators {
		validator, ok := iValidator.(Validator)
		if !ok {
			return errors.New("validator type error")
		}
		// parse validator check
		// FIXME: validate with current step's extracted variables
		// check, err := r.caseRunner.parser.Parse(
		// 	validator.Check, stepStruct.Variables)
		// if err != nil {
		// 	return errors.Wrap(err, "failed to parse validator check")
		// }
		// validator.Check, _ = check.(string)

		// parse validator expect
		validator.Expect, err = r.caseRunner.parser.Parse(
			validator.Expect, stepConfig.Variables)
		if err != nil {
			return errors.Wrap(err, "failed to parse validator expect")
		}
		parsedValidators = append(parsedValidators, validator)
	}
	stepConfig.Validators = parsedValidators

	return nil
}

// InitWithParameters updates session variables with given parameters.
// this is used for data driven
func (r *SessionRunner) InitWithParameters(parameters map[string]interface{}) {
	if len(parameters) == 0 {
		return
	}

	log.Info().Interface("parameters", parameters).Msg("update session variables")
	for k, v := range parameters {
		r.sessionVariables[k] = v
	}
}

func (r *SessionRunner) GetSessionVariables() map[string]interface{} {
	return r.sessionVariables
}

func (r *SessionRunner) GetTransactions() map[string]map[TransactionType]time.Time {
	return r.transactions
}
