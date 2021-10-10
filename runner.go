package httpboomer

import (
	"log"
	"net/http"
	"testing"

	"github.com/imroc/req"
)

func Run(t *testing.T, testcases ...ITestCase) error {
	return NewRunner().WithTestingT(t).SetDebug(true).Run(testcases...)
}

func NewRunner() *Runner {
	return &Runner{
		t:      &testing.T{},
		debug:  false, // default to turn off debug
		client: req.New(),
	}
}

type Runner struct {
	t      *testing.T
	debug  bool
	client *req.Req
}

func (r *Runner) WithTestingT(t *testing.T) *Runner {
	r.t = t
	return r
}

func (r *Runner) SetDebug(debug bool) *Runner {
	r.debug = debug
	return r
}

func (r *Runner) Run(testcases ...ITestCase) error {
	for _, testcase := range testcases {
		tcStruct, err := testcase.ToStruct()
		if err != nil {
			log.Printf("[Run] testcase.ToStruct() error: %v", err)
			return err
		}
		if err := r.runCase(tcStruct); err != nil {
			log.Printf("[Run] runCase error: %v", err)
			return err
		}
	}
	return nil
}

func (r *Runner) runCase(testcase *TestCase) error {
	config := &testcase.Config
	if err := r.parseConfig(config); err != nil {
		return err
	}

	log.Printf("Start to run testcase: %v", config.Name)

	extractedVariables := make(map[string]interface{})

	for _, step := range testcase.TestSteps {
		// override variables
		// step variables > extracted variables from previous steps
		stepVariables := mergeVariables(step.ToStruct().Variables, extractedVariables)
		// step variables > testcase config variables
		stepVariables = mergeVariables(stepVariables, config.Variables)

		// parse step variables
		parsedVariables, err := parseVariables(stepVariables)
		if err != nil {
			log.Printf("[parseConfig] parse variables: %v, error: %v", config.Variables, err)
			return err
		}
		step.ToStruct().Variables = parsedVariables

		stepData, err := r.runStep(step, config)
		if err != nil {
			return err
		}
		// update extracted variables
		for k, v := range stepData.ExportVars {
			extractedVariables[k] = v
		}
	}

	return nil
}

func (r *Runner) runStep(step IStep, config *TConfig) (stepData *StepData, err error) {
	log.Printf("run step begin: %v >>>>>>", step.Name())
	if tc, ok := step.(*testcaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Printf("run referenced testcase: %v", tc.step.Name)
		// TODO: override testcase config
		stepData, err = r.runStepTestCase(tc.step)
		if err != nil {
			return
		}
	} else {
		// run request
		tStep := parseStep(step, config)
		stepData, err = r.runStepRequest(tStep)
		if err != nil {
			return
		}
	}
	log.Printf("run step end: %v <<<<<<\n", step.Name())
	return
}

func (r *Runner) runStepRequest(step *TStep) (stepData *StepData, err error) {
	stepData = &StepData{
		Name:    step.Name,
		Success: false,
	}

	// prepare request args
	var v []interface{}
	if len(step.Request.Headers) > 0 {
		headers, err := parseHeaders(step.Request.Headers, step.Variables)
		if err != nil {
			return nil, err
		}
		v = append(v, req.Header(headers))
	}
	if len(step.Request.Params) > 0 {
		params, err := parseData(step.Request.Params, step.Variables)
		if err != nil {
			return nil, err
		}
		v = append(v, req.Param(params.(map[string]interface{})))
	}
	if step.Request.Data != nil {
		data, err := parseData(step.Request.Data, step.Variables)
		if err != nil {
			return nil, err
		}
		v = append(v, data)
	}
	if step.Request.JSON != nil {
		jsonData, err := parseData(step.Request.JSON, step.Variables)
		if err != nil {
			return nil, err
		}
		v = append(v, req.BodyJSON(jsonData))
	}

	for cookieName, cookieValue := range step.Request.Cookies {
		v = append(v, &http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		})
	}

	// do request action
	req.Debug = r.debug
	resp, err := r.client.Do(string(step.Request.Method), step.Request.URL, v...)
	if err != nil {
		return
	}
	defer resp.Response().Body.Close()

	// new response object
	respObj := NewResponseObject(r.t, resp)

	// extract variables from response
	extractors := step.Extract
	extractMapping := respObj.Extract(extractors)
	stepData.ExportVars = extractMapping

	// override step variables with extracted variables
	stepVariables := mergeVariables(step.Variables, extractMapping)

	// validate response
	err = respObj.Validate(step.Validators, stepVariables)
	if err != nil {
		return
	}

	stepData.Success = true
	return
}

func (r *Runner) runStepTestCase(step *TStep) (stepData *StepData, err error) {
	stepData = &StepData{
		Name:    step.Name,
		Success: false,
	}
	testcase := step.TestCase
	err = r.runCase(testcase)
	return
}

func (r *Runner) parseConfig(config *TConfig) error {
	// parse config variables
	parsedVariables, err := parseVariables(config.Variables)
	if err != nil {
		log.Printf("[parseConfig] parse variables: %v, error: %v", config.Variables, err)
		return err
	}
	config.Variables = parsedVariables

	// parse config name
	parsedName, err := parseString(config.Name, config.Variables)
	if err != nil {
		return err
	}
	config.Name = convertString(parsedName)

	// parse config base url
	parsedBaseURL, err := parseString(config.BaseURL, config.Variables)
	if err != nil {
		return err
	}
	config.BaseURL = convertString(parsedBaseURL)

	return nil
}

func (r *Runner) GetSummary() *TestCaseSummary {
	return &TestCaseSummary{}
}
