# gidevice

This module is initially forked from [electricbubble/gidevice@v0.6.2].

#### Devices

```go
package main

import (
	"log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatalln(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	for _, dev := range devices {
		log.Println(dev.Properties().SerialNumber, dev.Properties().ProductID, dev.Properties().DeviceID)
	}
}

```

### GetValue

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

type DeviceDetail struct {
	DeviceName                string `json:"DeviceName,omitempty"`
	DeviceColor               string `json:"DeviceColor,omitempty"`
	DeviceClass               string `json:"DeviceClass,omitempty"`
	ProductVersion            string `json:"ProductVersion,omitempty"`
	ProductType               string `json:"ProductType,omitempty"`
	ProductName               string `json:"ProductName,omitempty"`
	ModelNumber               string `json:"ModelNumber,omitempty"`
	SerialNumber              string `json:"SerialNumber,omitempty"`
	SIMStatus                 string `json:"SIMStatus,omitempty"`
	PhoneNumber               string `json:"PhoneNumber,omitempty"`
	CPUArchitecture           string `json:"CPUArchitecture,omitempty"`
	ProtocolVersion           string `json:"ProtocolVersion,omitempty"`
	RegionInfo                string `json:"RegionInfo,omitempty"`
	TelephonyCapability       bool   `json:"TelephonyCapability,omitempty"`
	TimeZone                  string `json:"TimeZone,omitempty"`
	UniqueDeviceID            string `json:"UniqueDeviceID,omitempty"`
	WiFiAddress               string `json:"WiFiAddress,omitempty"`
	WirelessBoardSerialNumber string `json:"WirelessBoardSerialNumber,omitempty"`
	BluetoothAddress          string `json:"BluetoothAddress,omitempty"`
	BuildVersion              string `json:"BuildVersion,omitempty"`
}

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatal(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("No Device")
	}

	d := devices[0]

	detail, err1 := d.GetValue("", "")
	if err1 != nil {
		fmt.Errorf("get %s device detail fail : %w", d.Properties().SerialNumber, err1)
	}

	data, _ := json.Marshal(detail)
	d1 := &DeviceDetail{}
	json.Unmarshal(data, d1)
	fmt.Println(d1)
}
```

#### DeveloperDiskImage

```go
package main

import (
	"encoding/base64"
	"log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatal(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("No Device")
	}

	d := devices[0]

	imageSignatures, err := d.Images()
	if err != nil {
		log.Fatalln(err)
	}

	for i, imgSign := range imageSignatures {
		log.Printf("[%d] %s\n", i+1, base64.StdEncoding.EncodeToString(imgSign))
	}

	dmgPath := "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/DeviceSupport/14.4/DeveloperDiskImage.dmg"
	signaturePath := "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/DeviceSupport/14.4/DeveloperDiskImage.dmg.signature"

	err = d.MountDeveloperDiskImage(dmgPath, signaturePath)
	if err != nil {
		log.Fatalln(err)
	}
}

```

#### App

```go
package main

import (
	"log"
	"path/filepath"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatalln(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatalln(err)
	}

	if len(devices) == 0 {
		log.Fatalln("No Device")
	}

	d := devices[0]

	bundleID := "com.apple.Preferences"
	pid, err := d.AppLaunch(bundleID)
	if err != nil {
		log.Fatalln(err)
	}

	err = d.AppKill(pid)
	if err != nil {
		log.Fatalln(err)
	}

	runningProcesses, err := d.AppRunningProcesses()
	if err != nil {
		log.Fatalln(err)
	}

	for _, process := range runningProcesses {
		if process.IsApplication {
			log.Printf("%4d\t%-24s\t%-36s\t%s\n", process.Pid, process.Name, filepath.Base(process.RealAppName), process.StartDate)
		}
	}
}

```

#### Screenshot

```go
package main

import (
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatalln(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatalln(err)
	}

	if len(devices) == 0 {
		log.Fatalln("No Device")
	}

	d := devices[0]

	raw, err := d.Screenshot()
	if err != nil {
		log.Fatalln(err)
	}

	img, format, err := image.Decode(raw)
	if err != nil {
		log.Fatalln(err)
	}
	userHomeDir, _ := os.UserHomeDir()
	file, err := os.Create(userHomeDir + "/Desktop/s1." + format)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() { _ = file.Close() }()
	switch format {
	case "png":
		err = png.Encode(file, img)
	case "jpeg":
		err = jpeg.Encode(file, img, nil)
	}
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(file.Name())
}

```

#### SimulateLocation

```go
package main

import (
	"log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatalln(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatalln(err)
	}

	if len(devices) == 0 {
		log.Fatalln("No Device")
	}

	d := devices[0]

	// https://api.map.baidu.com/lbsapi/getpoint/index.html
	if err = d.SimulateLocationUpdate(116.024067, 40.362639, gidevice.CoordinateSystemBD09); err != nil {
		log.Fatalln(err)
	}

	// https://developer.amap.com/tools/picker
	// https://lbs.qq.com/tool/getpoint/index.html
	// if err = d.SimulateLocationUpdate(120.116979, 30.252876, gidevice.CoordinateSystemGCJ02); err != nil {
	// 	log.Fatalln(err)
	// }

	// if err = d.SimulateLocationUpdate(121.499763, 31.239580,gidevice.CoordinateSystemWGS84); err != nil {
	// if err = d.SimulateLocationUpdate(121.499763, 31.239580); err != nil {
	// 	log.Fatalln(err)
	// }

	// err = d.SimulateLocationRecover()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}

```

#### XCTest

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatal(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("No Device")
	}

	d := devices[0]

	out, cancel, err := d.XCTest("com.leixipaopao.WebDriverAgentRunner.xctrunner")
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		for s := range out {
			fmt.Print(s)
		}
	}()

	<-done
	cancel()
	fmt.Println()
	log.Println("DONE")
}

```

#### Connect and Forward

```go
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
	"syscall"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
)

func main() {
	usbmux, err := gidevice.NewUsbmux()
	if err != nil {
		log.Fatal(err)
	}

	devices, err := usbmux.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("No Device")
	}

	d := devices[0]

	localPort, remotePort := 8100, 8100

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", localPort))

	go func(listener net.Listener) {
		for {
			var accept net.Conn
			if accept, err = listener.Accept(); err != nil {
				log.Println("accept:", err)
			}

			fmt.Println("accept", accept.RemoteAddr())

			rInnerConn, err := d.NewConnect(remotePort)
			if err != nil {
				log.Println(err)
				os.Exit(0)
			}

			rConn := rInnerConn.RawConn()
			_ = rConn.SetDeadline(time.Time{})

			go func(lConn net.Conn) {
				go func(lConn, rConn net.Conn) {
					if _, err := io.Copy(lConn, rConn); err != nil {
						//do sth
					}
				}(lConn, rConn)
				go func(lConn, rConn net.Conn) {
					if _, err := io.Copy(rConn, lConn); err != nil {
						//do sth
					}
				}(lConn, rConn)
			}(accept)
		}
	}(listener)

	done := make(chan os.Signal, syscall.SIGTERM)
	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
}
```

[electricbubble/gidevice@v0.6.2]: https://github.com/electricbubble/gidevice/tree/v0.6.2
