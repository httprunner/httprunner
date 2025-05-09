package option

// HookOptions contains options for action hooks
type HookOptions struct {
	// pre hook before action
	PreHook func()
	// post hook after action
	PostHook func()
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
