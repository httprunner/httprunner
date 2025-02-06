package uixt

import (
	"fmt"
	"os"
	"testing"
)

var androidStubDriver *stubAndroidDriver

func setupStubDriver(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	device.STUB = true
	androidStubDriver, err = device.NewStubDriver(Capabilities{})
	checkErr(t, err)
}

func TestHello(t *testing.T) {
	setupStubDriver(t)
	status, err := androidStubDriver.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(status)
}

func TestSource(t *testing.T) {
	setupStubDriver(t)
	source, err := androidStubDriver.Source()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(source)
}

func TestLogin(t *testing.T) {
	setupStubDriver(t)
	info, err := androidStubDriver.LoginNoneUI("com.ss.android.ugc.aweme", "12342316231", "8517", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestLogout(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.LogoutNoneUI("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSwipe(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.Swipe(878, 2375, 672, 2375)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTap(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.Tap(900, 400)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoubleTap(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.DoubleTap(500, 500)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLongPress(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.Swipe(1036, 1076, 1036, 1076, WithDuration(3))
	if err != nil {
		t.Fatal(err)
	}
}

func TestInput(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.Input("\"哈哈\"")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSave(t *testing.T) {
	setupStubDriver(t)
	raw, err := androidStubDriver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}
	source, err := androidStubDriver.Source()
	if err != nil {
		t.Fatal(err)
	}
	step := 14
	file, err := os.Create(fmt.Sprintf("/Users/bytedance/workcode/wings_algorithm/testcases/data/cases/0/%d.jpg", step))
	if err != nil {
		t.Fatal(err)
	}
	file.Write(raw.Bytes())

	file, err = os.Create(fmt.Sprintf("/Users/bytedance/workcode/wings_algorithm/testcases/data/cases/0/%d.json", step))
	if err != nil {
		t.Fatal(err)
	}
	file.Write([]byte(source))
}

func TestAppLaunch(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.AppLaunch("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAppTerminal(t *testing.T) {
	setupStubDriver(t)
	_, err := androidStubDriver.AppTerminate("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAppInfo(t *testing.T) {
	setupStubDriver(t)
	info, err := androidStubDriver.getLoginAppInfo("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}
