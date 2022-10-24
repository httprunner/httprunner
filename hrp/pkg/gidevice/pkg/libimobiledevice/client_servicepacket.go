package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"howett.net/plist"
)

func newServicePacketClient(innerConn InnerConn) *servicePacketClient {
	return &servicePacketClient{
		innerConn: innerConn,
	}
}

type servicePacketClient struct {
	innerConn InnerConn
}

func (c *servicePacketClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.newPacket(req, plist.XMLFormat)
}

func (c *servicePacketClient) NewBinaryPacket(req interface{}) (Packet, error) {
	return c.newPacket(req, plist.BinaryFormat)
}

func (c *servicePacketClient) newPacket(req interface{}, format int) (Packet, error) {
	pkt := new(servicePacket)
	if buf, err := plist.Marshal(req, format); err != nil {
		return nil, fmt.Errorf("plist packet marshal: %w", err)
	} else {
		pkt.body = buf
	}
	pkt.length = uint32(len(pkt.body))
	return pkt, nil
}

func (c *servicePacketClient) SendPacket(pkt Packet) (err error) {
	var raw []byte
	if raw, err = pkt.Pack(); err != nil {
		return fmt.Errorf("send packet: %w", err)
	}
	debugLog(fmt.Sprintf("--> %s\n", pkt))
	return c.innerConn.Write(raw)
}

func (c *servicePacketClient) ReceivePacket() (respPkt Packet, err error) {
	var bufLen []byte
	if bufLen, err = c.innerConn.Read(4); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}
	lenPkg := binary.BigEndian.Uint32(bufLen)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(bufLen)

	var buf []byte
	if buf, err = c.innerConn.Read(int(lenPkg)); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}
	buffer.Write(buf)

	if respPkt, err = new(servicePacket).Unpack(buffer); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}

	debugLog(fmt.Sprintf("<-- %s\n", respPkt))

	var reply LockdownBasicResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, fmt.Errorf("receive packet: %w", err)
	}

	if reply.Error != "" {
		return nil, fmt.Errorf("receive packet: %s", reply.Error)
	}

	return
}
