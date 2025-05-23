package hrp

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog/log"
)

// IAPI represents interface for api,
// includes API and APIPath.
type IAPI interface {
	GetPath() string
	ToAPI() (*API, error)
}

type API struct {
	Name          string                 `json:"name" yaml:"name"` // required
	Request       *Request               `json:"request,omitempty" yaml:"request,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SetupHooks    []string               `json:"setup_hooks,omitempty" yaml:"setup_hooks,omitempty"`
	TeardownHooks []string               `json:"teardown_hooks,omitempty" yaml:"teardown_hooks,omitempty"`
	Extract       map[string]string      `json:"extract,omitempty" yaml:"extract,omitempty"`
	Validators    []interface{}          `json:"validate,omitempty" yaml:"validate,omitempty"`
	Export        []string               `json:"export,omitempty" yaml:"export,omitempty"`
	Path          string
}

func (api *API) GetPath() string {
	return api.Path
}

func (api *API) ToAPI() (*API, error) {
	return api, nil
}

// APIPath implements IAPI interface.
type APIPath string

func (path *APIPath) GetPath() string {
	return fmt.Sprintf("%v", *path)
}

func (path *APIPath) ToAPI() (*API, error) {
	api := &API{}
	apiPath := path.GetPath()
	err := LoadFileObject(apiPath, api)
	if err != nil {
		return nil, err
	}
	// 1. deal with request body compatibility
	convertCompatRequestBody(api.Request)
	// 2. deal with validators compatibility
	err = convertCompatValidator(api.Validators)
	// 3. deal with extract expr including hyphen
	convertExtract(api.Extract)
	return api, err
}

// StepAPIWithOptionalArgs implements IStep interface.
type StepAPIWithOptionalArgs struct {
	StepConfig
	API interface{} `json:"api,omitempty" yaml:"api,omitempty"` // *APIPath or *API
}

// TeardownHook adds a teardown hook for current teststep.
func (s *StepAPIWithOptionalArgs) TeardownHook(hook string) *StepAPIWithOptionalArgs {
	s.TeardownHooks = append(s.TeardownHooks, hook)
	return s
}

// Export specifies variable names to export from referenced api for current step.
func (s *StepAPIWithOptionalArgs) Export(names ...string) *StepAPIWithOptionalArgs {
	api, ok := s.API.(*API)
	if ok {
		s.StepExport = append(api.Export, names...)
	}
	return s
}

func (s *StepAPIWithOptionalArgs) Name() string {
	if s.StepName != "" {
		return s.StepName
	}
	api, ok := s.API.(*API)
	if ok {
		return api.Name
	}
	return ""
}

func (s *StepAPIWithOptionalArgs) Type() StepType {
	return StepTypeAPI
}

func (s *StepAPIWithOptionalArgs) Config() *StepConfig {
	return &s.StepConfig
}

func (s *StepAPIWithOptionalArgs) Run(r *SessionRunner) (stepResult *StepResult, err error) {
	defer func() {
		stepResult.StepType = StepTypeAPI
	}()
	// extend request with referenced API
	api, _ := s.API.(*API)
	step := &StepRequestWithOptionalArgs{}
	// deep copy step to avoid data racing
	if err = copier.Copy(step, s.StepConfig); err != nil {
		log.Error().Err(err).Msg("copy step failed")
		return
	}
	extendWithAPI(step, api)
	stepResult, err = runStepRequest(r, step)
	return
}

// extend teststep with api, teststep will merge and override referenced api
func extendWithAPI(testStep *StepRequestWithOptionalArgs, overriddenStep *API) {
	// override api name
	if testStep.StepName == "" {
		testStep.StepName = overriddenStep.Name
	}
	// merge & override request
	testStep.Request = overriddenStep.Request
	// init upload
	if len(testStep.Request.Upload) != 0 {
		initUpload(testStep)
	}
	// merge & override variables
	testStep.Variables = mergeVariables(testStep.Variables, overriddenStep.Variables)
	// merge & override extractors
	testStep.StepConfig.Extract = mergeMap(testStep.StepConfig.Extract, overriddenStep.Extract)
	// merge & override validators
	testStep.Validators = mergeValidators(testStep.Validators, overriddenStep.Validators)
	// merge & override setupHooks
	testStep.SetupHooks = mergeSlices(testStep.SetupHooks, overriddenStep.SetupHooks)
	// merge & override teardownHooks
	testStep.TeardownHooks = mergeSlices(testStep.TeardownHooks, overriddenStep.TeardownHooks)
}
