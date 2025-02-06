package uixt

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/httprunner/funplugin/myexec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/internal/json"
	"github.com/httprunner/httprunner/v5/pkg/gadb"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var (
	DouyinServerPort = 32316

	// adb server
	AdbServerHost = "localhost"
	AdbServerPort = gadb.AdbServerPort // 5037

	// uiautomator2 server
	UIA2ServerHost            = "localhost"
	UIA2ServerPort            = 6790
	UIA2ServerPackageName     = "io.appium.uiautomator2.server"
	UIA2ServerTestPackageName = "io.appium.uiautomator2.server.test"

	EvalInstallerPackageName   = "sogou.mobile.explorer"
	InstallViaInstallerCommand = "am start -S -n sogou.mobile.explorer/.PackageInstallerActivity -d"
)

//go:embed evalite
var evalite embed.FS

const forwardToPrefix = "forward-to-"

func NewAndroidDevice(opts ...option.AndroidDeviceOption) (device *AndroidDevice, err error) {
	androidOptions := &option.AndroidDeviceConfig{
		UIA2IP:   UIA2ServerHost,
		UIA2Port: UIA2ServerPort,
	}
	for _, option := range opts {
		option(androidOptions)
	}
	deviceList, err := GetAndroidDevices(androidOptions.SerialNumber)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}

	if androidOptions.SerialNumber == "" && len(deviceList) > 1 {
		return nil, errors.Wrap(code.DeviceConnectionError, "more than one device connected, please specify the serial")
	}

	dev := deviceList[0]

	if androidOptions.SerialNumber == "" {
		selectSerial := dev.Serial()
		androidOptions.SerialNumber = selectSerial
		log.Warn().
			Str("serial", androidOptions.SerialNumber).
			Msg("android SerialNumber is not specified, select the first one")
	}

	device = &AndroidDevice{
		AndroidDeviceConfig: androidOptions,
		d:                   dev,
		logcat:              NewAdbLogcat(device.SerialNumber),
	}

	evalToolRaw, err := evalite.ReadFile("evalite")
	if err != nil {
		return nil, errors.Wrap(code.LoadFileError, err.Error())
	}
	err = dev.Push(bytes.NewReader(evalToolRaw), "/data/local/tmp/evalite", time.Now())
	if err != nil {
		return nil, errors.Wrap(code.DeviceShellExecError, err.Error())
	}
	log.Info().Str("serial", device.SerialNumber).Msg("init android device")
	return device, nil
}

func GetAndroidDevices(serial ...string) (devices []*gadb.Device, err error) {
	var adbClient gadb.Client
	if adbClient, err = gadb.NewClientWith(AdbServerHost, AdbServerPort); err != nil {
		return nil, err
	}

	if devices, err = adbClient.DeviceList(); err != nil {
		return nil, err
	}

	var deviceList []*gadb.Device
	// filter by serial
	for _, d := range devices {
		for _, s := range serial {
			if s != "" && s != d.Serial() {
				continue
			}
			deviceList = append(deviceList, d)
		}
	}

	if len(deviceList) == 0 {
		var err error
		if serial == nil || (len(serial) == 1 && serial[0] == "") {
			err = fmt.Errorf("no android device found")
		} else {
			err = fmt.Errorf("no android device found for serial %v", serial)
		}
		return nil, err
	}
	return deviceList, nil
}

type AndroidDevice struct {
	*option.AndroidDeviceConfig
	d      *gadb.Device
	logcat *AdbLogcat
}

func (dev *AndroidDevice) Init() error {
	dev.d.RunShellCommand("ime", "enable", UnicodeImePackageName)
	dev.d.RunShellCommand("rm", "-r", config.DeviceActionLogFilePath)

	if dev.UIA2 {
		// uiautomator2 server must be started before

		// check uiautomator server package installed
		if !dev.d.IsPackageInstalled(UIA2ServerPackageName) {
			return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
				"%s not installed", UIA2ServerPackageName)
		}
		if !dev.d.IsPackageInstalled(UIA2ServerTestPackageName) {
			return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
				"%s not installed", UIA2ServerTestPackageName)
		}

		// TODO: check uiautomator server package running
		// if dev.d.IsPackageRunning(UIA2ServerPackageName) {
		// 	return nil
		// }

		// start uiautomator2 server
		go func() {
			if err := dev.startUIA2Server(); err != nil {
				log.Error().Err(err).Msg("start UIA2 failed")
			}
		}()
		time.Sleep(5 * time.Second) // wait for uiautomator2 server start
	}
	return nil
}

func (dev *AndroidDevice) UUID() string {
	return dev.SerialNumber
}

func (dev *AndroidDevice) LogEnabled() bool {
	return dev.LogOn
}

func (dev *AndroidDevice) NewDriver(opts ...option.DriverOption) (driverExt *DriverExt, err error) {
	options := option.NewDriverOptions(opts...)

	var driver IWebDriver
	if dev.UIA2 || dev.LogOn {
		driver, err = dev.NewUSBDriver(options.Capabilities)
	} else if dev.STUB {
		driver, err = dev.NewStubDriver(options.Capabilities)
	} else {
		driver, err = NewADBDriver(dev)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to init UIA driver")
	}

	driverExt, err = newDriverExt(dev, driver, opts...)
	if err != nil {
		return nil, err
	}

	if dev.LogOn {
		err = driverExt.Driver.StartCaptureLog("hrp_adb_log")
		if err != nil {
			return nil, err
		}
	}

	return driverExt, nil
}

// NewUSBDriver creates new client via USB connected device, this will also start a new session.
func (dev *AndroidDevice) NewUSBDriver(capabilities option.Capabilities) (driver IWebDriver, err error) {
	localPort, err := dev.d.Forward(dev.UIA2Port)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				localPort, dev.UIA2Port, err))
	}

	rawURL := fmt.Sprintf("http://%s%d:%d/wd/hub",
		forwardToPrefix, localPort, dev.UIA2Port)
	uiaDriver, err := NewUIADriver(capabilities, rawURL)
	if err != nil {
		_ = dev.d.ForwardKill(localPort)
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}
	uiaDriver.adbClient = dev.d
	uiaDriver.logcat = dev.logcat

	return uiaDriver, nil
}

func (dev *AndroidDevice) NewStubDriver(capabilities option.Capabilities) (driver *StubAndroidDriver, err error) {
	socketLocalPort, err := dev.d.Forward(StubSocketName)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%s failed: %v",
				socketLocalPort, StubSocketName, err))
	}

	serverLocalPort, err := dev.d.Forward(DouyinServerPort)
	if err != nil {
		return nil, errors.Wrap(code.DeviceConnectionError,
			fmt.Sprintf("forward port %d->%d failed: %v",
				serverLocalPort, DouyinServerPort, err))
	}

	rawURL := fmt.Sprintf("http://%s%d:%d",
		forwardToPrefix, serverLocalPort, DouyinServerPort)

	stubDriver, err := newStubAndroidDriver(fmt.Sprintf("127.0.0.1:%d", socketLocalPort), rawURL)
	if err != nil {
		_ = dev.d.ForwardKill(socketLocalPort)
		_ = dev.d.ForwardKill(serverLocalPort)
		return nil, errors.Wrap(code.DeviceConnectionError, err.Error())
	}
	stubDriver.adbClient = dev.d
	stubDriver.logcat = dev.logcat

	return stubDriver, nil
}

// NewHTTPDriver creates new remote HTTP client, this will also start a new session.
func (dev *AndroidDevice) NewHTTPDriver(capabilities option.Capabilities) (driver IWebDriver, err error) {
	rawURL := fmt.Sprintf("http://%s:%d/wd/hub", dev.UIA2IP, dev.UIA2Port)
	uiaDriver, err := NewUIADriver(capabilities, rawURL)
	if err != nil {
		return nil, err
	}

	uiaDriver.adbClient = dev.d
	uiaDriver.logcat = dev.logcat
	return uiaDriver, nil
}

func (dev *AndroidDevice) StartPerf() error {
	// TODO
	return nil
}

func (dev *AndroidDevice) StopPerf() string {
	// TODO
	return ""
}

func (dev *AndroidDevice) StartPcap() error {
	// TODO
	return nil
}

func (dev *AndroidDevice) StopPcap() string {
	// TODO
	return ""
}

func (dev *AndroidDevice) Uninstall(packageName string) error {
	_, err := dev.d.RunShellCommand("uninstall", packageName)
	return err
}

func (dev *AndroidDevice) Install(apkPath string, opts ...option.InstallOption) error {
	installOpts := option.NewInstallOptions(opts...)
	brand, err := dev.d.Brand()
	if err != nil {
		return err
	}
	args := []string{}
	if installOpts.Reinstall {
		args = append(args, "-r")
	}
	if installOpts.GrantPermission {
		args = append(args, "-g")
	}
	if installOpts.Downgrade {
		args = append(args, "-d")
	}
	switch strings.ToLower(brand) {
	case "vivo":
		return dev.installVivoSilent(apkPath, args...)
	case "oppo", "realme", "oneplus":
		if dev.d.IsPackageInstalled(EvalInstallerPackageName) {
			return dev.installViaInstaller(apkPath, args...)
		}
		log.Warn().Msg("oppo not install eval installer")
		return dev.installCommon(apkPath, args...)
	default:
		return dev.installCommon(apkPath, args...)
	}
}

func (dev *AndroidDevice) installVivoSilent(apkPath string, args ...string) error {
	currentTime := builtin.GetCurrentDay()
	md5HashInBytes := md5.Sum([]byte(currentTime))
	verifyCode := hex.EncodeToString(md5HashInBytes[:])
	verifyCode = base64.StdEncoding.EncodeToString([]byte(verifyCode))
	verifyCode = verifyCode[:8]
	verifyCode = "-V" + verifyCode
	args = append([]string{verifyCode}, args...)
	_, err := dev.d.InstallAPK(apkPath, args...)
	return err
}

func (dev *AndroidDevice) installViaInstaller(apkPath string, args ...string) error {
	appRemotePath := "/data/local/tmp/" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".apk"
	err := dev.d.PushFile(apkPath, appRemotePath, time.Now())
	if err != nil {
		return err
	}
	done := make(chan error)
	defer func() {
		close(done)
	}()
	logcat := NewAdbLogcatWithCallback(dev.d.Serial(), func(line string) {
		re := regexp.MustCompile(`\{.*?}`)
		match := re.FindString(line)
		if match == "" {
			return
		}
		var result InstallResult
		err := json.Unmarshal([]byte(match), &result)
		if err != nil {
			log.Warn().Msg("parse Install msg line error: " + match)
			return
		}
		if result.Result == 0 {
			// 安装成功
			done <- nil
		} else {
			done <- errors.New(match)
		}
	})
	err = logcat.CatchLogcat("PackageInstallerCallback")
	if err != nil {
		return err
	}
	defer func() {
		_ = logcat.Stop()
	}()

	// 需要监听是否完成安装
	command := strings.Split(InstallViaInstallerCommand, " ")
	args = append(command, appRemotePath)
	_, err = dev.d.RunShellCommand("am", args[1:]...)
	if err != nil {
		return err
	}
	// 等待安装完成或超时
	timeout := 3 * time.Minute
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("installation timed out after %v", timeout)
	}
}

func (dev *AndroidDevice) installCommon(apkPath string, args ...string) error {
	_, err := dev.d.InstallAPK(apkPath, args...)
	return err
}

func (dev *AndroidDevice) GetCurrentWindow() (windowInfo WindowInfo, err error) {
	// adb shell dumpsys window | grep -E 'mCurrentFocus|mFocusedApp'
	output, err := dev.d.RunShellCommand("dumpsys", "window", "|", "grep", "-E", "'mCurrentFocus|mFocusedApp'")
	if err != nil {
		return WindowInfo{}, errors.Wrap(err, "get current window failed")
	}
	// mCurrentFocus=Window{a33bc55 u0 com.miui.home/com.miui.home.launcher.Launcher}
	reFocus := regexp.MustCompile(`mCurrentFocus=Window{.*? (\S+)/(\S+)}`)
	matches := reFocus.FindStringSubmatch(output)
	if len(matches) == 3 {
		windowInfo = WindowInfo{
			PackageName: matches[1],
			Activity:    matches[2],
		}
		return windowInfo, nil
	}
	// mFocusedApp=ActivityRecord{2db504f u0 com.miui.home/.launcher.Launcher t2}
	reApp := regexp.MustCompile(`mFocusedApp=ActivityRecord{.*? (\S+)/(\S+?)\s`)
	matches = reApp.FindStringSubmatch(output)
	if len(matches) == 3 {
		windowInfo = WindowInfo{
			PackageName: matches[1],
			Activity:    matches[2],
		}
		return windowInfo, nil
	}

	// adb shell dumpsys activity activities | grep mResumedActivity
	output, err = dev.d.RunShellCommand("dumpsys", "activity", "activities", "|", "grep", "mResumedActivity")
	if err != nil {
		return WindowInfo{}, errors.Wrap(err, "get current activity failed")
	}
	// mResumedActivity: ActivityRecord{2db504f u0 com.miui.home/.launcher.Launcher t2}
	reActivity := regexp.MustCompile(`mResumedActivity: ActivityRecord{.*? (\S+)/(\S+?)\s`)
	matches = reActivity.FindStringSubmatch(output)
	if len(matches) == 3 {
		windowInfo = WindowInfo{
			PackageName: matches[1],
			Activity:    matches[2],
		}
		return windowInfo, nil
	}

	return WindowInfo{}, errors.New("failed to extract current window")
}

func (dev *AndroidDevice) GetPackageInfo(packageName string) (AppInfo, error) {
	appInfo := AppInfo{
		Name: packageName,
	}
	// get package version
	appVersion, err := dev.getPackageVersion(packageName)
	if err == nil {
		appInfo.AppBaseInfo.VersionName = appVersion
	} else {
		log.Warn().Msg("failed to get package version")
		return appInfo, errors.Wrap(code.DeviceAppNotInstalled, err.Error())
	}

	// get package path
	packagePath, err := dev.getPackagePath(packageName)
	if err == nil {
		appInfo.AppBaseInfo.AppPath = packagePath
	} else {
		log.Warn().Msg("failed to get package path")
		return appInfo, errors.Wrap(code.DeviceAppNotInstalled, err.Error())
	}

	// get package md5
	packageMD5, err := dev.getPackageMD5(packagePath)
	if err == nil {
		appInfo.AppBaseInfo.AppMD5 = packageMD5
	} else {
		log.Warn().Msg("failed to get package md5")
		return appInfo, errors.Wrap(code.DeviceAppNotInstalled, err.Error())
	}

	log.Info().Interface("appInfo", appInfo).Msg("get package info")
	return appInfo, nil
}

func (dev *AndroidDevice) getPackageVersion(packageName string) (string, error) {
	output, err := dev.d.RunShellCommand("dumpsys", "package", packageName, "|", "grep", "versionName")
	if err != nil {
		return "", errors.Wrap(err, "get package version failed")
	}
	appVersion := ""
	re := regexp.MustCompile(`versionName=(.+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		appVersion = matches[1]
		return appVersion, nil
	}
	return "", errors.New("failed to get package version")
}

func (dev *AndroidDevice) getPackagePath(packageName string) (string, error) {
	output, err := dev.d.RunShellCommand("pm", "path", packageName)
	if err != nil {
		return "", errors.Wrap(err, "get package path failed")
	}
	re := regexp.MustCompile(`package:(.+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", errors.New("failed to get package path")
}

func (dev *AndroidDevice) getPackageMD5(packagePath string) (string, error) {
	output, err := dev.d.RunShellCommand("md5sum", packagePath)
	if err != nil {
		return "", errors.Wrap(err, "get package md5 failed")
	}
	matches := strings.Split(output, " ")
	if len(matches) > 1 {
		return matches[0], nil
	}
	return "", errors.New("failed to get package md5")
}

func (dev *AndroidDevice) startUIA2Server() error {
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Info().Str("package", UIA2ServerTestPackageName).
			Int("attempt", attempt).Msg("start uiautomator server")
		// $ adb shell am instrument -w $UIA2ServerTestPackageName
		// -w: wait for instrumentation to finish before returning.
		// Required for test runners.
		out, err := dev.d.RunShellCommand("am", "instrument", "-w", UIA2ServerTestPackageName)
		if err != nil {
			return errors.Wrap(err, "start uiautomator server failed")
		}
		if strings.Contains(out, "Process crashed") {
			log.Error().Msg("uiautomator server crashed, retrying...")
		}
	}

	return errors.Wrapf(code.MobileUIDriverAppCrashed,
		"uiautomator server crashed %d times", maxRetries)
}

func (dev *AndroidDevice) stopUIA2Server() error {
	_, err := dev.d.RunShellCommand("am", "force-stop", UIA2ServerPackageName)
	return err
}

type LineCallback func(string)

type AdbLogcat struct {
	serial string
	// logBuffer *bytes.Buffer
	errs     []error
	stopping chan struct{}
	done     chan struct{}
	cmd      *exec.Cmd
	callback LineCallback
	logs     []string
}

func NewAdbLogcatWithCallback(serial string, callback LineCallback) *AdbLogcat {
	return &AdbLogcat{
		serial: serial,
		// logBuffer: new(bytes.Buffer),
		stopping: make(chan struct{}),
		done:     make(chan struct{}),
		callback: callback,
		logs:     make([]string, 0),
	}
}

func NewAdbLogcat(serial string) *AdbLogcat {
	return &AdbLogcat{
		serial: serial,
		// logBuffer: new(bytes.Buffer),
		stopping: make(chan struct{}),
		done:     make(chan struct{}),
		logs:     make([]string, 0),
	}
}

// CatchLogcatContext starts logcat with timeout context
func (l *AdbLogcat) CatchLogcatContext(timeoutCtx context.Context) (err error) {
	if err = l.CatchLogcat(""); err != nil {
		return
	}
	go func() {
		select {
		case <-timeoutCtx.Done():
			_ = l.Stop()
		case <-l.stopping:
		}
	}()
	return
}

func (l *AdbLogcat) Stop() error {
	select {
	case <-l.stopping:
	default:
		close(l.stopping)
		<-l.done
		close(l.done)
	}
	return l.Errors()
}

func (l *AdbLogcat) Errors() (err error) {
	for _, e := range l.errs {
		if err != nil {
			err = fmt.Errorf("%v |[DeviceLogcatErr] %v", err, e)
		} else {
			err = fmt.Errorf("[DeviceLogcatErr] %v", e)
		}
	}
	return
}

func (l *AdbLogcat) CatchLogcat(filter string) (err error) {
	if l.cmd != nil {
		log.Warn().Msg("logcat already start")
		return nil
	}

	// FIXME: replace with gadb shell command
	// clear logcat
	if err = myexec.RunCommand("adb", "-s", l.serial, "shell", "logcat", "-c"); err != nil {
		return
	}
	args := []string{"-s", l.serial, "logcat", "--format", "time"}
	if filter != "" {
		args = append(args, "-s", filter)
	}
	// start logcat
	l.cmd = myexec.Command("adb", args...)
	// l.cmd.Stderr = l.logBuffer
	// l.cmd.Stdout = l.logBuffer
	reader, err := l.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = l.cmd.Start(); err != nil {
		return
	}
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			if l.callback != nil {
				l.callback(line) // Process each line with callback
			} else {
				l.logs = append(l.logs, line) // Store line if no callback
			}
		}
	}()
	go func() {
		<-l.stopping
		if e := reader.Close(); e != nil {
			log.Error().Err(e).Msg("close logcat reader failed")
		}
		if e := myexec.KillProcessesByGpid(l.cmd); e != nil {
			log.Error().Err(e).Msg("kill logcat process failed")
		}
		l.done <- struct{}{}
	}()

	return
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

type UiSelectorHelper struct {
	value *bytes.Buffer
}

func NewUiSelectorHelper() UiSelectorHelper {
	return UiSelectorHelper{value: bytes.NewBufferString("new UiSelector()")}
}

func (s UiSelectorHelper) String() string {
	return s.value.String() + ";"
}

// Text Set the search criteria to match the visible text displayed
// in a widget (for example, the text label to launch an app).
//
// The text for the element must match exactly with the string in your input
// argument. Matching is case-sensitive.
func (s UiSelectorHelper) Text(text string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.text("%s")`, text))
	return s
}

// TextMatches Set the search criteria to match the visible text displayed in a layout
// element, using a regular expression.
//
// The text in the widget must match exactly with the string in your
// input argument.
func (s UiSelectorHelper) TextMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.textMatches("%s")`, regex))
	return s
}

// TextStartsWith Set the search criteria to match visible text in a widget that is
// prefixed by the text parameter.
//
// The matching is case-insensitive.
func (s UiSelectorHelper) TextStartsWith(text string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.textStartsWith("%s")`, text))
	return s
}

// TextContains Set the search criteria to match the visible text in a widget
// where the visible text must contain the string in your input argument.
//
// The matching is case-sensitive.
func (s UiSelectorHelper) TextContains(text string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.textContains("%s")`, text))
	return s
}

// ClassName Set the search criteria to match the class property
// for a widget (for example, "android.widget.Button").
func (s UiSelectorHelper) ClassName(className string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.className("%s")`, className))
	return s
}

// ClassNameMatches Set the search criteria to match the class property
// for a widget, using a regular expression.
func (s UiSelectorHelper) ClassNameMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.classNameMatches("%s")`, regex))
	return s
}

// Description Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must match exactly
// with the string in your input argument.
//
// Matching is case-sensitive.
func (s UiSelectorHelper) Description(desc string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.description("%s")`, desc))
	return s
}

// DescriptionMatches Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must match exactly
// with the string in your input argument.
func (s UiSelectorHelper) DescriptionMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.descriptionMatches("%s")`, regex))
	return s
}

// DescriptionStartsWith Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must start
// with the string in your input argument.
//
// Matching is case-insensitive.
func (s UiSelectorHelper) DescriptionStartsWith(desc string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.descriptionStartsWith("%s")`, desc))
	return s
}

// DescriptionContains Set the search criteria to match the content-description
// property for a widget.
//
// The content-description is typically used
// by the Android Accessibility framework to
// provide an audio prompt for the widget when
// the widget is selected. The content-description
// for the widget must contain
// the string in your input argument.
//
// Matching is case-insensitive.
func (s UiSelectorHelper) DescriptionContains(desc string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.descriptionContains("%s")`, desc))
	return s
}

// ResourceId Set the search criteria to match the given resource ID.
func (s UiSelectorHelper) ResourceId(id string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.resourceId("%s")`, id))
	return s
}

// ResourceIdMatches Set the search criteria to match the resource ID
// of the widget, using a regular expression.
func (s UiSelectorHelper) ResourceIdMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.resourceIdMatches("%s")`, regex))
	return s
}

// Index Set the search criteria to match the widget by its node
// index in the layout hierarchy.
//
// The index value must be 0 or greater.
//
// Using the index can be unreliable and should only
// be used as a last resort for matching. Instead,
// consider using the `Instance(int)` method.
func (s UiSelectorHelper) Index(index int) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.index(%d)`, index))
	return s
}

// Instance Set the search criteria to match the
// widget by its instance number.
//
// The instance value must be 0 or greater, where
// the first instance is 0.
//
// For example, to simulate a user click on
// the third image that is enabled in a UI screen, you
// could specify a search criteria where the instance is
// 2, the `className(String)` matches the image
// widget class, and `enabled(boolean)` is true.
// The code would look like this:
//
//	`new UiSelector().className("android.widget.ImageView")
//	  .enabled(true).instance(2);`
func (s UiSelectorHelper) Instance(instance int) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.instance(%d)`, instance))
	return s
}

// Enabled Set the search criteria to match widgets that are enabled.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Enabled(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.enabled(%t)`, b))
	return s
}

// Focused Set the search criteria to match widgets that have focus.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Focused(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.focused(%t)`, b))
	return s
}

// Focusable Set the search criteria to match widgets that are focusable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Focusable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.focusable(%t)`, b))
	return s
}

// Scrollable Set the search criteria to match widgets that are scrollable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Scrollable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.scrollable(%t)`, b))
	return s
}

// Selected Set the search criteria to match widgets that
// are currently selected.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Selected(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.selected(%t)`, b))
	return s
}

// Checked Set the search criteria to match widgets that
// are currently checked (usually for checkboxes).
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Checked(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.checked(%t)`, b))
	return s
}

// Checkable Set the search criteria to match widgets that are checkable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Checkable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.checkable(%t)`, b))
	return s
}

// Clickable Set the search criteria to match widgets that are clickable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) Clickable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.clickable(%t)`, b))
	return s
}

// LongClickable Set the search criteria to match widgets that are long-clickable.
//
// Typically, using this search criteria alone is not useful.
// You should also include additional criteria, such as text,
// content-description, or the class name for a widget.
//
// If no other search criteria is specified, and there is more
// than one matching widget, the first widget in the tree
// is selected.
func (s UiSelectorHelper) LongClickable(b bool) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.longClickable(%t)`, b))
	return s
}

// packageName Set the search criteria to match the package name
// of the application that contains the widget.
func (s UiSelectorHelper) packageName(name string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.packageName(%s)`, name))
	return s
}

// PackageNameMatches Set the search criteria to match the package name
// of the application that contains the widget.
func (s UiSelectorHelper) PackageNameMatches(regex string) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.packageNameMatches(%s)`, regex))
	return s
}

// ChildSelector Adds a child UiSelector criteria to this selector.
//
// Use this selector to narrow the search scope to
// child widgets under a specific parent widget.
func (s UiSelectorHelper) ChildSelector(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.childSelector(%s)`, selector.value.String()))
	return s
}

func (s UiSelectorHelper) PatternSelector(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.patternSelector(%s)`, selector.value.String()))
	return s
}

func (s UiSelectorHelper) ContainerSelector(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.containerSelector(%s)`, selector.value.String()))
	return s
}

// FromParent Adds a child UiSelector criteria to this selector which is used to
// start search from the parent widget.
//
// Use this selector to narrow the search scope to
// sibling widgets as well all child widgets under a parent.
func (s UiSelectorHelper) FromParent(selector UiSelectorHelper) UiSelectorHelper {
	s.value.WriteString(fmt.Sprintf(`.fromParent(%s)`, selector.value.String()))
	return s
}
