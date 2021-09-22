package httpboomer

import (
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
	for _, step := range testcase.TestSteps {
		if err := step.Run(config); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runStep(step *TStep) error {
	var v []interface{}
	v = append(v, req.Header(step.Request.Headers))
	v = append(v, req.Param(step.Request.Params))

	for cookieName, cookieValue := range step.Request.Cookies {
		v = append(v, &http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		})
	}

	resp, err := r.Client.Do(string(step.Request.Method), step.Request.URL, v...)
	if err != nil {
		return err
	}
	resp.Response().Body.Close()
	return nil
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}
