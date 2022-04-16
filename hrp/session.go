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
	testCase           *TestCase
	hrpRunner          *HRPRunner
	parser             *Parser
	parsedConfig       *TConfig
	parametersIterator *ParametersIterator
	sessionVariables   map[string]interface{}
	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions map[string]map[transactionType]time.Time
	startTime    time.Time        // record start time of the testcase
	summary      *TestCaseSummary // record test case summary
}

func (r *SessionRunner) resetSession() {
	log.Info().Msg("reset session runner")
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
// givenVars is used for data driven
func (r *SessionRunner) Start(givenVars map[string]interface{}) error {
	config := r.testCase.Config
	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	// update config variables with given variables
	r.updateConfigVariables(givenVars)

	// reset session runner
	r.resetSession()

	// run step in sequential order
	for _, step := range r.testCase.TestSteps {
		log.Info().Str("step", step.Name()).
			Str("type", string(step.Type())).Msg("run step start")

		stepResult, err := step.Run(r)
		if err != nil {
			log.Error().
				Str("step", stepResult.Name).
				Str("type", string(stepResult.StepType)).
				Bool("success", false).
				Msg("run step end")

			if r.hrpRunner.failfast {
				return errors.Wrap(err, "abort running due to failfast setting")
			}
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

// updateConfigVariables updates config variables with given variables.
// this is used for data driven
func (r *SessionRunner) updateConfigVariables(parameters map[string]interface{}) {
	if parameters == nil {
		return
	}

	log.Info().Interface("parameters", parameters).Msg("update config variables")
	for k, v := range parameters {
		r.parsedConfig.Variables[k] = v
	}
}

// parseConfig parses testcase config, stores to parsedConfig.
func (r *SessionRunner) parseConfig() error {
	cfg := r.testCase.Config

	r.parsedConfig = &TConfig{}
	// deep copy config to avoid data racing
	if err := copier.Copy(r.parsedConfig, cfg); err != nil {
		log.Error().Err(err).Msg("copy testcase config failed")
		return err
	}

	// parse config variables
	parsedVariables, err := r.parser.ParseVariables(cfg.Variables)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	r.parsedConfig.Variables = parsedVariables

	// parse config name
	parsedName, err := r.parser.ParseString(cfg.Name, parsedVariables)
	if err != nil {
		return errors.Wrap(err, "parse config name failed")
	}
	r.parsedConfig.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := r.parser.ParseString(cfg.BaseURL, parsedVariables)
	if err != nil {
		return errors.Wrap(err, "parse config base url failed")
	}
	r.parsedConfig.BaseURL = convertString(parsedBaseURL)

	// ensure correction of think time config
	r.parsedConfig.ThinkTimeSetting.checkThinkTime()

	// parse testcase config parameters
	parametersIterator, err := initParametersIterator(r.parsedConfig)
	if err != nil {
		log.Error().Err(err).
			Interface("parameters", r.parsedConfig.Parameters).
			Interface("parametersSetting", r.parsedConfig.ParametersSetting).
			Msg("parse config parameters failed")
		return errors.Wrap(err, "parse testcase config parameters failed")
	}
	r.parametersIterator = parametersIterator

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
