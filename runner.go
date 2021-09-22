package httpboomer

import (
	"net/http"

	"github.com/imroc/req"
)

func HttpRunner() *Runner {
	return &Runner{}
}

type Runner struct {
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
	// config := testcase.Config
	for _, step := range testcase.TestSteps {
		if err := r.runStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runStep(req IStep) error {
	return req.Run()
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}

func (step *TStep) Run() error {

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
	resp, err := req.Do(string(step.Request.Method), step.Request.URL, v...)
	if err != nil {
		return err
	}
	resp.Response().Body.Close()
	return nil
}
