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
	parser           *parser
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

// Run runs the test steps in sequential order.
func (r *SessionRunner) Run() error {
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

func (r *SessionRunner) overrideVariables(step *TStep) (*TStep, error) {
	// copy step and config to avoid data racing
	copiedStep := &TStep{}
	if err := copier.Copy(copiedStep, step); err != nil {
		log.Error().Err(err).Msg("copy step data failed")
		return nil, err
	}

	stepVariables := copiedStep.Variables
	// override variables
	// step variables > session variables (extracted variables from previous steps)
	stepVariables = mergeVariables(stepVariables, r.sessionVariables)
	// step variables > testcase config variables
	stepVariables = mergeVariables(stepVariables, r.testCase.Config.Variables)

	// parse step variables
	parsedVariables, err := r.parser.parseVariables(stepVariables)
	if err != nil {
		log.Error().Interface("variables", r.testCase.Config.Variables).Err(err).Msg("parse step variables failed")
		return nil, err
	}
	copiedStep.Variables = parsedVariables // avoid data racing
	return copiedStep, nil
}

func (r *SessionRunner) overrideConfig(step *TStep) {
	// override headers
	if r.testCase.Config.Headers != nil {
		step.Request.Headers = mergeMap(step.Request.Headers, r.testCase.Config.Headers)
	}
	// parse step request url
	requestUrl, err := r.parser.parseString(step.Request.URL, step.Variables)
	if err != nil {
		log.Error().Err(err).Msg("parse request url failed")
		requestUrl = step.Variables
	}
	step.Request.URL = buildURL(r.testCase.Config.BaseURL, convertString(requestUrl)) // avoid data racing
}

func (r *SessionRunner) parseConfig(cfg *TConfig) error {
	// parse config variables
	parsedVariables, err := r.parser.parseVariables(cfg.Variables)
	if err != nil {
		log.Error().Interface("variables", cfg.Variables).Err(err).Msg("parse config variables failed")
		return err
	}
	cfg.Variables = parsedVariables

	// parse config name
	parsedName, err := r.parser.parseString(cfg.Name, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := r.parser.parseString(cfg.BaseURL, cfg.Variables)
	if err != nil {
		return err
	}
	cfg.BaseURL = convertString(parsedBaseURL)

	// ensure correction of think time config
	cfg.ThinkTimeSetting.checkThinkTime()

	return nil
}

func (r *SessionRunner) getSummary() *TestCaseSummary {
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
