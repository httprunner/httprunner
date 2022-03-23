package hrp

import (
	builtinJSON "encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/maja42/goval"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/funplugin"
	"github.com/httprunner/funplugin/shared"
	"github.com/httprunner/hrp/internal/builtin"
)

func newParser() *parser {
	return &parser{}
}

type parser struct {
	plugin funplugin.IPlugin // plugin is used to call functions
}

func buildURL(baseURL, stepURL string) string {
	uConfig, err := url.Parse(baseURL)
	if err != nil {
		log.Error().Str("baseURL", baseURL).Err(err).Msg("[buildURL] parse baseURL failed")
		return ""
	}

	uStep, err := uConfig.Parse(stepURL)
	if err != nil {
		log.Error().Str("stepURL", stepURL).Err(err).Msg("[buildURL] parse stepURL failed")
		return ""
	}

	// base url missed
	return uStep.String()
}

func (p *parser) parseHeaders(rawHeaders map[string]string, variablesMapping map[string]interface{}) (map[string]string, error) {
	parsedHeaders := make(map[string]string)
	headers, err := p.parseData(rawHeaders, variablesMapping)
	if err != nil {
		return rawHeaders, err
	}
	for k, v := range headers.(map[string]interface{}) {
		parsedHeaders[k] = convertString(v)
	}
	return parsedHeaders, nil
}

func convertString(raw interface{}) string {
	if value, ok := raw.(string); ok {
		return value
	} else {
		// raw is not string, e.g. int, float, etc.
		// convert to string
		return fmt.Sprintf("%v", raw)
	}
}

func (p *parser) parseData(raw interface{}, variablesMapping map[string]interface{}) (interface{}, error) {
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.String:
		// json.Number
		if rawValue, ok := raw.(builtinJSON.Number); ok {
			return parseJSONNumber(rawValue)
		}
		// other string
		value := rawValue.String()
		value = strings.TrimSpace(value)
		return p.parseString(value, variablesMapping)
	case reflect.Slice:
		parsedSlice := make([]interface{}, rawValue.Len())
		for i := 0; i < rawValue.Len(); i++ {
			parsedValue, err := p.parseData(rawValue.Index(i).Interface(), variablesMapping)
			if err != nil {
				return raw, err
			}
			parsedSlice[i] = parsedValue
		}
		return parsedSlice, nil
	case reflect.Map: // convert any map to map[string]interface{}
		parsedMap := make(map[string]interface{})
		for _, k := range rawValue.MapKeys() {
			parsedKey, err := p.parseString(k.String(), variablesMapping)
			if err != nil {
				return raw, err
			}
			v := rawValue.MapIndex(k)
			parsedValue, err := p.parseData(v.Interface(), variablesMapping)
			if err != nil {
				return raw, err
			}

			key := convertString(parsedKey)
			parsedMap[key] = parsedValue
		}
		return parsedMap, nil
	default:
		// other types, e.g. nil, int, float, bool
		return raw, nil
	}
}

func parseJSONNumber(raw builtinJSON.Number) (interface{}, error) {
	if strings.Contains(raw.String(), ".") {
		// float64
		return raw.Float64()
	} else {
		// int64
		return raw.Int64()
	}
}

const (
	regexVariable     = `[a-zA-Z_]\w*`    // variable name should start with a letter or underscore
	regexFunctionName = `[a-zA-Z_]\w*`    // function name should start with a letter or underscore
	regexNumber       = `^-?\d+(\.\d+)?$` // match number, e.g. 123, -123, 1.23, -1.23
)

var (
	regexCompileVariable = regexp.MustCompile(fmt.Sprintf(`\$\{(%s)\}|\$(%s)`, regexVariable, regexVariable))     // parse ${var} or $var
	regexCompileFunction = regexp.MustCompile(fmt.Sprintf(`\$\{(%s)\(([\$\w\.\-/\s=,]*)\)\}`, regexFunctionName)) // parse ${func1($a, $b)}
	regexCompileNumber   = regexp.MustCompile(regexNumber)                                                        // parse number
)

// parseString parse string with variables
func (p *parser) parseString(raw string, variablesMapping map[string]interface{}) (interface{}, error) {
	matchStartPosition := 0
	parsedString := ""
	remainedString := raw

	for matchStartPosition < len(raw) {
		// locate $ char position
		startPosition := strings.Index(remainedString, "$")
		if startPosition == -1 { // no $ found
			// append remained string
			parsedString += remainedString
			break
		}

		// found $, check if variable or function
		matchStartPosition += startPosition
		parsedString += remainedString[0:startPosition]
		remainedString = remainedString[startPosition:]

		// Notice: notation priority
		// $$ > ${func($a, $b)} > $var

		// search $$, use $$ to escape $ notation
		if strings.HasPrefix(remainedString, "$$") { // found $$
			matchStartPosition += 2
			parsedString += "$"
			remainedString = remainedString[2:]
			continue
		}

		// search function like ${func($a, $b)}
		funcMatched := regexCompileFunction.FindStringSubmatch(remainedString)
		if len(funcMatched) == 3 {
			funcName := funcMatched[1]
			argsStr := funcMatched[2]
			arguments, err := parseFunctionArguments(argsStr)
			if err != nil {
				return raw, err
			}
			parsedArgs, err := p.parseData(arguments, variablesMapping)
			if err != nil {
				return raw, err
			}

			result, err := p.callFunc(funcName, parsedArgs.([]interface{})...)
			if err != nil {
				log.Error().Str("funcName", funcName).Interface("arguments", arguments).
					Err(err).Msg("call function failed")
				return raw, err
			}
			log.Info().Str("funcName", funcName).Interface("arguments", arguments).
				Interface("output", result).Msg("call function success")

			if funcMatched[0] == raw {
				// raw_string is a function, e.g. "${add_one(3)}", return its eval value directly
				return result, nil
			}

			// raw_string contains one or many functions, e.g. "abc${add_one(3)}def"
			matchStartPosition += len(funcMatched[0])
			parsedString += fmt.Sprintf("%v", result)
			remainedString = raw[matchStartPosition:]
			log.Debug().
				Str("parsedString", parsedString).
				Int("matchStartPosition", matchStartPosition).
				Msg("[parseString] parse function")
			continue
		}

		// search variable like ${var} or $var
		varMatched := regexCompileVariable.FindStringSubmatch(remainedString)
		if len(varMatched) == 3 {
			var varName string
			if varMatched[1] != "" {
				varName = varMatched[1] // match ${var}
			} else {
				varName = varMatched[2] // match $var
			}
			varValue, ok := variablesMapping[varName]
			if !ok {
				return raw, fmt.Errorf("variable %s not found", varName)
			}

			if fmt.Sprintf("${%s}", varName) == raw || fmt.Sprintf("$%s", varName) == raw {
				// raw string is a variable, $var or ${var}, return its value directly
				return varValue, nil
			}

			matchStartPosition += len(varMatched[0])
			parsedString += fmt.Sprintf("%v", varValue)
			remainedString = raw[matchStartPosition:]
			log.Debug().
				Str("parsedString", parsedString).
				Int("matchStartPosition", matchStartPosition).
				Msg("[parseString] parse variable")
			continue
		}

		parsedString += remainedString
		break
	}

	return parsedString, nil
}

// callFunc calls function with arguments
// only support return at most one result value
func (p *parser) callFunc(funcName string, arguments ...interface{}) (interface{}, error) {
	// call with plugin function
	if p.plugin != nil && p.plugin.Has(funcName) {
		return p.plugin.Call(funcName, arguments...)
	}

	// get builtin function
	function, ok := builtin.Functions[funcName]
	if !ok {
		return nil, fmt.Errorf("function %s is not found", funcName)
	}
	fn := reflect.ValueOf(function)

	// call with builtin function
	return shared.CallFunc(fn, arguments...)
}

// merge two variables mapping, the first variables have higher priority
func mergeVariables(variables, overriddenVariables map[string]interface{}) map[string]interface{} {
	if overriddenVariables == nil {
		return variables
	}
	if variables == nil {
		return overriddenVariables
	}

	mergedVariables := make(map[string]interface{})
	for k, v := range overriddenVariables {
		mergedVariables[k] = v
	}
	for k, v := range variables {
		if fmt.Sprintf("${%s}", k) == v || fmt.Sprintf("$%s", k) == v {
			// e.g. {"base_url": "$base_url"}
			// or {"base_url": "${base_url}"}
			continue
		}

		mergedVariables[k] = v
	}
	return mergedVariables
}

// merge two map, the first map have higher priority
func mergeMap(m, overriddenMap map[string]string) map[string]string {
	if overriddenMap == nil {
		return m
	}
	if m == nil {
		return overriddenMap
	}

	mergedMap := make(map[string]string)
	for k, v := range overriddenMap {
		mergedMap[k] = v
	}
	for k, v := range m {
		mergedMap[k] = v
	}
	return mergedMap
}

// merge two validators slice, the first validators have higher priority
func mergeValidators(validators, overriddenValidators []interface{}) []interface{} {
	if validators == nil {
		return overriddenValidators
	}
	if overriddenValidators == nil {
		return validators
	}
	var mergedValidators []interface{}
	validators = append(validators, overriddenValidators...)
	for _, validator := range validators {
		flag := true
		for _, mergedValidator := range mergedValidators {
			if validator.(Validator).Check == mergedValidator.(Validator).Check {
				flag = false
				break
			}
		}
		if flag {
			mergedValidators = append(mergedValidators, validator)
		}
	}
	return mergedValidators
}

// merge two slices, the first slice have higher priority
func mergeSlices(slice, overriddenSlice []string) []string {
	if slice == nil {
		return overriddenSlice
	}
	if overriddenSlice == nil {
		return slice
	}

	for _, value := range overriddenSlice {
		if !builtin.Contains(slice, value) {
			slice = append(slice, value)
		}
	}
	return slice
}

// extend teststep with api, teststep will merge and override referenced api
func extendWithAPI(testStep *TStep, overriddenStep *API) {
	// override api name
	if testStep.Name == "" {
		testStep.Name = overriddenStep.Name
	}
	// merge & override request
	testStep.Request = overriddenStep.Request
	// merge & override variables
	testStep.Variables = mergeVariables(testStep.Variables, overriddenStep.Variables)
	// merge & override extractors
	testStep.Extract = mergeMap(testStep.Extract, overriddenStep.Extract)
	// merge & override validators
	testStep.Validators = mergeValidators(testStep.Validators, overriddenStep.Validators)
	// merge & override setupHooks
	testStep.SetupHooks = mergeSlices(testStep.SetupHooks, overriddenStep.SetupHooks)
	// merge & override teardownHooks
	testStep.TeardownHooks = mergeSlices(testStep.TeardownHooks, overriddenStep.TeardownHooks)
}

// extend referenced testcase with teststep, teststep config merge and override referenced testcase config
func extendWithTestCase(testStep *TStep, overriddenTestCase *TestCase) {
	// override testcase name
	if testStep.Name != "" {
		overriddenTestCase.Config.Name = testStep.Name
	}
	// merge & override variables
	overriddenTestCase.Config.Variables = mergeVariables(testStep.Variables, overriddenTestCase.Config.Variables)
	// merge & override extractors
	overriddenTestCase.Config.Export = mergeSlices(testStep.Export, overriddenTestCase.Config.Export)
}

var eval = goval.NewEvaluator()

// literalEval parse string to number if possible
func literalEval(raw string) (interface{}, error) {
	raw = strings.TrimSpace(raw)

	// return raw string if not number
	if !regexCompileNumber.Match([]byte(raw)) {
		return raw, nil
	}

	// eval string to number
	result, err := eval.Evaluate(raw, nil, nil)
	if err != nil {
		log.Error().Err(err).Msgf("[literalEval] eval %s failed", raw)
		return raw, err
	}
	return result, nil
}

func parseFunctionArguments(argsStr string) ([]interface{}, error) {
	argsStr = strings.TrimSpace(argsStr)
	if argsStr == "" {
		return []interface{}{}, nil
	}

	// split arguments by comma
	args := strings.Split(argsStr, ",")
	arguments := make([]interface{}, len(args))
	for index, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}

		// parse argument to number if possible
		arg, err := literalEval(arg)
		if err != nil {
			return nil, err
		}
		arguments[index] = arg
	}

	return arguments, nil
}

func (p *parser) parseVariables(variables map[string]interface{}) (map[string]interface{}, error) {
	parsedVariables := make(map[string]interface{})
	var traverseRounds int

	for len(parsedVariables) != len(variables) {
		for varName, varValue := range variables {
			// skip parsed variables
			if _, ok := parsedVariables[varName]; ok {
				continue
			}

			// extract variables from current value
			extractVarsSet := extractVariables(varValue)

			// check if reference variable itself
			// e.g.
			// variables = {"token": "abc$token"}
			// variables = {"key": ["$key", 2]}
			if _, ok := extractVarsSet[varName]; ok {
				log.Error().Interface("variables", variables).Msg("[parseVariables] variable self reference error")
				return variables, fmt.Errorf("variable self reference: %v", varName)
			}

			// check if reference variable not in variables mapping
			// e.g.
			// {"varA": "123$varB", "varB": "456$varC"} => $varC not defined
			// {"varC": "${sum_two($a, $b)}"} => $a, $b not defined
			var undefinedVars []string
			for extractVar := range extractVarsSet {
				if _, ok := variables[extractVar]; !ok { // not in variables mapping
					undefinedVars = append(undefinedVars, extractVar)
				}
			}
			if len(undefinedVars) > 0 {
				log.Error().Interface("undefinedVars", undefinedVars).Msg("[parseVariables] variable not defined error")
				return variables, fmt.Errorf("variable not defined: %v", undefinedVars)
			}

			parsedValue, err := p.parseData(varValue, parsedVariables)
			if err != nil {
				continue
			}
			parsedVariables[varName] = parsedValue
		}
		traverseRounds += 1
		// check if circular reference exists
		if traverseRounds > len(variables) {
			log.Error().Msg("[parseVariables] circular reference error, break infinite loop!")
			return variables, fmt.Errorf("circular reference")
		}
	}

	return parsedVariables, nil
}

type variableSet map[string]struct{}

func extractVariables(raw interface{}) variableSet {
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.String:
		return findallVariables(rawValue.String())
	case reflect.Slice:
		varSet := make(variableSet)
		for i := 0; i < rawValue.Len(); i++ {
			for extractVar := range extractVariables(rawValue.Index(i).Interface()) {
				varSet[extractVar] = struct{}{}
			}
		}
		return varSet
	case reflect.Map:
		varSet := make(variableSet)
		for _, key := range rawValue.MapKeys() {
			value := rawValue.MapIndex(key)
			for extractVar := range extractVariables(value.Interface()) {
				varSet[extractVar] = struct{}{}
			}
		}
		return varSet
	default:
		// other types, e.g. nil, int, float, bool
		return make(variableSet)
	}
}

func findallVariables(raw string) variableSet {
	matchStartPosition := 0
	remainedString := raw
	varSet := make(variableSet)

	for matchStartPosition < len(raw) {
		// locate $ char position
		startPosition := strings.Index(remainedString, "$")
		if startPosition == -1 { // no $ found
			return varSet
		}

		// found $, check if variable or function
		matchStartPosition += startPosition
		remainedString = remainedString[startPosition:]

		// Notice: notation priority
		// $$ > $var

		// search $$, use $$ to escape $ notation
		if strings.HasPrefix(remainedString, "$$") { // found $$
			matchStartPosition += 2
			remainedString = remainedString[2:]
			continue
		}

		// search variable like ${var} or $var
		varMatched := regexCompileVariable.FindStringSubmatch(remainedString)
		if len(varMatched) == 3 {
			var varName string
			if varMatched[1] != "" {
				varName = varMatched[1] // match ${var}
			} else {
				varName = varMatched[2] // match $var
			}
			varSet[varName] = struct{}{}

			matchStartPosition += len(varMatched[0])
			remainedString = raw[matchStartPosition:]
			continue
		}

		break
	}

	return varSet
}

func genCartesianProduct(paramsMap map[string]paramsType) paramsType {
	if len(paramsMap) == 0 {
		return nil
	}
	var params []paramsType
	for _, v := range paramsMap {
		params = append(params, v)
	}
	var cartesianProduct paramsType
	cartesianProduct = params[0]
	for i := 0; i < len(params)-1; i++ {
		var tempProduct paramsType
		for _, param1 := range cartesianProduct {
			for _, param2 := range params[i+1] {
				tempProduct = append(tempProduct, mergeVariables(param1, param2))
			}
		}
		cartesianProduct = tempProduct
	}
	return cartesianProduct
}

func parseParameters(parameters map[string]interface{}, variablesMapping map[string]interface{}) (map[string]paramsType, error) {
	if len(parameters) == 0 {
		return nil, nil
	}
	parsedParametersSlice := make(map[string]paramsType)
	var err error
	for k, v := range parameters {
		var parameterSlice paramsType
		rawValue := reflect.ValueOf(v)
		switch rawValue.Kind() {
		case reflect.String:
			// e.g. username-password: ${parameterize(examples/account.csv)} -> [{"username": "test1", "password": "111111"}, {"username": "test2", "password": "222222"}]
			var parsedParameterContent interface{}
			parsedParameterContent, err = newParser().parseString(rawValue.String(), variablesMapping)
			if err != nil {
				log.Error().Interface("parameterContent", rawValue).Msg("[parseParameters] parse parameter content error")
				return nil, err
			}
			parsedParameterRawValue := reflect.ValueOf(parsedParameterContent)
			if parsedParameterRawValue.Kind() != reflect.Slice {
				log.Error().Interface("parameterContent", parsedParameterRawValue).Msg("[parseParameters] parsed parameter content should be slice")
				return nil, errors.New("parsed parameter content should be slice")
			}
			parameterSlice, err = parseSlice(k, parsedParameterRawValue.Interface())
		case reflect.Slice:
			// e.g. user_agent: ["iOS/10.1", "iOS/10.2"] -> [{"user_agent": "iOS/10.1"}, {"user_agent": "iOS/10.2"}]
			parameterSlice, err = parseSlice(k, rawValue.Interface())
		default:
			log.Error().Interface("parameter", parameters).Msg("[parseParameters] parameter content should be slice or text(functions call)")
			return nil, errors.New("parameter content should be slice or text(functions call)")
		}
		if err != nil {
			return nil, err
		}
		parsedParametersSlice[k] = parameterSlice
	}
	return parsedParametersSlice, nil
}

func parseSlice(parameterName string, parameterContent interface{}) ([]map[string]interface{}, error) {
	parameterNameSlice := strings.Split(parameterName, "-")
	var parameterSlice []map[string]interface{}
	parameterContentSlice := reflect.ValueOf(parameterContent)
	if parameterContentSlice.Kind() != reflect.Slice {
		return nil, errors.New("parameterContent should be slice")
	}
	for i := 0; i < parameterContentSlice.Len(); i++ {
		parameterMap := make(map[string]interface{})
		elem := reflect.ValueOf(parameterContentSlice.Index(i).Interface())
		switch elem.Kind() {
		case reflect.Map:
			// e.g. "username-password": [{"username": "test1", "password": "passwd1", "other": "111"}, {"username": "test2", "password": "passwd2", "other": ""222}]
			// -> [{"username": "test1", "password": "passwd1"}, {"username": "test2", "password": "passwd2"}]
			for _, key := range parameterNameSlice {
				if _, ok := elem.Interface().(map[string]interface{})[key]; ok {
					parameterMap[key] = elem.MapIndex(reflect.ValueOf(key)).Interface()
				} else {
					log.Error().Interface("parameterNameSlice", parameterNameSlice).Msg("[parseParameters] parameter name not found")
					return nil, errors.New("parameter name not found")
				}
			}
		case reflect.Slice:
			// e.g. "username-password": [["test1", "passwd1"], ["test2", "passwd2"]]
			// -> [{"username": "test1", "password": "passwd1"}, {"username": "test2", "password": "passwd2"}]
			if len(parameterNameSlice) != elem.Len() {
				log.Error().Interface("parameterNameSlice", parameterNameSlice).Interface("parameterContent", elem.Interface()).Msg("[parseParameters] parameter name slice and parameter content slice should have the same length")
				return nil, errors.New("parameter name slice and parameter content slice should have the same length")
			} else {
				for j := 0; j < elem.Len(); j++ {
					parameterMap[parameterNameSlice[j]] = elem.Index(j).Interface()
				}
			}
		default:
			// e.g. "app_version": [3.1, 3.0]
			// -> [{"app_version": 3.1}, {"app_version": 3.0}]
			if len(parameterNameSlice) != 1 {
				log.Error().Interface("parameterNameSlice", parameterNameSlice).Msg("[parseParameters] parameter name slice should have only one element when parameter content is string")
				return nil, errors.New("parameter name slice should have only one element when parameter content is string")
			}
			parameterMap[parameterNameSlice[0]] = elem.Interface()
		}
		parameterSlice = append(parameterSlice, parameterMap)
	}
	return parameterSlice, nil
}

func initParameterIterator(cfg *TConfig, mode string) (err error) {
	var parameters map[string]paramsType
	parameters, err = parseParameters(cfg.Parameters, cfg.Variables)
	if err != nil {
		return err
	}
	// parse config parameters setting
	if cfg.ParametersSetting == nil {
		cfg.ParametersSetting = &TParamsConfig{Iterators: []*Iterator{}}
	}
	// boomer模式下不限制迭代次数
	if mode == "boomer" {
		cfg.ParametersSetting.Iteration = -1
	}
	rawValue := reflect.ValueOf(cfg.ParametersSetting.Strategy)
	switch rawValue.Kind() {
	case reflect.Map:
		// strategy: {"user_agent": "sequential", "username-password": "random"}, 每个参数对应一个迭代器，每个迭代器随机、顺序选取元素互不影响
		for k, v := range parameters {
			if _, ok := rawValue.Interface().(map[string]interface{})[k]; ok {
				// use strategy if configured
				cfg.ParametersSetting.Iterators = append(
					cfg.ParametersSetting.Iterators,
					newIterator(v, rawValue.MapIndex(reflect.ValueOf(k)).Interface().(string), cfg.ParametersSetting.Iteration),
				)
			} else {
				// use sequential strategy by default
				cfg.ParametersSetting.Iterators = append(
					cfg.ParametersSetting.Iterators,
					newIterator(v, strategySequential, cfg.ParametersSetting.Iteration),
				)
			}
		}
	case reflect.String:
		// strategy: random, 仅生成一个的迭代器，该迭代器在参数笛卡尔积slice中随机选取元素
		if len(rawValue.String()) == 0 {
			cfg.ParametersSetting.Strategy = strategySequential
		} else {
			cfg.ParametersSetting.Strategy = strings.ToLower(rawValue.String())
		}
		cfg.ParametersSetting.Iterators = append(
			cfg.ParametersSetting.Iterators,
			newIterator(genCartesianProduct(parameters), cfg.ParametersSetting.Strategy.(string), cfg.ParametersSetting.Iteration),
		)
	default:
		// default strategy: sequential, 仅生成一个的迭代器，该迭代器在参数笛卡尔积slice中顺序选取元素
		cfg.ParametersSetting.Strategy = strategySequential
		cfg.ParametersSetting.Iterators = append(
			cfg.ParametersSetting.Iterators,
			newIterator(genCartesianProduct(parameters), cfg.ParametersSetting.Strategy.(string), cfg.ParametersSetting.Iteration),
		)
	}
	return nil
}

func newIterator(parameters paramsType, strategy string, iteration int) *Iterator {
	iter := parameters.Iterator()
	iter.strategy = strategy
	if iteration > 0 {
		iter.iteration = iteration
	} else if iteration < 0 {
		iter.iteration = -1
	} else if iter.iteration == 0 {
		iter.iteration = 1
	}
	return iter
}
