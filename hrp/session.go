package hrp

import (
	_ "embed"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SessionRunner is used to run testcase and its steps.
// each testcase has its own SessionRunner instance and share session variables.
type SessionRunner struct {
	*testCaseRunner
	sessionVariables map[string]interface{}
	// transactions stores transaction timing info.
	// key is transaction name, value is map of transaction type and time, e.g. start time and end time.
	transactions      map[string]map[transactionType]time.Time
	startTime         time.Time               // record start time of the testcase
	summary           *TestCaseSummary        // record test case summary
	wsConn            *websocket.Conn         // one websocket connection each session
	pongResponseChan  chan string             // channel used to receive pong response message
	closeResponseChan chan *wsCloseRespObject // channel used to receive close response message
}

func (r *SessionRunner) resetSession() {
	log.Info().Msg("reset session runner")
	r.sessionVariables = make(map[string]interface{})
	r.transactions = make(map[string]map[transactionType]time.Time)
	r.startTime = time.Now()
	r.summary = newSummary()
	r.pongResponseChan = make(chan string, 1)
	r.closeResponseChan = make(chan *wsCloseRespObject, 1)
}

func (r *SessionRunner) GetParser() *Parser {
	return r.parser
}

func (r *SessionRunner) GetConfig() *TConfig {
	return r.parsedConfig
}

func (r *SessionRunner) HTTPStatOn() bool {
	return r.hrpRunner.httpStatOn
}

func (r *SessionRunner) LogOn() bool {
	return r.hrpRunner.requestsLogOn
}

// Start runs the test steps in sequential order.
// givenVars is used for data driven
func (r *SessionRunner) Start(givenVars map[string]interface{}) error {
	config := r.testCase.Config
	log.Info().Str("testcase", config.Name).Msg("run testcase start")

	// reset session runner
	r.resetSession()

	// update config variables with given variables
	r.updateSessionVariables(givenVars)

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

	// close websocket connection after all steps done
	defer func() {
		if r.wsConn != nil {
			log.Info().Str("testcase", config.Name).Msg("websocket disconnected")
			err := r.wsConn.Close()
			if err != nil {
				log.Error().Err(err).Msg("websocket disconnection failed")
			}
		}
	}()

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

// updateSessionVariables updates session variables with given variables.
// this is used for data driven
func (r *SessionRunner) updateSessionVariables(parameters map[string]interface{}) {
	if len(parameters) == 0 {
		return
	}

	log.Info().Interface("parameters", parameters).Msg("update session variables")
	for k, v := range parameters {
		r.sessionVariables[k] = v
	}
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
