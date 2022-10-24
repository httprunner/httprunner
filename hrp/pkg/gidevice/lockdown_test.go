//go:build localtest

package gidevice

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"testing"
	"time"
)

var lockdownSrv Lockdown

func setupLockdownSrv(t *testing.T) {
	setupDevice(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}
}

func Test_lockdown_QueryType(t *testing.T) {
	setupLockdownSrv(t)

	lockdownType, err := lockdownSrv.QueryType()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(lockdownType.Type)
}

func Test_lockdown_GetValue(t *testing.T) {
	setupLockdownSrv(t)

	// v, err := dev.GetValue("com.apple.mobile.iTunes", "")
	// v, err := dev.GetValue("com.apple.mobile.internal", "")
	v, err := dev.GetValue("com.apple.mobile.battery", "")
	// v, err := lockdownSrv.GetValue("", "ProductVersion")
	// v, err := lockdownSrv.GetValue("", "DeviceName")
	// v, err := lockdownSrv.GetValue("com.apple.mobile.iTunes", "")
	// v, err := lockdownSrv.GetValue("com.apple.mobile.battery", "")
	// v, err := lockdownSrv.GetValue("com.apple.disk_usage", "")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)
}

func Test_lockdown_SyslogRelayService(t *testing.T) {
	setupLockdownSrv(t)

	syslogRelaySrv, err := lockdownSrv.SyslogRelayService()
	if err != nil {
		t.Fatal(err)
	}
	syslogRelaySrv.Stop()

	lines := syslogRelaySrv.Lines()

	done := make(chan os.Signal, 1)

	go func() {
		for line := range lines {
			fmt.Println(line)
		}
		done <- os.Interrupt
		fmt.Println("DONE!!!")
	}()

	signal.Notify(done, os.Interrupt, os.Kill)

	<-done
	syslogRelaySrv.Stop()
	time.Sleep(time.Second)
}

func Test_lockdown_CrashReportMoverService(t *testing.T) {
	setupLockdownSrv(t)

	crashReportMoverSrv, err := lockdownSrv.CrashReportMoverService()
	if err != nil {
		t.Fatal(err)
	}

	filenames := make([]string, 0, 36)
	fn := func(cwd string, info *AfcFileInfo) {
		if cwd == "." {
			cwd = ""
		}
		filenames = append(filenames, path.Join(cwd, info.Name()))
		// fmt.Println(path.Join(cwd, name))
	}
	err = crashReportMoverSrv.walkDir(".", fn)
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range filenames {
		fmt.Println(n)
	}

	t.Log(len(filenames))
}
