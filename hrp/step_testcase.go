package hrp

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog/log"
)

// StepTestCaseWithOptionalArgs implements IStep interface.
type StepTestCaseWithOptionalArgs struct {
	StepConfig
	TestCase interface{} `json:"testcase,omitempty" yaml:"testcase,omitempty"` // *TestCasePath or *TestCase
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepTestCaseWithOptionalArgs) TeardownHook(hook string) *StepTestCaseWithOptionalArgs {
	s.TeardownHooks = append(s.TeardownHooks, hook)
	return s
}

// Export specifies variable names to export from referenced testcase for current step.
func (s *StepTestCaseWithOptionalArgs) Export(names ...string) *StepTestCaseWithOptionalArgs {
	s.StepExport = append(s.StepExport, names...)
	return s
}

func (s *StepTestCaseWithOptionalArgs) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	ts, ok := s.TestCase.(*TestCase)
	if ok {
		return ts.Config.Name
	}
	return ""
}

func (s *StepTestCaseWithOptionalArgs) Type() StepType {
	return stepTypeTestCase
}

func (s *StepTestCaseWithOptionalArgs) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepTestCaseWithOptionalArgs) Run(r *SessionRunner) (stepResult *StepResult, err error) {
	stepResult = &StepResult{
		Name:     s.StepName,
		StepType: stepTypeTestCase,
		Success:  false,
	}

	defer func() {
		// update testcase summary
		if err != nil {
			stepResult.Attachments = err.Error()
		}
	}()

	stepTestCase := s.TestCase.(*TestCase)

	// copy testcase to avoid data racing
	copiedTestCase := &TestCase{}
	if err := copier.Copy(copiedTestCase, stepTestCase); err != nil {
		log.Error().Err(err).Msg("copy step testcase failed")
		return stepResult, err
	}

	// override testcase config
	// override testcase name
	if s.StepName != "" {
		copiedTestCase.Config.Name = s.StepName
	}
	// merge & override extractors
	copiedTestCase.Config.Export = mergeSlices(s.StepExport, copiedTestCase.Config.Export)

	caseRunner, err := r.caseRunner.hrpRunner.NewCaseRunner(*copiedTestCase)
	if err != nil {
		log.Error().Err(err).Msg("create case runner failed")
		return stepResult, err
	}
	sessionRunner := caseRunner.NewSession()

	start := time.Now()
	var summary *TestCaseSummary
	// run referenced testcase with step variables
	summary, err = sessionRunner.Start(s.Variables)
	stepResult.Elapsed = time.Since(start).Milliseconds()

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
