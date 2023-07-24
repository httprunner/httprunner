package hrp

import (
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type TParamsConfig struct {
	PickOrder  iteratorPickOrder           `json:"pick_order,omitempty" yaml:"pick_order,omitempty"` // overall pick-order strategy
	Strategies map[string]iteratorStrategy `json:"strategies,omitempty" yaml:"strategies,omitempty"` // individual strategies for each parameters
	Limit      int                         `json:"limit,omitempty" yaml:"limit,omitempty"`
}

type iteratorPickOrder string

const (
	pickOrderSequential iteratorPickOrder = "sequential"
	pickOrderRandom     iteratorPickOrder = "random"
	pickOrderUnique     iteratorPickOrder = "unique"
)

/*
[

	{"username": "test1", "password": "111111"},
	{"username": "test2", "password": "222222"},

]
*/
type Parameters []map[string]interface{}

type iteratorStrategy struct {
	Name      string            `json:"name,omitempty" yaml:"name,omitempty"`
	PickOrder iteratorPickOrder `json:"pick_order,omitempty" yaml:"pick_order,omitempty"`
}

func (p *Parser) initParametersIterator(cfg *TConfig) (*ParametersIterator, error) {
	parameters, err := p.loadParameters(cfg.Parameters, cfg.Variables)
	if err != nil {
		return nil, err
	}
	return newParametersIterator(parameters, cfg.ParametersSetting), nil
}

func newParametersIterator(parameters map[string]Parameters, config *TParamsConfig) *ParametersIterator {
	if config == nil {
		config = &TParamsConfig{}
	}
	iterator := &ParametersIterator{
		data:                 parameters,
		hasNext:              true,
		sequentialParameters: nil,
		randomParameterNames: nil,
		limit:                config.Limit,
		index:                0,
	}

	if len(parameters) == 0 {
		iterator.data = map[string]Parameters{}
		iterator.limit = 1
		return iterator
	}

	parametersList := make([]Parameters, 0)
	for paramName := range parameters {
		// check parameter individual pick order strategy
		strategy, ok := config.Strategies[paramName]
		if !ok || strategy.PickOrder == "" {
			// default to overall pick order strategy
			strategy.PickOrder = config.PickOrder
		}

		// group parameters by pick order strategy
		if strategy.PickOrder == pickOrderRandom {
			iterator.randomParameterNames = append(iterator.randomParameterNames, paramName)
		} else {
			parametersList = append(parametersList, parameters[paramName])
		}
	}

	// generate cartesian product for sequential parameters
	iterator.sequentialParameters = genCartesianProduct(parametersList)

	if iterator.limit < 0 {
		log.Warn().Msg("parameters unlimited mode is only supported for load testing")
		iterator.limit = 0
	}
	if iterator.limit == 0 {
		// limit not set
		if len(iterator.sequentialParameters) > 0 {
			// use cartesian product of sequential parameters size as limit
			iterator.limit = len(iterator.sequentialParameters)
		} else {
			// all parameters are selected by random
			// only run once
			iterator.limit = 1
		}
	} else { // limit > 0
		log.Info().Int("limit", iterator.limit).Msg("set limit for parameters")
	}

	return iterator
}

type ParametersIterator struct {
	sync.Mutex
	data                 map[string]Parameters
	hasNext              bool       // cache query result
	sequentialParameters Parameters // cartesian product for sequential parameters
	randomParameterNames []string   // value is parameter names
	limit                int        // limit count for iteration
	index                int        // current iteration index
}

// SetUnlimitedMode is used for load testing
func (iter *ParametersIterator) SetUnlimitedMode() {
	log.Info().Msg("set parameters unlimited mode")
	iter.limit = -1
}

func (iter *ParametersIterator) HasNext() bool {
	if !iter.hasNext {
		return false
	}

	// unlimited mode
	if iter.limit == -1 {
		return true
	}

	// reached limit
	if iter.index >= iter.limit {
		// cache query result
		iter.hasNext = false
		return false
	}

	return true
}

func (iter *ParametersIterator) Next() map[string]interface{} {
	iter.Lock()
	defer iter.Unlock()

	if !iter.hasNext {
		return nil
	}

	var selectedParameters map[string]interface{}
	if len(iter.sequentialParameters) == 0 {
		selectedParameters = make(map[string]interface{})
	} else if iter.index < len(iter.sequentialParameters) {
		selectedParameters = iter.sequentialParameters[iter.index]
	} else {
		// loop back to the first sequential parameter
		index := iter.index % len(iter.sequentialParameters)
		selectedParameters = iter.sequentialParameters[index]
	}

	// merge with random parameters
	for _, paramName := range iter.randomParameterNames {
		randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
		randIndex := randSource.Intn(len(iter.data[paramName]))
		for k, v := range iter.data[paramName][randIndex] {
			selectedParameters[k] = v
		}
	}

	iter.index++
	if iter.limit > 0 && iter.index >= iter.limit {
		iter.hasNext = false
	}

	return selectedParameters
}

func (iter *ParametersIterator) outParameters() map[string]interface{} {
	res := map[string]interface{}{}
	for key, params := range iter.data {
		res[key] = params
	}
	return res
}

func genCartesianProduct(multiParameters []Parameters) Parameters {
	if len(multiParameters) == 0 {
		return nil
	}

	cartesianProduct := multiParameters[0]
	for i := 0; i < len(multiParameters)-1; i++ {
		var tempProduct Parameters
		for _, param1 := range cartesianProduct {
			for _, param2 := range multiParameters[i+1] {
				tempProduct = append(tempProduct, mergeVariables(param1, param2))
			}
		}
		cartesianProduct = tempProduct
	}

	return cartesianProduct
}

/*
	loadParameters loads parameters from multiple sources.

parameter value may be in three types:

	(1) data list, e.g. ["iOS/10.1", "iOS/10.2", "iOS/10.3"]
	(2) call built-in parameterize function, "${parameterize(account.csv)}"
	(3) call custom function in debugtalk.py, "${gen_app_version()}"

	configParameters = {
		"user_agent": ["iOS/10.1", "iOS/10.2", "iOS/10.3"],		// case 1
		"username-password": "${parameterize(account.csv)}", 	// case 2
		"app_version": "${gen_app_version()}", 					// case 3
	}

=>

	{
		"user_agent": [
			{"user_agent": "iOS/10.1"},
			{"user_agent": "iOS/10.2"},
			{"user_agent": "iOS/10.3"},
		],
		"username-password": [
			{"username": "test1", "password": "111111"},
			{"username": "test2", "password": "222222"},
		],
		"app_version": [
			{"app_version": "1.0.0"},
			{"app_version": "1.0.1"},
		]
	}
*/
func (p *Parser) loadParameters(configParameters map[string]interface{}, variablesMapping map[string]interface{}) (
	map[string]Parameters, error) {

	if len(configParameters) == 0 {
		return nil, nil
	}

	parsedParameters := make(map[string]Parameters)

	for k, v := range configParameters {
		var parametersRawList interface{}
		rawValue := reflect.ValueOf(v)

		switch rawValue.Kind() {
		case reflect.Slice:
			// case 1
			// e.g. user_agent: ["iOS/10.1", "iOS/10.2"]
			// => ["iOS/10.1", "iOS/10.2"]
			parametersRawList = rawValue.Interface()

		case reflect.String:
			// case 2 or case 3
			// e.g. username-password: ${parameterize(examples/hrp/account.csv)}
			// => [{"username": "test1", "password": "111111"}, {"username": "test2", "password": "222222"}]
			// => [["test1", "111111"], ["test2", "222222"]]
			// e.g. "app_version": "${gen_app_version()}"
			// => ["1.0.0", "1.0.1"]
			parsedParameterContent, err := p.ParseString(rawValue.String(), variablesMapping)
			if err != nil {
				log.Error().Err(err).
					Str("parametersRawContent", rawValue.String()).
					Msg("parse parameters content failed")
				return nil, err
			}

			parsedParameterRawValue := reflect.ValueOf(parsedParameterContent)
			if parsedParameterRawValue.Kind() != reflect.Slice {
				log.Error().
					Interface("parsedParameterContent", parsedParameterRawValue).
					Msg("parsed parameters content is not slice")
				return nil, errors.New("parsed parameters content should be slice")
			}
			parametersRawList = parsedParameterRawValue.Interface()

		default:
			log.Error().
				Interface("parameters", configParameters).
				Msg("config parameters raw value should be slice or string (functions call)")
			return nil, errors.New("config parameters raw value format error")
		}

		parameterSlice, err := convertParameters(k, parametersRawList)
		if err != nil {
			return nil, err
		}
		parsedParameters[k] = parameterSlice
	}
	return parsedParameters, nil
}

/*
	convert parameters to standard format

key and parametersRawList may be in three types:

case 1:

	key = "user_agent"
	parametersRawList = ["iOS/10.1", "iOS/10.2"]

case 2:

	key = "username-password"
	parametersRawList = [{"username": "test1", "password": "111111"}, {"username": "test2", "password": "222222"}]

case 3:

	key = "username-password"
	parametersRawList = [["test1", "111111"], ["test2", "222222"]]
*/
func convertParameters(key string, parametersRawList interface{}) (parameterSlice []map[string]interface{}, err error) {
	parametersRawSlice := reflect.ValueOf(parametersRawList)
	if parametersRawSlice.Kind() != reflect.Slice {
		return nil, errors.New("parameters raw value is not list")
	}

	// ["user_agent"], ["username", "password"], ["app_version"]
	parameterNames := strings.Split(key, "-")

	for i := 0; i < parametersRawSlice.Len(); i++ {
		parametersLine := make(map[string]interface{})
		elem := parametersRawSlice.Index(i)
		// e.g. Type: interface{} | []interface{}, convert interface{} to []interface{}
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}
		switch elem.Kind() {
		case reflect.Slice:
			// case 3
			// e.g. "username-password": ["test1", "111111"]
			// => {"username": "test1", "password": "111111"}
			if len(parameterNames) != elem.Len() {
				log.Error().
					Strs("parameterNames", parameterNames).
					Int("lineIndex", i).
					Interface("content", elem.Interface()).
					Msg("parameters line length does not match to names length")
				return nil, errors.New("parameters line length does not match to names length")
			}

			for j := 0; j < elem.Len(); j++ {
				parametersLine[parameterNames[j]] = elem.Index(j).Interface()
			}

		case reflect.Map:
			// case 2
			// e.g. "username-password": {"username": "test1", "password": "111111", "other": "111"}
			// => {"username": "test1", "password": "passwd1"}
			for _, name := range parameterNames {
				lineMap := elem.Interface().(map[string]interface{})
				if _, ok := lineMap[name]; ok {
					parametersLine[name] = elem.MapIndex(reflect.ValueOf(name)).Interface()
				} else {
					log.Error().
						Strs("parameterNames", parameterNames).
						Str("name", name).
						Msg("parameter name not found")
					return nil, errors.New("parameter name not found")
				}
			}

		default:
			// case 1
			// e.g. "user_agent": "iOS/10.1"
			// -> {"user_agent": "iOS/10.1"}
			if len(parameterNames) != 1 {
				log.Error().
					Strs("parameterNames", parameterNames).
					Int("lineIndex", i).
					Msg("parameters format error")
				return nil, errors.New("parameters format error")
			}
			parametersLine[parameterNames[0]] = elem.Interface()
		}
		parameterSlice = append(parameterSlice, parametersLine)
	}
	return parameterSlice, nil
}
