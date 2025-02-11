package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/ai"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var (
	serial   string
	runTimes int
)

func init() {
	serial = os.Getenv("UDID")
	numStr := os.Getenv("RunTimes")
	defaultNum := 20

	var err error
	runTimes, err = strconv.Atoi(numStr)
	if err != nil {
		runTimes = defaultNum
	}
	fmt.Printf("=== start running cases, serial=%s, runTimes=%d ===\n", serial, runTimes)
}

func launchAppDriver(pkgName string) (driverExt *uixt.XTDriver, err error) {
	device, _ := uixt.NewIOSDevice(option.WithUDID(serial))
	driver, err := device.NewDriver()
	if err != nil {
		return nil, err
	}

	//_, err = driver.AppTerminate(pkgName)
	//if err != nil {
	//	return nil, err
	//}
	//
	//err = driver.Homescreen()
	//if err != nil {
	//	return nil, err
	//}
	//
	//err = driver.AppLaunch(pkgName)
	//if err != nil {
	//	return nil, err
	//}

	time.Sleep(15 * time.Second)

	driverExt = uixt.NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))

	// 处理弹窗
	err = driverExt.ClosePopupsHandler()
	if err != nil {
		return nil, err
	}

	// 进入推荐页
	err = driverExt.TapByOCR("推荐", option.WithScope(0, 0, 1, 0.3))
	if err != nil {
		return nil, err
	}

	return driverExt, nil
}

func watchVideo(driver *uixt.XTDriver) (err error) {
	time.Sleep(3 * time.Second)
	err = driver.Swipe(0.7, 0.7, 0.7, 0.2)
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	// 点击进入某视频
	err = driver.TapXY(0.3, 0.5)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	// 点击播放区域，展现横屏图标
	err = driver.TapXY(0.5, 0.1)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// 切换横屏
	err = driver.TapByCV(
		option.WithScreenShotUITypes("fullScreen"))
	if err != nil {
		// 未找到横屏图标，该页面可能不是横版视频（直播|广告|Feed）
		// 退出回到推荐页
		driver.Back()
		return nil
	}

	// 观播 10s
	time.Sleep(10 * time.Second)

	// 返回视频页面
	err = driver.Back()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	// 返回推荐页
	err = driver.Back()
	if err != nil {
		return err
	}

	return nil
}

// build shell command
// go build -o bili_android examples/uitest/bilibili/cli.go
func main() {
	driver, err := launchAppDriver("tv.danmaku.bilianime")
	if err != nil {
		panic(err)
	}

	// 重复采集 XX 次
	for i := 0; i < runTimes; i++ {
		err = watchVideo(driver)
		if err != nil {
			panic(err)
		}
	}
}
