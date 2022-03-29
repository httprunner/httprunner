package hrp

import (
	"time"

	"github.com/jinzhu/copier"
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

func (s *StepTestCaseWithOptionalArgs) ToStruct() *TStep {
	return s.step
}

func (s *StepTestCaseWithOptionalArgs) Run(r *SessionRunner) (*StepResult, error) {
	copiedStep, err := r.overrideVariables(s.step)
	if err != nil {
		return nil, err
	}

	log.Info().Str("testcase", copiedStep.Name).Msg("run referenced testcase")
	stepResult := &StepResult{
		Name:     copiedStep.Name,
		StepType: stepTypeTestCase,
		Success:  false,
	}
	testcase := copiedStep.TestCase.(*TestCase)

	// copy testcase to avoid data racing
	copiedTestCase := &TestCase{}
	if err := copier.Copy(copiedTestCase, testcase); err != nil {
		log.Error().Err(err).Msg("copy testcase failed")
		return stepResult, err
	}
	// override testcase config
	extendWithTestCase(copiedStep, copiedTestCase)

	sessionRunner := r.hrpRunner.NewSessionRunner(copiedTestCase)

	start := time.Now()
	err = sessionRunner.Run()
	stepResult.Elapsed = time.Since(start).Milliseconds()
	if err != nil {
		log.Error().Err(err).Msg("run referenced testcase step failed")
		log.Info().Str("step", copiedStep.Name).Bool("success", false).Msg("run step end")
		stepResult.Attachment = err.Error()
		r.summary.Success = false
		return stepResult, err
	}
	summary := sessionRunner.getSummary()
	stepResult.Data = summary
	// export testcase export variables
	stepResult.ExportVars = sessionRunner.summary.InOut.ExportVars
	stepResult.Success = true

	// update extracted variables
	for k, v := range stepResult.ExportVars {
		r.sessionVariables[k] = v
	}

	// merge testcase summary
	r.summary.Records = append(r.summary.Records, summary.Records...)
	r.summary.Stat.Total += summary.Stat.Total
	r.summary.Stat.Successes += summary.Stat.Successes
	r.summary.Stat.Failures += summary.Stat.Failures

	log.Info().
		Str("step", copiedStep.Name).
		Bool("success", true).
		Interface("exportVars", stepResult.ExportVars).
		Msg("run step end")

	return stepResult, nil
}
