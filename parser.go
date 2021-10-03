package httpboomer

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/httprunner/httpboomer/builtin"
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
	regexVariable     = `[a-zA-Z_]\w*` // variable name should start with a letter or underscore
	regexFunctionName = `[a-zA-Z_]\w*` // function name should start with a letter or underscore
)

var (
	regexCompileVariable = regexp.MustCompile(fmt.Sprintf(`\$\{(%s)\}|\$(%s)`, regexVariable, regexVariable))     // parse ${var} or $var
	regexCompileFunction = regexp.MustCompile(fmt.Sprintf(`\$\{(%s)\(([\$\w\.\-/\s=,]*)\)\}`, regexFunctionName)) // parse ${func1($a, $b)}
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

		// search $$, use $$ to escape $ notation
		if strings.HasPrefix(remainedString, "$$") { // found $$
			matchStartPosition += 2
			parsedString += "$"
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

// callFunction call function with arguments
// only support return at most one result value
func callFunction(funcName string, arguments []interface{}) (interface{}, error) {
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

	argumentsValue := make([]reflect.Value, len(arguments))
	for index, argument := range arguments {
		argumentsValue[index] = reflect.ValueOf(argument)
	}

	log.Printf("[callFunction] func: %v, input arguments: %v", funcName, arguments)
	resultValues := funcValue.Call(argumentsValue)
	log.Printf("[callFunction] output results: %v", resultValues)

	if len(resultValues) == 0 {
		return nil, nil
	} else if len(resultValues) > 1 {
		// function should return at most one value
		err := fmt.Errorf("function %s should return at most one value", funcName)
		log.Printf("[callFunction] error: %s", err.Error())
		return nil, err
	}

	// convert reflect.Value to interface{}
	result := resultValues[0].Interface()
	return result, nil
}
