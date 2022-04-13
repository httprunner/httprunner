package hrp

import (
	_ "embed"
	"time"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SessionRunner is used to run testcase and its steps.
// each testcase has its own SessionRunner instance and share session variables.
type SessionRunner struct {
	testCase         *TestCase
	hrpRunner        *HRPRunner
	parser           *Parser
	parsedConfig     *TConfig
	sessionVariables map[string]interface{}
	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[transactionType]time.Time
	startTime    time.Time        // record start time of the testcase
	summary      *TestCaseSummary // record test case summary
}

func (r *SessionRunner) init() {
	log.Info().Msg("init session runner")
	r.parsedConfig = &TConfig{}
	r.sessionVariables = make(map[string]interface{})
	r.transactions = make(map[string]map[transactionType]time.Time)
	r.startTime = time.Now()
}

func (r *SessionRunner) GetParser() *Parser {
	return r.parser
}

func (r *SessionRunner) GetConfig() *TConfig {
	return r.parsedConfig
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
	if err := r.parseConfig(nil); err != nil {
		return err
	}

	r.startTime = time.Now()
	// run step in sequential order
	for _, step := range r.testCase.TestSteps {
		log.Info().Str("step", step.Name()).
			Str("type", string(step.Type())).Msg("run step start")

		stepResult, err := step.Run(r)
		if err != nil && r.hrpRunner.failfast {
			log.Error().
				Str("step", stepResult.Name).
				Str("type", string(stepResult.StepType)).
				Bool("success", false).
				Msg("run step end")
			return errors.Wrap(err, "abort running due to failfast setting")
		}

		// update extracted variables
		for k, v := range stepResult.ExportVars {
			r.sessionVariables[k] = v
		}

		log.Info().
			Str("step", stepResult.Name).
			Str("type", string(stepResult.StepType)).
			Bool("success", stepResult.Success).
			Interface("exportVars", stepResult.ExportVars).
			Msg("run step end")
	}

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return nil
}

// MergeStepVariables merges step variables with config variables and session variables
func (r *SessionRunner) MergeStepVariables(vars map[string]interface{}) (map[string]interface{}, error) {
	// override variables
	// step variables > session variables (extracted variables from previous steps)
	overrideVars := mergeVariables(vars, r.sessionVariables)
	// step variables > testcase config variables
	overrideVars = mergeVariables(overrideVars, r.parsedConfig.Variables)

	// parse step variables
	parsedVariables, err := r.parser.ParseVariables(overrideVars)
	if err != nil {
		log.Error().Interface("variables", r.parsedConfig.Variables).
			Err(err).Msg("parse step variables failed")
		return nil, err
	}
	return parsedVariables, nil
}

// parseConfig parses testcase config with given variables, stores to parsedConfig.
func (r *SessionRunner) parseConfig(variables map[string]interface{}) error {
	cfg := r.testCase.Config

	// deep copy config to avoid data racing
	if err := copier.Copy(r.parsedConfig, cfg); err != nil {
		log.Error().Err(err).Msg("copy testcase config failed")
		return err
	}

	// parse config variables
	mergedVars := mergeVariables(variables, cfg.Variables)
	parsedVariables, err := r.parser.ParseVariables(mergedVars)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	r.parsedConfig.Variables = parsedVariables

	// parse config name
	parsedName, err := r.parser.ParseString(cfg.Name, cfg.Variables)
	if err != nil {
		return err
	}
	r.parsedConfig.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := r.parser.ParseString(cfg.BaseURL, cfg.Variables)
	if err != nil {
		return err
	}
	r.parsedConfig.BaseURL = convertString(parsedBaseURL)

	// ensure correction of think time config
	r.parsedConfig.ThinkTimeSetting.checkThinkTime()

	return nil
}

func (r *SessionRunner) GetSummary() *TestCaseSummary {
	caseSummary := r.summary
	caseSummary.Name = r.parsedConfig.Name
	caseSummary.Time.StartAt = r.startTime
	caseSummary.Time.Duration = time.Since(r.startTime).Seconds()
	exportVars := make(map[string]interface{})
	for _, value := range r.parsedConfig.Export {
		exportVars[value] = r.sessionVariables[value]
	}
	caseSummary.InOut.ExportVars = exportVars
	caseSummary.InOut.ConfigVars = r.parsedConfig.Variables
	return caseSummary
}
