package libimobiledevice

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/lunixbochs/struc"
)

const PcapdServiceName = "com.apple.pcapd"

func filterPacket(pid int, procName string) func(*IOSPacketHeader) bool {
	return func(iph *IOSPacketHeader) bool {
		if pid > 0 {
			return iph.Pid == int32(pid) ||
				iph.Pid2 == int32(pid)
		}
		if procName != "" {
			return strings.HasPrefix(iph.ProcName, procName) ||
				strings.HasPrefix(iph.ProcName2, procName)
		}
		return true
	}
}

func NewPcapdClient(innerConn InnerConn, targetPID int, targetProcName string) *PcapdClient {
	return &PcapdClient{
		filter: filterPacket(targetPID, targetProcName),
		client: newServicePacketClient(innerConn),
	}
}

type PcapdClient struct {
	filter func(*IOSPacketHeader) bool
	client *servicePacketClient
}

func (c *PcapdClient) ReceivePacket() (respPkt Packet, err error) {
	var bufLen []byte
	if bufLen, err = c.client.innerConn.Read(4); err != nil {
		return nil, fmt.Errorf("lockdown(Pcapd) receive: %w", err)
	}
	lenPkg := binary.BigEndian.Uint32(bufLen)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(bufLen)

	var buf []byte
	if buf, err = c.client.innerConn.Read(int(lenPkg)); err != nil {
		return nil, fmt.Errorf("lockdown(Pcapd) receive: %w", err)
	}
	buffer.Write(buf)

	if respPkt, err = new(servicePacket).Unpack(buffer); err != nil {
		return nil, fmt.Errorf("lockdown(Pcapd) receive: %w", err)
	}

	debugLog(fmt.Sprintf("<-- %s\n", respPkt))

	return
}

const (
	PacketHeaderSize = uint32(95)
)

// ref: https://github.com/danielpaulus/go-ios/blob/fc943b9d236571f9775f5c593e3d49bb5bd67afd/ios/pcap/pcap.go#L27
type IOSPacketHeader struct {
	HdrSize        uint32  `struc:"uint32,big"`
	Version        uint8   `struc:"uint8,big"`
	PacketSize     uint32  `struc:"uint32,big"`
	Type           uint8   `struc:"uint8,big"`
	Unit           uint16  `struc:"uint16,big"`
	IO             uint8   `struc:"uint8,big"`
	ProtocolFamily uint32  `struc:"uint32,big"`
	FramePreLength uint32  `struc:"uint32,big"`
	FramePstLength uint32  `struc:"uint32,big"`
	IFName         string  `struc:"[16]byte"`
	Pid            int32   `struc:"int32,little"`
	ProcName       string  `struc:"[17]byte"`
	Unknown        uint32  `struc:"uint32,little"`
	Pid2           int32   `struc:"int32,little"`
	ProcName2      string  `struc:"[17]byte"`
	Unknown2       [8]byte `struc:"[8]byte"`
}

func (c *PcapdClient) GetPacket(buf []byte) ([]byte, error) {
	iph := IOSPacketHeader{}
	preader := bytes.NewReader(buf)
	_ = struc.Unpack(preader, &iph)

	if c.filter != nil {
		if !c.filter(&iph) {
			return nil, nil
		}
	}

	// support ios 15 beta4
	if iph.HdrSize > PacketHeaderSize {
		buf := make([]byte, iph.HdrSize-PacketHeaderSize)
		_, err := io.ReadFull(preader, buf)
		if err != nil {
			return []byte{}, err
		}
	}

	packet, err := ioutil.ReadAll(preader)
	if err != nil {
		return packet, err
	}
	if iph.FramePreLength == 0 {
		ext := []byte{
			0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe, 0xbe, 0xfe,
			0xbe, 0xfe, 0xbe, 0xfe, 0x08, 0x00,
		}
		return append(ext, packet...), nil
	}
	return packet, nil
}

type PcaprecHdrS struct {
	TsSec   int `struc:"uint32,little"` /* timestamp seconds */
	TsUsec  int `struc:"uint32,little"` /* timestamp microseconds */
	InclLen int `struc:"uint32,little"` /* number of octets of packet saved in file */
	OrigLen int `struc:"uint32,little"` /* actual length of packet */
}

func (c *PcapdClient) CreatePacket(packet []byte) ([]byte, error) {
	now := time.Now()
	phs := &PcaprecHdrS{
		int(now.Unix()),
		int(now.UnixNano()/1e3 - now.Unix()*1e6),
		len(packet),
		len(packet),
	}
	var buf bytes.Buffer
	err := struc.Pack(&buf, phs)
	if err != nil {
		return nil, err
	}

	buf.Write(packet)
	return buf.Bytes(), nil
}

func (c *PcapdClient) Close() {
	c.client.innerConn.Close()
}
