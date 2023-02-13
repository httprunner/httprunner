package uixt

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
)

type adbDriver struct {
	Driver

	adbClient gadb.Device
	logcat    *AdbLogcat
}

func NewAdbDriver() *adbDriver {
	log.Info().Msg("init adb driver")
	return &adbDriver{}
}

func (ad *adbDriver) NewSession(capabilities Capabilities) (sessionInfo SessionInfo, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) DeleteSession() (err error) {
	return errDriverNotImplemented
}

func (ad *adbDriver) Status() (deviceStatus DeviceStatus, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) DeviceInfo() (deviceInfo DeviceInfo, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Location() (location Location, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) BatteryInfo() (batteryInfo BatteryInfo, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) WindowSize() (size Size, err error) {
	// adb shell wm size
	resp, err := ad.adbClient.RunShellCommand("wm", "size")
	if err != nil {
		return
	}

	// Physical size: 1080x2340
	s := strings.Trim(strings.Split(resp, ": ")[1], "\n")
	ss := strings.Split(s, "x")
	width, _ := strconv.Atoi(ss[0])
	height, _ := strconv.Atoi(ss[1])
	size = Size{Width: width, Height: height}
	return
}

func (ad *adbDriver) Screen() (screen Screen, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Scale() (scale float64, err error) {
	return 1, nil
}

// PressBack simulates a short press on the BACK button.
func (ad *adbDriver) PressBack(options ...DataOption) (err error) {
	// adb shell input keyevent 4
	_, err = ad.adbClient.RunShellCommand("input", "keyevent", fmt.Sprintf("%d", KCBack))
	return
}

func (ad *adbDriver) StartCamera() (err error) {
	if _, err = ad.adbClient.RunShellCommand("rm", "-r", "/sdcard/DCIM/Camera"); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	var version string
	if version, err = ad.adbClient.RunShellCommand("getprop", "ro.build.version.release"); err != nil {
		return err
	}
	if version == "11" || version == "12" {
		if _, err = ad.adbClient.RunShellCommand("am", "start", "-a", "android.media.action.STILL_IMAGE_CAMERA"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ad.adbClient.RunShellCommand("input", "swipe", "750", "1000", "250", "1000"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ad.adbClient.RunShellCommand("input", "keyevent", fmt.Sprintf("%d", KCCamera)); err != nil {
			return err
		}
		return
	} else {
		if _, err = ad.adbClient.RunShellCommand("am", "start", "-a", "android.media.action.VIDEO_CAPTURE"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ad.adbClient.RunShellCommand("input", "keyevent", fmt.Sprintf("%d", KCCamera)); err != nil {
			return err
		}
		return
	}
}

func (ad *adbDriver) StopCamera() (err error) {
	err = ad.PressBack()
	if err != nil {
		return err
	}
	err = ad.Homescreen()
	if err != nil {
		return err
	}

	// kill samsung shell command
	if _, err = ad.AppTerminate("com.sec.android.app.camera"); err != nil {
		return err
	}
	// kill other camera (huawei mi)
	if _, err = ad.AppTerminate("com.android.camera2"); err != nil {
		return err
	}
	return
}

func (ad *adbDriver) Homescreen() (err error) {
	return ad.PressKeyCode(KCHome, KMEmpty)
}

func (ad *adbDriver) PressKeyCode(keyCode KeyCode, metaState KeyMeta) (err error) {
	// adb shell input keyevent <keyCode>
	_, err = ad.adbClient.RunShellCommand(
		"input", "keyevent", fmt.Sprintf("%d", keyCode))
	return
}

func (ad *adbDriver) AppLaunch(bundleId string) (err error) {
	// 不指定 Activity 名称启动（启动主 Activity）
	// adb shell monkey -p <packagename> -c android.intent.category.LAUNCHER 1
	sOutput, err := ad.adbClient.RunShellCommand(
		"monkey", "-p", bundleId, "-c", "android.intent.category.LAUNCHER", "1",
	)
	if err != nil {
		return err
	}
	if strings.Contains(sOutput, "monkey aborted") {
		return fmt.Errorf("app launch: %s", strings.TrimSpace(sOutput))
	}
	return nil
}

func (ad *adbDriver) AppTerminate(bundleId string) (successful bool, err error) {
	// 强制停止应用，停止 <packagename> 相关的进程
	// adb shell am force-stop <packagename>
	_, err = ad.adbClient.RunShellCommand("am", "force-stop", bundleId)
	return err == nil, err
}

func (ad *adbDriver) Tap(x, y int, options ...DataOption) error {
	return ad.TapFloat(float64(x), float64(y), options...)
}

func (ad *adbDriver) TapFloat(x, y float64, options ...DataOption) (err error) {
	dataOptions := NewDataOptions(options...)

	if len(dataOptions.Offset) == 2 {
		x += float64(dataOptions.Offset[0])
		y += float64(dataOptions.Offset[1])
	}

	// adb shell input tap x y
	_, err = ad.adbClient.RunShellCommand(
		"input", "tap", fmt.Sprintf("%.1f", x), fmt.Sprintf("%.1f", y))
	return
}

func (ad *adbDriver) DoubleTap(x, y int) error {
	return ad.DoubleTapFloat(float64(x), float64(y))
}

func (ad *adbDriver) DoubleTapFloat(x, y float64) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) TouchAndHold(x, y int, second ...float64) (err error) {
	return ad.TouchAndHoldFloat(float64(x), float64(y), second...)
}

func (ad *adbDriver) TouchAndHoldFloat(x, y float64, second ...float64) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Drag(fromX, fromY, toX, toY int, options ...DataOption) error {
	return ad.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ad *adbDriver) DragFloat(fromX, fromY, toX, toY float64, options ...DataOption) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Swipe(fromX, fromY, toX, toY int, options ...DataOption) error {
	return ad.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ad *adbDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...DataOption) error {
	// adb shell input swipe fromX fromY toX toY
	_, err := ad.adbClient.RunShellCommand(
		"input", "swipe",
		fmt.Sprintf("%.1f", fromX), fmt.Sprintf("%.1f", fromY),
		fmt.Sprintf("%.1f", toX), fmt.Sprintf("%.1f", toY),
	)
	return err
}

func (ad *adbDriver) ForceTouch(x, y int, pressure float64, second ...float64) error {
	return ad.ForceTouchFloat(float64(x), float64(y), pressure, second...)
}

func (ad *adbDriver) ForceTouchFloat(x, y, pressure float64, second ...float64) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) SetPasteboard(contentType PasteboardType, content string) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) GetPasteboard(contentType PasteboardType) (raw *bytes.Buffer, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) SendKeys(text string, options ...DataOption) (err error) {
	// adb shell input text <text>
	_, err = ad.adbClient.RunShellCommand("input", "text", text)
	return
}

func (ad *adbDriver) Input(text string, options ...DataOption) (err error) {
	return ad.SendKeys(text, options...)
}

func (ad *adbDriver) PressButton(devBtn DeviceButton) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Rotation() (rotation Rotation, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) SetRotation(rotation Rotation) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Screenshot() (raw *bytes.Buffer, err error) {
	// adb shell screencap -p
	resp, err := ad.adbClient.RunShellCommandWithBytes(
		"screencap", "-p",
	)
	if err == nil {
		return bytes.NewBuffer(resp), nil
	}
	return nil, err
}

func (ad *adbDriver) Source(srcOpt ...SourceOption) (source string, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) AccessibleSource() (source string, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) HealthCheck() (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) GetAppiumSettings() (settings map[string]interface{}, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) SetAppiumSettings(settings map[string]interface{}) (ret map[string]interface{}, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) IsHealthy() (healthy bool, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) StartCaptureLog(identifier ...string) (err error) {
	log.Info().Msg("start adb log recording")

	// clear logcat
	if _, err = ad.adbClient.RunShellCommand("logcat", "--clear"); err != nil {
		return err
	}

	// start logcat
	err = ad.logcat.CatchLogcat()
	if err != nil {
		err = errors.Wrap(code.AndroidCaptureLogError,
			fmt.Sprintf("start adb log recording failed: %v", err))
		return err
	}
	return nil
}

func (ad *adbDriver) StopCaptureLog() (result interface{}, err error) {
	log.Info().Msg("stop adb log recording")
	err = ad.logcat.Stop()
	if err != nil {
		log.Error().Err(err).Msg("failed to get adb log recording")
		err = errors.Wrap(code.AndroidCaptureLogError,
			fmt.Sprintf("get adb log recording failed: %v", err))
		return "", err
	}
	content := ad.logcat.logBuffer.String()
	return ConvertPoints(content), nil
}
