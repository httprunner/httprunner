package option

// HookOptions contains options for action hooks
type HookOptions struct {
	// pre hook before action
	PreHook func()
	// post hook after action
	PostHook func()
}

func (o *HookOptions) GetHookOptions() []ActionOption {
	options := make([]ActionOption, 0)
	if o == nil {
		return options
	}

	if o.PreHook != nil {
		options = append(options, WithPreHook(o.PreHook))
	}
	if o.PostHook != nil {
		options = append(options, WithPostHook(o.PostHook))
	}

	return options
}

// WithPreHook sets the pre hook before action
func WithPreHook(preHook func()) ActionOption {
	return func(o *ActionOptions) {
		o.PreHook = preHook
	}
}

// WithPostHook sets the post hook after action
func WithPostHook(postHook func()) ActionOption {
	return func(o *ActionOptions) {
		o.PostHook = postHook
	}
}

// WithHooks sets the pre hook and post hook
func WithHooks(preHook func(), postHook func()) ActionOption {
	return func(o *ActionOptions) {
		o.PreHook = preHook
		o.PostHook = postHook
	}
}
