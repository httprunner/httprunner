package uixt

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var (
	iOSStubDriver IDriverExt
	iOSDevice     *IOSDevice
)

func setupiOSStubDriver(t *testing.T) {
	var err error
	iOSDevice, err = NewIOSDevice(
		option.WithWDAPort(8700),
		option.WithWDAMjpegPort(8800),
		option.WithIOSStub(false))
	checkErr(t, err)
	iOSStubDriver, err = iOSDevice.NewDriver()
	checkErr(t, err)
}

func TestIOSLogin(t *testing.T) {
	setupiOSStubDriver(t)
	info, err := iOSStubDriver.GetDriver().LoginNoneUI("", "12342316231", "8517", "")
	checkErr(t, err)
	t.Log(info)
}

func TestIOSLogout(t *testing.T) {
	setupiOSStubDriver(t)
	err := iOSStubDriver.GetDriver().LogoutNoneUI("")
	checkErr(t, err)
}

func TestIOSIsLogin(t *testing.T) {
	setupiOSStubDriver(t)
	err := iOSStubDriver.GetDriver().LogoutNoneUI("")
	checkErr(t, err)
}

func TestIOSSource(t *testing.T) {
	setupiOSStubDriver(t)
	source, err := iOSStubDriver.GetDriver().Source()
	checkErr(t, err)
	t.Log(source)
}

func TestIOSForeground(t *testing.T) {
	setupiOSStubDriver(t)
	app, err := iOSStubDriver.GetDriver().GetForegroundApp()
	checkErr(t, err)
	t.Log(app)
}

func TestIOSSwipe(t *testing.T) {
	setupiOSStubDriver(t)
	iOSStubDriver.GetDriver().Swipe(540, 0, 540, 1000)
}

func TestIOSSave(t *testing.T) {
	setupiOSStubDriver(t)
	raw, err := iOSStubDriver.GetDriver().Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	source, err := iOSStubDriver.GetDriver().Source()
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
	setupiOSStubDriver(t)
	localPort, err := builtin.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	err = iOSDevice.forward(localPort, 8800)
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
