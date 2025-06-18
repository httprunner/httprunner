package hrp

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
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
	return &s.StepConfig
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
		StepType:    step.Type(),
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

// Call custom function, used for pre/post action hook
func Call(desc string, fn func(), opts ...option.ActionOption) error {
	actionOptions := option.NewActionOptions(opts...)

	startTime := time.Now()
	defer func() {
		log.Info().Str("desc", desc).
			Int64("duration(ms)", time.Since(startTime).Milliseconds()).
			Msg("function called")
	}()

	if actionOptions.Timeout == 0 {
		// wait for function to finish
		fn()
		return nil
	}

	// set timeout for function execution
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
		// function completed within timeout
		return nil
	case <-time.After(time.Duration(actionOptions.Timeout) * time.Second):
		return fmt.Errorf("function execution exceeded timeout of %d seconds", actionOptions.Timeout)
	}
}
