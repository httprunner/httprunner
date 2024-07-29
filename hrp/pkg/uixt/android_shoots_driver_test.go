package uixt

import "testing"

var shootsDriver *ShootsAndroidDriver

func setupShootsDriver(t *testing.T) {
	device, err := NewAndroidDevice()
	checkErr(t, err)
	device.SHOOTS = true
	shootsDriver, err = device.NewShootsDriver(Capabilities{})
	checkErr(t, err)
}

func TestHello(t *testing.T) {
	setupShootsDriver(t)
	status, err := shootsDriver.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(status)
}

func TestSource(t *testing.T) {
	setupShootsDriver(t)
	source, err := shootsDriver.Source()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(source)
}

func TestIsLogin(t *testing.T) {
	setupShootsDriver(t)
	res, err := shootsDriver.isLogin("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestLogin(t *testing.T) {
	setupShootsDriver(t)
	err := shootsDriver.LoginNoneUI("com.ss.android.ugc.aweme", "12342316231", "8517")
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogout(t *testing.T) {
	setupShootsDriver(t)
	err := shootsDriver.LogoutNoneUI("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}
