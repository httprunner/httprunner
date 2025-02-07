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

const (
	EvalInstallerPackageName   = "sogou.mobile.explorer"
	InstallViaInstallerCommand = "am start -S -n sogou.mobile.explorer/.PackageInstallerActivity -d"
)

//go:embed evalite
var evalite embed.FS

func NewAndroidDevice(opts ...option.AndroidDeviceOption) (device *AndroidDevice, err error) {
	androidOptions := option.NewAndroidDeviceOptions(opts...)

	// get all attached android devices
	adbClient, err := gadb.NewClientWith(
		androidOptions.AdbServerHost, androidOptions.AdbServerPort)
	if err != nil {
		return nil, err
	}
	devices, err := adbClient.DeviceList()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, errors.Wrapf(code.DeviceConnectionError,
			"no attached android devices")
	}

	// filter device by serial
	var gadbDevice *gadb.Device
	if androidOptions.SerialNumber == "" {
		if len(devices) > 1 {
			return nil, errors.Wrap(code.DeviceConnectionError,
				"more than one device connected, please specify the serial")
		}
		gadbDevice = devices[0]
		androidOptions.SerialNumber = gadbDevice.Serial()
		log.Warn().Str("serial", androidOptions.SerialNumber).
			Msg("android SerialNumber is not specified, select the attached one")
	} else {
		for _, d := range devices {
			if d.Serial() == androidOptions.SerialNumber {
				gadbDevice = d
				break
			}
		}
		if gadbDevice == nil {
			return nil, errors.Wrapf(code.DeviceConnectionError,
				"android device %s not attached", androidOptions.SerialNumber)
		}
	}

	device = &AndroidDevice{
		Device:               gadbDevice,
		AndroidDeviceOptions: androidOptions,
		Logcat:               NewAdbLogcat(androidOptions.SerialNumber),
	}
	log.Info().Str("serial", device.SerialNumber).Msg("init android device")
	return device, nil
}

type AndroidDevice struct {
	*gadb.Device
	*option.AndroidDeviceOptions
	Logcat *AdbLogcat
}

func (dev *AndroidDevice) Setup() error {
	dev.RunShellCommand("ime", "enable", UnicodeImePackageName)
	dev.RunShellCommand("rm", "-r", config.DeviceActionLogFilePath)

	// setup evalite
	evalToolRaw, err := evalite.ReadFile("evalite")
	if err != nil {
		return errors.Wrap(code.LoadFileError, err.Error())
	}
	err = dev.Push(bytes.NewReader(evalToolRaw), "/data/local/tmp/evalite", time.Now())
	if err != nil {
		return errors.Wrap(code.DeviceShellExecError, err.Error())
	}

	if dev.UIA2 {
		// uiautomator2 server must be started before

		// check uiautomator server package installed
		if !dev.IsPackageInstalled(dev.UIA2ServerPackageName) {
			return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
				"%s not installed", dev.UIA2ServerPackageName)
		}
		if !dev.IsPackageInstalled(dev.UIA2ServerTestPackageName) {
			return errors.Wrapf(code.MobileUIDriverAppNotInstalled,
				"%s not installed", dev.UIA2ServerTestPackageName)
		}

		// TODO: check uiautomator server package running
		// if dev.IsPackageRunning(UIA2ServerPackageName) {
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

func (dev *AndroidDevice) Teardown() error {
	return nil
}

func (dev *AndroidDevice) UUID() string {
	return dev.SerialNumber
}

func (dev *AndroidDevice) LogEnabled() bool {
	return dev.LogOn
}

func (dev *AndroidDevice) NewDriver(opts ...option.DriverOption) (driverExt *DriverExt, err error) {
	var driver IDriver
	if dev.UIA2 || dev.LogOn {
		driver, err = NewUIA2Driver(dev)
	} else if dev.STUB {
		driver, err = NewStubDriver(dev)
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

func (dev *AndroidDevice) Install(apkPath string, opts ...option.InstallOption) error {
	installOpts := option.NewInstallOptions(opts...)
	brand, err := dev.Brand()
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
		if dev.IsPackageInstalled(EvalInstallerPackageName) {
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
	_, err := dev.InstallAPK(apkPath, args...)
	return err
}

func (dev *AndroidDevice) installViaInstaller(apkPath string, args ...string) error {
	appRemotePath := "/data/local/tmp/" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".apk"
	err := dev.PushFile(apkPath, appRemotePath, time.Now())
	if err != nil {
		return err
	}
	done := make(chan error)
	defer func() {
		close(done)
	}()
	logcat := NewAdbLogcatWithCallback(dev.Serial(), func(line string) {
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
	_, err = dev.RunShellCommand("am", args[1:]...)
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
	_, err := dev.InstallAPK(apkPath, args...)
	return err
}

func (dev *AndroidDevice) Uninstall(packageName string) error {
	_, err := dev.RunShellCommand("uninstall", packageName)
	return err
}

func (dev *AndroidDevice) GetCurrentWindow() (windowInfo WindowInfo, err error) {
	// adb shell dumpsys window | grep -E 'mCurrentFocus|mFocusedApp'
	output, err := dev.RunShellCommand("dumpsys", "window", "|", "grep", "-E", "'mCurrentFocus|mFocusedApp'")
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
	output, err = dev.RunShellCommand("dumpsys", "activity", "activities", "|", "grep", "mResumedActivity")
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
	output, err := dev.RunShellCommand("dumpsys", "package", packageName, "|", "grep", "versionName")
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
	output, err := dev.RunShellCommand("pm", "path", packageName)
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
	output, err := dev.RunShellCommand("md5sum", packagePath)
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
		log.Info().Str("package", dev.UIA2ServerTestPackageName).
			Int("attempt", attempt).Msg("start uiautomator server")
		// $ adb shell am instrument -w $UIA2ServerTestPackageName
		// -w: wait for instrumentation to finish before returning.
		// Required for test runners.
		out, err := dev.RunShellCommand("am", "instrument", "-w", dev.UIA2ServerTestPackageName)
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
	_, err := dev.RunShellCommand("am", "force-stop", dev.UIA2ServerPackageName)
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
