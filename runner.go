package hrp

import (
	"net/http"
	"testing"

	"github.com/imroc/req"
	"github.com/pkg/errors"
)

// run API test with default configs
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
	log.Info().Msg("[init] WithTestingT")
	r.t = t
	return r
}

func (r *Runner) SetDebug(debug bool) *Runner {
	log.Info().Bool("debug", debug).Msg("[init] SetDebug")
	r.debug = debug
	return r
}

func (r *Runner) SetProxyUrl(proxyUrl string) *Runner {
	log.Info().Str("proxyUrl", proxyUrl).Msg("[init] SetProxyUrl")
	r.client.SetProxyUrl(proxyUrl)
	return r
}

func (r *Runner) Run(testcases ...ITestCase) error {
	for _, iTestCase := range testcases {
		testcase, err := iTestCase.ToTestCase()
		if err != nil {
			log.Error().Err(err).Msg("[Run] convert ITestCase interface to TestCase struct failed")
			return err
		}
		if err := r.runCase(testcase); err != nil {
			log.Error().Err(err).Msg("[Run] run testcase failed")
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

	log.Info().Str("testcase", config.Name).Msg("run testcase start")

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
			log.Error().Interface("variables", config.Variables).Err(err).Msg("parse step variables failed")
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

	log.Info().Str("testcase", config.Name).Msg("run testcase end")
	return nil
}

func (r *Runner) runStep(step IStep, config *TConfig) (stepData *StepData, err error) {
	log.Info().Str("step", step.Name()).Msg("run step start")
	if tc, ok := step.(*testcaseWithOptionalArgs); ok {
		// run referenced testcase
		log.Info().Str("testcase", tc.step.Name).Msg("run referenced testcase")
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
	log.Info().
		Str("step", step.Name()).
		Bool("success", stepData.Success).
		Interface("exportVars", stepData.ExportVars).
		Msg("run step end")
	return
}

func (r *Runner) runStepRequest(step *TStep) (stepData *StepData, err error) {
	stepData = &StepData{
		Name:           step.Name,
		Success:        false,
		ResponseLength: 0,
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
	if step.Request.Body != nil {
		data, err := parseData(step.Request.Body, step.Variables)
		if err != nil {
			return nil, err
		}
		switch data.(type) {
		case map[string]interface{}: // post json
			v = append(v, req.BodyJSON(data))
		default: // post raw data
			v = append(v, data)
		}
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
	respObj, err := NewResponseObject(r.t, resp)
	if err != nil {
		err = errors.Wrap(err, "init ResponseObject error")
		return
	}

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
	stepData.ResponseLength = resp.Response().ContentLength
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
		log.Error().Interface("variables", config.Variables).Err(err).Msg("parse config variables failed")
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
