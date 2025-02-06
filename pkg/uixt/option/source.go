package option

import "strings"

// SourceOption Configure the format or attribute of the Source
type SourceOption map[string]interface{}

func NewSourceOption() SourceOption {
	return make(SourceOption)
}

// WithFormatAsJson Application elements tree in form of json string
func (opt SourceOption) WithFormatAsJson() SourceOption {
	opt["format"] = "json"
	return opt
}

func (opt SourceOption) WithProcessName(processName string) SourceOption {
	opt["processName"] = processName
	return opt
}

// WithFormatAsXml Application elements tree in form of xml string
func (opt SourceOption) WithFormatAsXml() SourceOption {
	opt["format"] = "xml"
	return opt
}

// WithFormatAsDescription Application elements tree in form of internal XCTest debugDescription string
func (opt SourceOption) WithFormatAsDescription() SourceOption {
	opt["format"] = "description"
	return opt
}

// WithScope Allows to provide XML scope.
//
//	only `xml` is supported.
func (opt SourceOption) WithScope(scope string) SourceOption {
	if vFormat, ok := opt["format"]; ok && vFormat != "xml" {
		return opt
	}
	opt["scope"] = scope
	return opt
}

// WithExcludedAttributes Excludes the given attribute names.
// only `xml` is supported.
func (opt SourceOption) WithExcludedAttributes(attributes []string) SourceOption {
	if vFormat, ok := opt["format"]; ok && vFormat != "xml" {
		return opt
	}
	opt["excluded_attributes"] = strings.Join(attributes, ",")
	return opt
}
