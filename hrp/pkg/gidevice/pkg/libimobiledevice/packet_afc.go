package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var afcHeader = []byte{0x43, 0x46, 0x41, 0x36, 0x4C, 0x50, 0x41, 0x41}

var _ Packet = (*afcPacket)(nil)

type afcPacket struct {
	entireLen uint64
	thisLen   uint64
	packetNum uint64
	operation uint64
}

func (p *afcPacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(afcHeader)

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, p.entireLen)
	buf.Write(b)
	binary.LittleEndian.PutUint64(b, p.thisLen)
	buf.Write(b)
	binary.LittleEndian.PutUint64(b, p.packetNum)
	buf.Write(b)
	binary.LittleEndian.PutUint64(b, p.operation)
	buf.Write(b)

	return buf.Bytes(), nil
}

func (p *afcPacket) Unpack(buffer *bytes.Buffer) (Packet, error) {
	return p.unpack(buffer)
}

func (p *afcPacket) unpack(buffer *bytes.Buffer) (*afcPacket, error) {
	magic := make([]byte, 8)
	if _, err := buffer.Read(magic); err != nil {
		return nil, fmt.Errorf("afc packet unpack: %w", err)
	}
	if bytes.Compare(magic, afcHeader) != 0 {
		return nil, errors.New("afc packet unpack: header not match")
	}

	respPkt := new(afcPacket)
	if err := binary.Read(buffer, binary.LittleEndian, &respPkt.entireLen); err != nil {
		return nil, fmt.Errorf("afc packet unpack: %w", err)
	}
	if err := binary.Read(buffer, binary.LittleEndian, &respPkt.thisLen); err != nil {
		return nil, fmt.Errorf("afc packet unpack: %w", err)
	}
	if err := binary.Read(buffer, binary.LittleEndian, &respPkt.packetNum); err != nil {
		return nil, fmt.Errorf("afc packet unpack: %w", err)
	}
	if err := binary.Read(buffer, binary.LittleEndian, &respPkt.operation); err != nil {
		return nil, fmt.Errorf("afc packet unpack: %w", err)
	}
	return respPkt, nil
}

func (p *afcPacket) Unmarshal(v interface{}) error {
	// switch msg := v.(type) {
	// case *AfcMessage:
	// 	// msg.EntireLen = p.entireLen
	// 	// msg.ThisLen = p.thisLen
	// 	// msg.PacketNum = p.packetNum
	// 	msg.Operation = p.operation
	// default:
	// 	return errors.New("the type of the method parameter must be '*AfcMessage'")
	// }
	// return nil
	panic("never use (afcPacket)")
}

func (p *afcPacket) String() string {
	return fmt.Sprintf(
		"EntireLen: %d, ThisLen: %d, PacketNum: %d, Operation: %X\n",
		p.entireLen, p.thisLen, p.packetNum, p.operation,
	)
}
