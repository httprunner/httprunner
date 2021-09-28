package httpboomer

import (
	"log"
	"net/http"
	"testing"

	"github.com/imroc/req"
)

func Test(t *testing.T, testcases ...*TestCase) error {
	return NewRunner().WithTestingT(t).SetDebug(true).Run(testcases...)
}

func NewRunner() *Runner {
	return &Runner{
		t:      &testing.T{},
		debug:  false, // default to turn off debug
		Client: req.New(),
	}
}

type Runner struct {
	t      *testing.T
	debug  bool
	Client *req.Req
}

func (r *Runner) WithTestingT(t *testing.T) *Runner {
	r.t = t
	return r
}

func (r *Runner) SetDebug(debug bool) *Runner {
	r.debug = debug
	return r
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
		if err := r.runStepTestCase(tc.step); err != nil {
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
	// prepare request args
	var v []interface{}
	if len(step.Request.Headers) > 0 {
		v = append(v, req.Header(step.Request.Headers))
	}
	if len(step.Request.Params) > 0 {
		v = append(v, req.Param(step.Request.Params))
	}
	if step.Request.Data != nil {
		v = append(v, step.Request.Data)
	}
	if step.Request.JSON != nil {
		v = append(v, req.BodyJSON(step.Request.JSON))
	}

	for cookieName, cookieValue := range step.Request.Cookies {
		v = append(v, &http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		})
	}

	// do request action
	req.Debug = r.debug
	resp, err := r.Client.Do(string(step.Request.Method), step.Request.URL, v...)
	if err != nil {
		return err
	}
	defer resp.Response().Body.Close()

	// validate response
	respObj := NewResponseObject(r.t, resp)
	err = respObj.Validate(step.Validators, step.Variables)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) runStepTestCase(step *TStep) error {
	testcase := step.TestCase
	return r.runCase(testcase)
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}
