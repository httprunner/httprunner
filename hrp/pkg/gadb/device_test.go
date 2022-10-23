package gadb

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDevice_State(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		state, err := dev.State()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dev.Serial(), state)
	}
}

func TestDevice_DevicePath(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		devPath, err := dev.DevicePath()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dev.Serial(), devPath)
	}
}

func TestDevice_Product(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		product := dev.Product()
		t.Log(dev.Serial(), product)
	}
}

func TestDevice_Model(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		t.Log(dev.Serial(), dev.Model())
	}
}

func TestDevice_Usb(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		t.Log(dev.Serial(), dev.Usb(), dev.IsUsb())
	}

}

func TestDevice_DeviceInfo(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	for i := range devices {
		dev := devices[i]
		t.Log(dev.DeviceInfo())
	}
}

func TestDevice_Forward(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	localPort := 61000
	err = devices[0].Forward(localPort, 6790)
	if err != nil {
		t.Fatal(err)
	}

	err = devices[0].ForwardKill(localPort)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_ForwardList(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	for i := range devices {
		dev := devices[i]
		forwardList, err := dev.ForwardList()
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dev.serial, "->", forwardList)
	}
}

func TestDevice_ForwardKill(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	SetDebug(true)

	err = devices[0].ForwardKill(6790)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_RunShellCommand(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	// for i := range devices {
	// 	dev := devices[i]
	// 	// cmdOutput, err := dev.RunShellCommand(`pm list packages  | grep  "bili"`)
	// 	// cmdOutput, err := dev.RunShellCommand(`pm list packages`, `| grep "bili"`)
	// 	// cmdOutput, err := dev.RunShellCommand("dumpsys activity | grep mFocusedActivity")
	// 	cmdOutput, err := dev.RunShellCommand("monkey", "-p", "tv.danmaku.bili", "-c", "android.intent.category.LAUNCHER", "1")
	// 	if err != nil {
	// 		t.Fatal(dev.serial, err)
	// 	}
	// 	t.Log("\n"+dev.serial, cmdOutput)
	// }

	// SetDebug(true)

	dev := devices[len(devices)-1]
	dev = devices[0]

	// cmdOutput, err := dev.RunShellCommand("monkey", "-p", "tv.danmaku.bili", "-c", "android.intent.category.LAUNCHER", "1")
	cmdOutput, err := dev.RunShellCommand("ls /sdcard")
	// cmdOutput, err := dev.RunShellCommandWithBytes("screencap -p")
	if err != nil {
		t.Fatal(dev.serial, err)
	}
	t.Log("\n⬇️"+dev.serial+"⬇️\n", cmdOutput)

}

func TestDevice_EnableAdbOverTCP(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	dev := devices[len(devices)-1]
	dev = devices[0]

	SetDebug(true)

	err = dev.EnableAdbOverTCP()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_List(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	dev := devices[len(devices)-1]
	dev = devices[0]

	SetDebug(true)

	// fileEntries, err := dev.List("/sdcard")
	fileEntries, err := dev.List("/sdcard/Download")
	if err != nil {
		t.Fatal(err)
	}

	for i := range fileEntries {
		t.Log(fileEntries[i].Name, "\t", fileEntries[i].IsDir())
	}
}

func TestDevice_Push(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	dev := devices[len(devices)-1]
	dev = devices[0]

	SetDebug(true)

	file, _ := os.Open("/Users/hero/Documents/temp/MuMu共享文件夹/test.txt")
	err = dev.PushFile(file, "/sdcard/Download/push.txt", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	err = dev.Push(strings.NewReader("world"), "/sdcard/Download/hello.txt", time.Now())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDevice_Pull(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	dev := devices[len(devices)-1]
	dev = devices[0]

	SetDebug(true)

	buffer := bytes.NewBufferString("")
	err = dev.Pull("/sdcard/Download/hello.txt", buffer)
	if err != nil {
		t.Fatal(err)
	}

	userHomeDir, _ := os.UserHomeDir()
	if err = ioutil.WriteFile(userHomeDir+"/Desktop/hello.txt", buffer.Bytes(), DefaultFileMode); err != nil {
		t.Fatal(err)
	}
}
