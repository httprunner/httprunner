package driver_ext

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/httprunner/httprunner/v5/uixt"
)

func setupAndroidStubDriver(t *testing.T) *StubAndroidDriver {
	device, err := uixt.NewAndroidDevice()
	require.Nil(t, err)
	device.Options.UIA2 = false
	device.Options.LogOn = false
	driver, err := NewStubAndroidDriver(device)
	require.Nil(t, err)
	return driver
}

func TestAndroidStubDriver_LoginNoneUI(t *testing.T) {
	androidStubDriver := setupAndroidStubDriver(t)
	info, err := androidStubDriver.LoginNoneUI("com.ss.android.ugc.aweme", "12343418541", "", "im112233")
	assert.Nil(t, err)
	t.Logf("login info: %+v", info)
}
