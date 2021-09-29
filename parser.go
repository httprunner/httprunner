package httpboomer

import (
	"fmt"
	"log"
	"net/url"
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

func parseData(raw interface{}, variablesMapping map[string]interface{}) interface{} {
	switch v := raw.(type) {
	case string:
		v = strings.TrimSpace(v)
		return parseString(v, variablesMapping)
	default:
		return raw
	}
}

var (
	regexVariable        = `[a-zA-Z_]\w*`                                                                     // variable name should start with a letter or underscore
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

		// search variable like ${var} or $var
		varMatched := regexCompileVariable.FindStringSubmatch(raw[matchStartPosition:])
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
