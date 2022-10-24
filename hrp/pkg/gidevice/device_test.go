//go:build localtest

package gidevice

import (
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"
)

var dev Device

func setupDevice(t *testing.T) {
	setupUsbmux(t)
	devices, err := um.Devices()
	if err != nil {
		t.Fatal(err)
	}

	if len(devices) == 0 {
		t.Fatal("No Device")
	}

	dev = devices[0]
}
func Test_device_ReadPairRecord(t *testing.T) {
	setupDevice(t)

	pairRecord, err := dev.ReadPairRecord()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(pairRecord.HostID, pairRecord.SystemBUID, pairRecord.WiFiMACAddress)
}

func Test_device_NewConnect(t *testing.T) {
	setupDevice(t)

	if _, err := dev.NewConnect(LockdownPort); err != nil {
		t.Fatal(err)
	}
}

func Test_device_DeletePairRecord(t *testing.T) {
	setupDevice(t)

	if err := dev.DeletePairRecord(); err != nil {
		t.Fatal(err)
	}

}

func Test_device_SavePairRecord(t *testing.T) {
	setupLockdownSrv(t)

	pairRecord, err := lockdownSrv.Pair()
	if err != nil {
		t.Fatal(err)
	}

	err = dev.SavePairRecord(pairRecord)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_device_XCTest(t *testing.T) {
	setupLockdownSrv(t)

	bundleID = "com.leixipaopao.WebDriverAgentRunner.xctrunner"
	out, cancel, err := dev.XCTest(bundleID)
	// out, cancel, err := dev.XCTest(bundleID, WithXCTestEnv(map[string]interface{}{"USE_PORT": 8222, "MJPEG_SERVER_PORT": 8333}))
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		for s := range out {
			fmt.Print(s)
		}
		done <- os.Interrupt
	}()

	for {
		select {
		case <-done:
			cancel()
			fmt.Println()
			t.Log("DONE")
			return
		}
	}
}

func Test_device_AppInstall(t *testing.T) {
	setupLockdownSrv(t)

	ipaPath := "/private/tmp/derivedDataPath/Build/Products/Release-iphoneos/WebDriverAgentRunner-Runner.ipa"
	err := dev.AppInstall(ipaPath)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_device_AppUninstall(t *testing.T) {
	setupLockdownSrv(t)

	bundleID = "com.leixipaopao.WebDriverAgentRunner.xctrunner"
	err := dev.AppUninstall(bundleID)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_device_Syslog(t *testing.T) {
	setupLockdownSrv(t)

	dev.SyslogStop()

	lines, err := dev.Syslog()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan os.Signal, 1)

	go func() {
		for line := range lines {
			fmt.Println(line)
		}
		done <- os.Interrupt
		t.Log("DONE!!!")
	}()

	signal.Notify(done, os.Interrupt, os.Kill)

	// <-done
	time.Sleep(3 * time.Second)
	dev.SyslogStop()
	time.Sleep(200 * time.Millisecond)
}

func Test_device_Reboot(t *testing.T) {
	setupDevice(t)
	dev.Reboot()
}

func Test_device_Shutdown(t *testing.T) {
	setupDevice(t)
	dev.Shutdown()
}

func Test_device_InstallationProxyBrowse(t *testing.T) {
	setupDevice(t)

	list, err := dev.InstallationProxyBrowse(
		WithApplicationType(ApplicationTypeUser),
		WithReturnAttributes("CFBundleDisplayName", "CFBundleIdentifier", "SequenceNumber", "SequenceNumber"),
	)
	// list, err := dev.InstallationProxyBrowse()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(list))

	for _, l := range list {
		t.Logf("%#v", l)
	}
}
