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
	"github.com/httprunner/httprunner/v4/hrp/internal/config"
	"github.com/httprunner/httprunner/v4/hrp/internal/version"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
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
				Actions: make(map[uixt.ActionMethod]int),
			},
		},
		Time: &TestCaseTime{
			StartAt: config.StartTime,
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
		s.rootDir = config.RootDir
	}

	// merge action stats
	for action, count := range caseSummary.Stat.Actions {
		if _, ok := s.Stat.TestSteps.Actions[action]; !ok {
			s.Stat.TestSteps.Actions[action] = 0
		}
		s.Stat.TestSteps.Actions[action] += count
	}
}

func (s *Summary) SetupDirPath() (path string, err error) {
	dirPath := filepath.Join(s.rootDir, config.ResultsDir)
	err = builtin.EnsureFolderExists(dirPath)
	if err != nil {
		return "", err
	}
	return dirPath, nil
}

func (s *Summary) GenHTMLReport() error {
	reportsDir, err := s.SetupDirPath()
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

func (s *Summary) GenSummary() (path string, err error) {
	reportsDir, err := s.SetupDirPath()
	if err != nil {
		return "", err
	}

	path = filepath.Join(reportsDir, "summary.json")
	err = builtin.Dump2JSON(s, path)
	if err != nil {
		return "", err
	}
	return path, nil
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
	Total     int                       `json:"total" yaml:"total"`
	Successes int                       `json:"successes" yaml:"successes"`
	Failures  int                       `json:"failures" yaml:"failures"`
	Actions   map[uixt.ActionMethod]int `json:"actions" yaml:"actions"` // record action stats
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
			Actions: make(map[uixt.ActionMethod]int),
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
	case stepTypeTestCase:
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

func newSessionData() *SessionData {
	return &SessionData{
		ReqResps: &ReqResps{},
	}
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
