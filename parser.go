package hrp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/maja42/goval"
	log "github.com/sirupsen/logrus"

	"github.com/httprunner/hrp/builtin"
)

func parseStep(step IStep, config *TConfig) *TStep {
	tStep := step.ToStruct()
	tStep.Request.URL = buildURL(config.BaseURL, tStep.Request.URL)
	return tStep
}

func buildURL(baseURL, stepURL string) string {
	uConfig, err := url.Parse(baseURL)
	if err != nil {
		log.Errorf("[buildURL] baseURL: %v, error: %v", baseURL, err)
		return ""
	}

	uStep, err := uConfig.Parse(stepURL)
	if err != nil {
		log.Errorf("[buildURL] baseURL: %v, error: %v", baseURL, err)
		return ""
	}

	// base url missed
	return uStep.String()
}

func parseHeaders(rawHeaders map[string]string, variablesMapping map[string]interface{}) (map[string]string, error) {
	parsedHeaders := make(map[string]string)
	headers, err := parseData(rawHeaders, variablesMapping)
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

func parseData(raw interface{}, variablesMapping map[string]interface{}) (interface{}, error) {
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.String:
		// json.Number
		if rawValue, ok := raw.(json.Number); ok {
			return parseJSONNumber(rawValue)
		}
		// other string
		value := rawValue.String()
		value = strings.TrimSpace(value)
		return parseString(value, variablesMapping)
	case reflect.Slice:
		parsedSlice := make([]interface{}, rawValue.Len())
		for i := 0; i < rawValue.Len(); i++ {
			parsedValue, err := parseData(rawValue.Index(i).Interface(), variablesMapping)
			if err != nil {
				return raw, err
			}
			parsedSlice[i] = parsedValue
		}
		return parsedSlice, nil
	case reflect.Map: // convert any map to map[string]interface{}
		parsedMap := make(map[string]interface{})
		for _, k := range rawValue.MapKeys() {
			parsedKey, err := parseString(k.String(), variablesMapping)
			if err != nil {
				return raw, err
			}
			v := rawValue.MapIndex(k)
			parsedValue, err := parseData(v.Interface(), variablesMapping)
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

func parseJSONNumber(raw json.Number) (interface{}, error) {
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
func parseString(raw string, variablesMapping map[string]interface{}) (interface{}, error) {
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
			parsedArgs, err := parseData(arguments, variablesMapping)
			if err != nil {
				return raw, err
			}

			result, err := callFunc(funcName, parsedArgs.([]interface{})...)
			if err != nil {
				return raw, err
			}

			if funcMatched[0] == raw {
				// raw_string is a function, e.g. "${add_one(3)}", return its eval value directly
				return result, nil
			}

			// raw_string contains one or many functions, e.g. "abc${add_one(3)}def"
			matchStartPosition += len(funcMatched[0])
			parsedString += fmt.Sprintf("%v", result)
			remainedString = raw[matchStartPosition:]
			log.Infof("[parseString] parsedString: %v, matchStartPosition: %v", parsedString, matchStartPosition)
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
			log.Infof("[parseString] parsedString: %v, matchStartPosition: %v", parsedString, matchStartPosition)
			continue
		}

		parsedString += remainedString
		break
	}

	return parsedString, nil
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

// callFunc call function with arguments
// only support return at most one result value
func callFunc(funcName string, arguments ...interface{}) (interface{}, error) {
	function, ok := builtin.Functions[funcName]
	if !ok {
		// function not found
		return nil, fmt.Errorf("function %s is not found", funcName)
	}

	funcValue := reflect.ValueOf(function)
	if funcValue.Kind() != reflect.Func {
		// function not valid
		return nil, fmt.Errorf("function %s is invalid", funcName)
	}

	if funcValue.Type().NumIn() != len(arguments) {
		// function arguments not match
		return nil, fmt.Errorf("function %s arguments number not match", funcName)
	}

	argumentsValue := make([]reflect.Value, len(arguments))
	for index, argument := range arguments {
		argumentValue := reflect.ValueOf(argument)
		expectArgumentType := funcValue.Type().In(index)
		actualArgumentType := reflect.TypeOf(argument)

		// type match
		if expectArgumentType == actualArgumentType {
			argumentsValue[index] = argumentValue
			continue
		}

		// type not match, check if convertible
		if !actualArgumentType.ConvertibleTo(expectArgumentType) {
			// function argument type not match and not convertible
			err := fmt.Errorf("function %s argument %d type is neither match nor convertible, expect %v, actual %v",
				funcName, index, expectArgumentType, actualArgumentType)
			log.Errorf("[callFunction] error: %s", err.Error())
			return nil, err
		}
		// convert argument to expect type
		argumentsValue[index] = argumentValue.Convert(expectArgumentType)
	}

	log.Infof("[callFunction] func: %v, input arguments: %v", funcName, arguments)
	resultValues := funcValue.Call(argumentsValue)
	log.Infof("[callFunction] output values: %v", resultValues)

	if len(resultValues) > 1 {
		// function should return at most one value
		err := fmt.Errorf("function %s should return at most one value", funcName)
		log.Errorf("[callFunction] error: %s", err.Error())
		return nil, err
	}

	// no return value
	if len(resultValues) == 0 {
		return nil, nil
	}

	// return one value
	// convert reflect.Value to interface{}
	result := resultValues[0].Interface()
	log.Infof("[callFunction] output result: %+v(%T)", result, result)
	return result, nil
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
		log.Errorf("[literalEval] eval %s error: %s", raw, err.Error())
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

func parseVariables(variables map[string]interface{}) (map[string]interface{}, error) {
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
				log.Errorf("[parseVariables] variable self reference error: %v", variables)
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
				log.Errorf("[parseVariables] variable not defined error: %v", undefinedVars)
				return variables, fmt.Errorf("variable not defined: %v", undefinedVars)
			}

			parsedValue, err := parseData(varValue, parsedVariables)
			if err != nil {
				continue
			}
			parsedVariables[varName] = parsedValue
		}
		traverseRounds += 1
		// check if circular reference exists
		if traverseRounds > len(variables) {
			log.Errorf("[parseVariables] circular reference error, break infinite loop!")
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
