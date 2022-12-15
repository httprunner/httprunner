package hrp

import (
	"bufio"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
)

func newOutSummary() *Summary {
	platForm := &Platform{
		HttprunnerVersion: version.VERSION,
		GoVersion:         runtime.Version(),
		Platform:          fmt.Sprintf("%v-%v", runtime.GOOS, runtime.GOARCH),
	}
	return &Summary{
		Success: true,
		Stat:    &Stat{},
		Time: &TestCaseTime{
			StartAt: env.StartTime,
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

func (s *Summary) appendCaseSummary(caseSummary *TestCaseSummary) {
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
	s.Success = s.Success && caseSummary.Success

	// specify output reports dir
	if len(s.Details) == 1 {
		s.rootDir = caseSummary.RootDir
	} else if s.rootDir != caseSummary.RootDir {
		// if multiple testcases have different root path, use current working dir
		s.rootDir = env.RootDir
	}
}

func (s *Summary) genHTMLReport() error {
	reportsDir := filepath.Join(s.rootDir, env.ResultsDir)
	err := builtin.EnsureFolderExists(reportsDir)
	if err != nil {
		return err
	}

	reportPath := filepath.Join(reportsDir, "report.html")
	file, err := os.OpenFile(reportPath, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		log.Error().Err(err).Msg("open file failed")
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	tmpl := template.Must(template.New("report").Parse(reportTemplate))
	err = tmpl.Execute(writer, s)
	if err != nil {
		log.Error().Err(err).Msg("execute applies a parsed template to the specified data object failed")
		return err
	}
	err = writer.Flush()
	if err == nil {
		log.Info().Str("path", reportPath).Msg("generate HTML report")
	} else {
		log.Error().Str("path", reportPath).Msg("generate HTML report failed")
	}
	return err
}

func (s *Summary) genSummary() error {
	reportsDir := filepath.Join(s.rootDir, env.ResultsDir)
	err := builtin.EnsureFolderExists(reportsDir)
	if err != nil {
		return err
	}

	summaryPath := filepath.Join(reportsDir, "summary.json")
	err = builtin.Dump2JSON(s, summaryPath)
	if err != nil {
		return err
	}
	return nil
}

//go:embed internal/scaffold/templates/report/template.html
var reportTemplate string

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
	Total     int `json:"total" yaml:"total"`
	Successes int `json:"successes" yaml:"successes"`
	Failures  int `json:"failures" yaml:"failures"`
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
	RootDir string         `json:"root_dir" yaml:"root_dir"`
}

type TestCaseInOut struct {
	ConfigVars map[string]interface{} `json:"config_vars" yaml:"config_vars"`
	ExportVars map[string]interface{} `json:"export_vars" yaml:"export_vars"`
}

func newSessionData() *SessionData {
	return &SessionData{
		Success:  false,
		ReqResps: &ReqResps{},
	}
}

type SessionData struct {
	Success    bool                `json:"success" yaml:"success"`
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

func newSummary() *TestCaseSummary {
	return &TestCaseSummary{
		Success: true,
		Stat:    &TestStepStat{},
		Time:    &TestCaseTime{},
		InOut:   &TestCaseInOut{},
		Records: []*StepResult{},
	}
}
