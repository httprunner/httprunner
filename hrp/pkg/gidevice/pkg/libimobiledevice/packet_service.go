package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"howett.net/plist"
)

var _ Packet = (*servicePacket)(nil)

type servicePacket struct {
	length uint32
	body   []byte
}

func (p *servicePacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, p.length)
	buf.Write(b)

	buf.Write(p.body)

	return buf.Bytes(), nil
}

func (p *servicePacket) Unpack(buffer *bytes.Buffer) (pkt Packet, err error) {
	respPkt := new(servicePacket)
	if err = binary.Read(buffer, binary.BigEndian, &respPkt.length); err != nil {
		return nil, fmt.Errorf("packet (service) unpack: %w", err)
	}
	respPkt.body = buffer.Bytes()

	return respPkt, nil
}

func (p *servicePacket) Unmarshal(v interface{}) (err error) {
	_, err = plist.Unmarshal(p.body, v)
	return
}

func (p *servicePacket) String() string {
	return fmt.Sprintf("Length: %d\n%s", p.length, p.body)
}
