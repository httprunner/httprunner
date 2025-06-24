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
	"sort"
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
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/sdk"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/mcphost"
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
		mcpConfigPath: "",
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
	mcpConfigPath    string // MCP config file path
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

// SetMCPConfigPath configures the MCP config path.
func (r *HRPRunner) SetMCPConfigPath(mcpConfigPath string) *HRPRunner {
	log.Info().Str("mcpConfigPath", mcpConfigPath).Msg("[init] SetMCPConfigPath")
	r.mcpConfigPath = mcpConfigPath
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

	// defer summary saving and HTML report generation
	// this ensures they run regardless of how the function exits
	defer func() {
		s.Time.Duration = time.Since(s.Time.StartAt).Seconds()
		log.Info().Int("duration(s)", int(s.Time.Duration)).Msg("run testcase finished")

		// save summary
		if r.saveTests {
			if summaryPath, saveErr := s.GenSummary(); saveErr != nil {
				log.Error().Err(saveErr).Msg("failed to save summary")
			} else {
				log.Info().Str("path", summaryPath).Msg("summary saved successfully")
			}
		}

		// generate HTML report
		if r.genHTMLReport {
			if reportErr := s.GenHTMLReport(); reportErr != nil {
				log.Error().Err(reportErr).Msg("failed to generate HTML report")
			} else {
				log.Info().Msg("HTML report generated successfully")
			}
		}
	}()

	// load all testcases
	testCases, err := LoadTestCases(testcases...)
	if err != nil {
		log.Error().Err(err).Msg("failed to load testcases")
		return err
	}

	// collect all MCP hosts for cleanup
	var mcpHosts []*mcphost.MCPHost

	// quit all plugins and close MCP hosts
	defer func() {
		pluginMap.Range(func(key, value interface{}) bool {
			if plugin, ok := value.(funplugin.IPlugin); ok {
				plugin.Quit()
			}
			return true
		})

		// Close all MCP hosts with timeout
		if len(mcpHosts) > 0 {
			done := make(chan struct{})
			go func() {
				defer close(done)
				for _, host := range mcpHosts {
					if host != nil {
						host.Shutdown()
					}
				}
			}()

			// Wait for cleanup with timeout
			select {
			case <-done:
				log.Debug().Msg("All MCP hosts cleaned up successfully")
			case <-time.After(10 * time.Second):
				log.Warn().Msg("MCP hosts cleanup timeout")
			}
		}
	}()

	var runErr error
	// run testcase one by one
	for _, testcase := range testCases {
		// check for interrupt signal before processing each testcase
		select {
		case <-r.interruptSignal:
			log.Warn().Msg("interrupted in main runner")
			return errors.Wrap(code.InterruptError, "main runner interrupted")
		default:
		}

		// each testcase has its own case runner
		caseRunner, err := NewCaseRunner(*testcase, r)
		if err != nil {
			log.Error().Err(err).Msg("[Run] init case runner failed")
			return err
		}

		// collect MCP host for cleanup
		if caseRunner.parser.MCPHost != nil {
			mcpHosts = append(mcpHosts, caseRunner.parser.MCPHost)
		}

		for it := caseRunner.parametersIterator; it.HasNext(); {
			// check for interrupt signal before each iteration
			select {
			case <-r.interruptSignal:
				log.Warn().Msg("interrupted in parameter iteration")
				return errors.Wrap(code.InterruptError, "parameter iteration interrupted")
			default:
			}

			// case runner can run multiple times with different parameters
			// each run has its own session runner
			sessionRunner := caseRunner.NewSession()
			caseSummary, err := sessionRunner.Start(it.Next())
			s.AddCaseSummary(caseSummary)
			if err != nil {
				log.Error().Err(err).Msg("[Run] run testcase failed")
				if r.failfast {
					return err
				}
				runErr = err
			}
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
		TestCase:  testcase,
		hrpRunner: hrpRunner,
		parser:    NewParser(),
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

	// init MCP servers
	mcpConfigPath := hrpRunner.mcpConfigPath
	if mcpConfigPath == "" {
		mcpConfigPath = config.MCPConfigPath
	}
	if mcpConfigPath != "" {
		mcpHost, err := mcphost.NewMCPHost(mcpConfigPath, false)
		if err != nil {
			return nil, errors.Wrapf(err, "init mcp config %s failed", mcpConfigPath)
		}
		caseRunner.parser.MCPHost = mcpHost
		log.Info().Str("mcpConfigPath", mcpConfigPath).Msg("mcp server loaded")
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
	if parsedConfig.AIOptions != nil {
		aiOpts = parsedConfig.AIOptions.Options()
	}

	var driverConfigs []uixt.DriverCacheConfig
	// parse android devices config
	for _, androidDeviceOptions := range parsedConfig.Android {
		err := r.parseDeviceConfig(androidDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse android config failed: %v", err))
		}
		driverConfigs = append(driverConfigs, uixt.DriverCacheConfig{
			Platform:   "android",
			Serial:     androidDeviceOptions.SerialNumber,
			AIOptions:  aiOpts,
			DeviceOpts: option.FromAndroidOptions(androidDeviceOptions),
		})
	}
	// parse iOS devices config
	for _, iosDeviceOptions := range parsedConfig.IOS {
		err := r.parseDeviceConfig(iosDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse ios config failed: %v", err))
		}
		driverConfigs = append(driverConfigs, uixt.DriverCacheConfig{
			Platform:   "ios",
			Serial:     iosDeviceOptions.UDID,
			AIOptions:  aiOpts,
			DeviceOpts: option.FromIOSOptions(iosDeviceOptions),
		})
	}
	// parse harmony devices config
	for _, harmonyDeviceOptions := range parsedConfig.Harmony {
		err := r.parseDeviceConfig(harmonyDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse harmony config failed: %v", err))
		}
		driverConfigs = append(driverConfigs, uixt.DriverCacheConfig{
			Platform:   "harmony",
			Serial:     harmonyDeviceOptions.ConnectKey,
			AIOptions:  aiOpts,
			DeviceOpts: option.FromHarmonyOptions(harmonyDeviceOptions),
		})
	}
	// parse browser devices config
	for _, browserDeviceOptions := range parsedConfig.Browser {
		err := r.parseDeviceConfig(browserDeviceOptions, parsedConfig.Variables)
		if err != nil {
			return nil, errors.Wrap(code.InvalidCaseError,
				fmt.Sprintf("parse browser config failed: %v", err))
		}
		driverConfigs = append(driverConfigs, uixt.DriverCacheConfig{
			Platform:   "browser",
			Serial:     browserDeviceOptions.BrowserID,
			AIOptions:  aiOpts,
			DeviceOpts: option.FromBrowserOptions(browserDeviceOptions),
		})
	}

	// init XTDriver and register to unified cache
	for _, driverConfig := range driverConfigs {
		driver, err := uixt.GetOrCreateXTDriver(driverConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "init %s XTDriver failed", driverConfig.Platform)
		}

		// Set MCP clients if MCPHost is available
		if r.parser.MCPHost != nil {
			mcpClients := r.parser.MCPHost.GetAllClients()
			driver.SetMCPClients(mcpClients)
			log.Debug().Str("serial", driverConfig.Serial).
				Int("mcp_clients", len(mcpClients)).
				Msg("Set MCP clients for XTDriver")
		}
	}

	return parsedConfig, nil
}

// RegisterUIXTDriver is used to register a external driver to the unified cache
func (r *CaseRunner) RegisterUIXTDriver(serial string, driver *uixt.XTDriver) error {
	if err := uixt.RegisterXTDriver(serial, driver); err != nil {
		log.Error().Err(err).Str("serial", serial).Msg("register XTDriver failed")
		return err
	}
	log.Info().Str("serial", serial).Msg("register XTDriver success")
	return nil
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

		// Save JSON case content to results directory
		if config.Path != "" {
			if err := saveJSONCase(config.Path); err != nil {
				log.Warn().Err(err).Str("path", config.Path).Msg("save JSON case failed")
			}
		}

		// TODO: move to mobile ui step
		// Collect logs from cached drivers
		for _, cached := range uixt.ListCachedDrivers() {
			// add WDA/UIA logs to summary
			logs := map[string]interface{}{
				"uuid": cached.Key,
			}

			client := cached.Item
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

const (
	RUN_STEP_START = "run step start"
	RUN_STEP_END   = "run step end"
)

// executionTask holds the necessary information for a single step execution.
type executionTask struct {
	stepName   string
	parameters map[string]interface{}
}

// formatParameters formats parameter values into a string for display in step names.
// e.g. {"foo": "bar", "age": 18} -> "bar-18"
func formatParameters(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	// sort keys to ensure consistent order
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var values []string
	for _, k := range keys {
		values = append(values, fmt.Sprintf("%v", params[k]))
	}
	return strings.Join(values, "-")
}

// generateExecutionTasks generates a list of execution tasks based on step parameters and loops.
func (r *SessionRunner) generateExecutionTasks(step IStep) ([]executionTask, error) {
	stepConfig := step.Config()
	stepName := step.Name()

	// determine effective loop times
	loopTimes := stepConfig.Loops
	if loopTimes <= 0 {
		loopTimes = 1 // default to 1 if not set
	}

	// initialize parameters iterator
	parametersIterator, err := r.caseRunner.parser.InitParametersIterator(&TConfig{
		Parameters:        stepConfig.Parameters,
		ParametersSetting: stepConfig.ParametersSetting,
		Variables:         stepConfig.Variables,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize parameters iterator")
	}

	// collect all parameter combinations first
	var allParameters []map[string]interface{}
	if parametersIterator != nil {
		for parametersIterator.HasNext() {
			allParameters = append(allParameters, parametersIterator.Next())
		}
	}

	// if no parameters are specified, but loop times are set,
	// we should run the step loopTimes with empty parameters.
	if len(allParameters) == 0 && loopTimes > 0 {
		allParameters = append(allParameters, make(map[string]interface{}))
	}

	// generate execution tasks
	var tasks []executionTask
	for loopIndex := 1; loopIndex <= loopTimes; loopIndex++ {
		for _, params := range allParameters {
			// determine step name based on parameters and loops
			currentStepName := stepName
			hasParameters := len(params) > 0
			hasLoops := loopTimes > 1

			if hasParameters {
				paramStr := formatParameters(params)
				if hasLoops {
					currentStepName = fmt.Sprintf("%s [loop_%d_params_%s]", stepName, loopIndex, paramStr)
				} else {
					currentStepName = fmt.Sprintf("%s [params_%s]", stepName, paramStr)
				}
			} else if hasLoops {
				currentStepName = fmt.Sprintf("%s_loop_%d", stepName, loopIndex)
			}

			tasks = append(tasks, executionTask{
				stepName:   currentStepName,
				parameters: params,
			})
		}
	}

	return tasks, nil
}

func (r *SessionRunner) RunStep(step IStep) (stepResult *StepResult, err error) {
	// check for interrupt signal before running step
	select {
	case <-r.caseRunner.hrpRunner.interruptSignal:
		log.Warn().Msg("interrupted in RunStep")
		return nil, errors.Wrap(code.InterruptError, "RunStep interrupted")
	default:
	}

	// parse step struct
	if err = r.ParseStep(step); err != nil {
		log.Error().Err(err).Msg("parse step struct failed")
		if r.caseRunner.hrpRunner.failfast {
			return nil, errors.Wrap(err, "parse step struct failed")
		}
	}

	stepName := step.Name()
	stepType := string(step.Type())

	log.Info().Str("step", stepName).Str("type", stepType).Msg(RUN_STEP_START)

	// execute step with parameters iterator
	tasks, err := r.generateExecutionTasks(step)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate execution tasks")
	}

	var stepResults []*StepResult

	// execute with loops as outer iteration
	for _, task := range tasks {
		// execute step with merged variables
		stepResult, err := r.executeStepWithVariables(step, task.stepName, task.parameters)
		if err != nil {
			if r.caseRunner.hrpRunner.failfast {
				return nil, errors.Wrap(err, "execute step failed")
			}
			log.Error().Err(err).Str("step", task.stepName).Msg("execute step failed")
		}

		stepResults = append(stepResults, stepResult)
	}

	// return the last step result, or nil if no steps were executed
	if len(stepResults) > 0 {
		// add all step results to summary
		for _, result := range stepResults {
			r.summary.AddStepResult(result)
			// update extracted variables from the last result
			for k, v := range result.ExportVars {
				r.sessionVariables[k] = v
			}
		}

		// log final result
		lastResult := stepResults[len(stepResults)-1]
		if lastResult.Success {
			log.Info().Str("step", stepName).
				Str("type", stepType).
				Bool("success", true).
				Int64("elapsed(ms)", lastResult.Elapsed).
				Interface("exportVars", lastResult.ExportVars).
				Msg(RUN_STEP_END)
		} else {
			log.Error().Str("step", stepName).
				Str("type", stepType).
				Bool("success", false).
				Int64("elapsed(ms)", lastResult.Elapsed).
				Msg(RUN_STEP_END)
		}

		return lastResult, nil
	}

	return nil, errors.New("no steps were executed")
}

// executeStepWithVariables executes a single step with given parameters
// parameters will override step variables with the same name
func (r *SessionRunner) executeStepWithVariables(step IStep, stepName string, parameters map[string]interface{}) (stepResult *StepResult, err error) {
	stepConfig := step.Config()

	// backup original variables
	originalVariables := make(map[string]interface{})
	for k, v := range stepConfig.Variables {
		originalVariables[k] = v
	}

	// merge parameters into step variables
	// parameters have higher priority than variables
	for k, v := range parameters {
		stepConfig.Variables[k] = v
	}

	// execute step
	stepResult, err = step.Run(r)
	stepResult.Name = stepName

	// restore original variables to avoid side effects
	stepConfig.Variables = originalVariables

	return stepResult, err
}

func (r *SessionRunner) GetSummary() *TestCaseSummary {
	r.summary.Time.Duration = time.Since(r.summary.Time.StartAt).Seconds()
	return r.summary
}

// GenerateReport generates report for the testcase.
func (r *SessionRunner) GenerateReport() error {
	summary := NewSummary()
	caseSummary := r.GetSummary()
	summary.AddCaseSummary(caseSummary)
	summary.Time.Duration = time.Since(caseSummary.Time.StartAt).Seconds()
	return summary.GenHTMLReport()
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

// saveJSONCase saves the original JSON case content to the results directory
func saveJSONCase(casePath string) error {
	// Read the original JSON case content
	path := TestCasePath(casePath)
	testCase, err := path.GetTestCase()
	if err != nil {
		return errors.Wrap(err, "load JSON case failed")
	}

	// remove environs from testcase config
	tConfig := testCase.Config.(*TConfig)
	tConfig.Environs = nil

	// save JSON case to results directory
	jsonCasePath := config.GetConfig().CaseFilePath()
	err = testCase.Dump2JSON(jsonCasePath)
	if err != nil {
		return errors.Wrap(err, "dump JSON case failed")
	}
	return nil
}
