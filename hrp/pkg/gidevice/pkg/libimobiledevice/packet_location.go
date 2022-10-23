package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

var _ Packet = (*locationPacket)(nil)

type locationPacket struct {
	lon float64
	lat float64
}

func (l *locationPacket) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		return nil, fmt.Errorf("packet (location) pack: %w", err)
	}

	latS := []byte(strconv.FormatFloat(l.lat, 'E', -1, 64))
	if err := binary.Write(buf, binary.BigEndian, uint32(len(latS))); err != nil {
		return nil, fmt.Errorf("packet (location) pack: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, latS); err != nil {
		return nil, fmt.Errorf("packet (location) pack: %w", err)
	}

	lonS := []byte(strconv.FormatFloat(l.lon, 'E', -1, 64))
	if err := binary.Write(buf, binary.BigEndian, uint32(len(lonS))); err != nil {
		return nil, fmt.Errorf("packet (location) pack: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, lonS); err != nil {
		return nil, fmt.Errorf("packet (location) pack: %w", err)
	}

	return buf.Bytes(), nil
}

func (l *locationPacket) Unpack(buffer *bytes.Buffer) (Packet, error) {
	panic("never use (location)")
}

func (l *locationPacket) Unmarshal(v interface{}) error {
	panic("never use (location)")
}

func (l *locationPacket) String() string {
	return fmt.Sprintf("lon: %v, lat: %v\n", l.lon, l.lat)
}
