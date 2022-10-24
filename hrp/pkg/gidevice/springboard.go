package gidevice

import (
	"bytes"
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

func newSpringBoard(client *libimobiledevice.SpringBoardClient) *springboard {
	return &springboard{
		client: client,
	}
}

type springboard struct {
	client *libimobiledevice.SpringBoardClient
}

func (s springboard) GetIconPNGData(bundleId string) (raw *bytes.Buffer, err error) {
	var pkt libimobiledevice.Packet
	req := map[string]interface{}{
		"command":  "getIconPNGData",
		"bundleId": bundleId,
	}
	if pkt, err = s.client.NewBinaryPacket(req); err != nil {
		return
	}
	if err = s.client.SendPacket(pkt); err != nil {
		return nil, err
	}
	var respPkt libimobiledevice.Packet
	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return nil, err
	}
	var reply libimobiledevice.IconPNGDataResponse
	raw = new(bytes.Buffer)
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}
	if _, err = raw.Write(reply.PNGData); err != nil {
		return nil, err
	}
	return
}

func (s springboard) GetInterfaceOrientation() (orientation libimobiledevice.OrientationState, err error) {
	var pkt libimobiledevice.Packet
	req := map[string]interface{}{
		"command": "getInterfaceOrientation",
	}
	if pkt, err = s.client.NewBinaryPacket(req); err != nil {
		return
	}
	if err = s.client.SendPacket(pkt); err != nil {
		return 0, err
	}
	var respPkt libimobiledevice.Packet
	if respPkt, err = s.client.ReceivePacket(); err != nil {
		return 0, err
	}
	var reply libimobiledevice.InterfaceOrientationResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return 0, fmt.Errorf("receive packet: %w", err)
	}
	orientation = reply.Orientation
	return
}
