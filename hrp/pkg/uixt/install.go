package uixt

type InstallOptions struct {
	Reinstall       bool
	GrantPermission bool
	Downgrade       bool
	RetryTime       int
}

type InstallOption func(o *InstallOptions)

func NewInstallOptions(options ...InstallOption) *InstallOptions {
	installOptions := &InstallOptions{}
	for _, option := range options {
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

func WithRetryTime(retryTime int) InstallOption {
	return func(o *InstallOptions) {
		o.RetryTime = retryTime
	}
}

type InstallResult struct {
	Result    int    `json:"result"`
	ErrorCode int    `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}
