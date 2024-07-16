package uixt

import "testing"

var driver *ShootsAndroidDriver

func setupAndroid(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	device.SHOOTS = true
	driver, err = device.NewShootsDriver(Capabilities{})
	checkErr(t, err)
}

func TestHello(t *testing.T) {
	setupAndroid(t)
	status, err := driver.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(status)
}

func TestSource(t *testing.T) {
	setupAndroid(t)
	source, err := driver.Source()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(source)
}

func TestLogin(t *testing.T) {
	setupAndroid(t)
	res, err := driver.isLogin("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
