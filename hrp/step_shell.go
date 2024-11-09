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
	StepConfig
	Shell *Shell `json:"shell,omitempty" yaml:"shell,omitempty"`
}

func (s *StepShell) Name() string {
	return s.StepName
}

func (s *StepShell) Type() StepType {
	return stepTypeShell
}

func (s *StepShell) Config() *StepConfig {
	return &StepConfig{
		StepName:  s.StepName,
		Variables: s.Variables,
	}
}

func (s *StepShell) Run(r *SessionRunner) (*StepResult, error) {
	return runStepShell(r, s)
}

// Validate switches to step validation.
func (s *StepShell) Validate() *StepShellValidation {
	return &StepShellValidation{
		StepConfig: s.StepConfig,
		Shell:      s.Shell,
	}
}

// StepShellValidation implements IStep interface.
type StepShellValidation struct {
	StepConfig
	Shell *Shell `json:"shell,omitempty" yaml:"shell,omitempty"`
}

func (s *StepShellValidation) Name() string {
	return s.StepName
}

func (s *StepShellValidation) Type() StepType {
	return stepTypeShell + stepTypeSuffixValidation
}

func (s *StepShellValidation) Config() *StepConfig {
	return &StepConfig{
		StepName:  s.StepName,
		Variables: s.Variables,
	}
}

func (s *StepShellValidation) Run(r *SessionRunner) (*StepResult, error) {
	return runStepShell(r, s)
}

func (s *StepShellValidation) AssertExitCode(expected int) *StepShellValidation {
	s.Shell.ExpectExitCode = expected
	return s
}

func runStepShell(r *SessionRunner, step IStep) (stepResult *StepResult, err error) {
	var shell *Shell
	switch stepShell := step.(type) {
	case *StepShell:
		shell = stepShell.Shell
	case *StepShellValidation:
		shell = stepShell.Shell
	default:
		return nil, errors.New("invalid shell step type")
	}

	log.Info().
		Str("name", step.Name()).
		Str("type", string(stepTypeShell)).
		Str("content", shell.String).
		Msg("run shell string")

	stepResult = &StepResult{
		Name:        step.Name(),
		StepType:    stepTypeShell,
		Success:     false,
		Elapsed:     0,
		ContentSize: 0,
	}

	vars := r.caseRunner.Config.Get().Variables
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
