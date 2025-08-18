package gadb

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
)

type DeviceFileInfo struct {
	Name         string
	Mode         os.FileMode
	Size         uint32
	LastModified time.Time
}

func (info DeviceFileInfo) IsDir() bool {
	return (info.Mode & (1 << 14)) == (1 << 14)
}

const DefaultFileMode = os.FileMode(0o664)

type DeviceState string

const (
	StateUnknown      DeviceState = "UNKNOWN"
	StateOnline       DeviceState = "online"
	StateOffline      DeviceState = "offline"
	StateDisconnected DeviceState = "disconnected"
	StateBootloader   DeviceState = "bootloader"
	StateRecovery     DeviceState = "recovery"
	StateUnauthorized DeviceState = "unauthorized"
)

var deviceStateStrings = map[string]DeviceState{
	"":             StateDisconnected, // no devices/emulators found
	"offline":      StateOffline,
	"bootloader":   StateBootloader,
	"recovery":     StateRecovery,
	"unauthorized": StateUnauthorized,
	"device":       StateOnline,
}

func deviceStateConv(k string) (deviceState DeviceState) {
	var ok bool
	if deviceState, ok = deviceStateStrings[k]; !ok {
		return StateUnknown
	}
	return
}

type DeviceForward struct {
	Serial  string
	Local   string
	Remote  string
	Reverse bool
	// LocalProtocol string
	// RemoteProtocol string
}

type Device struct {
	adbClient Client
	serial    string
	attrs     map[string]string
	feat      Features
}

func (d *Device) HasFeature(name Feature) bool {
	feats, err := d.GetFeatures()
	if err != nil || len(feats) == 0 {
		return false
	}
	return feats.HasFeature(name)
}

func (d *Device) GetFeatures() (features Features, err error) {
	if len(d.feat) > 0 {
		return d.feat, nil
	}
	return d.features()
}

func (d *Device) features() (features Features, err error) {
	res, err := d.executeCommand("host:features")
	if err != nil {
		return nil, err
	}
	if len(res) > 4 {
		// stip hash
		res = res[4:]
	}
	fs := strings.Split(string(res), ",")
	features = make(Features, len(fs))
	for _, f := range fs {
		features[Feature(f)] = struct{}{}
	}
	d.feat = features
	return features, nil
}

func (d *Device) HasAttribute(key string) bool {
	_, ok := d.attrs[key]
	return ok
}

func (d *Device) Product() (string, error) {
	if d.HasAttribute("product") {
		return d.attrs["product"], nil
	}
	return "", errors.New("does not have attribute: product")
}

func (d *Device) Model() (string, error) {
	if d.HasAttribute("model") {
		return d.attrs["model"], nil
	}
	return "", errors.New("does not have attribute: model")
}

func (d *Device) Brand() (string, error) {
	if d.HasAttribute("brand") {
		return d.attrs["brand"], nil
	}
	brand, err := d.RunShellCommand("getprop", "ro.product.brand")
	brand = strings.TrimSpace(brand)
	if err != nil {
		return "", errors.New("does not have attribute: brand")
	}
	d.attrs["brand"] = brand
	return brand, nil
}

func (d *Device) Usb() (string, error) {
	if d.HasAttribute("usb") {
		return d.attrs["usb"], nil
	}
	return "", errors.New("does not have attribute: usb")
}

func (d *Device) SystemVersion() (string, error) {
	if d.HasAttribute("systemVersion") {
		return d.attrs["systemVersion"], nil
	}
	systemVersion, err := d.RunShellCommand("getprop", "ro.build.version.release")
	systemVersion = strings.TrimSpace(systemVersion)
	if err != nil {
		return "", errors.New("get android system version failed")
	}
	d.attrs["systemVersion"] = systemVersion
	return systemVersion, nil
}

func (d *Device) SdkVersion() (string, error) {
	if d.HasAttribute("sdkVersion") {
		return d.attrs["sdkVersion"], nil
	}
	sdkVersion, err := d.RunShellCommand("getprop", "ro.build.version.sdk")
	if err != nil {
		return "", errors.New("get android sdk version failed")
	}
	sdkVersion = strings.TrimSpace(sdkVersion)
	d.attrs["sdkVersion"] = sdkVersion
	return sdkVersion, nil
}

func (d *Device) transportId() (string, error) {
	if d.HasAttribute("transport_id") {
		return d.attrs["transport_id"], nil
	}
	return "", errors.New("does not have attribute: transport_id")
}

func (d *Device) DeviceInfo() map[string]string {
	return d.attrs
}

func (d *Device) Serial() string {
	// 	resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:get-serialno", d.serial))
	return d.serial
}

func (d *Device) IsUsb() (bool, error) {
	usb, err := d.Usb()
	if err != nil {
		return false, err
	}

	return usb != "", nil
}

func (d *Device) State() (DeviceState, error) {
	resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:get-state", d.serial))
	return deviceStateConv(resp), err
}

func (d *Device) DevicePath() (string, error) {
	resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:get-devpath", d.serial))
	return resp, err
}

func (d *Device) Forward(remoteInterface interface{}, noRebind ...bool) (port int, err error) {
	var remote string
	switch r := remoteInterface.(type) {
	// for unix sockets
	case string:
		remote = r
	case int:
		remote = fmt.Sprintf("tcp:%d", r)
	}

	forwardList, err := d.ForwardList()
	if err != nil {
		return
	}
	for _, forwardItem := range forwardList {
		if forwardItem.Remote == remote {
			return strconv.Atoi(forwardItem.Local[4:])
		}
	}
	localPort, err := builtin.GetFreePort()
	if err != nil {
		return
	}
	local := fmt.Sprintf("tcp:%d", localPort)

	command := fmt.Sprintf("host-serial:%s:forward:%s;%s", d.serial, local, remote)
	if len(noRebind) != 0 && noRebind[0] {
		command = fmt.Sprintf("host-serial:%s:forward:norebind:%s;%s", d.serial, local, remote)
	}

	_, err = d.adbClient.executeCommand(command, true)
	return localPort, nil
}

func (d *Device) ForwardList() (deviceForwardList []DeviceForward, err error) {
	var forwardList []DeviceForward
	if forwardList, err = d.adbClient.ForwardList(); err != nil {
		return nil, err
	}

	deviceForwardList = make([]DeviceForward, 0, len(deviceForwardList))
	for i := range forwardList {
		if forwardList[i].Serial == d.serial {
			deviceForwardList = append(deviceForwardList, forwardList[i])
		}
	}
	// resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:list-forward", d.serial))
	return
}

func (d *Device) ForwardKill(localPort int) (err error) {
	local := fmt.Sprintf("tcp:%d", localPort)
	_, err = d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:killforward:%s", d.serial, local), true)
	return
}

func (d *Device) ReverseForward(localPort int, remoteInterface interface{}, noRebind ...bool) (err error) {
	var command string
	var remote string
	local := fmt.Sprintf("tcp:%d", localPort)
	switch r := remoteInterface.(type) {
	// for unix sockets
	case string:
		remote = r
	case int:
		remote = fmt.Sprintf("tcp:%d", r)
	}

	if len(noRebind) != 0 && noRebind[0] {
		command = fmt.Sprintf("reverse:forward:norebind:%s;%s", remote, local)
	} else {
		command = fmt.Sprintf("reverse:forward:%s;%s", remote, local)
	}
	_, err = d.executeCommand(command, true)
	return
}

func (d *Device) ReverseForwardList() (deviceForwardList []DeviceForward, err error) {
	res, err := d.executeCommand("reverse:list-forward")
	if err != nil {
		return nil, err
	}
	resStr := string(res)
	lines := strings.Split(resStr, "\n")
	for _, line := range lines {
		groups := strings.Split(line, " ")
		if len(groups) == 3 {
			deviceForwardList = append(deviceForwardList, DeviceForward{
				Reverse: true,
				Serial:  d.serial,
				Remote:  groups[1],
				Local:   groups[2],
			})
		}
	}
	return
}

func (d *Device) ReverseForwardKill(remoteInterface interface{}) error {
	remote := ""
	switch r := remoteInterface.(type) {
	case string:
		remote = r
	case int:
		remote = fmt.Sprintf("tcp:%d", r)
	}
	_, err := d.executeCommand(fmt.Sprintf("reverse:killforward:%s", remote), true)
	return err
}

func (d *Device) RunStubCommand(command []byte, processName string) (res string, err error) {
	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return "", err
	}
	defer func() { _ = tp.Close() }()

	if err = tp.SendWithCheck(fmt.Sprintf("localabstract:%s", processName)); err != nil {
		return "", err
	}

	if err = tp.SendBytes(command); err != nil {
		return "", err
	}

	lenBuf, err := tp.ReadBytesN(4)
	if err != nil {
		return "", err
	}
	length := binary.LittleEndian.Uint32(lenBuf)
	result, err := tp.ReadBytesN(int(length) - 4)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (d *Device) ReverseForwardKillAll() error {
	_, err := d.executeCommand("reverse:killforward-all")
	return err
}

func (d *Device) RunShellCommand(cmd string, args ...string) (string, error) {
	raw, err := d.RunShellCommandWithBytes(cmd, args...)
	if err != nil && errors.Cause(err) == nil {
		err = errors.Wrap(code.DeviceShellExecError, err.Error())
	}
	return string(raw), err
}

func (d *Device) RunShellCommandWithBytes(cmd string, args ...string) ([]byte, error) {
	if d.HasFeature(FeatShellV2) {
		return d.RunShellCommandV2WithBytes(cmd, args...)
	}
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	if strings.TrimSpace(cmd) == "" {
		return nil, errors.New("adb shell: command cannot be empty")
	}

	startTime := time.Now()
	defer func() {
		// log elapsed seconds for shell execution
		log.Debug().Str("cmd",
			fmt.Sprintf("adb -s %s shell %s", d.serial, cmd)).
			Int64("elapsed(ms)", time.Since(startTime).Milliseconds()).
			Msg("run adb shell")
	}()

	raw, err := d.executeCommand(fmt.Sprintf("shell:%s", cmd))
	return raw, err
}

// RunShellCommandV2WithBytes shell v2, 支持后台运行而不会阻断
func (d *Device) RunShellCommandV2WithBytes(cmd string, args ...string) ([]byte, error) {
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	if strings.TrimSpace(cmd) == "" {
		return nil, errors.New("adb shell: command cannot be empty")
	}

	startTime := time.Now()
	defer func() {
		// log elapsed seconds for shell execution
		log.Debug().Str("cmd",
			fmt.Sprintf("adb -s %s shell %s", d.serial, cmd)).
			Int64("elapsed(ms)", time.Since(startTime).Milliseconds()).
			Msg("run adb shell in v2")
	}()

	raw, err := d.executeCommand(fmt.Sprintf("shell,v2,raw:%s", cmd))
	if err != nil {
		return raw, err
	}
	return d.parseV2CommandWithBytes(raw)
}

func (d *Device) parseV2CommandWithBytes(input []byte) (res []byte, err error) {
	if len(input) == 0 {
		return input, nil
	}
	reader := bytes.NewReader(input)
	sizeBuf := make([]byte, 4)
	var (
		resBuf   []byte
		exitCode int
	)
loop:
	for {
		msgCode, err := reader.ReadByte()
		if err != nil {
			return input, err
		}
		switch msgCode {
		case 0x01, 0x02: // STDOUT, STDERR
			_, err = io.ReadFull(reader, sizeBuf)
			if err != nil {
				return input, err
			}
			size := binary.LittleEndian.Uint32(sizeBuf)
			if cap(resBuf) < int(size) {
				resBuf = make([]byte, int(size))
			}
			_, err = io.ReadFull(reader, resBuf[:size])
			if err != nil {
				return input, err
			}
			res = append(res, resBuf[:size]...)
		case 0x03: // EXIT
			_, err = io.ReadFull(reader, sizeBuf)
			if err != nil {
				return input, err
			}
			size := binary.LittleEndian.Uint32(sizeBuf)
			if cap(resBuf) < int(size) {
				resBuf = make([]byte, int(size))
			}
			ec, err := reader.ReadByte()
			if err != nil {
				return input, err
			}
			exitCode = int(ec)
			break loop
		default:
			return input, nil
		}
	}
	if exitCode != 0 {
		return nil, errors.New(string(res))
	}
	return res, nil
}

func (d *Device) EnableAdbOverTCP(port ...int) (err error) {
	if len(port) == 0 {
		port = []int{AdbDaemonPort}
	}

	_, err = d.executeCommand(fmt.Sprintf("tcpip:%d", port[0]), true)
	return
}

func (d *Device) createDeviceTransport(readTimeout ...time.Duration) (tp transport, err error) {
	if tp, err = newTransport(fmt.Sprintf("%s:%d", d.adbClient.host, d.adbClient.port), readTimeout...); err != nil {
		return transport{}, err
	}

	err = tp.SendWithCheck(fmt.Sprintf("host:transport:%s", d.serial))
	return
}

func (d *Device) executeCommand(command string, onlyVerifyResponse ...bool) (raw []byte, err error) {
	if len(onlyVerifyResponse) == 0 {
		onlyVerifyResponse = []bool{false}
	}

	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return nil, err
	}
	defer func() { _ = tp.Close() }()

	if err = tp.SendWithCheck(command); err != nil {
		return nil, err
	}

	if onlyVerifyResponse[0] {
		return
	}

	raw, err = tp.ReadBytesAll()
	return
}

func (d *Device) List(remotePath string) (devFileInfos []DeviceFileInfo, err error) {
	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return nil, err
	}
	defer func() { _ = tp.Close() }()

	var sync syncTransport
	if sync, err = tp.CreateSyncTransport(); err != nil {
		return nil, err
	}
	defer func() { _ = sync.Close() }()

	if err = sync.Send("LIST", remotePath); err != nil {
		return nil, err
	}

	devFileInfos = make([]DeviceFileInfo, 0)

	var entry DeviceFileInfo
	for entry, err = sync.ReadDirectoryEntry(); err == nil; entry, err = sync.ReadDirectoryEntry() {
		if entry == (DeviceFileInfo{}) {
			break
		}
		devFileInfos = append(devFileInfos, entry)
	}

	return
}

func (d *Device) PushFile(localPath, remotePath string, modification ...time.Time) (err error) {
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	if len(modification) == 0 {
		var stat os.FileInfo
		if stat, err = localFile.Stat(); err != nil {
			return err
		}
		modification = []time.Time{stat.ModTime()}
	}

	return d.Push(localFile, remotePath, modification[0], DefaultFileMode)
}

func (d *Device) Push(source io.Reader, remotePath string, modification time.Time, mode ...os.FileMode) (err error) {
	if len(mode) == 0 {
		mode = []os.FileMode{DefaultFileMode}
	}

	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return err
	}
	defer func() { _ = tp.Close() }()

	var sync syncTransport
	if sync, err = tp.CreateSyncTransport(); err != nil {
		return err
	}
	defer func() { _ = sync.Close() }()

	data := fmt.Sprintf("%s,%d", remotePath, mode[0])
	if err = sync.Send("SEND", data); err != nil {
		return err
	}

	if err = sync.SendStream(source); err != nil {
		return
	}

	if err = sync.SendStatus("DONE", uint32(modification.Unix())); err != nil {
		return
	}

	if err = sync.VerifyStatus(); err != nil {
		return
	}
	return
}

func (d *Device) Pull(remotePath string, dest io.Writer) (err error) {
	var tp transport
	if tp, err = d.createDeviceTransport(); err != nil {
		return err
	}
	defer func() { _ = tp.Close() }()

	var sync syncTransport
	if sync, err = tp.CreateSyncTransport(); err != nil {
		return err
	}
	defer func() { _ = sync.Close() }()

	if err = sync.Send("RECV", remotePath); err != nil {
		return err
	}

	err = sync.WriteStream(dest)
	return
}

func (d *Device) PullFolder(remotePath string, localPath string) (err error) {
	// Check if remote path exists and is a directory
	fileInfos, err := d.List(remotePath)
	if err != nil {
		return fmt.Errorf("failed to list remote directory: %w", err)
	}

	// Create local directory if it doesn't exist
	if err = os.MkdirAll(localPath, 0o755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Pull each file/directory recursively
	for _, fileInfo := range fileInfos {
		remoteItemPath := remotePath + "/" + fileInfo.Name
		localItemPath := localPath + "/" + fileInfo.Name

		if fileInfo.IsDir() {
			// Recursively pull subdirectory
			if err = d.PullFolder(remoteItemPath, localItemPath); err != nil {
				return fmt.Errorf("failed to pull subdirectory %s: %w", remoteItemPath, err)
			}
		} else {
			// Pull file
			if err = d.PullFile(remoteItemPath, localItemPath); err != nil {
				return fmt.Errorf("failed to pull file %s: %w", remoteItemPath, err)
			}
		}
	}

	return nil
}

func (d *Device) PullFile(remotePath string, localPath string) (err error) {
	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Use existing Pull method to pull file content
	if err = d.Pull(remotePath, localFile); err != nil {
		return fmt.Errorf("failed to pull file content: %w", err)
	}

	return nil
}

func (d *Device) installViaABBExec(apk io.ReadSeeker, args ...string) (raw []byte, err error) {
	var (
		tp       transport
		filesize int64
	)
	timeout := 8
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Minute)
	defer cancel()

	filesize, err = apk.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if tp, err = d.createDeviceTransport(4 * time.Minute); err != nil {
		return nil, err
	}
	defer func() { _ = tp.Close() }()
	go func() {
		<-ctx.Done()
		_ = tp.Close()
	}()
	cmd := "abb_exec:package\x00install\x00-t"
	for _, arg := range args {
		cmd += "\x00" + arg
	}
	cmd += fmt.Sprintf("\x00-S\x00%d", filesize)
	if err = tp.SendWithCheck(cmd); err != nil {
		return nil, err
	}

	_, err = apk.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(tp.Conn(), apk)
	if err != nil {
		return nil, err
	}
	raw, err = tp.ReadBytesAll()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, fmt.Errorf("installation timed out after %d minutes", timeout)
	}
	return
}

func (d *Device) InstallAPK(apkPath string, args ...string) (string, error) {
	apkFile, err := os.Open(apkPath)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("open apk file %s failed", apkPath))
	}
	defer apkFile.Close()

	haserr := func(ret string) bool {
		return strings.Contains(ret, "Failure")
	}
	// 该方法掉线不会返回error。导致误认为安装成功
	if d.HasFeature(FeatAbbExec) {
		raw, err := d.installViaABBExec(apkFile)
		if err != nil {
			return "", fmt.Errorf("error installing: %v", err)
		}
		if haserr(string(raw)) {
			return "", errors.New(string(raw))
		}
		return string(raw), err
	}

	remote := fmt.Sprintf("/data/local/tmp/%s.apk", builtin.GenNameWithTimestamp("gadb_remote_%d"))
	err = d.Push(apkFile, remote, time.Now())
	if err != nil {
		return "", fmt.Errorf("error pushing: %v", err)
	}
	args = append([]string{"install"}, args...)
	args = append(args, "-f", remote)
	res, err := d.RunShellCommand("pm", args...)
	if err != nil {
		return "", errors.Wrap(err, "install apk failed")
	}
	if haserr(res) {
		return "", errors.New(res)
	}

	return res, nil
}

func (d *Device) Uninstall(packageName string, keepData ...bool) (string, error) {
	if len(keepData) == 0 {
		keepData = []bool{false}
	}
	packageName = strings.TrimSpace(packageName)
	if len(packageName) == 0 {
		return "", fmt.Errorf("invalid package name")
	}
	args := []string{"uninstall"}
	if keepData[0] {
		args = append(args, "-k")
	}
	args = append(args, packageName)
	return d.RunShellCommand("pm", args...)
}

func (d *Device) ListPackages() ([]string, error) {
	args := []string{"list", "packages"}
	resRaw, err := d.RunShellCommand("pm", args...)
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(resRaw, "\n")
	var packages []string
	for _, line := range lines {
		packageName := strings.TrimPrefix(line, "package:")
		packages = append(packages, packageName)
	}
	return packages, nil
}

func (d *Device) IsPackageInstalled(packageName string) bool {
	packages, err := d.ListPackages()
	if err != nil {
		return false
	}
	packageName = strings.TrimSpace(packageName)
	if len(packageName) == 0 {
		return false
	}
	return builtin.Contains(packages, packageName)
}

func (d *Device) IsPackageRunning(packageName string) bool {
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		log.Error().Msg("package name is empty, skip checking package running")
		return false
	}

	// Use ps -ef command with grep to check if package is running
	// ps -ef shows full command line which includes package name as argument
	// This works for both regular apps and instrumentation test processes
	output, err := d.RunShellCommand("ps -ef | grep " + packageName + " | grep -v grep")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

func (d *Device) ScreenCap() ([]byte, error) {
	if d.HasFeature(FeatShellV2) {
		return d.RunShellCommandV2WithBytes("screencap", "-p")
	}

	// for shell v1, screenshot buffer maybe truncated
	// thus we firstly save it to local file and then pull it
	tempPath := fmt.Sprintf("/data/local/tmp/screenshot_%d.png",
		time.Now().Unix())
	_, err := d.RunShellCommandWithBytes("screencap", "-p", tempPath)
	if err != nil {
		return nil, errors.Wrap(err, "screencap failed")
	}

	// remove temp file
	defer func() {
		go d.RunShellCommand("rm", tempPath)
	}()

	buffer := bytes.NewBuffer(nil)
	err = d.Pull(tempPath, buffer)
	if err != nil {
		return nil, errors.Wrap(err, "pull video failed")
	}
	return buffer.Bytes(), nil
}

func (d *Device) ScreenRecord(ctx context.Context) ([]byte, error) {
	videoPath := fmt.Sprintf("/sdcard/screenrecord_%d.mp4", time.Now().Unix())

	done := make(chan error, 1)
	go func() {
		_, err := d.RunShellCommandWithBytes("screenrecord", videoPath)
		done <- err
	}()

	select {
	case <-ctx.Done():
		// timeout or cancelled
		pid, err := d.RunShellCommand("pidof", "screenrecord")
		if err == nil && pid != "" {
			// 发送 SIGINT 信号终止录屏
			_, _ = d.RunShellCommand("kill", "-2", strings.TrimSpace(pid))
		}
		<-done // 等待进程完全退出
	case err := <-done:
		// adb screenrecord will exit on reached 180s
		if err != nil {
			return nil, errors.Wrap(err, "screenrecord failed")
		}
	}

	// remove temp file
	defer func() {
		go d.RunShellCommand("rm", videoPath)
	}()

	buffer := bytes.NewBuffer(nil)
	err := d.Pull(videoPath, buffer)
	if err != nil {
		return nil, errors.Wrap(err, "pull video failed")
	}
	return buffer.Bytes(), nil
}
