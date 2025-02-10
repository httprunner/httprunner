package driver_ext

import (
	"fmt"
	"os"
	"testing"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var shootsAndroidDriver *ShootsAndroidDriver

func setupShootsAndroidDriver(t *testing.T) {
	device, err := uixt.NewAndroidDevice()
	checkErr(t, err)
	shootsAndroidDriver, err = NewShootsAndroidDriver(device)
	checkErr(t, err)
}

func checkErr(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg, err)
		}
	}
}

func TestHello(t *testing.T) {
	setupShootsAndroidDriver(t)
	status, err := shootsAndroidDriver.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(status)
}

func TestSource(t *testing.T) {
	setupShootsAndroidDriver(t)
	source, err := shootsAndroidDriver.Source()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(source)
}

func TestLogin(t *testing.T) {
	setupShootsAndroidDriver(t)
	info, err := shootsAndroidDriver.LoginNoneUI("com.ss.android.ugc.aweme", "12342316231", "8517", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}

func TestLogout(t *testing.T) {
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.LogoutNoneUI("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSwipe(t *testing.T) {
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.Swipe(878, 2375, 672, 2375)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTap(t *testing.T) {
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.Tap(900, 400)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoubleTap(t *testing.T) {
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.DoubleTap(500, 500)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLongPress(t *testing.T) {
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.Swipe(1036, 1076, 1036, 1076,
		option.WithDuration(3))
	if err != nil {
		t.Fatal(err)
	}
}

func TestInput(t *testing.T) {
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.Input("\"哈哈\"")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSave(t *testing.T) {
	setupShootsAndroidDriver(t)
	raw, err := shootsAndroidDriver.Screenshot()
	if err != nil {
		t.Fatal(err)
	}
	source, err := shootsAndroidDriver.Source()
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
	setupShootsAndroidDriver(t)
	err := shootsAndroidDriver.AppLaunch("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAppTerminal(t *testing.T) {
	setupShootsAndroidDriver(t)
	_, err := shootsAndroidDriver.AppTerminate("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAppInfo(t *testing.T) {
	setupShootsAndroidDriver(t)
	info, err := shootsAndroidDriver.getLoginAppInfo("com.ss.android.ugc.aweme")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info)
}
