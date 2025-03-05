package option

type InstallOptions struct {
	Reinstall       bool
	GrantPermission bool
	Downgrade       bool
	RetryTimes      int
}

type InstallOption func(o *InstallOptions)

func NewInstallOptions(opts ...InstallOption) *InstallOptions {
	installOptions := &InstallOptions{}
	for _, option := range opts {
		option(installOptions)
	}
	return installOptions
}

func WithReinstall(reinstall bool) InstallOption {
	return func(o *InstallOptions) {
		o.Reinstall = reinstall
	}
}

func WithGrantPermission(grantPermission bool) InstallOption {
	return func(o *InstallOptions) {
		o.GrantPermission = grantPermission
	}
}

func WithDowngrade(downgrade bool) InstallOption {
	return func(o *InstallOptions) {
		o.Downgrade = downgrade
	}
}

func WithRetryTimes(retryTimes int) InstallOption {
	return func(o *InstallOptions) {
		o.RetryTimes = retryTimes
	}
}
