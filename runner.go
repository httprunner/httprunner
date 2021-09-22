package httpboomer

import (
	"log"
	"net/http"

	"github.com/imroc/req"
)

var defaultRunner = NewRunner()

func Test(testcases ...*TestCase) error {
	return defaultRunner.Run(testcases...)
}

func NewRunner() *Runner {
	return &Runner{
		Client: req.New(),
	}
}

type Runner struct {
	Client *req.Req
}

func (r *Runner) Run(testcases ...*TestCase) error {
	for _, testcase := range testcases {
		if err := r.runCase(testcase); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runCase(testcase *TestCase) error {
	config := &testcase.Config
	log.Printf("Start to run testcase: %v", config.Name)
	for _, step := range testcase.TestSteps {
		tStep := parseStep(step, config)
		if err := r.runStep(tStep); err != nil {
			return err
		}
	}
	return nil
}

func parseStep(step IStep, config *TConfig) *TStep {
	tStep := step.ToStruct()
	tStep.Request.URL = config.BaseURL + tStep.Request.URL
	return tStep
}

func (r *Runner) runStep(step *TStep) error {
	log.Printf("run step begin: %v >>>>>>", step.Name)
	var v []interface{}
	v = append(v, req.Header(step.Request.Headers))
	v = append(v, req.Param(step.Request.Params))

	for cookieName, cookieValue := range step.Request.Cookies {
		v = append(v, &http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		})
	}

	req.Debug = true
	resp, err := r.Client.Do(string(step.Request.Method), step.Request.URL, v...)
	if err != nil {
		return err
	}
	resp.Response().Body.Close()
	log.Printf("run step end: %v <<<<<<\n", step.Name)
	return nil
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}
