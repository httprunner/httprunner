package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const ScreenshotServiceName = "com.apple.mobile.screenshotr"

func NewScreenshotClient(innerConn InnerConn) *ScreenshotClient {
	return &ScreenshotClient{
		client: newServicePacketClient(innerConn),
	}
}

type ScreenshotClient struct {
	client *servicePacketClient
}

func (c *ScreenshotClient) NewBinaryPacket(req interface{}) (Packet, error) {
	return c.client.NewBinaryPacket(req)
}

func (c *ScreenshotClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *ScreenshotClient) ReceivePacket() (respPkt Packet, err error) {
	var bufLen []byte
	if bufLen, err = c.client.innerConn.Read(4); err != nil {
		return nil, fmt.Errorf("lockdown(Screenshot) receive: %w", err)
	}
	lenPkg := binary.BigEndian.Uint32(bufLen)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(bufLen)

	var buf []byte
	if buf, err = c.client.innerConn.Read(int(lenPkg)); err != nil {
		return nil, fmt.Errorf("lockdown(Screenshot) receive: %w", err)
	}
	buffer.Write(buf)

	if respPkt, err = new(servicePacket).Unpack(buffer); err != nil {
		return nil, fmt.Errorf("lockdown(Screenshot) receive: %w", err)
	}

	debugLog(fmt.Sprintf("<-- %s\n", respPkt))

	return
}
