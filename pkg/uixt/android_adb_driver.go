package uixt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/utf7"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
)

const (
	AdbKeyBoardPackageName = "com.android.adbkeyboard/.AdbIME"
	UnicodeImePackageName  = "io.appium.settings/.UnicodeIME"
)

type adbDriver struct {
	Driver

	adbClient *gadb.Device
	logcat    *AdbLogcat
}

func NewAdbDriver() *adbDriver {
	log.Info().Msg("init adb driver")
	driver := &adbDriver{}
	driver.NewSession(nil)
	return driver
}

func (ad *adbDriver) runShellCommand(cmd string, args ...string) (output string, err error) {
	driverResult := &DriverResult{
		RequestMethod: "adb",
		RequestUrl:    cmd,
		RequestBody:   strings.Join(args, " "),
		RequestTime:   time.Now(),
	}
	defer func() {
		driverResult.ResponseDuration = time.Since(driverResult.RequestTime).Milliseconds()
		if err != nil {
			driverResult.Success = false
			driverResult.Error = err.Error()
		} else {
			driverResult.Success = true
		}
		ad.session.addRequestResult(driverResult)
	}()

	// adb shell screencap -p
	if cmd == "screencap" {
		resp, err := ad.adbClient.ScreenCap()
		if err == nil {
			driverResult.ResponseBody = "OMITTED"
			return string(resp), nil
		}
		return "", errors.Wrap(err, "adb screencap failed")
	}

	output, err = ad.adbClient.RunShellCommand(cmd, args...)
	driverResult.ResponseBody = strings.TrimSpace(output)
	return output, err
}

func (ad *adbDriver) NewSession(capabilities Capabilities) (sessionInfo SessionInfo, err error) {
	ad.Driver.session.Reset()
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

func (ad *adbDriver) getWindowSize() (size Size, err error) {
	// adb shell wm size
	output, err := ad.runShellCommand("wm", "size")
	if err != nil {
		return size, errors.Wrap(err, "get window size failed by adb shell")
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

func (ad *adbDriver) WindowSize() (size Size, err error) {
	if !ad.windowSize.IsNil() {
		// use cached window size
		return ad.windowSize, nil
	}

	size, err = ad.getWindowSize()
	if err != nil {
		return
	}

	orientation, err2 := ad.Orientation()
	if err2 != nil {
		// Notice: do not return err if get window orientation failed
		orientation = OrientationPortrait
		log.Warn().Err(err2).Msgf(
			"get window orientation failed, use default %s", orientation)
	}
	if orientation != OrientationPortrait {
		size.Width, size.Height = size.Height, size.Width
	}

	ad.windowSize = size // cache window size
	return size, nil
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
	output, err := ad.runShellCommand("date", "+%s")
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
	_, err = ad.runShellCommand("input", "keyevent", fmt.Sprintf("%d", KCBack))
	if err != nil {
		return errors.Wrap(err, "press back failed")
	}
	return nil
}

func (ad *adbDriver) StartCamera() (err error) {
	if _, err = ad.runShellCommand("rm", "-r", "/sdcard/DCIM/Camera"); err != nil {
		return errors.Wrap(err, "remove /sdcard/DCIM/Camera failed")
	}
	time.Sleep(5 * time.Second)
	var version string
	if version, err = ad.runShellCommand("getprop", "ro.build.version.release"); err != nil {
		return err
	}
	if version == "11" || version == "12" {
		if _, err = ad.runShellCommand("am", "start", "-a", "android.media.action.STILL_IMAGE_CAMERA"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ad.runShellCommand("input", "swipe", "750", "1000", "250", "1000"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ad.runShellCommand("input", "keyevent", fmt.Sprintf("%d", KCCamera)); err != nil {
			return err
		}
		return
	} else {
		if _, err = ad.runShellCommand("am", "start", "-a", "android.media.action.VIDEO_CAPTURE"); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
		if _, err = ad.runShellCommand("input", "keyevent", fmt.Sprintf("%d", KCCamera)); err != nil {
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

func (ad *adbDriver) Orientation() (orientation Orientation, err error) {
	output, err := ad.runShellCommand("dumpsys", "input", "|", "grep", "'SurfaceOrientation'")
	if err != nil {
		return
	}
	re := regexp.MustCompile(`SurfaceOrientation: (\d)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 { // 确保找到了匹配项
		if matches[1] == "0" || matches[1] == "2" {
			return OrientationPortrait, nil
		} else if matches[1] == "1" || matches[1] == "3" {
			return OrientationLandscapeLeft, nil
		}
	}
	err = fmt.Errorf("not found SurfaceOrientation value")
	return
}

func (ad *adbDriver) Homescreen() (err error) {
	return ad.PressKeyCodes(KCHome, KMEmpty)
}

func (ad *adbDriver) Unlock() (err error) {
	// Notice: brighten should be executed before unlock
	// brighten android device screen
	if err := ad.PressKeyCodes(KCWakeup, KMEmpty); err != nil {
		log.Error().Err(err).Msg("brighten android device screen failed")
	}
	// unlock android device screen
	if err := ad.PressKeyCodes(KCMenu, KMEmpty); err != nil {
		log.Error().Err(err).Msg("press menu key to unlock screen failed")
	}

	// swipe up to unlock
	return ad.Swipe(500, 1500, 500, 500)
}

func (ad *adbDriver) Backspace(count int, options ...ActionOption) (err error) {
	if count == 0 {
		return nil
	}
	if count == 1 {
		return ad.PressKeyCode(KCDel)
	}
	keyArray := make([]KeyCode, count)

	for i := range keyArray {
		keyArray[i] = KCDel
	}
	return ad.combinationKey(keyArray)
}

func (ad *adbDriver) combinationKey(keyCodes []KeyCode) (err error) {
	if len(keyCodes) == 1 {
		return ad.PressKeyCode(keyCodes[0])
	}
	strKeyCodes := make([]string, len(keyCodes))
	for i, keycode := range keyCodes {
		strKeyCodes[i] = fmt.Sprintf("%d", keycode)
	}
	_, err = ad.runShellCommand(
		"input", append([]string{"keycombination"}, strKeyCodes...)...)
	return
}

func (ad *adbDriver) PressKeyCode(keyCode KeyCode) (err error) {
	return ad.PressKeyCodes(keyCode, KMEmpty)
}

func (ad *adbDriver) PressKeyCodes(keyCode KeyCode, metaState KeyMeta) (err error) {
	// adb shell input keyevent [--longpress] KEYCODE [METASTATE]
	if metaState != KMEmpty {
		// press key with metastate, e.g. KMShiftOn/KMCtrlOn
		_, err = ad.runShellCommand(
			"input", "keyevent", "--longpress",
			fmt.Sprintf("%d", keyCode),
			fmt.Sprintf("%d", metaState))
	} else {
		_, err = ad.runShellCommand(
			"input", "keyevent",
			fmt.Sprintf("%d", keyCode))
	}
	return
}

func (ad *adbDriver) AppLaunch(packageName string) (err error) {
	// 不指定 Activity 名称启动（启动主 Activity）
	// adb shell monkey -p <packagename> -c android.intent.category.LAUNCHER 1
	sOutput, err := ad.runShellCommand(
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
	_, err = ad.runShellCommand("am", "force-stop", packageName)
	if err != nil {
		return false, errors.Wrap(err, "force-stop app failed")
	}

	return true, nil
}

func (ad *adbDriver) Tap(x, y float64, options ...ActionOption) error {
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
	_, err := ad.runShellCommand(
		"input", "tap", xStr, yStr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("tap <%s, %s> failed", xStr, yStr))
	}
	return nil
}

func (ad *adbDriver) DoubleTap(x, y float64, options ...ActionOption) error {
	// adb shell input tap x y
	xStr := fmt.Sprintf("%.1f", x)
	yStr := fmt.Sprintf("%.1f", y)
	_, err := ad.runShellCommand(
		"input", "tap", xStr, yStr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("tap <%s, %s> failed", xStr, yStr))
	}
	time.Sleep(time.Duration(100) * time.Millisecond)
	_, err = ad.runShellCommand(
		"input", "tap", xStr, yStr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("tap <%s, %s> failed", xStr, yStr))
	}
	return nil
}

func (ad *adbDriver) TouchAndHold(x, y float64, options ...ActionOption) (err error) {
	actionOptions := NewActionOptions(options...)

	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}
	x += actionOptions.getRandomOffset()
	y += actionOptions.getRandomOffset()
	duration := 1000.0
	if actionOptions.Duration > 0 {
		duration = actionOptions.Duration * 1000
	}
	// adb shell input swipe fromX fromY toX toY
	_, err = ad.runShellCommand(
		"input", "swipe",
		fmt.Sprintf("%.1f", x), fmt.Sprintf("%.1f", y),
		fmt.Sprintf("%.1f", x), fmt.Sprintf("%.1f", y),
		fmt.Sprintf("%d", int(duration)),
	)
	if err != nil {
		return errors.Wrap(err, "long press failed")
	}
	return nil
}

func (ad *adbDriver) Drag(fromX, fromY, toX, toY float64, options ...ActionOption) (err error) {
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
	duration := 200.0
	if actionOptions.Duration > 0 {
		duration = actionOptions.Duration * 1000
	}
	command := "swipe"
	if actionOptions.PressDuration > 0 {
		command = "draganddrop"
	}
	// adb shell input swipe fromX fromY toX toY
	_, err = ad.runShellCommand(
		"input", command,
		fmt.Sprintf("%.1f", fromX), fmt.Sprintf("%.1f", fromY),
		fmt.Sprintf("%.1f", toX), fmt.Sprintf("%.1f", toY),
		fmt.Sprintf("%d", int(duration)),
	)
	if err != nil {
		return errors.Wrap(err, "drag failed")
	}
	return nil
}

func (ad *adbDriver) Swipe(fromX, fromY, toX, toY float64, options ...ActionOption) error {
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
	_, err := ad.runShellCommand(
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
	err = ad.SendUnicodeKeys(text, options...)
	if err == nil {
		return
	}
	err = ad.InputText(text, options...)
	return
}

func (ad *adbDriver) InputText(text string, options ...ActionOption) error {
	// adb shell input text <text>
	_, err := ad.runShellCommand("input", "text", text)
	if err != nil {
		return errors.Wrap(err, "send keys failed")
	}
	return nil
}

func (ad *adbDriver) SendUnicodeKeys(text string, options ...ActionOption) (err error) {
	// If the Unicode IME is not installed, fall back to the old interface.
	// There might be differences in the tracking schemes across different phones, and it is pending further verification.
	// In release version: without the Unicode IME installed, the test cannot execute.
	if !ad.IsUnicodeIMEInstalled() {
		return fmt.Errorf("appium unicode ime not installed")
	}
	currentIme, err := ad.GetIme()
	if err != nil {
		return
	}
	if currentIme != UnicodeImePackageName {
		defer func() {
			_ = ad.SetIme(currentIme)
		}()
		err = ad.SetIme(UnicodeImePackageName)
		if err != nil {
			log.Warn().Err(err).Msgf("set Unicode Ime failed")
			return
		}
	}
	encodedStr, err := utf7.Encoding.NewEncoder().String(text)
	if err != nil {
		log.Warn().Err(err).Msgf("encode text with modified utf7 failed")
		return
	}
	err = ad.InputText("\""+strings.ReplaceAll(encodedStr, "\"", "\\\"")+"\"", options...)
	return
}

func (ad *adbDriver) IsAdbKeyBoardInstalled() bool {
	output, err := ad.runShellCommand("ime", "list", "-a")
	if err != nil {
		return false
	}
	return strings.Contains(output, AdbKeyBoardPackageName)
}

func (ad *adbDriver) IsUnicodeIMEInstalled() bool {
	output, err := ad.runShellCommand("ime", "list", "-s")
	if err != nil {
		return false
	}
	return strings.Contains(output, UnicodeImePackageName)
}

func (ad *adbDriver) ListIme() []string {
	output, err := ad.runShellCommand("ime", "list", "-s")
	if err != nil {
		return []string{}
	}
	return strings.Split(output, "\n")
}

func (ad *adbDriver) SendKeysByAdbKeyBoard(text string) (err error) {
	defer func() {
		// Reset to default, don't care which keyboard was chosen before switch:
		if _, resetErr := ad.runShellCommand("ime", "reset"); resetErr != nil {
			log.Error().Err(err).Msg("failed to reset ime")
		}
	}()

	// Enable ADBKeyBoard from adb
	if _, err = ad.runShellCommand("ime", "enable", AdbKeyBoardPackageName); err != nil {
		log.Error().Err(err).Msg("failed to enable adbKeyBoard")
		return
	}
	// Switch to ADBKeyBoard from adb
	if _, err = ad.runShellCommand("ime", "set", AdbKeyBoardPackageName); err != nil {
		log.Error().Err(err).Msg("failed to set adbKeyBoard")
		return
	}
	time.Sleep(time.Second)
	// input Quoted text
	text = strings.ReplaceAll(text, " ", "\\ ")
	if _, err = ad.runShellCommand("am", "broadcast", "-a", "ADB_INPUT_TEXT", "--es", "msg", text); err != nil {
		log.Error().Err(err).Msg("failed to input by adbKeyBoard")
		return
	}
	if _, err = ad.runShellCommand("input", "keyevent", fmt.Sprintf("%d", KCEnter)); err != nil {
		log.Error().Err(err).Msg("failed to input keyevent enter")
		return
	}
	time.Sleep(time.Second)
	return
}

func (ad *adbDriver) Input(text string, options ...ActionOption) (err error) {
	return ad.SendKeys(text, options...)
}

func (ad *adbDriver) Clear(packageName string) error {
	if _, err := ad.runShellCommand("pm", "clear", packageName); err != nil {
		log.Error().Str("packageName", packageName).Err(err).Msg("failed to clear package cache")
		return err
	}

	return nil
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
	resp, err := ad.runShellCommand("screencap", "-p")
	if err != nil {
		return nil, errors.Wrap(err, "adb screencap failed")
	}

	return bytes.NewBuffer([]byte(resp)), nil
}

func (ad *adbDriver) Source(srcOpt ...SourceOption) (source string, err error) {
	_, err = ad.runShellCommand("rm", "-rf", "/sdcard/window_dump.xml")
	if err != nil {
		return
	}
	// 高版本报错 ERROR: null root node returned by UiTestAutomationBridge.
	_, err = ad.runShellCommand("uiautomator", "dump")
	if err != nil {
		return
	}
	source, err = ad.runShellCommand("cat", "/sdcard/window_dump.xml")
	if err != nil {
		return
	}
	return
}

func (ad *adbDriver) LoginNoneUI(packageName, phoneNumber string, captcha, password string) (info AppLoginInfo, err error) {
	return info, errDriverNotImplemented
}

func (ad *adbDriver) LogoutNoneUI(packageName string) error {
	return errDriverNotImplemented
}

func (ad *adbDriver) sourceTree(srcOpt ...SourceOption) (sourceTree *Hierarchy, err error) {
	source, err := ad.Source()
	if err != nil {
		return
	}
	sourceTree = new(Hierarchy)
	err = xml.Unmarshal([]byte(source), sourceTree)
	if err != nil {
		return
	}
	return
}

func (ad *adbDriver) TapByText(text string, options ...ActionOption) error {
	sourceTree, err := ad.sourceTree()
	if err != nil {
		return err
	}
	return ad.tapByTextUsingHierarchy(sourceTree, text, options...)
}

func (ad *adbDriver) tapByTextUsingHierarchy(hierarchy *Hierarchy, text string, options ...ActionOption) error {
	bounds := ad.searchNodes(hierarchy.Layout, text, options...)
	actionOptions := NewActionOptions(options...)
	if len(bounds) == 0 {
		if actionOptions.IgnoreNotFoundError {
			log.Info().Msg("not found element by text " + text)
			return nil
		}
		return errors.New("not found element by text " + text)
	}
	for _, bound := range bounds {
		width, height := bound.Center()
		err := ad.Tap(width, height, options...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ad *adbDriver) TapByTexts(actions ...TapTextAction) error {
	sourceTree, err := ad.sourceTree()
	if err != nil {
		return err
	}

	for _, action := range actions {
		err := ad.tapByTextUsingHierarchy(sourceTree, action.Text, action.Options...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ad *adbDriver) searchNodes(nodes []Layout, text string, options ...ActionOption) []Bounds {
	actionOptions := NewActionOptions(options...)
	var results []Bounds
	for _, node := range nodes {
		result := ad.searchNodes(node.Layout, text, options...)
		results = append(results, result...)
		if actionOptions.Regex {
			// regex on, check if match regex
			if !regexp.MustCompile(text).MatchString(node.Text) {
				continue
			}
		} else {
			// regex off, check if match exactly
			if node.Text != text {
				ad.searchNodes(node.Layout, text, options...)
				continue
			}
		}
		if node.Bounds != nil {
			results = append(results, *node.Bounds)
		}
	}
	return results
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
	// start logcat
	err = ad.logcat.CatchLogcat("iesqaMonitor:V")
	if err != nil {
		err = errors.Wrap(code.DeviceCaptureLogError,
			fmt.Sprintf("start adb log recording failed: %v", err))
		return err
	}
	return nil
}

func (ad *adbDriver) StopCaptureLog() (result interface{}, err error) {
	defer func() {
		log.Info().Msg("stop adb log recording")
		err = ad.logcat.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to get adb log recording")
		}
	}()
	if err != nil {
		log.Error().Err(err).Msg("failed to close adb log writer")
	}
	pointRes := ConvertPoints(ad.logcat.logs)

	// 没有解析到打点日志，走兜底逻辑
	if len(pointRes) == 0 {
		log.Info().Msg("action log is null, use action file >>>")
		logFilePathPrefix := fmt.Sprintf("%v/data", config.ActionLogFilePath)
		files := []string{}
		ad.adbClient.RunShellCommand("pull", config.DeviceActionLogFilePath, config.ActionLogFilePath)
		err = filepath.Walk(config.ActionLogFilePath, func(path string, info fs.FileInfo, err error) error {
			// 只是需要日志文件
			if ok := strings.Contains(path, logFilePathPrefix); ok {
				files = append(files, path)
			}
			return nil
		})
		// 先保持原有状态码不变，这里不return error
		if err != nil {
			log.Error().Err(err).Msg("read log file fail")
			return pointRes, nil
		}

		if len(files) != 1 {
			log.Error().Err(err).Msg("log file count error")
			return pointRes, nil
		}

		reader, err := os.Open(files[0])
		if err != nil {
			log.Info().Msg("open File error")
			return pointRes, nil
		}
		defer func() {
			_ = reader.Close()
		}()

		var lines []string // 创建一个空的字符串数组来存储文件的每一行

		// 使用 bufio.NewScanner 读取文件
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			lines = append(lines, scanner.Text()) // 将每行文本添加到字符串数组
		}

		if err := scanner.Err(); err != nil {
			return pointRes, nil
		}

		pointRes = ConvertPoints(lines)
	}
	return pointRes, nil
}

func (ad *adbDriver) GetSession() *DriverSession {
	return &ad.Driver.session
}

func (ad *adbDriver) GetDriverResults() []*DriverResult {
	return nil
}

func (ad *adbDriver) GetForegroundApp() (app AppInfo, err error) {
	packageInfo, err := ad.runShellCommand(
		"CLASSPATH=/data/local/tmp/evalite", "app_process", "/",
		"com.bytedance.iesqa.eval_process.PackageService", "2>/dev/null")
	if err != nil {
		return app, err
	}
	err = json.Unmarshal([]byte(strings.TrimSpace(packageInfo)), &app)
	if err != nil {
		log.Error().Err(err).Str("packageInfo", packageInfo).Msg("get foreground app failed")
	}
	return
}

func (ad *adbDriver) SetIme(imeRegx string) error {
	imeList := ad.ListIme()
	ime := ""
	for _, imeName := range imeList {
		if regexp.MustCompile(imeRegx).MatchString(imeName) {
			ime = imeName
			break
		}
	}
	if ime == "" {
		return fmt.Errorf("failed to set ime by %s, ime list: %v", imeRegx, imeList)
	}
	brand, _ := ad.adbClient.Brand()
	packageName := strings.Split(ime, "/")[0]
	res, err := ad.runShellCommand("ime", "set", ime)
	log.Info().Str("funcName", "SetIme").Interface("ime", ime).
		Interface("output", res).Msg("set ime")
	if err != nil {
		return err
	}

	if strings.ToLower(brand) == "oppo" {
		time.Sleep(1 * time.Second)
		pid, _ := ad.runShellCommand("pidof", packageName)
		if strings.TrimSpace(pid) == "" {
			appInfo, err := ad.GetForegroundApp()
			_ = ad.AppLaunch(packageName)
			if err == nil && packageName != UnicodeImePackageName {
				time.Sleep(10 * time.Second)
				nextAppInfo, err := ad.GetForegroundApp()
				log.Info().Str("beforeFocusedPackage", appInfo.PackageName).Str("afterFocusedPackage", nextAppInfo.PackageName).Msg("")
				if err == nil && nextAppInfo.PackageName != appInfo.PackageName {
					_ = ad.PressKeyCodes(KCBack, KMEmpty)
				}
			}
		}
	}
	// even if the shell command has returned,
	// as there might be a situation where the input method has not been completely switched yet
	// Listen to the following message.
	// InputMethodManagerService: onServiceConnected, name:ComponentInfo{io.appium.settings/io.appium.settings.UnicodeIME}, token:android.os.Binder@44f825
	// But there is no such log on Vivo.
	time.Sleep(3 * time.Second)
	return nil
}

func (ad *adbDriver) GetIme() (ime string, err error) {
	currentIme, err := ad.runShellCommand("settings", "get", "secure", "default_input_method")
	if err != nil {
		log.Warn().Err(err).Msgf("get default ime failed")
		return
	}
	currentIme = strings.TrimSpace(currentIme)
	return currentIme, nil
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

func (ad *adbDriver) RecordScreen(folderPath string, duration time.Duration) (videoPath string, err error) {
	// 获取当前时间戳
	timestamp := time.Now().Format("20060102_150405") + fmt.Sprintf("_%03d", time.Now().UnixNano()/1e6%1000)
	// 创建文件名
	fileName := fmt.Sprintf("%s/%s.mp4", folderPath, timestamp)
	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("Error creating directory")
	}

	// 创建一个文件
	file, err := os.Create(fileName)
	if err != nil {
		log.Error().Err(err)
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	// scrcpy -s 7d21bb91 --record=file.mp4 -N
	cmd := exec.Command(
		"scrcpy",
		"-s", ad.adbClient.Serial(),
		fmt.Sprintf("--record=%s", fileName),
		"-N",
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	// 启动命令
	if err := cmd.Start(); err != nil {
		log.Error().Err(err)
		return "", err
	}
	timer := time.After(duration)

	done := make(chan error)
	go func() {
		// 等待 ffmpeg 命令执行完毕
		done <- cmd.Wait()
	}()
	select {
	case <-timer:
		// 超时，停止 scrcpy 进程
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Error().Err(err)
		}
	case err := <-done:
		// ffmpeg 正常结束
		if err != nil {
			log.Error().Err(err)
			return "", err
		}
	}
	return filepath.Abs(fileName)
}

func (ad *adbDriver) TearDown() error {
	return nil
}
