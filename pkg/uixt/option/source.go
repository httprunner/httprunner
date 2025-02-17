package option

import "strings"

func NewSourceOptions(opts ...SourceOption) *SourceOptions {
	options := &SourceOptions{}
	for _, option := range opts {
		option(options)
	}
	return options
}

type SourceOptions struct {
	Format             SourceFormat `json:"format,omitempty"`
	ProcessName        string       `json:"processName,omitempty"`
	Scope              string       `json:"scope,omitempty"`
	ExcludedAttributes string       `json:"excluded_attributes,omitempty"`
}

func (o *SourceOptions) Query() string {
	query := []string{}
	if o.Format != "" {
		query = append(query, "format="+string(o.Format))
	}
	if o.ProcessName != "" {
		query = append(query, "processName="+o.ProcessName)
	}
	if o.Scope != "" {
		query = append(query, "scope="+o.Scope)
	}
	if o.ExcludedAttributes != "" {
		query = append(query, "excluded_attributes="+o.ExcludedAttributes)
	}
	return strings.Join(query, "&")
}

type SourceOption func(o *SourceOptions)

type SourceFormat string

const (
	SourceFormatJSON        SourceFormat = "json"
	SourceFormatXML         SourceFormat = "xml"
	SourceFormatDescription SourceFormat = "description"
)

// WithFormat specify Application elements tree format
// `json` or `xml` or `description`
func WithFormat(format SourceFormat) SourceOption {
	return func(o *SourceOptions) {
		o.Format = format
	}
}

func WithProcessName(name string) SourceOption {
	return func(o *SourceOptions) {
		o.ProcessName = name
	}
}

// WithSourceScope Allows to provide XML scope.
// only `xml` is supported.
func WithSourceScope(scope string) SourceOption {
	return func(o *SourceOptions) {
		o.Scope = scope
	}
}

// WithExcludedAttributes Excludes the given attribute names.
// only `xml` is supported.
func WithExcludedAttributes(attributes []string) SourceOption {
	return func(o *SourceOptions) {
		o.ExcludedAttributes = strings.Join(attributes, ",")
	}
}
