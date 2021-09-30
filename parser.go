package httpboomer

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"
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
	for k, v := range rawHeaders {
		parsedValue := parseString(v, variablesMapping)
		if value, ok := parsedValue.(string); ok {
			parsedHeaders[k] = value
		} else {
			// parsed value is not string, e.g. int, float, etc.
			// convert to string
			parsedHeaders[k] = fmt.Sprintf("%v", parsedValue)
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
	regexVariable = `[a-zA-Z_]\w*` // variable name should start with a letter or underscore
)

var (
	regexCompileVariable = regexp.MustCompile(fmt.Sprintf(`\$\{(%s)\}|\$(%s)`, regexVariable, regexVariable)) // parse ${var} or $var
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
