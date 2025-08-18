package uixt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/utf7"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

func NewADBDriver(device *AndroidDevice) (*ADBDriver, error) {
	log.Info().Interface("device", device).Msg("init android adb driver")
	driver := &ADBDriver{
		Device:  device,
		Session: NewDriverSession(),
	}
	// setup driver
	if err := driver.Setup(); err != nil {
		return nil, err
	}
	return driver, nil
}

type ADBDriver struct {
	Device  *AndroidDevice
	Session *DriverSession

	// cache to avoid repeated query
	windowSize types.Size
}

func (ad *ADBDriver) runShellCommand(cmd string, args ...string) (output string, err error) {
	driverResult := &DriverRequests{
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
		ad.Session.addRequestResult(driverResult)
	}()

	// adb shell screencap -p
	if cmd == "screencap" {
		resp, err := ad.Device.ScreenCap()
		if err == nil {
			driverResult.ResponseBody = "OMITTED"
			return string(resp), nil
		}
		return "", errors.Wrap(err, "adb screencap failed")
	}

	output, err = ad.Device.RunShellCommand(cmd, args...)
	driverResult.ResponseBody = strings.TrimSpace(output)
	return output, err
}

func (ad *ADBDriver) InitSession(capabilities option.Capabilities) error {
	log.Warn().Msg("InitSession not implemented in ADBDriver")
	return nil
}

func (ad *ADBDriver) DeleteSession() error {
	log.Warn().Msg("DeleteSession not implemented in ADBDriver")
	return nil
}

func (ad *ADBDriver) Status() (deviceStatus types.DeviceStatus, err error) {
	log.Warn().Msg("Status not implemented in ADBDriver")
	return
}

func (ad *ADBDriver) GetDevice() IDevice {
	return ad.Device
}

func (ad *ADBDriver) DeviceInfo() (deviceInfo types.DeviceInfo, err error) {
	log.Warn().Msg("DeviceInfo not implemented in ADBDriver")
	return
}

func (ad *ADBDriver) BatteryInfo() (batteryInfo types.BatteryInfo, err error) {
	log.Warn().Msg("BatteryInfo not implemented in ADBDriver")
	return
}

func (ad *ADBDriver) getWindowSize() (size types.Size, err error) {
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
			return types.Size{Width: width, Height: height}, nil
		}
	}
	err = errors.New("physical window size not found by adb")
	return
}

func (ad *ADBDriver) WindowSize() (size types.Size, err error) {
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
		orientation = types.OrientationPortrait
		log.Warn().Err(err2).Msgf(
			"get window orientation failed, use default %s", orientation)
	}
	if orientation != types.OrientationPortrait {
		size.Width, size.Height = size.Height, size.Width
	}

	ad.windowSize = size // cache window size
	return size, nil
}

// Back simulates a short press on the BACK button.
func (ad *ADBDriver) Back() (err error) {
	log.Info().Msg("ADBDriver.Back")
	// adb shell input keyevent 4
	_, err = ad.runShellCommand("input", "keyevent", fmt.Sprintf("%d", KCBack))
	if err != nil {
		return errors.Wrap(err, "press back failed")
	}
	return nil
}

func (ad *ADBDriver) Orientation() (orientation types.Orientation, err error) {
	output, err := ad.runShellCommand("dumpsys", "input", "|", "grep", "'SurfaceOrientation'")
	if err != nil {
		return
	}
	re := regexp.MustCompile(`SurfaceOrientation: (\d)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 { // 确保找到了匹配项
		if matches[1] == "0" || matches[1] == "2" {
			return types.OrientationPortrait, nil
		} else if matches[1] == "1" || matches[1] == "3" {
			return types.OrientationLandscapeLeft, nil
		}
	}
	err = fmt.Errorf("not found SurfaceOrientation value")
	return
}

func (ad *ADBDriver) Home() (err error) {
	log.Info().Msg("ADBDriver.Home")
	return ad.PressKeyCode(KCHome, KMEmpty)
}

func (ad *ADBDriver) Unlock() (err error) {
	log.Info().Msg("ADBDriver.Unlock")
	// Notice: brighten should be executed before unlock
	// brighten android device screen
	if err := ad.PressKeyCode(KCWakeup, KMEmpty); err != nil {
		log.Error().Err(err).Msg("brighten android device screen failed")
	}
	// unlock android device screen
	if err := ad.PressKeyCode(KCMenu, KMEmpty); err != nil {
		log.Error().Err(err).Msg("press menu key to unlock screen failed")
	}

	// swipe up to unlock
	return ad.Swipe(500, 1500, 500, 500)
}

func (ad *ADBDriver) Backspace(count int, opts ...option.ActionOption) (err error) {
	log.Info().Int("count", count).Msg("ADBDriver.Backspace")
	if count == 0 {
		return nil
	}
	if count == 1 {
		return ad.PressKeyCode(KCDel, KMEmpty)
	}
	keyArray := make([]KeyCode, count)

	for i := range keyArray {
		keyArray[i] = KCDel
	}
	return ad.combinationKey(keyArray)
}

func (ad *ADBDriver) combinationKey(keyCodes []KeyCode) (err error) {
	if len(keyCodes) == 1 {
		return ad.PressKeyCode(keyCodes[0], KMEmpty)
	}
	strKeyCodes := make([]string, len(keyCodes))
	for i, keycode := range keyCodes {
		strKeyCodes[i] = fmt.Sprintf("%d", keycode)
	}
	_, err = ad.runShellCommand(
		"input", append([]string{"keycombination"}, strKeyCodes...)...)
	return
}

func (ad *ADBDriver) PressKeyCode(keyCode KeyCode, metaState KeyMeta) (err error) {
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

func (ad *ADBDriver) AppLaunch(packageName string) (err error) {
	log.Info().Str("packageName", packageName).Msg("ADBDriver.AppLaunch")
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

	return postHandler(ad, option.ACTION_SetTouchInfo,
		option.NewActionOptions(option.WithAntiRisk(true)))
}

func (ad *ADBDriver) AppTerminate(packageName string) (successful bool, err error) {
	log.Info().Str("packageName", packageName).Msg("ADBDriver.AppTerminate")
	// 强制停止应用，停止 <packagename> 相关的进程
	// adb shell am force-stop <packagename>
	_, err = ad.runShellCommand("am", "force-stop", packageName)
	if err != nil {
		return false, errors.Wrap(err, "force-stop app failed")
	}

	return true, nil
}

func (ad *ADBDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("ADBDriver.TapXY")
	absX, absY, err := convertToAbsolutePoint(ad, x, y)
	if err != nil {
		return err
	}
	return ad.TapAbsXY(absX, absY, opts...)
}

func (ad *ADBDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("ADBDriver.TapAbsXY")
	actionOptions := option.NewActionOptions(opts...)
	x, y, err := preHandler_TapAbsXY(ad, actionOptions, x, y)
	if err != nil {
		return err
	}
	defer postHandler(ad, option.ACTION_TapAbsXY, actionOptions)

	// adb shell input tap x y
	xStr := fmt.Sprintf("%.1f", x)
	yStr := fmt.Sprintf("%.1f", y)
	_, err = ad.runShellCommand("input", "tap", xStr, yStr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("tap <%s, %s> failed", xStr, yStr))
	}
	return nil
}

func (ad *ADBDriver) DoubleTap(x, y float64, opts ...option.ActionOption) error {
	log.Info().Float64("x", x).Float64("y", y).Msg("ADBDriver.DoubleTap")
	actionOptions := option.NewActionOptions(opts...)
	x, y, err := preHandler_DoubleTap(ad, actionOptions, x, y)
	if err != nil {
		return err
	}
	defer postHandler(ad, option.ACTION_DoubleTapXY, actionOptions)

	// adb shell input tap x y
	xStr := fmt.Sprintf("%.1f", x)
	yStr := fmt.Sprintf("%.1f", y)
	_, err = ad.runShellCommand(
		"input", "tap", xStr, yStr, ";",
		"sleep", "0.05", ";",
		"input", "tap", xStr, yStr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("double tap <%s, %s> failed", xStr, yStr))
	}
	return nil
}

func (ad *ADBDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	log.Info().Float64("x", x).Float64("y", y).Msg("ADBDriver.TouchAndHold")
	actionOptions := option.NewActionOptions(opts...)
	x, y = actionOptions.ApplyTapOffset(x, y)
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

func (ad *ADBDriver) Drag(fromX, fromY, toX, toY float64, opts ...option.ActionOption) (err error) {
	log.Info().Float64("fromX", fromX).Float64("fromY", fromY).
		Float64("toX", toX).Float64("toY", toY).Msg("ADBDriver.Drag")

	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err = preHandler_Drag(ad, actionOptions, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	defer postHandler(ad, option.ACTION_Drag, actionOptions)

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
		return errors.Wrap(err, "adb drag failed")
	}
	return nil
}

func (ad *ADBDriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	log.Info().Float64("fromX", fromX).Float64("fromY", fromY).
		Float64("toX", toX).Float64("toY", toY).Msg("ADBDriver.Swipe")

	actionOptions := option.NewActionOptions(opts...)
	fromX, fromY, toX, toY, err := preHandler_Swipe(ad, option.ACTION_SwipeCoordinate, actionOptions, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	defer postHandler(ad, option.ACTION_SwipeCoordinate, actionOptions)

	// adb shell input swipe fromX fromY toX toY
	_, err = ad.runShellCommand(
		"input", "swipe",
		fmt.Sprintf("%.1f", fromX), fmt.Sprintf("%.1f", fromY),
		fmt.Sprintf("%.1f", toX), fmt.Sprintf("%.1f", toY),
	)
	if err != nil {
		return errors.Wrap(err, "adb swipe failed")
	}
	return nil
}

func (ad *ADBDriver) ForceTouch(x, y int, pressure float64, second ...float64) error {
	return ad.ForceTouchFloat(float64(x), float64(y), pressure, second...)
}

func (ad *ADBDriver) ForceTouchFloat(x, y, pressure float64, second ...float64) (err error) {
	log.Warn().Msg("ForceTouchFloat not implemented in ADBDriver")
	return
}

func (ad *ADBDriver) Input(text string, opts ...option.ActionOption) error {
	log.Info().Str("text", text).Msg("ADBDriver.Input")
	err := ad.SendUnicodeKeys(text, opts...)
	if err == nil {
		return nil
	}
	// adb shell input text <text>
	return ad.input(text, opts...)
}

func (ad *ADBDriver) input(text string, _ ...option.ActionOption) error {
	_, err := ad.runShellCommand("input", "text", text)
	if err != nil {
		return errors.Wrap(err, "send keys failed")
	}
	return nil
}

func (ad *ADBDriver) SendUnicodeKeys(text string, opts ...option.ActionOption) (err error) {
	log.Info().Str("text", text).Msg("ADBDriver.SendUnicodeKeys")
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
	if currentIme != option.UnicodeImePackageName {
		defer func() {
			_ = ad.SetIme(currentIme)
		}()
		err = ad.SetIme(option.UnicodeImePackageName)
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
	err = ad.input("\""+strings.ReplaceAll(encodedStr, "\"", "\\\"")+"\"", opts...)
	return
}

func (ad *ADBDriver) IsAdbKeyBoardInstalled() bool {
	output, err := ad.runShellCommand("ime", "list", "-a")
	if err != nil {
		return false
	}
	return strings.Contains(output, option.AdbKeyBoardPackageName)
}

func (ad *ADBDriver) IsUnicodeIMEInstalled() bool {
	output, err := ad.runShellCommand("ime", "list", "-s")
	if err != nil {
		return false
	}
	return strings.Contains(output, option.UnicodeImePackageName)
}

func (ad *ADBDriver) ListIme() []string {
	output, err := ad.runShellCommand("ime", "list", "-s")
	if err != nil {
		return []string{}
	}
	return strings.Split(output, "\n")
}

func (ad *ADBDriver) SendKeysByAdbKeyBoard(text string) (err error) {
	defer func() {
		// Reset to default, don't care which keyboard was chosen before switch:
		if _, resetErr := ad.runShellCommand("ime", "reset"); resetErr != nil {
			log.Error().Err(err).Msg("failed to reset ime")
		}
	}()

	// Enable ADBKeyBoard from adb
	if _, err = ad.runShellCommand("ime", "enable", option.AdbKeyBoardPackageName); err != nil {
		log.Error().Err(err).Msg("failed to enable adbKeyBoard")
		return
	}
	// Switch to ADBKeyBoard from adb
	if _, err = ad.runShellCommand("ime", "set", option.AdbKeyBoardPackageName); err != nil {
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

func (ad *ADBDriver) AppClear(packageName string) error {
	log.Info().Str("packageName", packageName).Msg("ADBDriver.AppClear")
	if _, err := ad.runShellCommand("pm", "clear", packageName); err != nil {
		log.Error().Str("packageName", packageName).Err(err).Msg("failed to clear package cache")
		return err
	}

	return nil
}

func (ad *ADBDriver) Rotation() (rotation types.Rotation, err error) {
	log.Warn().Msg("Rotation not implemented in ADBDriver")
	return
}

func (ad *ADBDriver) SetRotation(rotation types.Rotation) (err error) {
	log.Warn().Msg("SetRotation not implemented in ADBDriver")
	return
}

func (ad *ADBDriver) ScreenShot(opts ...option.ActionOption) (raw *bytes.Buffer, err error) {
	resp, err := ad.Device.ScreenCap()
	if err != nil {
		return nil, errors.Wrapf(code.DeviceScreenShotError,
			"adb screencap failed %v", err)
	}
	raw = bytes.NewBuffer(resp)
	return raw, nil
}

func (ad *ADBDriver) Source(srcOpt ...option.SourceOption) (source string, err error) {
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

func (ad *ADBDriver) sourceTree(srcOpt ...option.SourceOption) (sourceTree *Hierarchy, err error) {
	source, err := ad.Source(srcOpt...)
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

func (ad *ADBDriver) tapByTextUsingHierarchy(hierarchy *Hierarchy, text string, opts ...option.ActionOption) error {
	bounds := ad.searchNodes(hierarchy.Layout, text, opts...)
	actionOptions := option.NewActionOptions(opts...)
	if len(bounds) == 0 {
		if actionOptions.IgnoreNotFoundError {
			log.Info().Msg("not found element by text " + text)
			return nil
		}
		return errors.New("not found element by text " + text)
	}
	for _, bound := range bounds {
		width, height := bound.Center()
		err := ad.TapXY(width, height, opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ud *ADBDriver) TapByXpath(xpath string, opts ...option.ActionOption) (err error) {
	source, err := ud.Source()
	if err != nil {
		log.Error().Err(err).Msg("failed to get source")
		return err
	}
	doc, err := xmlquery.Parse(strings.NewReader(source))
	if err != nil {
		log.Error().Err(err).Msg("failed to parse source")
		return err
	}
	targetNodes := xmlquery.Find(doc, xpath)
	if len(targetNodes) > 0 {
		bounds := targetNodes[0].SelectAttr("bounds")
		re := regexp.MustCompile(`\[(\d+),(\d+)\]\[(\d+),(\d+)\]`)

		matches := re.FindStringSubmatch(bounds)
		if len(matches) != 5 {
			return fmt.Errorf("failed to parse bounds: %s", bounds)
		}

		x1, _ := strconv.Atoi(matches[1])
		y1, _ := strconv.Atoi(matches[2])
		x2, _ := strconv.Atoi(matches[3])
		y2, _ := strconv.Atoi(matches[4])

		centerX := float64(x1+x2) / 2
		centerY := float64(y1+y2) / 2
		log.Info().Str("xpath", xpath).Str("bounds", bounds).Msg("find node by xpath success")
		return ud.TapAbsXY(centerX, centerY, opts...)
	}

	log.Error().Str("xpath", xpath).Msg("failed to find node by xpath")
	return errors.New("failed to find node by xpath")
}

func (ad *ADBDriver) searchNodes(nodes []Layout, text string, opts ...option.ActionOption) []Bounds {
	actionOptions := option.NewActionOptions(opts...)
	var results []Bounds
	for _, node := range nodes {
		result := ad.searchNodes(node.Layout, text, opts...)
		results = append(results, result...)
		if actionOptions.Regex {
			// regex on, check if match regex
			if !regexp.MustCompile(text).MatchString(node.Text) {
				continue
			}
		} else {
			// regex off, check if match exactly
			if node.Text != text {
				ad.searchNodes(node.Layout, text, opts...)
				continue
			}
		}
		if node.Bounds != nil {
			results = append(results, *node.Bounds)
		}
	}
	return results
}

func (ad *ADBDriver) StartCaptureLog(identifier ...string) (err error) {
	log.Info().Msg("start adb log recording")
	// start logcat
	err = ad.Device.Logcat.CatchLogcat("iesqaMonitor:V")
	if err != nil {
		err = errors.Wrap(code.DeviceCaptureLogError,
			fmt.Sprintf("start adb log recording failed: %v", err))
		return err
	}
	return nil
}

func (ad *ADBDriver) StopCaptureLog() (result interface{}, err error) {
	defer func() {
		log.Info().Msg("stop adb log recording")
		err = ad.Device.Logcat.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to get adb log recording")
		}
	}()
	if err != nil {
		log.Error().Err(err).Msg("failed to close adb log writer")
	}
	pointRes := ConvertPoints(ad.Device.Logcat.logs)
	// 没有解析到打点日志，走兜底逻辑
	if len(pointRes) == 0 {
		log.Info().Msg("action log is null, use action file >>>")
		actionLogDirPath := config.GetConfig().ActionLogDirPath()
		files := []string{}
		actionLogRegStr := `.*data_\d+\.txt`
		ad.Device.PullFolder(config.DeviceActionLogFilePath, actionLogDirPath)
		err = filepath.Walk(actionLogDirPath, func(path string, info fs.FileInfo, err error) error {
			// 只是需要日志文件
			if ok, _ := regexp.MatchString(actionLogRegStr, path); ok {
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

func (ad *ADBDriver) GetSession() *DriverSession {
	return ad.Session
}

func (ad *ADBDriver) ForegroundInfo() (app types.AppInfo, err error) {
	// Get foreground app package info using evalite service
	packageInfo, err := ad.getForegroundPackageInfo()
	if err != nil {
		log.Error().Err(err).Msg("failed to get foreground app info")
		return app, err
	}

	// Parse package info JSON
	packageInfo = strings.TrimSpace(packageInfo)
	if packageInfo == "" {
		err = errors.New("foreground app output is empty")
		log.Error().Err(err).Msg("get foreground app info failed")
		return app, err
	}
	if err = json.Unmarshal([]byte(packageInfo), &app); err != nil {
		log.Error().Err(err).Str("packageInfo", packageInfo).Msg("failed to parse package info")
		return app, err
	}
	return app, nil
}

// getForegroundPackageInfo executes the evalite service command to get foreground app info
func (ad *ADBDriver) getForegroundPackageInfo() (string, error) {
	const maxRetries = 2
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		packageInfo, err := ad.runShellCommand("CLASSPATH=/data/local/tmp/evalite",
			"app_process", "/", "com.bytedance.iesqa.eval_process.PackageService", "2>/dev/null")
		if err == nil {
			return packageInfo, nil
		}
		lastErr = err
		log.Warn().Err(err).Int("attempt", i+1).Msg("failed to get foreground package info, retrying")
	}

	return "", lastErr
}

func (ad *ADBDriver) SetIme(imeRegx string) error {
	log.Info().Str("imeRegx", imeRegx).Msg("ADBDriver.SetIme")
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
	brand, _ := ad.Device.Brand()
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
			appInfo, err := ad.ForegroundInfo()
			_ = ad.AppLaunch(packageName)
			if err == nil && packageName != option.UnicodeImePackageName {
				time.Sleep(10 * time.Second)
				nextAppInfo, err := ad.ForegroundInfo()
				log.Info().Str("beforeFocusedPackage", appInfo.PackageName).Str("afterFocusedPackage", nextAppInfo.PackageName).Msg("")
				if err == nil && nextAppInfo.PackageName != appInfo.PackageName {
					_ = ad.PressKeyCode(KCBack, KMEmpty)
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

func (ad *ADBDriver) GetIme() (ime string, err error) {
	currentIme, err := ad.runShellCommand("settings", "get", "secure", "default_input_method")
	if err != nil {
		log.Warn().Err(err).Msgf("get default ime failed")
		return
	}
	currentIme = strings.TrimSpace(currentIme)
	return currentIme, nil
}

func (ad *ADBDriver) ScreenRecord(opts ...option.ActionOption) (videoPath string, err error) {
	log.Info().Msg("ADBDriver.ScreenRecord")
	options := option.NewActionOptions(opts...)

	var filePath string
	if options.ScreenRecordPath != "" {
		filePath = options.ScreenRecordPath
	} else {
		timestamp := time.Now().Format("20060102_150405") + fmt.Sprintf("_%03d", time.Now().UnixNano()/1e6%1000)
		filePath = filepath.Join(config.GetConfig().ScreenShotsPath(), fmt.Sprintf("%s.mp4", timestamp))
	}

	var ctx context.Context
	if options.Context != nil {
		ctx = options.Context
	} else {
		ctx = context.Background()
	}

	var cancel context.CancelFunc
	duration := options.ScreenRecordDuration
	if duration == 0 {
		duration = options.Duration
	}
	if duration != 0 {
		ctx, cancel = context.WithTimeout(ctx,
			time.Duration(duration*float64(time.Second)))
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// get android system version
	var sysVersion int
	if systemVersion, err := ad.Device.SystemVersion(); err == nil {
		if version, err := strconv.Atoi(systemVersion); err == nil {
			sysVersion = version
		}
	}
	if sysVersion == 0 {
		log.Warn().Err(err).Msg("get android system version failed")
	}

	var useAdbScreenRecord bool
	audioOn := options.ScreenRecordWithAudio
	if options.ScreenRecordWithScrcpy {
		useAdbScreenRecord = false
	} else if !audioOn {
		log.Info().Bool("audioOn", audioOn).Msg("screen record with adb screenrecord by default")
		useAdbScreenRecord = true
	} else if sysVersion != 0 && sysVersion < 11 {
		// scrcpy audio forwarding is supported for devices with Android 11 or higher
		// https://github.com/Genymobile/scrcpy/blob/master/doc/audio.md
		log.Warn().Bool("audioOn", audioOn).Int("version", sysVersion).
			Msg("Audio disabled, it is only supported for Android >= 11, use adb screenrecord")
		useAdbScreenRecord = true
	}

	defer func() {
		if err == nil {
			filePath, err = filepath.Abs(filePath)
			if err != nil {
				err = errors.Wrap(err, "get absolute path failed")
			} else {
				log.Info().Str("path", filePath).Msg("screen record success")
			}
		}
	}()

	if useAdbScreenRecord {
		// screen record with adb screenrecord
		// adb screenrecord duration is limited in range [1,180] seconds
		res, err := ad.Device.ScreenRecord(ctx)
		if err != nil {
			return "", errors.Wrap(err, "screen record failed")
		}
		if err := os.WriteFile(filePath, res, 0o644); err != nil {
			return "", errors.Wrap(err, "write screen record file failed")
		}
		return filePath, nil
	}

	// screen record with scrcpy
	log.Info().Float64("duration(s)", duration).Msg("screen record with scrcpy")

	// start scrcpy
	cmd := exec.Command(
		"scrcpy",
		"-s", ad.Device.Serial(),
		fmt.Sprintf("--record=%s", filePath),
		"--record-format=mp4",
		"--max-fps=30",
		"--no-playback", // Disable video and audio playback on the computer
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return "", errors.Wrap(err, "start screen record failed")
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// timeout or cancelled
		log.Info().Msg("screen recording stopped")
		if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Error().Err(err).Msg("failed to stop scrcpy process")
			_ = cmd.Process.Kill() // 强制结束进程
		}
		<-done // 等待进程完全退出
	case err := <-done:
		log.Info().Msg("scrcpy exited")
		if err != nil {
			return "", errors.Wrap(err, "screen record with scrcpy failed")
		}
	}

	return filePath, nil
}

func (ad *ADBDriver) Setup() error {
	log.Warn().Msg("Setup not implemented in ADBDriver")
	return nil
}

func (ad *ADBDriver) TearDown() error {
	log.Warn().Msg("TearDown not implemented in ADBDriver")
	return nil
}

func (ad *ADBDriver) OpenUrl(url string) (err error) {
	_, err = ad.runShellCommand(
		"am", "start", "-W", "-a", "android.intent.action.VIEW",
		"-d", fmt.Sprintf("'%s'", url))
	return
}

var androidButtonMap = map[types.DeviceButton]string{
	types.DeviceButtonBack:       "KEYCODE_BACK",
	types.DeviceButtonHome:       "KEYCODE_HOME",
	types.DeviceButtonEnter:      "KEYCODE_ENTER",
	types.DeviceButtonVolumeUp:   "KEYCODE_VOLUME_UP",
	types.DeviceButtonVolumeDown: "KEYCODE_VOLUME_DOWN",
}

func (ad *ADBDriver) PressButton(button types.DeviceButton) error {
	buttonName, ok := androidButtonMap[button]
	if !ok {
		return fmt.Errorf("unsupported button: %s", button)
	}
	_, err := ad.runShellCommand("input", "keyevent", buttonName)
	return err
}

func (ad *ADBDriver) PushImage(localPath string) error {
	log.Info().Str("localPath", localPath).Msg("ADBDriver.PushImage")
	remoteDir := "/sdcard/DCIM/Camera/"
	return ad.PushFile(localPath, remoteDir)
}

// PullImages pulls all images from device's DCIM/Camera directory to local directory
func (ad *ADBDriver) PullImages(localDir string) error {
	log.Info().Str("localDir", localDir).Msg("ADBDriver.PullImages")
	remoteDir := "/sdcard/DCIM/Camera/"

	// create local directory if not exists
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	files, err := ad.Device.List(remoteDir)
	if err != nil {
		return fmt.Errorf("failed to list directory %s: %w", remoteDir, err)
	}

	for _, file := range files {
		// filter image files by extension
		ext := strings.ToLower(path.Ext(file.Name))
		if !isImageFile(ext) {
			continue
		}

		remotePath := path.Join(remoteDir, file.Name)
		localPath := path.Join(localDir, file.Name)

		// check if file already exists
		if _, err := os.Stat(localPath); err == nil {
			log.Debug().Str("localPath", localPath).Msg("file already exists, skipping")
			continue
		}

		// create local file
		f, err := os.Create(localPath)
		if err != nil {
			log.Error().Err(err).Str("localPath", localPath).Msg("failed to create local file")
			continue
		}
		defer f.Close()

		// pull image file
		if err := ad.Device.Pull(remotePath, f); err != nil {
			log.Error().Err(err).
				Str("remotePath", remotePath).
				Str("localPath", localPath).
				Msg("failed to pull image")
			continue // continue with next file
		}
		log.Info().
			Str("remotePath", remotePath).
			Str("localPath", localPath).
			Msg("image pulled successfully")
	}
	return nil
}

// isImageFile checks if the file extension is an image format
func isImageFile(ext string) bool {
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".heic": true,
	}
	return imageExts[ext]
}

func (ad *ADBDriver) ClearImages() error {
	log.Info().Msg("ADBDriver.ClearImages")
	_, _ = ad.Device.RunShellCommand("rm", "-rf", "/sdcard/DCIM/Camera/*")
	return nil
}

func (ad *ADBDriver) PushFile(localPath string, remoteDir string) error {
	log.Info().Str("localPath", localPath).Str("remoteDir", remoteDir).Msg("ADBDriver.PushFile")
	remotePath := path.Join(remoteDir, path.Base(localPath))
	if err := ad.Device.PushFile(localPath, remotePath); err != nil {
		return err
	}
	// refresh
	_, _ = ad.Device.RunShellCommand("am", "broadcast",
		"-a", "android.intent.action.MEDIA_SCANNER_SCAN_FILE",
		"-d", fmt.Sprintf("file://%s", remotePath))
	return nil
}

func (ad *ADBDriver) PullFiles(localDir string, remoteDirs ...string) error {
	log.Info().Str("localDir", localDir).Strs("remoteDirs", remoteDirs).Msg("ADBDriver.PullFiles")

	// create local directory if not exists
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	for _, remoteDir := range remoteDirs {
		files, err := ad.Device.List(remoteDir)
		if err != nil {
			return fmt.Errorf("failed to list directory %s: %w", remoteDir, err)
		}

		for _, file := range files {
			remotePath := path.Join(remoteDir, file.Name)
			localPath := path.Join(localDir, file.Name)

			// check if file already exists
			if _, err := os.Stat(localPath); err == nil {
				log.Debug().Str("localPath", localPath).Msg("file already exists, skipping")
				continue
			}

			// create local file
			f, err := os.Create(localPath)
			if err != nil {
				log.Error().Err(err).Str("localPath", localPath).Msg("failed to create local file")
				continue
			}
			defer f.Close()

			// pull image file
			if err := ad.Device.Pull(remotePath, f); err != nil {
				log.Error().Err(err).
					Str("remotePath", remotePath).
					Str("localPath", localPath).
					Msg("failed to pull file")
				continue // continue with next file
			}
			log.Info().
				Str("remotePath", remotePath).
				Str("localPath", localPath).
				Msg("file pulled successfully")
		}
	}
	return nil
}

func (ad *ADBDriver) ClearFiles(paths ...string) error {
	log.Info().Strs("paths", paths).Msg("ADBDriver.ClearFiles")
	for _, path := range paths {
		_, _ = ad.Device.RunShellCommand("rm", "-rf", path)
	}
	return nil
}

func (ad *ADBDriver) GetPasteboard() (content string, err error) {
	/**
	adb shell am broadcast -n  io.appium.settings/.receivers.ClipboardReceiver -a io.appium.settings.clipboard.get
	Broadcasting: Intent { act=io.appium.settings.clipboard.get flg=0x400000 cmp=io.appium.settings/.receivers.ClipboardReceiver }
	Broadcast completed: result=-1, data="SEhHRw=="
		**/

	// Check and switch input method if needed, similar to SendUnicodeKeys
	currentIme, err := ad.GetIme()
	if err != nil {
		return "", err
	}

	// If current IME is not the required one, switch temporarily and restore later
	if currentIme != option.UnicodeImePackageName {
		defer func() {
			_ = ad.SetIme(currentIme)
		}()

		err = ad.SetIme(option.UnicodeImePackageName)
		if err != nil {
			log.Warn().Err(err).Msgf("set Unicode Ime failed for clipboard operation")
			// Continue anyway, might still work with current IME
		}
	}

	res, err := ad.Device.RunShellCommand("am", "broadcast", "-n", option.AppiumSettingsPackageName+"/.receivers.ClipboardReceiver", "-a", option.AppiumSettingsPackageName+".clipboard.get")
	if err != nil {
		return "", err
	}

	// Parse the response to extract the base64 encoded data
	dataPrefix := "data=\""
	dataIndex := strings.Index(res, dataPrefix)
	if dataIndex == -1 {
		return "", fmt.Errorf("clipboard data not found in response: %s", res)
	}

	dataStart := dataIndex + len(dataPrefix)
	dataEnd := strings.Index(res[dataStart:], "\"")
	if dataEnd == -1 {
		return "", fmt.Errorf("malformed clipboard data in response: %s", res)
	}

	base64Data := res[dataStart : dataStart+dataEnd]
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode clipboard content: %w", err)
	}

	return string(decodedData), nil
}

type ExportPoint struct {
	Start     int         `json:"start" yaml:"start"`
	End       int         `json:"end" yaml:"end"`
	From      interface{} `json:"from" yaml:"from"`
	To        interface{} `json:"to" yaml:"to"`
	Operation string      `json:"operation" yaml:"operation"`
	Ext       string      `json:"ext" yaml:"ext"`
	RunTime   int         `json:"run_time,omitempty" yaml:"run_time,omitempty"`
}

func ConvertPoints(lines []string) (eps []ExportPoint) {
	log.Info().Msg("ConvertPoints")
	log.Info().Msg(strings.Join(lines, "\n"))
	for _, line := range lines {
		if strings.Contains(line, "ext") {
			idx := strings.Index(line, "{")
			if idx == -1 {
				continue
			}
			line = line[idx:]
			p := ExportPoint{}
			err := json.Unmarshal([]byte(line), &p)
			if err != nil {
				log.Error().Msg("failed to parse point data")
				continue
			}
			log.Info().Msg(line)
			eps = append(eps, p)
		}
	}
	return
}

func (ad *ADBDriver) HoverBySelector(selector string, options ...option.ActionOption) (err error) {
	return err
}

func (ad *ADBDriver) TapBySelector(text string, opts ...option.ActionOption) error {
	log.Info().Str("text", text).Msg("ADBDriver.TapByHierarchy")
	sourceTree, err := ad.sourceTree()
	if err != nil {
		return err
	}
	return ad.tapByTextUsingHierarchy(sourceTree, text, opts...)
}

func (ad *ADBDriver) SecondaryClick(x, y float64) (err error) {
	return err
}

func (ad *ADBDriver) SecondaryClickBySelector(selector string, options ...option.ActionOption) (err error) {
	return err
}
