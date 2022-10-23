package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

var (
	_ Packet = (*dtxMessagePayloadPacket)(nil)
	_ Packet = (*dtxMessageHeaderPacket)(nil)
	_ Packet = (*dtxMessagePacket)(nil)
)

type dtxMessagePayloadPacket struct {
	Flags           uint32
	AuxiliaryLength uint32
	TotalLength     uint64
}

func (p *dtxMessagePayloadPacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p.Flags)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.AuxiliaryLength)
	buf.Write(b)

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, p.TotalLength)
	buf.Write(b)

	return buf.Bytes(), nil
}

func (p *dtxMessagePayloadPacket) Unpack(buffer *bytes.Buffer) (pkt Packet, err error) {
	return p.unpack(buffer)
}

func (p *dtxMessagePayloadPacket) unpack(buffer *bytes.Buffer) (pkt *dtxMessagePayloadPacket, err error) {
	respPkt := new(dtxMessagePayloadPacket)
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.Flags); err != nil {
		return nil, fmt.Errorf("packet (DTXMessagePayloadHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.AuxiliaryLength); err != nil {
		return nil, fmt.Errorf("packet (DTXMessagePayloadHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.TotalLength); err != nil {
		return nil, fmt.Errorf("packet (DTXMessagePayloadHeader) unpack: %w", err)
	}
	return respPkt, nil
}

func (p *dtxMessagePayloadPacket) Unmarshal(v interface{}) error {
	panic("never use (dtxMessagePayloadHeader)")
}

func (p *dtxMessagePayloadPacket) String() string {
	return fmt.Sprintf("DTXMessagePayloadHeader Flags: %d, AuxiliaryLength: %d, TotalLength: %d\n",
		p.Flags, p.AuxiliaryLength, p.TotalLength,
	)
}

type dtxMessageHeaderPacket struct {
	Magic             uint32
	CB                uint32
	FragmentId        uint16
	FragmentCount     uint16
	Length            uint32
	Identifier        uint32
	ConversationIndex uint32
	ChannelCode       uint32
	ExpectsReply      uint32
}

func (p *dtxMessageHeaderPacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p.Magic)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.CB)
	buf.Write(b)

	b = make([]byte, 2)
	binary.LittleEndian.PutUint16(b, p.FragmentId)
	buf.Write(b)
	binary.LittleEndian.PutUint16(b, p.FragmentCount)
	buf.Write(b)

	b = make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p.Length)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.Identifier)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.ConversationIndex)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.ChannelCode)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.ExpectsReply)
	buf.Write(b)

	return buf.Bytes(), nil
}

func (p *dtxMessageHeaderPacket) Unpack(buffer *bytes.Buffer) (pkt Packet, err error) {
	return p.unpack(buffer)
}

func (p *dtxMessageHeaderPacket) unpack(buffer *bytes.Buffer) (pkt *dtxMessageHeaderPacket, err error) {
	respPkt := new(dtxMessageHeaderPacket)
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.Magic); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.CB); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.FragmentId); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.FragmentCount); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.Length); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.Identifier); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.ConversationIndex); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.ChannelCode); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.ExpectsReply); err != nil {
		return nil, fmt.Errorf("packet (DTXMessageHeader) unpack: %w", err)
	}
	return respPkt, nil
}

func (p *dtxMessageHeaderPacket) Unmarshal(v interface{}) error {
	panic("never use (DTXMessageHeader)")
}

func (p *dtxMessageHeaderPacket) String() string {
	return fmt.Sprintf("DTXMessageHeader Magic: %d, CB: %d, FragmentId: %d, FragmentCount: %d\n"+
		"Length: %d, Identifier: %d, ConversationIndex: %d, ChannelCode: %d, ExpectsReply: %d\n",
		p.Magic, p.CB, p.FragmentId, p.FragmentCount,
		p.Length, p.Identifier, p.ConversationIndex, p.ChannelCode, p.ExpectsReply,
	)
}

type dtxMessagePacket struct {
	Header  *dtxMessageHeaderPacket
	Payload *dtxMessagePayloadPacket
	Aux     []byte
	Sel     []byte
}

func (p *dtxMessagePacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	raw, err := p.Header.Pack()
	if err != nil {
		return nil, fmt.Errorf("packet (DTXMessagePacket) pack: %w", err)
	}
	buf.Write(raw)

	if raw, err = p.Payload.Pack(); err != nil {
		return nil, fmt.Errorf("packet (DTXMessagePacket) pack: %w", err)
	}
	buf.Write(raw)

	if p.Aux != nil || len(p.Aux) != 0 {
		buf.Write(p.Aux)
	}
	if p.Sel != nil || len(p.Sel) != 0 {
		buf.Write(p.Sel)
	}

	return buf.Bytes(), nil
}

func (p *dtxMessagePacket) Unpack(buffer *bytes.Buffer) (Packet, error) {
	panic("implement me")
}

func (p *dtxMessagePacket) Unmarshal(v interface{}) error {
	panic("never use (DTXMessagePacket)")
}

func (p *dtxMessagePacket) String() string {
	return fmt.Sprintf(
		"DTXMessagePacket %s\n%s\n"+
			"%s\n%s\n",
		p.Header.String(), p.Payload.String(),
		// p.Aux, p.Sel,
		hex.Dump(p.Aux), hex.Dump(p.Sel),
	)
}
