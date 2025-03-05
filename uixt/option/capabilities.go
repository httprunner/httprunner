package option

type AlertAction string

const (
	AlertActionAccept  AlertAction = "accept"
	AlertActionDismiss AlertAction = "dismiss"
)

type Capabilities map[string]interface{}

func NewCapabilities() Capabilities {
	return make(Capabilities)
}

// WithDefaultAlertAction
func (caps Capabilities) WithDefaultAlertAction(alertAction AlertAction) Capabilities {
	caps["defaultAlertAction"] = alertAction
	return caps
}

// WithMaxTypingFrequency
//
//	Defaults to `60`.
func (caps Capabilities) WithMaxTypingFrequency(n int) Capabilities {
	if n <= 0 {
		n = 60
	}
	caps["maxTypingFrequency"] = n
	return caps
}

// WithWaitForIdleTimeout
//
//	Defaults to `10`
func (caps Capabilities) WithWaitForIdleTimeout(second float64) Capabilities {
	caps["waitForIdleTimeout"] = second
	return caps
}

// WithShouldUseTestManagerForVisibilityDetection If set to YES will ask TestManagerDaemon for element visibility
//
//	Defaults to  `false`
func (caps Capabilities) WithShouldUseTestManagerForVisibilityDetection(b bool) Capabilities {
	caps["shouldUseTestManagerForVisibilityDetection"] = b
	return caps
}

// WithShouldUseCompactResponses If set to YES will use compact (standards-compliant) & faster responses
//
//	Defaults to `true`
func (caps Capabilities) WithShouldUseCompactResponses(b bool) Capabilities {
	caps["shouldUseCompactResponses"] = b
	return caps
}

// WithElementResponseAttributes If shouldUseCompactResponses == NO,
// is the comma-separated list of fields to return with each element.
//
//	Defaults to `type,label`.
func (caps Capabilities) WithElementResponseAttributes(s string) Capabilities {
	caps["elementResponseAttributes"] = s
	return caps
}

// WithShouldUseSingletonTestManager
//
//	Defaults to `true`
func (caps Capabilities) WithShouldUseSingletonTestManager(b bool) Capabilities {
	caps["shouldUseSingletonTestManager"] = b
	return caps
}

// WithDisableAutomaticScreenshots
//
//	Defaults to `true`
func (caps Capabilities) WithDisableAutomaticScreenshots(b bool) Capabilities {
	caps["disableAutomaticScreenshots"] = b
	return caps
}

// WithShouldTerminateApp
//
//	Defaults to `true`
func (caps Capabilities) WithShouldTerminateApp(b bool) Capabilities {
	caps["shouldTerminateApp"] = b
	return caps
}

// WithEventloopIdleDelaySec
// Delays the invocation of '-[XCUIApplicationProcess setEventLoopHasIdled:]' by the timer interval passed.
// which is skipped on setting it to zero.
func (caps Capabilities) WithEventloopIdleDelaySec(second float64) Capabilities {
	caps["eventloopIdleDelaySec"] = second
	return caps
}
