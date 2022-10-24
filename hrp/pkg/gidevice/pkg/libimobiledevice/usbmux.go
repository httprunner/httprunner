package libimobiledevice

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"time"

	"howett.net/plist"
)

var DefaultDeadlineTimeout = 30 * time.Second

const (
	BundleID         = "electricbubble.libimobiledevice"
	ProgramName      = "libimobiledevice"
	ClientVersion    = "libimobiledevice-beta"
	LibUSBMuxVersion = 3
)

type ReplyCode uint64

const (
	ReplyCodeOK ReplyCode = iota
	ReplyCodeBadCommand
	ReplyCodeBadDevice
	ReplyCodeConnectionRefused
	_ // ignore `4`
	_ // ignore `5`
	ReplyCodeBadVersion
)

func (rc ReplyCode) String() string {
	switch rc {
	case ReplyCodeOK:
		return "ok"
	case ReplyCodeBadCommand:
		return "bad command"
	case ReplyCodeBadDevice:
		return "bad device"
	case ReplyCodeConnectionRefused:
		return "connection refused"
	case ReplyCodeBadVersion:
		return "bad version"
	default:
		return "unknown reply code: " + strconv.Itoa(int(rc))
	}
}

type ProtoVersion uint32

// proto_version == 1
//    construct message plist
// else `0`? res == `RESULT_BADVERSION`
//    binary packet

const (
	ProtoVersionBinary ProtoVersion = iota
	ProtoVersionPlist
)

type ProtoMessageType uint32

const (
	_ ProtoMessageType = iota
	ProtoMessageTypeResult
	ProtoMessageTypeConnect
	ProtoMessageTypeListen
	ProtoMessageTypeDeviceAdd
	ProtoMessageTypeDeviceRemove
	ProtoMessageTypeDevicePaired
	_ // `7`
	ProtoMessageTypePlist
)

type MessageType string

const (
	MessageTypeResult           MessageType = "Result"
	MessageTypeConnect          MessageType = "Connect"
	MessageTypeListen           MessageType = "Listen"
	MessageTypeDeviceAdd        MessageType = "Attached"
	MessageTypeDeviceRemove     MessageType = "Detached"
	MessageTypeReadBUID         MessageType = "ReadBUID"
	MessageTypeReadPairRecord   MessageType = "ReadPairRecord"
	MessageTypeSavePairRecord   MessageType = "SavePairRecord"
	MessageTypeDeletePairRecord MessageType = "DeletePairRecord"
	MessageTypeDeviceList       MessageType = "ListDevices"
)

type BaseDevice struct {
	MessageType MessageType      `plist:"MessageType"`
	DeviceID    int              `plist:"DeviceID"`
	Properties  DeviceProperties `plist:"Properties"`
}

type DeviceProperties struct {
	DeviceID        int    `plist:"DeviceID"`
	ConnectionType  string `plist:"ConnectionType"`
	ConnectionSpeed int    `plist:"ConnectionSpeed"`
	ProductID       int    `plist:"ProductID"`
	LocationID      int    `plist:"LocationID"`
	SerialNumber    string `plist:"SerialNumber"`
	UDID            string `plist:"UDID"`
	USBSerialNumber string `plist:"USBSerialNumber"`

	EscapedFullServiceName string `plist:"EscapedFullServiceName"`
	InterfaceIndex         int    `plist:"InterfaceIndex"`
	NetworkAddress         []byte `plist:"NetworkAddress"`
}

func NewUsbmuxClient(timeout ...time.Duration) (c *UsbmuxClient, err error) {
	if len(timeout) == 0 {
		timeout = []time.Duration{DefaultDeadlineTimeout}
	}
	c = &UsbmuxClient{version: ProtoVersionPlist}
	var conn net.Conn
	if conn, err = rawDial(timeout[0]); err != nil {
		return nil, fmt.Errorf("usbmux connect: %w", err)
	}

	c.innerConn = newInnerConn(conn, timeout[0])
	return
}

type UsbmuxClient struct {
	innerConn InnerConn
	version   ProtoVersion
	tag       uint32
}

func (c *UsbmuxClient) NewBasicRequest(msgType MessageType) *BasicRequest {
	return &BasicRequest{
		MessageType:         msgType,
		BundleID:            BundleID,
		ProgramName:         ProgramName,
		ClientVersionString: ClientVersion,
		LibUSBMuxVersion:    LibUSBMuxVersion,
	}
}

func (c *UsbmuxClient) NewConnectRequest(deviceID, port int) *ConnectRequest {
	return &ConnectRequest{
		BasicRequest: *c.NewBasicRequest(MessageTypeConnect),
		DeviceID:     deviceID,
		PortNumber:   ((port << 8) & 0xFF00) | (port >> 8),
	}
}

func (c *UsbmuxClient) NewReadPairRecordRequest(udid string) *ReadPairRecordRequest {
	return &ReadPairRecordRequest{
		BasicRequest: *c.NewBasicRequest(MessageTypeReadPairRecord),
		PairRecordID: udid,
	}
}

func (c *UsbmuxClient) NewSavePairRecordRequest(udid string, deviceID int, data []byte) *SavePairRecordRequest {
	return &SavePairRecordRequest{
		BasicRequest:   *c.NewBasicRequest(MessageTypeSavePairRecord),
		PairRecordID:   udid,
		PairRecordData: data,
		DeviceID:       deviceID,
	}
}

func (c *UsbmuxClient) NewDeletePairRecordRequest(udid string) *DeletePairRecordRequest {
	return &DeletePairRecordRequest{
		BasicRequest: *c.NewBasicRequest(MessageTypeDeletePairRecord),
		PairRecordID: udid,
	}
}

func (c *UsbmuxClient) NewPacket(protoMsgType ProtoMessageType) Packet {
	return c.newPacket(protoMsgType)
}

func (c *UsbmuxClient) newPacket(protoMsgType ProtoMessageType) *packet {
	c.tag++
	pkt := &packet{
		version: c.version,
		msgType: protoMsgType,
		tag:     c.tag,
	}
	return pkt
}

func (c *UsbmuxClient) NewPlistPacket(req interface{}) (Packet, error) {
	pkt := c.newPacket(ProtoMessageTypePlist)
	if buf, err := plist.Marshal(req, plist.XMLFormat); err != nil {
		return nil, fmt.Errorf("plist packet marshal: %w", err)
	} else {
		pkt.body = buf
	}
	pkt.length = uint32(len(pkt.body) + 4*4)
	return pkt, nil
}

func (c *UsbmuxClient) SendPacket(pkt Packet) (err error) {
	var raw []byte
	if raw, err = pkt.Pack(); err != nil {
		return fmt.Errorf("usbmux send: %w", err)
	}
	// debugLog(fmt.Sprintf("--> Length: %d, Version: %d, Type: %d, Tag: %d\n%s\n", pkt.Length(), pkt.Version(), pkt.Type(), pkt.Tag(), pkt.Body()))
	debugLog(fmt.Sprintf("--> %s\n", pkt))
	return c.innerConn.Write(raw)
}

func (c *UsbmuxClient) ReceivePacket() (respPkt Packet, err error) {
	var bufLen []byte
	if bufLen, err = c.innerConn.Read(4); err != nil {
		return nil, fmt.Errorf("usbmux receive: %w", err)
	}
	lenPkg := binary.LittleEndian.Uint32(bufLen)

	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(bufLen)

	var buf []byte
	if buf, err = c.innerConn.Read(int(lenPkg - 4)); err != nil {
		return nil, fmt.Errorf("usbmux receive: %w", err)
	}
	buffer.Write(buf)

	if respPkt, err = new(packet).Unpack(buffer); err != nil {
		return nil, fmt.Errorf("usbmux receive: %w", err)
	}

	// debugLog(fmt.Sprintf("<-- Length: %d, Version: %d, Type: %d, Tag: %d\n%s\n", respPkt.Length(), respPkt.Version(), respPkt.Type(), respPkt.Tag(), respPkt.Body()))
	debugLog(fmt.Sprintf("<-- %s\n", respPkt))

	reply := struct {
		MessageType string    `plist:"MessageType"`
		Number      ReplyCode `plist:"Number"`
	}{}
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, fmt.Errorf("usbmux receive: %w", err)
	}

	if reply.Number != ReplyCodeOK {
		return nil, fmt.Errorf("usbmux receive: %s", reply.Number.String())
	}

	return
}

func (c *UsbmuxClient) Close() {
	c.innerConn.Close()
}

func (c *UsbmuxClient) RawConn() net.Conn {
	return c.innerConn.RawConn()
}

func (c *UsbmuxClient) InnerConn() InnerConn {
	return c.innerConn
}

func rawDial(timeout time.Duration) (net.Conn, error) {
	dialer := net.Dialer{
		Timeout: timeout,
	}

	var network, address string
	switch runtime.GOOS {
	case "darwin", "android", "linux":
		network, address = "unix", "/var/run/usbmuxd"
	case "windows":
		network, address = "tcp", "127.0.0.1:27015"
	default:
		return nil, fmt.Errorf("raw dial: unsupported system: %s", runtime.GOOS)
	}

	return dialer.Dial(network, address)
}

type InnerConn interface {
	Write(data []byte) (err error)
	Read(length int) (data []byte, err error)
	Handshake(version []int, pairRecord *PairRecord) (err error)
	DismissSSL() (err error)
	Close()
	RawConn() net.Conn
	Timeout(time.Duration)
}

func newInnerConn(conn net.Conn, timeout time.Duration) InnerConn {
	return &safeConn{
		conn:    conn,
		timeout: timeout,
	}
}

type safeConn struct {
	conn    net.Conn
	sslConn *tls.Conn
	timeout time.Duration
}

func (c *safeConn) Write(data []byte) (err error) {
	conn := c.RawConn()
	if c.timeout <= 0 {
		err = conn.SetWriteDeadline(time.Time{})
	} else {
		err = conn.SetWriteDeadline(time.Now().Add(c.timeout))
	}
	if err != nil {
		return err
	}

	for totalSent := 0; totalSent < len(data); {
		var sent int
		if sent, err = conn.Write(data[totalSent:]); err != nil {
			return err
		}
		if sent == 0 {
			return err
		}
		totalSent += sent
	}
	return
}

func (c *safeConn) Read(length int) (data []byte, err error) {
	conn := c.RawConn()
	if c.timeout <= 0 {
		err = conn.SetReadDeadline(time.Time{})
	} else {
		err = conn.SetReadDeadline(time.Now().Add(c.timeout))
	}
	if err != nil {
		return nil, err
	}

	data = make([]byte, 0, length)
	for len(data) < length {
		buf := make([]byte, length-len(data))
		_n, _err := 0, error(nil)
		if _n, _err = conn.Read(buf); _err != nil && _n == 0 {
			return nil, _err
		}
		data = append(data, buf[:_n]...)
	}
	return
}

func (c *safeConn) Handshake(version []int, pairRecord *PairRecord) (err error) {
	minVersion := uint16(tls.VersionTLS11)
	maxVersion := uint16(tls.VersionTLS11)

	if version[0] > 10 {
		minVersion = tls.VersionTLS11
		maxVersion = tls.VersionTLS13
	}

	cert, err := tls.X509KeyPair(pairRecord.RootCertificate, pairRecord.RootPrivateKey)
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
	}

	c.sslConn = tls.Client(c.conn, config)

	if err = c.sslConn.Handshake(); err != nil {
		return err
	}

	return
}

func (c *safeConn) DismissSSL() (err error) {
	if c.sslConn != nil {
		// err = c.sslConn.CloseWrite()
		// if err = c.sslConn.CloseWrite(); err != nil {
		// 	return err
		// }
		c.sslConn = nil
	}
	return
}

func (c *safeConn) Close() {
	if c.sslConn != nil {
		if err := c.sslConn.Close(); err != nil {
			debugLog(fmt.Sprintf("close: %s", err))
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			debugLog(fmt.Sprintf("close: %s", err))
		}
	}
}

// RawConn `sslConn` first
func (c *safeConn) RawConn() net.Conn {
	if c.sslConn != nil {
		return c.sslConn
	}
	return c.conn
}

func (c *safeConn) Timeout(duration time.Duration) {
	c.timeout = duration
}
