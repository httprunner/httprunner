package gadb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
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

func (d *Device) Product() string {
	return d.attrs["product"]
}

func (d *Device) Model() string {
	return d.attrs["model"]
}

func (d *Device) Usb() string {
	return d.attrs["usb"]
}

func (d *Device) transportId() string {
	return d.attrs["transport_id"]
}

func (d *Device) DeviceInfo() map[string]string {
	return d.attrs
}

func (d *Device) Serial() string {
	// 	resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:get-serialno", d.serial))
	return d.serial
}

func (d *Device) IsUsb() bool {
	return d.Usb() != ""
}

func (d *Device) State() (DeviceState, error) {
	resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:get-state", d.serial))
	return deviceStateConv(resp), err
}

func (d *Device) DevicePath() (string, error) {
	resp, err := d.adbClient.executeCommand(fmt.Sprintf("host-serial:%s:get-devpath", d.serial))
	return resp, err
}

func (d *Device) Forward(localPort int, remoteInterface interface{}, noRebind ...bool) (err error) {
	command := ""
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
		command = fmt.Sprintf("host-serial:%s:forward:norebind:%s;%s", d.serial, local, remote)
	} else {
		command = fmt.Sprintf("host-serial:%s:forward:%s;%s", d.serial, local, remote)
	}

	_, err = d.adbClient.executeCommand(command, true)
	return
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

func (d *Device) ReverseForwardKillAll() error {
	_, err := d.executeCommand("reverse:killforward-all")
	return err
}

func (d *Device) RunShellCommand(cmd string, args ...string) (string, error) {
	raw, err := d.RunShellCommandWithBytes(cmd, args...)
	if err != nil && errors.Cause(err) == nil {
		err = errors.Wrap(code.AndroidShellExecError, err.Error())
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

func (d *Device) createDeviceTransport() (tp transport, err error) {
	if tp, err = newTransport(fmt.Sprintf("%s:%d", d.adbClient.host, d.adbClient.port)); err != nil {
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

func (d *Device) PushFile(local *os.File, remotePath string, modification ...time.Time) (err error) {
	if len(modification) == 0 {
		var stat os.FileInfo
		if stat, err = local.Stat(); err != nil {
			return err
		}
		modification = []time.Time{stat.ModTime()}
	}

	return d.Push(local, remotePath, modification[0], DefaultFileMode)
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

func (d *Device) installViaABBExec(apk io.ReadSeeker) (raw []byte, err error) {
	var (
		tp       transport
		filesize int64
	)
	filesize, err = apk.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if tp, err = d.createDeviceTransport(); err != nil {
		return nil, err
	}
	defer func() { _ = tp.Close() }()

	cmd := fmt.Sprintf("abb_exec:package\x00install\x00-t\x00-S\x00%d", filesize)
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
	return
}

func (d *Device) InstallAPK(apk io.ReadSeeker) (string, error) {
	haserr := func(ret string) bool {
		return strings.Contains(ret, "Failure")
	}
	if d.HasFeature(FeatAbbExec) {
		raw, err := d.installViaABBExec(apk)
		if err != nil {
			return "", fmt.Errorf("error installing: %v", err)
		}
		if haserr(string(raw)) {
			return "", errors.New(string(raw))
		}
		return string(raw), err
	}

	remote := fmt.Sprintf("/data/local/tmp/%s.apk", builtin.GenNameWithTimestamp("gadb_remote_%d"))
	err := d.Push(apk, remote, time.Now())
	if err != nil {
		return "", fmt.Errorf("error pushing: %v", err)
	}

	res, err := d.RunShellCommand("pm", "install", "-f", remote)
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
	packageName = strings.ReplaceAll(packageName, " ", "")
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
		return nil, err
	}

	buffer := bytes.NewBuffer(nil)
	err = d.Pull(tempPath, buffer)
	return buffer.Bytes(), err
}
