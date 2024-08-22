package uixt

import "testing"

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

func TestIsLogin(t *testing.T) {
	setupStubDriver(t)
	res, err := androidStubDriver.isLogin("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestLogin(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.LoginNoneUI("com.ss.android.ugc.aweme", "12342316231", "8517")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogout(t *testing.T) {
	setupStubDriver(t)
	err := androidStubDriver.LogoutNoneUI("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}
