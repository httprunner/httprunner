package hrp

import (
	"fmt"
	"os"
	"time"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// StepFunction implements IStep interface.
type StepFunction struct {
	StepConfig
	Fn func()
}

func (s *StepFunction) Name() string {
	return s.StepName
}

func (s *StepFunction) Type() StepType {
	return StepTypeFunction
}

func (s *StepFunction) Config() *StepConfig {
	return &StepConfig{
		StepName:  s.StepName,
		Variables: s.Variables,
	}
}

func (s *StepFunction) Run(r *SessionRunner) (*StepResult, error) {
	return runStepFunction(r, s)
}

func runStepFunction(r *SessionRunner, step IStep) (stepResult *StepResult, err error) {
	var fn func()
	switch stepFn := step.(type) {
	case *StepFunction:
		fn = stepFn.Fn
	default:
		return nil, errors.New("unexpected function step type")
	}

	log.Info().
		Str("name", step.Name()).
		Str("type", string(StepTypeFunction)).
		Msg("run function")

	start := time.Now()
	stepResult = &StepResult{
		Name:        step.Name(),
		StepType:    StepTypeFunction,
		Success:     false,
		ContentSize: 0,
		StartTime:   start.Unix(),
	}
	defer func() {
		attachments := uixt.Attachments{}
		if err != nil {
			attachments["error"] = err.Error()
		}
		stepResult.Attachments = attachments
		stepResult.Elapsed = time.Since(start).Milliseconds()
	}()

	vars := r.caseRunner.Config.Get().Variables
	for key, value := range vars {
		os.Setenv(key, fmt.Sprintf("%v", value))
	}

	// exec function
	fn()

	stepResult.Success = true
	return stepResult, nil
}
