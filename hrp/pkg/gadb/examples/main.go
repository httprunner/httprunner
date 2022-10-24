package main

import (
	"log"
	"os"
	"strings"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
)

func main() {
	adbClient, err := gadb.NewClient()
	checkErr(err, "fail to connect adb server")

	devices, err := adbClient.DeviceList()
	checkErr(err)

	if len(devices) == 0 {
		log.Fatalln("list of devices is empty")
	}

	dev := devices[0]

	userHomeDir, _ := os.UserHomeDir()
	apk, err := os.Open(userHomeDir + "/Desktop/xuexi_android_10002068.apk")
	checkErr(err)

	log.Println("starting to push apk")

	remotePath := "/data/local/tmp/xuexi_android_10002068.apk"
	err = dev.PushFile(apk, remotePath)
	checkErr(err, "adb push")

	log.Println("push completed")

	log.Println("starting to install apk")

	shellOutput, err := dev.RunShellCommand("pm install", remotePath)
	checkErr(err, "pm install")
	if !strings.Contains(shellOutput, "Success") {
		log.Fatalln("fail to install: ", shellOutput)
	}

	log.Println("install completed")

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
