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
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
	"github.com/httprunner/httprunner/v5/pkg/uixt/types"
)

func NewADBDriver(device *AndroidDevice) (*ADBDriver, error) {
	log.Info().Interface("device", device).Msg("init android adb driver")
	driver := &ADBDriver{
		Device:  device,
		Session: &Session{},
	}
	driver.InitSession(nil)
	return driver, nil
}

type ADBDriver struct {
	Device  *AndroidDevice
	Session *Session
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
	ad.Session = &Session{}
	ad.Session.Reset()
	return nil
}

func (ad *ADBDriver) DeleteSession() (err error) {
	return types.ErrDriverNotImplemented
}

func (ad *ADBDriver) Status() (deviceStatus types.DeviceStatus, err error) {
	err = types.ErrDriverNotImplemented
	return
}

func (ad *ADBDriver) GetDevice() IDevice {
	return ad.Device
}

func (ad *ADBDriver) DeviceInfo() (deviceInfo types.DeviceInfo, err error) {
	err = types.ErrDriverNotImplemented
	return
}

func (ad *ADBDriver) BatteryInfo() (batteryInfo types.BatteryInfo, err error) {
	err = types.ErrDriverNotImplemented
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
	if !ad.Session.windowSize.IsNil() {
		// use cached window size
		return ad.Session.windowSize, nil
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

	ad.Session.windowSize = size // cache window size
	return size, nil
}

// Back simulates a short press on the BACK button.
func (ad *ADBDriver) Back() (err error) {
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
	return ad.PressKeyCode(KCHome, KMEmpty)
}

func (ad *ADBDriver) Unlock() (err error) {
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

func (ad *ADBDriver) AppTerminate(packageName string) (successful bool, err error) {
	// 强制停止应用，停止 <packagename> 相关的进程
	// adb shell am force-stop <packagename>
	_, err = ad.runShellCommand("am", "force-stop", packageName)
	if err != nil {
		return false, errors.Wrap(err, "force-stop app failed")
	}

	return true, nil
}

func (ad *ADBDriver) TapXY(x, y float64, opts ...option.ActionOption) error {
	absX, absY, err := convertToAbsolutePoint(ad, x, y)
	if err != nil {
		return err
	}
	return ad.TapAbsXY(absX, absY, opts...)
}

func (ad *ADBDriver) TapAbsXY(x, y float64, opts ...option.ActionOption) error {
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

func (ad *ADBDriver) DoubleTapXY(x, y float64, opts ...option.ActionOption) error {
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

func (ad *ADBDriver) TouchAndHold(x, y float64, opts ...option.ActionOption) (err error) {
	actionOptions := option.NewActionOptions(opts...)

	if len(actionOptions.Offset) == 2 {
		x += float64(actionOptions.Offset[0])
		y += float64(actionOptions.Offset[1])
	}
	x += actionOptions.GetRandomOffset()
	y += actionOptions.GetRandomOffset()
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
	absFromX, absFromY, absToX, absToY, err := convertToAbsoluteCoordinates(ad, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}

	actionOptions := option.NewActionOptions(opts...)
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
		fmt.Sprintf("%.1f", absFromX), fmt.Sprintf("%.1f", absFromY),
		fmt.Sprintf("%.1f", absToX), fmt.Sprintf("%.1f", absToY),
		fmt.Sprintf("%d", int(duration)),
	)
	if err != nil {
		return errors.Wrap(err, "adb drag failed")
	}
	return nil
}

func (ad *ADBDriver) Swipe(fromX, fromY, toX, toY float64, opts ...option.ActionOption) error {
	absFromX, absFromY, absToX, absToY, err := convertToAbsoluteCoordinates(ad, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}

	// adb shell input swipe fromX fromY toX toY
	_, err = ad.runShellCommand(
		"input", "swipe",
		fmt.Sprintf("%.1f", absFromX), fmt.Sprintf("%.1f", absFromY),
		fmt.Sprintf("%.1f", absToX), fmt.Sprintf("%.1f", absToY),
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
	err = types.ErrDriverNotImplemented
	return
}

func (ad *ADBDriver) Input(text string, opts ...option.ActionOption) error {
	err := ad.SendUnicodeKeys(text, opts...)
	if err == nil {
		return nil
	}
	// adb shell input text <text>
	_, err = ad.runShellCommand("input", "text", text)
	if err != nil {
		return errors.Wrap(err, "send keys failed")
	}
	return nil
}

func (ad *ADBDriver) SendUnicodeKeys(text string, opts ...option.ActionOption) (err error) {
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
	err = ad.Input("\""+strings.ReplaceAll(encodedStr, "\"", "\\\"")+"\"", opts...)
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
	if _, err := ad.runShellCommand("pm", "clear", packageName); err != nil {
		log.Error().Str("packageName", packageName).Err(err).Msg("failed to clear package cache")
		return err
	}

	return nil
}

func (ad *ADBDriver) Rotation() (rotation types.Rotation, err error) {
	err = types.ErrDriverNotImplemented
	return
}

func (ad *ADBDriver) SetRotation(rotation types.Rotation) (err error) {
	err = types.ErrDriverNotImplemented
	return
}

func (ad *ADBDriver) ScreenShot(opts ...option.ActionOption) (raw *bytes.Buffer, err error) {
	resp, err := ad.runShellCommand("screencap", "-p")
	if err != nil {
		return nil, errors.Wrapf(code.DeviceScreenShotError,
			"adb screencap failed %v", err)
	}
	raw = bytes.NewBuffer([]byte(resp))

	actionOptions := option.NewActionOptions(opts...)
	if actionOptions.ScreenShotFileName != "" {
		// save screenshot to file
		path, err := saveScreenShot(raw, actionOptions.ScreenShotFileName)
		if err != nil {
			return nil, errors.Wrapf(code.DeviceScreenShotError,
				"save screenshot file failed %v", err)
		}
		log.Info().Str("path", path).Msg("screenshot saved")
	}

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

func (ad *ADBDriver) TapByText(text string, opts ...option.ActionOption) error {
	sourceTree, err := ad.sourceTree()
	if err != nil {
		return err
	}
	return ad.tapByTextUsingHierarchy(sourceTree, text, opts...)
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

func (ad *ADBDriver) TapByTexts(actions ...TapTextAction) error {
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
		logFilePathPrefix := fmt.Sprintf("%v/data", config.ActionLogFilePath)
		files := []string{}
		ad.Device.RunShellCommand("pull", config.DeviceActionLogFilePath, config.ActionLogFilePath)
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

func (ad *ADBDriver) GetSession() *Session {
	return ad.Session
}

func (ad *ADBDriver) ForegroundInfo() (app types.AppInfo, err error) {
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

func (ad *ADBDriver) SetIme(imeRegx string) error {
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

func (ad *ADBDriver) ScreenRecord(duration time.Duration) (videoPath string, err error) {
	timestamp := time.Now().Format("20060102_150405") + fmt.Sprintf("_%03d", time.Now().UnixNano()/1e6%1000)
	fileName := filepath.Join(config.ScreenShotsPath, fmt.Sprintf("%s.mp4", timestamp))

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
		"-s", ad.Device.Serial(),
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

func (ad *ADBDriver) Setup() error {
	return nil
}

func (ad *ADBDriver) TearDown() error {
	return nil
}
