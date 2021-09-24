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
		r.runStep(step, config)
	}
	return nil
}

func (r *Runner) runStep(step IStep, config *TConfig) error {
	log.Printf("run step begin: %v >>>>>>", step.Name())
	if tc, ok := step.(*testcaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Printf("run referenced testcase: %v", tc.step.Name)
		// TODO: override testcase config
		if err := r.runCase(tc.step.TestCase); err != nil {
			return err
		}
	} else {
		// run request
		tStep := parseStep(step, config)
		if err := r.runStepRequest(tStep); err != nil {
			return err
		}
	}
	log.Printf("run step end: %v <<<<<<\n", step.Name())
	return nil
}

func (r *Runner) runStepRequest(step *TStep) error {
	var v []interface{}
	v = append(v, req.Header(step.Request.Headers))
	v = append(v, req.Param(step.Request.Params))
	v = append(v, step.Request.Data)
	v = append(v, req.BodyJSON(step.Request.JSON))

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
	defer resp.Response().Body.Close()
	return nil
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}
