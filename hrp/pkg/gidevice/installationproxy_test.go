//go:build localtest

package gidevice

import (
	"testing"
)

var installationProxySrv InstallationProxy

func setupInstallationProxySrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if installationProxySrv, err = lockdownSrv.InstallationProxyService(); err != nil {
		t.Fatal(err)
	}
}

func Test_installationProxy_Browse(t *testing.T) {
	setupInstallationProxySrv(t)

	// currentList, err := installationProxySrv.Browse(WithMetaData(true))
	// currentList, err := installationProxySrv.Browse(WithReturnAttributes("CFBundleIdentifier", "SequenceNumber", "SequenceNumber"))
	// currentList, err := installationProxySrv.Browse(WithApplicationType(ApplicationTypeSystem))
	// currentList, err := installationProxySrv.Browse(WithApplicationType(ApplicationTypeSystem), WithReturnAttributes("ApplicationType", "ApplicationType"))
	// currentList, err := dev.InstallationProxyBrowse()
	currentList, err := installationProxySrv.Browse()
	// currentList, err := installationProxySrv.Browse(WithBundleIDs("com.apple.MusicUIService"), WithBundleIDs("com.apple.Home.HomeControlService"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(currentList))

	for _, cl := range currentList {
		app, ok := cl.(map[string]interface{})
		if ok {
			t.Log(app)
		} else {
			t.Log(cl)
		}
	}
}

func Test_installationProxy_Lookup(t *testing.T) {
	setupInstallationProxySrv(t)

	// lookupResult, err := installationProxySrv.Lookup()
	// lookupResult, err := dev.InstallationProxyLookup(
	lookupResult, err := installationProxySrv.Lookup(
		// WithApplicationType(ApplicationTypeUser),
		// WithApplicationType(ApplicationTypeSystem),
		// WithReturnAttributes("CFBundleDevelopmentRegion"),
		// WithReturnAttributes("CFBundleDisplayName", "CFBundleIdentifier"),
		// WithBundleIDs("com.apple.mobilephone"),
		WithBundleIDs("com.leixipaopao.WebDriverAgentRunner.xctrunner"),
	)
	if err != nil {
		t.Fatal(err)
	}

	ret := lookupResult.(map[string]interface{})
	t.Log(len(ret))

	for k, v := range ret {
		t.Log(k, "-->", v)
	}
}
