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

	adbClient *gadb.Device
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
	output, err := ad.adbClient.RunShellCommand("wm", "size")
	if err != nil {
		return size, errors.Wrap(err, "get window size failed with adb")
	}

	// output may contain both Physical and Override size, use Override if existed
	// Physical size: 1080x2340
	// Override size: 1080x2220

	matchedSizeType := "Physical"
	if strings.Contains(output, "Override") {
		matchedSizeType = "Override"
	}

	var resolution string
	sizeList := strings.Split(output, "\n")
	log.Trace().Msgf("window size: %v", sizeList)
	for _, size := range sizeList {
		if strings.Contains(size, matchedSizeType) {
			resolution = strings.Split(size, ": ")[1]
			// 1080x2340
			ss := strings.Split(resolution, "x")
			width, _ := strconv.Atoi(ss[0])
			height, _ := strconv.Atoi(ss[1])
			return Size{Width: width, Height: height}, nil
		}
	}

	err = errors.New("physical window size not found by adb")
	return
}

func (ad *adbDriver) Screen() (screen Screen, err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Scale() (scale float64, err error) {
	return 1, nil
}

func (ad *adbDriver) GetTimestamp() (timestamp int64, err error) {
	// adb shell date +%s
	output, err := ad.adbClient.RunShellCommand("date", "+%s")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get timestamp by adb")
	}

	timestamp, err = strconv.ParseInt(strings.TrimSpace(output), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "convert timestamp failed")
	}

	return timestamp, nil
}

// PressBack simulates a short press on the BACK button.
func (ad *adbDriver) PressBack(options ...ActionOption) (err error) {
	// adb shell input keyevent 4
	_, err = ad.adbClient.RunShellCommand("input", "keyevent", fmt.Sprintf("%d", KCBack))
	if err != nil {
		return errors.Wrap(err, "press back failed")
	}
	return nil
}

func (ad *adbDriver) StartCamera() (err error) {
	if _, err = ad.adbClient.RunShellCommand("rm", "-r", "/sdcard/DCIM/Camera"); err != nil {
		return errors.Wrap(err, "remove /sdcard/DCIM/Camera failed")
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

func (ad *adbDriver) AppLaunch(packageName string) (err error) {
	// 不指定 Activity 名称启动（启动主 Activity）
	// adb shell monkey -p <packagename> -c android.intent.category.LAUNCHER 1
	sOutput, err := ad.adbClient.RunShellCommand(
		"monkey", "-p", packageName, "-c", "android.intent.category.LAUNCHER", "1",
	)
	if err != nil {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("monkey launch failed: %v", err))
	}
	if strings.Contains(sOutput, "monkey aborted") {
		return errors.Wrap(code.MobileUILaunchAppError,
			fmt.Sprintf("monkey aborted: %s", strings.TrimSpace(sOutput)))
	}
	return nil
}

func (ad *adbDriver) AppTerminate(packageName string) (successful bool, err error) {
	// 强制停止应用，停止 <packagename> 相关的进程
	// adb shell am force-stop <packagename>
	_, err = ad.adbClient.RunShellCommand("am", "force-stop", packageName)
	if err != nil {
		return false, errors.Wrap(err, "force-stop app failed")
	}

	return true, nil
}

func (ad *adbDriver) Tap(x, y int, options ...ActionOption) error {
	return ad.TapFloat(float64(x), float64(y), options...)
}

func (ad *adbDriver) TapFloat(x, y float64, options ...ActionOption) (err error) {
	actionOptions := NewActionOptions(options...)

	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}
	x += actionOptions.getRandomOffset()
	y += actionOptions.getRandomOffset()

	// adb shell input tap x y
	xStr := fmt.Sprintf("%.1f", x)
	yStr := fmt.Sprintf("%.1f", y)
	_, err = ad.adbClient.RunShellCommand(
		"input", "tap", xStr, yStr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("tap <%s, %s> failed", xStr, yStr))
	}
	return nil
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

func (ad *adbDriver) Drag(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return ad.DragFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ad *adbDriver) DragFloat(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
	err = errDriverNotImplemented
	return
}

func (ad *adbDriver) Swipe(fromX, fromY, toX, toY int, options ...ActionOption) error {
	return ad.SwipeFloat(float64(fromX), float64(fromY), float64(toX), float64(toY), options...)
}

func (ad *adbDriver) SwipeFloat(fromX, fromY, toX, toY float64, options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)

	if len(actionOptions.Offset) == 4 {
		fromX += float64(actionOptions.Offset[0])
		fromY += float64(actionOptions.Offset[1])
		toX += float64(actionOptions.Offset[2])
		toY += float64(actionOptions.Offset[3])
	}
	fromX += actionOptions.getRandomOffset()
	fromY += actionOptions.getRandomOffset()
	toX += actionOptions.getRandomOffset()
	toY += actionOptions.getRandomOffset()

	// adb shell input swipe fromX fromY toX toY
	_, err := ad.adbClient.RunShellCommand(
		"input", "swipe",
		fmt.Sprintf("%.1f", fromX), fmt.Sprintf("%.1f", fromY),
		fmt.Sprintf("%.1f", toX), fmt.Sprintf("%.1f", toY),
	)
	if err != nil {
		return errors.Wrap(err, "swipe failed")
	}
	return nil
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

func (ad *adbDriver) SendKeys(text string, options ...ActionOption) (err error) {
	// adb shell input text <text>
	_, err = ad.adbClient.RunShellCommand("input", "text", text)
	if err != nil {
		return errors.Wrap(err, "send keys failed")
	}
	return nil
}

func (ad *adbDriver) Input(text string, options ...ActionOption) (err error) {
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
	resp, err := ad.adbClient.ScreenCap()
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
	if _, err = ad.adbClient.RunShellCommand("logcat", "-c"); err != nil {
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
	log.Info().Str("logcat content", content).Msg("display logcat content")
	return ConvertPoints(content), nil
}

func (ad *adbDriver) GetForegroundApp() (app AppInfo, err error) {
	// adb shell dumpsys activity activities
	output, err := ad.adbClient.RunShellCommand("dumpsys", "activity", "activities")
	if err != nil {
		log.Error().Err(err).Msg("failed to dumpsys activities")
		return AppInfo{}, errors.Wrap(err, "dumpsys activities failed")
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// grep mResumedActivity|ResumedActivity
		if strings.HasPrefix(trimmedLine, "mResumedActivity:") || strings.HasPrefix(trimmedLine, "ResumedActivity:") {
			// mResumedActivity: ActivityRecord{9656d74 u0 com.android.settings/.Settings t407}
			// ResumedActivity: ActivityRecord{8265c25 u0 com.android.settings/.Settings t73}
			strs := strings.Split(trimmedLine, " ")
			for _, str := range strs {
				if strings.Contains(str, "/") {
					// com.android.settings/.Settings
					s := strings.Split(str, "/")
					app := AppInfo{
						AppBaseInfo: AppBaseInfo{
							PackageName: s[0],
							Activity:    s[1],
						},
					}
					return app, nil
				}
			}
		}
	}

	return AppInfo{}, errors.Wrap(code.MobileUIAssertForegroundAppError, "get foreground app failed")
}

func (ad *adbDriver) AssertForegroundApp(packageName string, activityType ...string) error {
	log.Debug().Str("package_name", packageName).
		Strs("activity_type", activityType).
		Msg("assert android foreground package and activity")

	app, err := ad.GetForegroundApp()
	if err != nil {
		log.Warn().Err(err).Msg("get foreground app failed, skip app/activity assertion")
		return nil // Notice: ignore error when get foreground app failed
	}

	// assert package
	if app.PackageName != packageName {
		log.Error().
			Interface("foreground_app", app.AppBaseInfo).
			Str("expected_package", packageName).
			Msg("assert package failed")
		return errors.Wrap(code.MobileUIAssertForegroundAppError,
			"assert foreground package failed")
	}

	if len(activityType) == 0 {
		return nil
	}

	// assert activity
	expectActivityType := activityType[0]
	activities, ok := androidActivities[packageName]
	if !ok {
		msg := fmt.Sprintf("activities not configured for package %s", packageName)
		log.Error().Msg(msg)
		return errors.Wrap(code.MobileUIAssertForegroundActivityError, msg)
	}

	expectActivities, ok := activities[expectActivityType]
	if !ok {
		msg := fmt.Sprintf("activity type %s not configured for package %s",
			expectActivityType, packageName)
		log.Error().Msg(msg)
		return errors.Wrap(code.MobileUIAssertForegroundActivityError, msg)
	}

	// assertion
	for _, expectActivity := range expectActivities {
		if strings.HasSuffix(app.Activity, expectActivity) {
			// assert activity success
			return nil
		}
	}

	// assert activity failed
	log.Error().
		Interface("foreground_app", app.AppBaseInfo).
		Str("expected_activity_type", expectActivityType).
		Strs("expected_activities", expectActivities).
		Msg("assert activity failed")
	return errors.Wrap(code.MobileUIAssertForegroundActivityError,
		"assert foreground activity failed")
}

var androidActivities = map[string]map[string][]string{
	// DY
	"com.ss.android.ugc.aweme": {
		"feed": []string{".splash.SplashActivity"},
		"live": []string{".live.LivePlayActivity"},
	},
	// DY lite
	"com.ss.android.ugc.aweme.lite": {
		"feed": []string{".splash.SplashActivity"},
		"live": []string{".live.LivePlayActivity"},
	},
	// KS
	"com.smile.gifmaker": {
		"feed": []string{
			"com.yxcorp.gifshow.HomeActivity",
			"com.yxcorp.gifshow.detail.PhotoDetailActivity",
		},
		"live": []string{
			"com.kuaishou.live.core.basic.activity.LiveSlideActivity",
			"com.yxcorp.gifshow.detail.PhotoDetailActivity",
		},
	},
	// KS lite
	"com.kuaishou.nebula": {
		"feed": []string{
			"com.yxcorp.gifshow.HomeActivity",
			"com.yxcorp.gifshow.detail.PhotoDetailActivity",
		},
		"live": []string{
			"com.kuaishou.live.core.basic.activity.LiveSlideActivity",
			"com.yxcorp.gifshow.detail.PhotoDetailActivity",
		},
	},
	// TODO: SPH, XHS
}
