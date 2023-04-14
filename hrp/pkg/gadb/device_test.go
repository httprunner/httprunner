//go:build localtest

package gadb

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
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

func TestDevice_ReverseForward(t *testing.T) {
	adbClient, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}

	devices, err := adbClient.DeviceList()
	if err != nil {
		t.Fatal(err)
	}

	localPort := 5005
	err = devices[0].ReverseForward(localPort, "localabstract:scrcpy")
	if err != nil {
		t.Fatal(err)
	}
	err = devices[0].ReverseForward(localPort, "localabstract:scrcpy1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = devices[0].ReverseForwardList()
	if err != nil {
		t.Fatal(err)
	}

	err = devices[0].ReverseForwardKill("localabstract:scrcpy1")
	if err != nil {
		t.Fatal(err)
	}
	err = devices[0].ReverseForwardKillAll()
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

func TestDevice_RunShellCommandBackgroundWithBytes(t *testing.T) {
	type fields struct {
		adbClient Client
		serial    string
		attrs     map[string]string
	}
	type args struct {
		cmd  string
		args []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "runShellCommandBackground",
			fields: fields{
				adbClient: func() Client {
					c, _ := NewClient()
					return c
				}(),
				serial: "63c1ee94",
			},
			args: args{
				cmd: "nohup sleep 10 2>/dev/null 1>/dev/null &",
				// cmd:  "sleep 10",

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Device{
				adbClient: tt.fields.adbClient,
				serial:    tt.fields.serial,
				attrs:     tt.fields.attrs,
			}
			got, err := d.RunShellCommandV2WithBytes(tt.args.cmd, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Device.RunShellCommandBackgroundWithBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Device.RunShellCommandBackgroundWithBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_InstallAPK(t *testing.T) {
	apk, _ := os.Open("test.apk")
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

	res, err := dev.InstallAPK(apk)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestDevice_HasFeature(t *testing.T) {
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

	t.Log(dev.GetFeatures())
}
