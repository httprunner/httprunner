package driver_ext

import (
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"testing"
)

var (
	iOSStubDriver *StubIOSDriver
)

func checkErr(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg, err)
		}
	}
}

func setupIOSStubDriver(t *testing.T) {
	iOSDevice, err := uixt.NewIOSDevice(option.WithWDAPort(8700), option.WithWDAMjpegPort(8800), option.WithResetHomeOnStartup(false))
	checkErr(t, err)
	iOSStubDriver, err = NewStubIOSDriver(iOSDevice)
	checkErr(t, err)
}

func TestIOSStubDriver_LoginNoneUI(t *testing.T) {
	setupIOSStubDriver(t)
	info, err := iOSStubDriver.LoginNoneUI("com.ss.iphone.ugc.AwemeInhouse", "12343418541", "", "im112233")
	checkErr(t, err)
	t.Logf("login info: %+v", info)
}
