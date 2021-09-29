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
	regexCompileVariable = regexp.MustCompile(`\$\{(\w+)\}|\$(\w+)`) // parse ${var} or $var
)

// parseString parse string with variables
func parseString(raw string, variablesMapping map[string]interface{}) interface{} {
	matchStartPosition := strings.Index(raw, "$")
	if matchStartPosition == -1 { // no $ found
		return raw
	}
	parsedString := raw[0:matchStartPosition]

	for matchStartPosition < len(raw) {
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

			parsedString += fmt.Sprintf("%v", varValue)
			matchStartPosition += len(varMatched[0])
			log.Printf("[parseString] parsedString: %v, matchStartPosition: %v", parsedString, matchStartPosition)
			continue
		}

		currentPosition := matchStartPosition
		var remainedString string
		// find next $ location
		nextStartPosition := strings.Index(raw[currentPosition+1:], "$")
		if nextStartPosition == -1 { // no $ found
			remainedString = raw[currentPosition:]
			// break loop
			matchStartPosition = len(raw)
		} else { // found next $
			matchStartPosition = nextStartPosition
			remainedString = raw[currentPosition:nextStartPosition]
		}

		// append remained string
		parsedString += remainedString
	}

	return parsedString
}
