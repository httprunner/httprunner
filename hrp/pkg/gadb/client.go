package gadb

import (
	"fmt"
	"strconv"
	"strings"
)

const AdbServerPort = 5037
const AdbDaemonPort = 5555

type Client struct {
	host string
	port int
}

func NewClient() (Client, error) {
	return NewClientWith("localhost")
}

func NewClientWith(host string, port ...int) (adbClient Client, err error) {
	if len(port) == 0 {
		port = []int{AdbServerPort}
	}
	adbClient.host = host
	adbClient.port = port[0]

	var tp transport
	if tp, err = adbClient.createTransport(); err != nil {
		return Client{}, err
	}
	defer func() { _ = tp.Close() }()

	return
}

func (c Client) ServerVersion() (version int, err error) {
	var resp string
	if resp, err = c.executeCommand("host:version"); err != nil {
		return 0, err
	}

	var v int64
	if v, err = strconv.ParseInt(resp, 16, 64); err != nil {
		return 0, err
	}

	version = int(v)
	return
}

func (c Client) DeviceSerialList() (serials []string, err error) {
	var resp string
	if resp, err = c.executeCommand("host:devices"); err != nil {
		return
	}

	lines := strings.Split(resp, "\n")
	serials = make([]string, 0, len(lines))

	for i := range lines {
		fields := strings.Fields(lines[i])
		if len(fields) < 2 {
			continue
		}
		serials = append(serials, fields[0])
	}

	return
}

func (c Client) DeviceList() (devices []Device, err error) {
	var resp string
	if resp, err = c.executeCommand("host:devices-l"); err != nil {
		return
	}

	lines := strings.Split(resp, "\n")
	devices = make([]Device, 0, len(lines))

	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 || len(fields[0]) == 0 {
			debugLog(fmt.Sprintf("can't parse: %s", line))
			continue
		}

		sliceAttrs := fields[2:]
		mapAttrs := map[string]string{}
		for _, field := range sliceAttrs {
			split := strings.Split(field, ":")
			key, val := split[0], split[1]
			mapAttrs[key] = val
		}
		devices = append(devices, Device{adbClient: c, serial: fields[0], attrs: mapAttrs})
	}

	return
}

func (c Client) ForwardList() (deviceForward []DeviceForward, err error) {
	var resp string
	if resp, err = c.executeCommand("host:list-forward"); err != nil {
		return nil, err
	}

	lines := strings.Split(resp, "\n")
	deviceForward = make([]DeviceForward, 0, len(lines))

	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		deviceForward = append(deviceForward, DeviceForward{Serial: fields[0], Local: fields[1], Remote: fields[2]})
	}

	return
}

func (c Client) ForwardKillAll() (err error) {
	_, err = c.executeCommand("host:killforward-all", true)
	return
}

func (c Client) Connect(ip string, port ...int) (err error) {
	if len(port) == 0 {
		port = []int{AdbDaemonPort}
	}

	var resp string
	if resp, err = c.executeCommand(fmt.Sprintf("host:connect:%s:%d", ip, port[0])); err != nil {
		return err
	}
	if !strings.HasPrefix(resp, "connected to") && !strings.HasPrefix(resp, "already connected to") {
		return fmt.Errorf("adb connect: %s", resp)
	}
	return
}

func (c Client) Disconnect(ip string, port ...int) (err error) {
	cmd := fmt.Sprintf("host:disconnect:%s", ip)
	if len(port) != 0 {
		cmd = fmt.Sprintf("host:disconnect:%s:%d", ip, port[0])
	}

	var resp string
	if resp, err = c.executeCommand(cmd); err != nil {
		return err
	}
	if !strings.HasPrefix(resp, "disconnected") {
		return fmt.Errorf("adb disconnect: %s", resp)
	}
	return
}

func (c Client) DisconnectAll() (err error) {
	var resp string
	if resp, err = c.executeCommand("host:disconnect:"); err != nil {
		return err
	}

	if !strings.HasPrefix(resp, "disconnected everything") {
		return fmt.Errorf("adb disconnect all: %s", resp)
	}
	return
}

func (c Client) KillServer() (err error) {
	var tp transport
	if tp, err = c.createTransport(); err != nil {
		return err
	}
	defer func() { _ = tp.Close() }()

	err = tp.Send("host:kill")
	return
}

func (c Client) createTransport() (tp transport, err error) {
	return newTransport(fmt.Sprintf("%s:%d", c.host, c.port))
}

func (c Client) executeCommand(command string, onlyVerifyResponse ...bool) (resp string, err error) {
	if len(onlyVerifyResponse) == 0 {
		onlyVerifyResponse = []bool{false}
	}

	var tp transport
	if tp, err = c.createTransport(); err != nil {
		return "", err
	}
	defer func() { _ = tp.Close() }()

	if err = tp.Send(command); err != nil {
		return "", err
	}
	if err = tp.VerifyResponse(); err != nil {
		return "", err
	}

	if onlyVerifyResponse[0] {
		return
	}

	if resp, err = tp.UnpackString(); err != nil {
		return "", err
	}
	return
}
