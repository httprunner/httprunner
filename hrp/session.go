package hrp

import (
	_ "embed"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SessionRunner is used to run testcase and its steps.
// each testcase has its own SessionRunner instance and share session variables.
type SessionRunner struct {
	testCase         *TestCase
	hrpRunner        *HRPRunner
	parser           *Parser
	sessionVariables map[string]interface{}
	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[transactionType]time.Time
	startTime    time.Time        // record start time of the testcase
	summary      *TestCaseSummary // record test case summary
}

func (r *SessionRunner) init() {
	log.Info().Msg("init session runner")
	r.sessionVariables = make(map[string]interface{})
	r.transactions = make(map[string]map[transactionType]time.Time)
	r.startTime = time.Now()
	r.summary.Name = r.testCase.Config.Name
}

func (r *SessionRunner) GetParser() *Parser {
	return r.parser
}

func (r *SessionRunner) GetConfig() *TConfig {
	return r.testCase.Config
}

func (r *SessionRunner) LogOn() bool {
	return r.hrpRunner.requestsLogOn
}

// Start runs the test steps in sequential order.
func (r *SessionRunner) Start() error {
	config := r.testCase.Config
	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	// init session runner
	r.init()

	// init plugin
	var err error
	if r.parser.plugin, err = initPlugin(config.Path, r.hrpRunner.pluginLogOn); err != nil {
		return err
	}
	defer func() {
		if r.parser.plugin != nil {
			r.parser.plugin.Quit()
		}
	}()

	// parse config
	if err := r.parseConfig(config); err != nil {
		return err
	}

	r.startTime = time.Now()
	// run step in sequential order
	for _, step := range r.testCase.TestSteps {
		_, err := step.Run(r)
		if err != nil && r.hrpRunner.failfast {
			return errors.Wrap(err, "abort running due to failfast setting")
		}
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return nil
}

func (r *SessionRunner) UpdateSession(vars map[string]interface{}) {
	for k, v := range vars {
		r.sessionVariables[k] = v
	}
}

// UpdateSummary appends step result to summary
func (r *SessionRunner) UpdateSummary(stepResult *StepResult) {
	r.summary.Records = append(r.summary.Records, stepResult)
	r.summary.Stat.Total += 1
	if stepResult.Success {
		r.summary.Stat.Successes += 1
	} else {
		r.summary.Stat.Failures += 1
		// update summary result to failed
		r.summary.Success = false
	}
}

// MergeStepVariables merges step variables with config variables and session variables
func (r *SessionRunner) MergeStepVariables(vars map[string]interface{}) (map[string]interface{}, error) {
	// override variables
	// step variables > session variables (extracted variables from previous steps)
	overrideVars := mergeVariables(vars, r.sessionVariables)
	// step variables > testcase config variables
	overrideVars = mergeVariables(overrideVars, r.testCase.Config.Variables)

	// parse step variables
	parsedVariables, err := r.parser.ParseVariables(overrideVars)
	if err != nil {
		log.Error().Interface("variables", r.testCase.Config.Variables).
			Err(err).Msg("parse step variables failed")
		return nil, err
	}
	return parsedVariables, nil
}

func (r *SessionRunner) parseConfig(cfg *TConfig) error {
	// parse config variables
	parsedVariables, err := r.parser.ParseVariables(cfg.Variables)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	cfg.Variables = parsedVariables

	// parse config name
	parsedName, err := r.parser.ParseString(cfg.Name, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := r.parser.ParseString(cfg.BaseURL, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.BaseURL = convertString(parsedBaseURL)

	// ensure correction of think time config
	cfg.ThinkTimeSetting.checkThinkTime()

	return nil
}

func (r *SessionRunner) GetSummary() *TestCaseSummary {
	caseSummary := r.summary
	caseSummary.Time.StartAt = r.startTime
	caseSummary.Time.Duration = time.Since(r.startTime).Seconds()
	exportVars := make(map[string]interface{})
	for _, value := range r.testCase.Config.Export {
		exportVars[value] = r.sessionVariables[value]
	}
	caseSummary.InOut.ExportVars = exportVars
	caseSummary.InOut.ConfigVars = r.testCase.Config.Variables
	return caseSummary
}
