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

func (s *StepTestCaseWithOptionalArgs) Struct() *TStep {
	return s.step
}

func (s *StepTestCaseWithOptionalArgs) Run(r *SessionRunner) (*StepResult, error) {
	stepVariables, err := r.MergeStepVariables(s.step.Variables)
	if err != nil {
		return nil, err
	}
	s.step.Variables = stepVariables

	log.Info().Str("testcase", s.step.Name).Msg("run referenced testcase")
	stepResult := &StepResult{
		Name:     s.step.Name,
		StepType: stepTypeTestCase,
		Success:  false,
	}
	testcase := s.step.TestCase.(*TestCase)

	// copy testcase to avoid data racing
	copiedTestCase := &TestCase{}
	if err := copier.Copy(copiedTestCase, testcase); err != nil {
		log.Error().Err(err).Msg("copy testcase failed")
		return stepResult, err
	}
	// override testcase config
	extendWithTestCase(s.step, copiedTestCase)

	sessionRunner := r.hrpRunner.NewSessionRunner(copiedTestCase)

	start := time.Now()
	err = sessionRunner.Start()
	stepResult.Elapsed = time.Since(start).Milliseconds()
	if err != nil {
		log.Error().Err(err).Msg("run referenced testcase step failed")
		log.Info().Str("step", s.step.Name).Bool("success", false).Msg("run step end")
		stepResult.Attachment = err.Error()
		r.summary.Success = false
		return stepResult, err
	}
	summary := sessionRunner.GetSummary()
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
		Str("step", s.step.Name).
		Bool("success", true).
		Interface("exportVars", stepResult.ExportVars).
		Msg("run step end")

	return stepResult, nil
}

// extend referenced testcase with teststep, teststep config merge and override referenced testcase config
func extendWithTestCase(testStep *TStep, overriddenTestCase *TestCase) {
	// override testcase name
	if testStep.Name != "" {
		overriddenTestCase.Config.Name = testStep.Name
	}
	// merge & override variables
	overriddenTestCase.Config.Variables = mergeVariables(testStep.Variables, overriddenTestCase.Config.Variables)
	// merge & override extractors
	overriddenTestCase.Config.Export = mergeSlices(testStep.Export, overriddenTestCase.Config.Export)
}
