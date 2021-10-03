package httpboomer

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/httprunner/httpboomer/builtin"
	"github.com/maja42/goval"
)

func parseStep(step IStep, config *TConfig) *TStep {
	tStep := step.ToStruct()
	tStep.Request.URL = buildURL(config.BaseURL, tStep.Request.URL)
	return tStep
}

func buildURL(baseURL, stepURL string) string {
	uConfig, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("[buildURL] baseURL: %v, error: %v", baseURL, err)
		return ""
	}

	uStep, err := uConfig.Parse(stepURL)
	if err != nil {
		log.Fatalf("[buildURL] baseURL: %v, error: %v", baseURL, err)
		return ""
	}

	// base url missed
	return uStep.String()
}

func parseHeaders(rawHeaders map[string]string, variablesMapping map[string]interface{}) map[string]string {
	parsedHeaders := make(map[string]string)
	headers := parseData(rawHeaders, variablesMapping).(map[string]interface{})
	for k, v := range headers {
		if value, ok := v.(string); ok {
			parsedHeaders[k] = value
		} else {
			parsedHeaders[k] = fmt.Sprintf("%v", v)
		}
	}
	return parsedHeaders
}

func parseData(raw interface{}, variablesMapping map[string]interface{}) interface{} {
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.String:
		value := rawValue.String()
		value = strings.TrimSpace(value)
		return parseString(value, variablesMapping)
	case reflect.Slice:
		parsedSlice := make([]interface{}, rawValue.Len())
		for i := 0; i < rawValue.Len(); i++ {
			parsedSlice[i] = parseData(rawValue.Index(i).Interface(), variablesMapping)
		}
		return parsedSlice
	case reflect.Map: // convert any map to map[string]interface{}
		parsedMap := make(map[string]interface{})
		for _, k := range rawValue.MapKeys() {
			parsedKey := parseString(k.String(), variablesMapping)
			v := rawValue.MapIndex(k)
			parsedValue := parseData(v.Interface(), variablesMapping)

			if key, ok := parsedKey.(string); ok {
				parsedMap[key] = parsedValue
			} else {
				// parsed key is not string, e.g. int, float, etc.
				// convert to string
				parsedMap[fmt.Sprintf("%v", parsedKey)] = parsedValue
			}
		}
		return parsedMap
	default:
		// other types, e.g. nil, int, float, bool
		return raw
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
func parseString(raw string, variablesMapping map[string]interface{}) interface{} {
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
				return raw
			}
			parsedArgs := parseData(arguments, variablesMapping).([]interface{})

			result, err := callFunc(funcName, parsedArgs...)
			if err != nil {
				return raw
			}

			if funcMatched[0] == raw {
				// raw_string is a function, e.g. "${add_one(3)}", return its eval value directly
				return result
			}

			// raw_string contains one or many functions, e.g. "abc${add_one(3)}def"
			matchStartPosition += len(funcMatched[0])
			parsedString += fmt.Sprintf("%v", result)
			remainedString = raw[matchStartPosition:]
			log.Printf("[parseString] parsedString: %v, matchStartPosition: %v", parsedString, matchStartPosition)
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
			varValue := variablesMapping[varName]

			if fmt.Sprintf("${%s}", varName) == raw || fmt.Sprintf("$%s", varName) == raw {
				// raw string is a variable, $var or ${var}, return its value directly
				return varValue
			}

			matchStartPosition += len(varMatched[0])
			parsedString += fmt.Sprintf("%v", varValue)
			remainedString = raw[matchStartPosition:]
			log.Printf("[parseString] parsedString: %v, matchStartPosition: %v", parsedString, matchStartPosition)
			continue
		}

		parsedString += remainedString
		break
	}

	return parsedString
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
		// ensure each argument type match
		expectArgumentType := funcValue.Type().In(index)
		actualArgumentType := reflect.TypeOf(argument)
		if expectArgumentType != actualArgumentType {
			// function argument type not match
			err := fmt.Errorf("function %s argument %d type not match, expect %v, actual %v",
				funcName, index, expectArgumentType, actualArgumentType)
			log.Printf("[callFunction] error: %s", err.Error())
			return nil, err
		}
		argumentsValue[index] = reflect.ValueOf(argument)
	}

	log.Printf("[callFunction] func: %v, input arguments: %v", funcName, arguments)
	resultValues := funcValue.Call(argumentsValue)
	log.Printf("[callFunction] output values: %v", resultValues)

	if len(resultValues) > 1 {
		// function should return at most one value
		err := fmt.Errorf("function %s should return at most one value", funcName)
		log.Printf("[callFunction] error: %s", err.Error())
		return nil, err
	}

	// no return value
	if len(resultValues) == 0 {
		return nil, nil
	}

	// return one value
	// convert reflect.Value to interface{}
	result := resultValues[0].Interface()
	log.Printf("[callFunction] output result: %+v(%T)", result, result)
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
		log.Printf("[literalEval] eval %s error: %s", raw, err.Error())
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
