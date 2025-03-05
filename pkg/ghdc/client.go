package ghdc

import (
	"fmt"
	"log"
	"strings"
)

const HdcServerPort = 8710

type Client struct {
	host string
	port int
}

func NewClient() (Client, error) {
	return NewClientWith("localhost")
}

func NewClientWith(host string, port ...int) (hdClient Client, err error) {
	if len(port) == 0 {
		port = []int{HdcServerPort}
	}
	hdClient.host = host
	hdClient.port = port[0]

	var tp transport
	if tp, err = hdClient.createTransport(); err != nil {
		return Client{}, err
	}
	defer func() { _ = tp.Close() }()

	return
}

func (c Client) ServerVersion() (version string, err error) {
	return c.executeCommand("version")
}

func (c Client) DeviceSerialList() (serials []string, err error) {
	var resp string
	if resp, err = c.executeCommand("list targets"); err != nil {
		return
	}

	lines := strings.Split(resp, "\n")
	serials = make([]string, 0, len(lines))

	for i := range lines {
		if lines[i] == "" {
			continue
		}
		fields := strings.Fields(lines[i])
		serials = append(serials, fields[0])
	}

	return
}

func (c Client) DeviceList() (devices []Device, err error) {
	var resp string
	if resp, err = c.executeCommand("list targets -v"); err != nil {
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
		if len(fields) < 4 || len(fields[0]) == 0 {
			debugLog(fmt.Sprintf("can't parse: %s", line))
			continue
		}
		if fields[2] == "Offline" {
			log.Printf("device [%s] Offline ", fields[0])
			continue
		}
		if fields[1] == "UART" {
			continue
		}
		slickAttrs := make(map[string]string)
		slickAttrs["usb"] = fields[1]
		slickAttrs["device_status"] = fields[2]
		device, err := NewDevice(c, fields[0], slickAttrs)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return
}

func (c Client) ForwardList() (deviceForward []DeviceForward, err error) {
	var resp string
	if resp, err = c.executeCommand("fport ls"); err != nil {
		return nil, err
	}

	lines := strings.Split(resp, "\n")
	deviceForward = make([]DeviceForward, 0, len(lines))

	for i := range lines {
		line := strings.TrimSpace(lines[i])
		line = strings.ReplaceAll(line, "'", "")
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		deviceForward = append(deviceForward, DeviceForward{Local: fields[0], Remote: fields[1]})
	}

	return
}

func (c Client) Connect(ip string, port ...int) (err error) {
	if len(port) == 0 {
		port = []int{HdcServerPort}
	}

	var resp string
	if resp, err = c.executeCommand(fmt.Sprintf("tconn %s:%d", ip, port[0])); err != nil {
		return err
	}
	if !strings.HasPrefix(resp, "connected to") && !strings.HasPrefix(resp, "already connected to") {
		return fmt.Errorf("hd connect: %s", resp)
	}
	return
}

func (c Client) KillServer() (err error) {
	var tp transport
	if tp, err = c.createTransport(); err != nil {
		return err
	}
	defer func() { _ = tp.Close() }()

	err = tp.SendCommand("kill")
	return
}

func (c Client) createTransport() (tp transport, err error) {
	return newTransport(fmt.Sprintf("%s:%d", c.host, c.port), false, "")
}

func (c Client) executeCommand(command string) (resp string, err error) {
	var tp transport
	if tp, err = c.createTransport(); err != nil {
		return "", err
	}
	defer func() { _ = tp.Close() }()

	if err = tp.SendCommand(command); err != nil {
		return "", err
	}
	return tp.ReadStringAll()
}
