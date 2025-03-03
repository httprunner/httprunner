package driver_ext

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var iOSStubDriver *StubIOSDriver

func setupIOSStubDriver(t *testing.T) {
	iOSDevice, err := uixt.NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithResetHomeOnStartup(false))
	assert.Nil(t, err)
	iOSStubDriver, err = NewStubIOSDriver(iOSDevice)
	assert.Nil(t, err)
}

func TestIOSStubDriver_LoginNoneUI(t *testing.T) {
	setupIOSStubDriver(t)
	info, err := iOSStubDriver.LoginNoneUI("com.ss.iphone.ugc.AwemeInhouse", "12343418541", "", "im112233")
	assert.Nil(t, err)
	t.Logf("login info: %+v", info)
}
