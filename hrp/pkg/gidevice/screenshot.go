package gidevice

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ Screenshot = (*screenshot)(nil)

func newScreenshot(client *libimobiledevice.ScreenshotClient) *screenshot {
	return &screenshot{
		client:    client,
		exchanged: false,
	}
}

type screenshot struct {
	client    *libimobiledevice.ScreenshotClient
	exchanged bool
}

func (s *screenshot) Take() (raw *bytes.Buffer, err error) {
	if err = s.exchange(); err != nil {
		return nil, err
	}

	// link service
	req := []interface{}{
		"DLMessageProcessMessage",
		map[string]interface{}{
			"MessageType": "ScreenShotRequest",
		},
	}

	var pkt libimobiledevice.Packet
	if pkt, err = s.client.NewBinaryPacket(req); err != nil {
		return nil, err
	}

	if err = s.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return nil, err
	}

	var resp []interface{}
	if err = respPkt.Unmarshal(&resp); err != nil {
		return nil, err
	}

	if resp[0].(string) != "DLMessageProcessMessage" {
		return nil, fmt.Errorf("message device not ready %s %s", resp[3], resp[4])
	}

	raw = new(bytes.Buffer)

	screen := resp[1].(map[string]interface{})
	var data []byte
	ok := false
	if data, ok = screen["ScreenShotData"].([]byte); !ok {
		return nil, errors.New("`ScreenShotData` not ready")
	}
	if _, err = raw.Write(data); err != nil {
		return nil, err
	}
	return
}

func (s *screenshot) exchange() (err error) {
	if s.exchanged {
		return
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return err
	}

	var resp []interface{}
	if err = respPkt.Unmarshal(&resp); err != nil {
		return err
	}

	req := []interface{}{
		"DLMessageVersionExchange",
		"DLVersionsOk",
		resp[1],
	}

	var pkt libimobiledevice.Packet
	if pkt, err = s.client.NewBinaryPacket(req); err != nil {
		return err
	}

	if err = s.client.SendPacket(pkt); err != nil {
		return err
	}

	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return err
	}

	if err = respPkt.Unmarshal(&resp); err != nil {
		return err
	}

	if resp[3].(string) != "DLMessageDeviceReady" {
		return fmt.Errorf("message device not ready %s", resp[3])
	}

	s.exchanged = true
	return
}
