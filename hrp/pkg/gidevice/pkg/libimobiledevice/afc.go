package libimobiledevice

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const AfcServiceName = "com.apple.afc"

func NewAfcClient(innerConn InnerConn) *AfcClient {
	return &AfcClient{
		innerConn: innerConn,
	}
}

type AfcClient struct {
	innerConn InnerConn
	packetNum uint64
}

func (c *AfcClient) newPacket(operation uint64, data, payload []byte) Packet {
	c.packetNum++
	pkt := &afcPacket{
		operation: operation,
		packetNum: c.packetNum,
		entireLen: 40,
		thisLen:   40,
	}
	if data != nil {
		n := uint64(len(data))
		pkt.entireLen += n
		pkt.thisLen += n
	}
	if payload != nil {
		pkt.entireLen += uint64(len(payload))
	}
	return pkt
}

func (c *AfcClient) Send(operation uint64, data, payload []byte) (err error) {
	pkt := c.newPacket(operation, data, payload)
	var raw []byte
	if raw, err = pkt.Pack(); err != nil {
		return fmt.Errorf("send packet (afc): %w", err)
	}

	buf := new(bytes.Buffer)
	buf.Write(raw)
	if data != nil {
		debugLog(fmt.Sprintf("--> %s ...afc data...\n", pkt))
		buf.Write(data)
	} else {
		debugLog(fmt.Sprintf("--> %s\n", pkt))
	}

	if err = c.innerConn.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("send packet (afc): %w", err)
	}

	if payload != nil {
		if err = c.innerConn.Write(payload); err != nil {
			return fmt.Errorf("send packet (afc): %w", err)
		}
	}

	return
}

func (c *AfcClient) Receive() (respMsg *AfcMessage, err error) {
	var bufHeader []byte
	if bufHeader, err = c.innerConn.Read(40); err != nil {
		return nil, fmt.Errorf("receive packet (afc): %w", err)
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bufHeader)
	var respPkt *afcPacket
	if respPkt, err = new(afcPacket).unpack(buffer); err != nil {
		return nil, fmt.Errorf("receive packet (afc): %w", err)
	}

	respMsg = new(AfcMessage)
	respMsg.Operation = respPkt.operation

	buffer.Reset()
	if respPkt.entireLen > 40 {
		length := int(respPkt.entireLen - 40)
		var bufDataAndPayload []byte
		if bufDataAndPayload, err = c.innerConn.Read(length); err != nil {
			return nil, fmt.Errorf("receive packet (afc): %w", err)
		}
		buffer.Write(bufDataAndPayload)
	}

	bufData := make([]byte, respPkt.thisLen-40)
	if _, err = buffer.Read(bufData); err != nil {
		return nil, fmt.Errorf("receive packet (afc buffer): %w", err)
	}
	respMsg.Data = bufData
	respMsg.Payload = buffer.Bytes()

	debugLog(fmt.Sprintf("<-- %s\n%s\n%s", respPkt, hex.Dump(respMsg.Data), hex.Dump(respMsg.Payload)))

	return
}
