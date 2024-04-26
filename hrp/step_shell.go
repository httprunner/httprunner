package hrp

import (
	"fmt"
	"os"
	"time"

	"github.com/httprunner/funplugin/myexec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Shell struct {
	String         string `json:"string" yaml:"string"`
	ExpectExitCode int    `json:"expect_exit_code" yaml:"expect_exit_code"`
}

// StepShell implements IStep interface.
type StepShell struct {
	step *TStep
}

func (s *StepShell) Name() string {
	return s.step.Name
}

func (s *StepShell) Type() StepType {
	return stepTypeShell
}

func (s *StepShell) Struct() *TStep {
	return s.step
}

func (s *StepShell) Run(r *SessionRunner) (*StepResult, error) {
	return runStepShell(r, s.step)
}

// Validate switches to step validation.
func (s *StepShell) Validate() *StepShellValidation {
	return &StepShellValidation{
		step: s.step,
	}
}

// StepShellValidation implements IStep interface.
type StepShellValidation struct {
	step *TStep
}

func (s *StepShellValidation) Name() string {
	return s.step.Name
}

func (s *StepShellValidation) Type() StepType {
	return stepTypeShell + stepTypeSuffixValidation
}

func (s *StepShellValidation) Struct() *TStep {
	return s.step
}

func (s *StepShellValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepShell(r, s.step)
}

func (s *StepShellValidation) AssertExitCode(expected int) *StepShellValidation {
	s.step.Shell.ExpectExitCode = expected
	return s
}

func runStepShell(r *SessionRunner, step *TStep) (stepResult *StepResult, err error) {
	shell := step.Shell
	log.Info().
		Str("name", step.Name).
		Str("type", string(stepTypeShell)).
		Str("content", shell.String).
		Msg("run shell string")

	stepResult = &StepResult{
		Name:        step.Name,
		StepType:    stepTypeShell,
		Success:     false,
		Elapsed:     0,
		ContentSize: 0,
	}

	vars := r.caseRunner.parsedConfig.Variables
	for key, value := range vars {
		os.Setenv(key, fmt.Sprintf("%v", value))
	}

	start := time.Now()
	exitCode, err := myexec.RunShell(shell.String)
	stepResult.Elapsed = time.Since(start).Milliseconds()
	if err != nil {
		if exitCode == shell.ExpectExitCode {
			// get expected error
			log.Warn().Err(err).
				Int("exitCode", exitCode).
				Msg("get expected error, ignore")
			stepResult.Success = true
			return stepResult, nil
		}

		err = errors.Wrap(err, "exec shell string failed")
		return
	}

	// validate response
	if exitCode != shell.ExpectExitCode {
		err = fmt.Errorf("unexpected exit code %d, expect %d",
			exitCode, shell.ExpectExitCode)
		return
	}

	stepResult.Success = true
	return stepResult, nil
}
