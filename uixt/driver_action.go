package uixt

import (
	"github.com/httprunner/httprunner/v5/uixt/option"
)

type MobileAction struct {
	Method  option.ActionMethod   `json:"method,omitempty" yaml:"method,omitempty"`
	Params  interface{}           `json:"params,omitempty" yaml:"params,omitempty"`
	Fn      func()                `json:"-" yaml:"-"` // used for function action, not serialized
	Options *option.ActionOptions `json:"options,omitempty" yaml:"options,omitempty"`
	option.ActionOptions
}

func (ma MobileAction) GetOptions() []option.ActionOption {
	var actionOptionList []option.ActionOption
	// Notice: merge options from ma.Options and ma.ActionOptions
	if ma.Options != nil {
		actionOptionList = append(actionOptionList, ma.Options.Options()...)
	}
	actionOptionList = append(actionOptionList, ma.ActionOptions.Options()...)
	return actionOptionList
}
