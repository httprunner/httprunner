//go:build localtest

package gidevice

import (
	"testing"
)

var (
	instrumentsSrv Instruments
	bundleID       = "com.apple.Preferences"
)

func setupInstrumentsSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if instrumentsSrv, err = lockdownSrv.InstrumentsService(); err != nil {
		t.Fatal(err)
	}
}

func Test_instruments_AppLaunch(t *testing.T) {
	setupInstrumentsSrv(t)

	// bundleID = "com.leixipaopao.WebDriverAgentRunner.xctrunner"

	// pid, err := dev.AppLaunch(bundleID)
	pid, err := instrumentsSrv.AppLaunch(bundleID)
	// pid, err := instrumentsSrv.AppLaunch(bundleID, WithKillExisting(true))
	// pid, err := instrumentsSrv.AppLaunch(bundleID, WithKillExisting(true), WithArguments([]interface{}{"-AppleLanguages", "(Russian)"}))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pid)
}

func Test_instruments_AppKill(t *testing.T) {
	setupInstrumentsSrv(t)

	pid, err := instrumentsSrv.AppLaunch(bundleID)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pid)

	// if err = dev.AppKill(pid); err != nil {
	if err = instrumentsSrv.AppKill(pid); err != nil {
		t.Fatal(err)
	}
}

func Test_instruments_AppRunningProcesses(t *testing.T) {
	setupInstrumentsSrv(t)

	// processes, err := dev.AppRunningProcesses()
	processes, err := instrumentsSrv.AppRunningProcesses()
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range processes {
		t.Log(p.IsApplication, "\t", p.Pid, "\t", p.Name, "\t", p.RealAppName, "\t", p.StartDate)
	}
}

func Test_instruments_AppList(t *testing.T) {
	setupInstrumentsSrv(t)

	// apps, err := dev.AppList()
	apps, err := instrumentsSrv.AppList()
	if err != nil {
		t.Fatal(err)
	}

	for _, app := range apps {
		t.Logf("%v\t%v\t%v\t%v\t%v\n", app.Type, app.DisplayName, app.ExecutableName, app.AppExtensionUUIDs, app.BundlePath)
	}
}

func Test_instruments_DeviceInfo(t *testing.T) {
	setupInstrumentsSrv(t)

	devInfo, err := instrumentsSrv.DeviceInfo()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(devInfo.Description)
	t.Log(devInfo.DisplayName)
	t.Log(devInfo.Identifier)
	t.Log(devInfo.Version)
	t.Log(devInfo.ProductType)
	t.Log(devInfo.ProductVersion)
	t.Log(devInfo.XRDeviceClassName)
}
