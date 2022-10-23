package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"howett.net/plist"
)

var _ Packet = (*packet)(nil)

type packet struct {
	length  uint32
	version ProtoVersion
	msgType ProtoMessageType
	tag     uint32
	body    []byte
}

func (p *packet) Pack() ([]byte, error) {
	buf := new(bytes.Buffer)

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p.length)
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, uint32(p.version))
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, uint32(p.msgType))
	buf.Write(b)
	binary.LittleEndian.PutUint32(b, p.tag)
	buf.Write(b)

	buf.Write(p.body)

	return buf.Bytes(), nil
}

func (p *packet) Unpack(buffer *bytes.Buffer) (pkt Packet, err error) {
	respPkt := new(packet)
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.length); err != nil {
		return nil, fmt.Errorf("packet unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.version); err != nil {
		return nil, fmt.Errorf("packet unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.msgType); err != nil {
		return nil, fmt.Errorf("packet unpack: %w", err)
	}
	if err = binary.Read(buffer, binary.LittleEndian, &respPkt.tag); err != nil {
		return nil, fmt.Errorf("packet unpack: %w", err)
	}
	respPkt.body = buffer.Bytes()
	return respPkt, nil
}

func (p *packet) Unmarshal(v interface{}) (err error) {
	_, err = plist.Unmarshal(p.body, v)
	return
}

func (p *packet) String() string {
	return fmt.Sprintf(
		"Length: %d, Version: %d, Type: %d, Tag: %d\n%s",
		p.length, p.version, p.msgType, p.tag, p.body,
	)
}

type (
	BasicRequest struct {
		MessageType         MessageType `plist:"MessageType"`
		BundleID            string      `plist:"BundleID,omitempty"`
		ProgramName         string      `plist:"ProgName,omitempty"`
		ClientVersionString string      `plist:"ClientVersionString"`
		LibUSBMuxVersion    uint        `plist:"kLibUSBMuxVersion"`
	}

	ConnectRequest struct {
		BasicRequest
		DeviceID   int `plist:"DeviceID"`
		PortNumber int `plist:"PortNumber"`
	}

	ReadPairRecordRequest struct {
		BasicRequest
		PairRecordID string `plist:"PairRecordID"`
	}

	SavePairRecordRequest struct {
		BasicRequest
		PairRecordID   string `plist:"PairRecordID"`
		PairRecordData []byte `plist:"PairRecordData"`
		DeviceID       int    `plist:"DeviceID"`
	}

	DeletePairRecordRequest struct {
		BasicRequest
		PairRecordID string `plist:"PairRecordID"`
	}
)

type PairRecord struct {
	DeviceCertificate []byte `plist:"DeviceCertificate"`
	EscrowBag         []byte `plist:"EscrowBag,omitempty"`
	HostCertificate   []byte `plist:"HostCertificate"`
	HostPrivateKey    []byte `plist:"HostPrivateKey,omitempty"`
	HostID            string `plist:"HostID"`
	RootCertificate   []byte `plist:"RootCertificate"`
	RootPrivateKey    []byte `plist:"RootPrivateKey,omitempty"`
	SystemBUID        string `plist:"SystemBUID"`
	WiFiMACAddress    string `plist:"WiFiMACAddress,omitempty"`
}
