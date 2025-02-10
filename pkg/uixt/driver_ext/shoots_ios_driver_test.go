package driver_ext

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var (
	shootsIOSDriver uixt.IDriverExt
	iOSDevice       *uixt.IOSDevice
)

func setupShootsIOSDriver(t *testing.T) {
	var err error
	iOSDevice, err = uixt.NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800))
	checkErr(t, err)
	shootsIOSDriver, err = iOSDevice.NewDriver()
	checkErr(t, err)
}

func TestIOSLogin(t *testing.T) {
	setupShootsIOSDriver(t)
	info, err := shootsIOSDriver.(*ShootsIOSDriver).LoginNoneUI("", "12342316231", "8517", "")
	checkErr(t, err)
	t.Log(info)
}

func TestIOSLogout(t *testing.T) {
	setupShootsIOSDriver(t)
	err := shootsIOSDriver.(*ShootsIOSDriver).LogoutNoneUI("")
	checkErr(t, err)
}

func TestIOSIsLogin(t *testing.T) {
	setupShootsIOSDriver(t)
	err := shootsIOSDriver.(*ShootsIOSDriver).LogoutNoneUI("")
	checkErr(t, err)
}

func TestIOSSource(t *testing.T) {
	setupShootsIOSDriver(t)
	source, err := shootsIOSDriver.GetDriver().Source()
	checkErr(t, err)
	t.Log(source)
}

func TestIOSForeground(t *testing.T) {
	setupShootsIOSDriver(t)
	app, err := shootsIOSDriver.GetDriver().GetForegroundApp()
	checkErr(t, err)
	t.Log(app)
}

func TestIOSSwipe(t *testing.T) {
	setupShootsIOSDriver(t)
	shootsIOSDriver.GetDriver().Swipe(540, 0, 540, 1000)
}

func TestIOSSave(t *testing.T) {
	setupShootsIOSDriver(t)
	raw, err := shootsIOSDriver.GetDriver().Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	source, err := shootsIOSDriver.GetDriver().Source()
	if err != nil {
		t.Fatal(err)
	}
	step := 7
	file, err := os.Create(fmt.Sprintf("/Users/bytedance/workcode/wings_algorithm/testcases/data/cases/ios/4159417_cvcn02okg4g0/%d.jpg", step))
	if err != nil {
		t.Fatal(err)
	}
	file.Write(raw.Bytes())

	file, err = os.Create(fmt.Sprintf("/Users/bytedance/workcode/wings_algorithm/testcases/data/cases/ios/4159417_cvcn02okg4g0/%d.json", step))
	if err != nil {
		t.Fatal(err)
	}
	file.Write([]byte(source))
}

func TestListen(t *testing.T) {
	setupShootsIOSDriver(t)
	localPort, err := builtin.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	err = iOSDevice.Forward(localPort, 8800)
	if err != nil {
		t.Fatal(err)
	}
	addr := fmt.Sprintf("0.0.0.0:%d", localPort)
	l, err := net.Listen("tcp", addr)
	if err == nil {
		l.Close() // 端口成功绑定后立即释放，返回该端口号
	} else {
		t.Fatal(err)
	}
}
