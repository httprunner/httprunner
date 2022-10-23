package hrp

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// StepTestCaseWithOptionalArgs implements IStep interface.
type StepTestCaseWithOptionalArgs struct {
	step *TStep
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepTestCaseWithOptionalArgs) TeardownHook(hook string) *StepTestCaseWithOptionalArgs {
	s.step.TeardownHooks = append(s.step.TeardownHooks, hook)
	return s
}

// Export specifies variable names to export from referenced testcase for current step.
func (s *StepTestCaseWithOptionalArgs) Export(names ...string) *StepTestCaseWithOptionalArgs {
	s.step.Export = append(s.step.Export, names...)
	return s
}

func (s *StepTestCaseWithOptionalArgs) Name() string {
	if s.step.Name != "" {
		return s.step.Name
	}
	ts, ok := s.step.TestCase.(*TestCase)
	if ok {
		return ts.Config.Name
	}
	return ""
}

func (s *StepTestCaseWithOptionalArgs) Type() StepType {
	return stepTypeTestCase
}

func (s *StepTestCaseWithOptionalArgs) Struct() *TStep {
	return s.step
}

func (s *StepTestCaseWithOptionalArgs) Run(r *SessionRunner) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:     s.step.Name,
		StepType: stepTypeTestCase,
		Success:  false,
	}

	// merge step variables with session variables
	stepVariables, err := r.ParseStepVariables(s.step.Variables)
	if err != nil {
		err = errors.Wrap(err, "parse step variables failed")
		return
	}

	defer func() {
		// update testcase summary
		if err != nil {
			stepResult.Attachments = err.Error()
		}
	}()

	stepTestCase := s.step.TestCase.(*TestCase)

	// copy testcase to avoid data racing
	copiedTestCase := &TestCase{}
	if err := copier.Copy(copiedTestCase, stepTestCase); err != nil {
		log.Error().Err(err).Msg("copy step testcase failed")
		return stepResult, err
	}

	// override testcase config
	// override testcase name
	if s.step.Name != "" {
		copiedTestCase.Config.Name = s.step.Name
	}
	// merge & override extractors
	copiedTestCase.Config.Export = mergeSlices(s.step.Export, copiedTestCase.Config.Export)

	caseRunner, err := r.caseRunner.hrpRunner.NewCaseRunner(copiedTestCase)
	if err != nil {
		log.Error().Err(err).Msg("create case runner failed")
		return stepResult, err
	}
	sessionRunner := caseRunner.NewSession()

	start := time.Now()
	// run referenced testcase with step variables
	err = sessionRunner.Start(stepVariables)
	stepResult.Elapsed = time.Since(start).Milliseconds()

	summary, err2 := sessionRunner.GetSummary()
	if err2 != nil {
		log.Error().Err(err2).Msg("get summary failed")
		if err != nil {
			err = errors.Wrap(err, err2.Error())
		} else {
			err = err2
		}
	}
	// update step names
	for _, record := range summary.Records {
		record.Name = fmt.Sprintf("%s - %s", stepResult.Name, record.Name)
	}
	stepResult.Data = summary.Records
	// export testcase export variables
	stepResult.ExportVars = summary.InOut.ExportVars

	if err == nil {
		stepResult.Success = true
	}
	return stepResult, err
}
