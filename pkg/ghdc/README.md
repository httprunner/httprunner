# ghdc

ghdc 是一个用于与鸿蒙设备进行交互的工具，封装了各种 HDC（鸿蒙的 ADB）命令和 UI 自动化能力。

## 目录结构
ghdc \
├── client.go&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;封装 hdc list targets 等非设备命令 \
├── device.go&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;封装 hdc -t connectkey shell 等指定设备的命令 \
└── uidevice.go&nbsp;封装设备所有自动化能力

## hdc 命令调用
目前支持的能力:
- 获取设备
- 文件传输
- shell 命令
- 端口挂载
- 属性获取（brand, model, 版本等）
- 截图

`hdc`与鸿蒙的关系和`adb`与安卓关系一致，其架构与 adb server 相同，分为 client 和 server。`hdc start` 会启动一个 hdc server，监听本地的 8710 端口。我们也是和该 8710 端口通信执行 HDC 命令。目前支持常用的 hdc 能力，并可扩展至所有 hdc 能力。

```shell
hdc -m -s ::ffff:127.0.0.1:8710
```
hdc 命令分为两类：
- 不指定设备的命令：例如 `hdc list targets`。逻辑封装在 client.go 中。与 gadb 类似，所有的命令通过 RPC 方式与 hdc server 直接通信。
- 指定设备的命令：例如 `hdc -t connectkey shell`。逻辑封装在 device.go 中。这些命令会在与 server 建连时指定执行的设备。

## UI Test 自动化能力
目前支持的能力:
- 点击\滑动\输入
- 按键操作
- 手势操作
- TouchDown/TouchMove/TouchUp 屏幕操作
- 屏幕旋转
- 音量设置
- 图片流获取
- 控件信息监听
- 简单控件操作

Harmony Next 内置了 UI Test 服务，提供了所有常用的自动化能力。并且这个服务也部分开源，支持二次开发。由于协议未开源，我们通过逆向工程绕过 JS API，直接通过 socket 与 UI Test 服务进行通信，操作手机。代码能力封装在 uidevice.go 中。

UI Test 协议分为两类：
- 无 session 单次返回长连接：例如点击、滑动等，发出命令即可，返回一个执行成功或失败。
- 有 session 多次返回长连接：例如获取屏幕图片流、监听控件信息变化等。

## 使用方法
```go
package main

import (
	"github.com/httprunner/httprunner/v5/pkg/ghdc"
	"log"
)

func main() {
	client, err := ghdc.NewClient()
	checkErr(err, "fail to connect hdc server")

	devices, err := client.DeviceList()
	checkErr(err)

	if len(devices) == 0 {
		log.Fatalln("list of devices is empty")
	}
	dev := devices[0]
	driver, err := ghdc.NewUIDriver(dev)
	checkErr(err, "fail to init device uiDriver")

	err = driver.Touch(225, 1715)
	checkErr(err)
}

func checkErr(err error, msg ...string) {
	if err == nil {
		return
	}

	var output string
	if len(msg) != 0 {
		output = msg[0] + " "
	}
	output += err.Error()
	log.Fatalln(output)
}
```