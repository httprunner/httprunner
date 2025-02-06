package option

import "github.com/httprunner/funplugin"

type DriverOptions struct {
	Capabilities     Capabilities
	Plugin           funplugin.IPlugin
	WithImageService bool
	WithResultFolder bool
	WithUIAction     bool
}

func NewDriverOptions(opts ...DriverOption) *DriverOptions {
	driverOptions := &DriverOptions{
		WithImageService: true,
		WithResultFolder: true,
		WithUIAction:     true,
	}
	for _, option := range opts {
		option(driverOptions)
	}
	return driverOptions
}

type DriverOption func(*DriverOptions)

func WithDriverCapabilities(capabilities Capabilities) DriverOption {
	return func(options *DriverOptions) {
		options.Capabilities = capabilities
	}
}

func WithDriverImageService(withImageService bool) DriverOption {
	return func(options *DriverOptions) {
		options.WithImageService = withImageService
	}
}

func WithDriverResultFolder(withResultFolder bool) DriverOption {
	return func(options *DriverOptions) {
		options.WithResultFolder = withResultFolder
	}
}

func WithUIAction(withUIAction bool) DriverOption {
	return func(options *DriverOptions) {
		options.WithUIAction = withUIAction
	}
}

func WithDriverPlugin(plugin funplugin.IPlugin) DriverOption {
	return func(options *DriverOptions) {
		options.Plugin = plugin
	}
}
