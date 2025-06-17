package hrp

import (
	_ "embed"
	"fmt"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/version"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func NewSummary() *Summary {
	platForm := &Platform{
		HttprunnerVersion: version.VERSION,
		GoVersion:         runtime.Version(),
		Platform:          fmt.Sprintf("%v-%v", runtime.GOOS, runtime.GOARCH),
	}
	return &Summary{
		Success: true,
		Stat: &Stat{
			TestSteps: TestStepStat{
				Actions: make(map[option.ActionName]int),
			},
		},
		Time: &TestCaseTime{
			StartAt: config.GetConfig().StartTime,
		},
		Platform: platForm,
	}
}

// Summary stores tests summary for current task execution, maybe include one or multiple testcases
type Summary struct {
	Success  bool               `json:"success" yaml:"success"`
	Stat     *Stat              `json:"stat" yaml:"stat"`
	Time     *TestCaseTime      `json:"time" yaml:"time"`
	Platform *Platform          `json:"platform" yaml:"platform"`
	Details  []*TestCaseSummary `json:"details" yaml:"details"`
	rootDir  string
}

func (s *Summary) AddCaseSummary(caseSummary *TestCaseSummary) {
	log.Info().Str("name", caseSummary.Name).Msg("add case summary")
	s.Success = s.Success && caseSummary.Success
	s.Stat.TestCases.Total += 1
	s.Stat.TestSteps.Total += caseSummary.Stat.Total
	if caseSummary.Success {
		s.Stat.TestCases.Success += 1
	} else {
		s.Stat.TestCases.Fail += 1
	}
	s.Stat.TestSteps.Successes += caseSummary.Stat.Successes
	s.Stat.TestSteps.Failures += caseSummary.Stat.Failures
	s.Details = append(s.Details, caseSummary)

	// specify output reports dir
	if len(s.Details) == 1 {
		s.rootDir = caseSummary.RootDir
	} else if s.rootDir != caseSummary.RootDir {
		// if multiple testcases have different root path, use current working dir
		s.rootDir = config.GetConfig().RootDir
	}

	// merge action stats
	for action, count := range caseSummary.Stat.Actions {
		if _, ok := s.Stat.TestSteps.Actions[action]; !ok {
			s.Stat.TestSteps.Actions[action] = 0
		}
		s.Stat.TestSteps.Actions[action] += count
	}
}

func (s *Summary) GenHTMLReport() error {
	// Find summary.json and hrp.log files
	summaryPath := config.GetConfig().SummaryFilePath()
	logPath := config.GetConfig().LogFilePath()
	reportPath := config.GetConfig().ReportFilePath()

	// Check if summary.json exists, if not create it first
	if !builtin.FileExists(summaryPath) {
		if _, err := s.GenSummary(); err != nil {
			return fmt.Errorf("failed to generate summary.json: %w", err)
		}
	}

	return GenerateHTMLReportFromFiles(summaryPath, logPath, reportPath)
}

func (s *Summary) GenSummary() (path string, err error) {
	path = config.GetConfig().SummaryFilePath()
	err = builtin.Dump2JSON(s, path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (s *Summary) GetResultsPath() string {
	return config.GetConfig().ResultsPath()
}

func (s *Summary) GetSummaryFilePath() string {
	return config.GetConfig().SummaryFilePath()
}

func (s *Summary) GetLogFilePath() string {
	return config.GetConfig().LogFilePath()
}

func (s *Summary) GetReportFilePath() string {
	return config.GetConfig().ReportFilePath()
}

type Stat struct {
	TestCases TestCaseStat `json:"testcases" yaml:"testcases"`
	TestSteps TestStepStat `json:"teststeps" yaml:"teststeps"`
}

type TestCaseStat struct {
	Total   int `json:"total" yaml:"total"`
	Success int `json:"success" yaml:"success"`
	Fail    int `json:"fail" yaml:"fail"`
}

type TestStepStat struct {
	Total     int                       `json:"total" yaml:"total"`
	Successes int                       `json:"successes" yaml:"successes"`
	Failures  int                       `json:"failures" yaml:"failures"`
	Actions   map[option.ActionName]int `json:"actions" yaml:"actions"` // record action stats
}

type TestCaseTime struct {
	StartAt  time.Time `json:"start_at,omitempty" yaml:"start_at,omitempty"`
	Duration float64   `json:"duration,omitempty" yaml:"duration,omitempty"`
}

type Platform struct {
	HttprunnerVersion string `json:"httprunner_version" yaml:"httprunner_version"`
	GoVersion         string `json:"go_version" yaml:"go_version"`
	Platform          string `json:"platform" yaml:"platform"`
}

func NewCaseSummary() *TestCaseSummary {
	return &TestCaseSummary{
		Success: true,
		Stat: &TestStepStat{
			Actions: make(map[option.ActionName]int),
		},
		Time: &TestCaseTime{
			StartAt: time.Now(),
		},
		InOut:   &TestCaseInOut{},
		Records: []*StepResult{},
	}
}

// TestCaseSummary stores tests summary for one testcase
type TestCaseSummary struct {
	Name    string         `json:"name" yaml:"name"`
	Success bool           `json:"success" yaml:"success"`
	CaseId  string         `json:"case_id,omitempty" yaml:"case_id,omitempty"` // TODO
	Stat    *TestStepStat  `json:"stat" yaml:"stat"`
	Time    *TestCaseTime  `json:"time" yaml:"time"`
	InOut   *TestCaseInOut `json:"in_out" yaml:"in_out"`
	Logs    []interface{}  `json:"logs,omitempty" yaml:"logs,omitempty"`
	Records []*StepResult  `json:"records" yaml:"records"`
	RootDir string         `json:"root_dir,omitempty" yaml:"root_dir,omitempty"`
}

// AddStepResult updates summary of StepResult.
func (s *TestCaseSummary) AddStepResult(stepResult *StepResult) {
	switch stepResult.StepType {
	case StepTypeTestCase:
		// record requests of testcase step
		records, ok := stepResult.Data.([]*StepResult)
		if !ok {
			log.Warn().
				Interface("data", stepResult.Data).
				Msg("get unexpected testcase step data")
			return
		}
		s.Success = s.Success && stepResult.Success
		for _, result := range records {
			s.addSingleStepResult(result)
		}
	default:
		s.addSingleStepResult(stepResult)
	}
}

func (s *TestCaseSummary) addSingleStepResult(stepResult *StepResult) {
	s.Success = s.Success && stepResult.Success
	s.Stat.Total += 1
	if stepResult.Success {
		s.Stat.Successes += 1
	} else {
		s.Stat.Failures += 1
	}
	s.Records = append(s.Records, stepResult)
}

type TestCaseInOut struct {
	ConfigVars map[string]interface{} `json:"config_vars" yaml:"config_vars"`
	ExportVars map[string]interface{} `json:"export_vars" yaml:"export_vars"`
}

type SessionData struct {
	ReqResps   *ReqResps           `json:"req_resps" yaml:"req_resps"`
	Address    *Address            `json:"address,omitempty" yaml:"address,omitempty"` // TODO
	Validators []*ValidationResult `json:"validators,omitempty" yaml:"validators,omitempty"`
}

type ReqResps struct {
	Request  interface{} `json:"request" yaml:"request"`
	Response interface{} `json:"response" yaml:"response"`
}

type Address struct {
	ClientIP   string `json:"client_ip,omitempty" yaml:"client_ip,omitempty"`
	ClientPort string `json:"client_port,omitempty" yaml:"client_port,omitempty"`
	ServerIP   string `json:"server_ip,omitempty" yaml:"server_ip,omitempty"`
	ServerPort string `json:"server_port,omitempty" yaml:"server_port,omitempty"`
}

type ValidationResult struct {
	Validator
	CheckValue  interface{} `json:"check_value" yaml:"check_value"`
	CheckResult string      `json:"check_result" yaml:"check_result"`
}
