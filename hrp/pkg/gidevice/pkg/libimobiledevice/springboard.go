package libimobiledevice

type IconPNGDataResponse struct {
	PNGData []byte `plist:"pngData"`
}

type InterfaceOrientationResponse struct {
	Orientation OrientationState `plist:"interfaceOrientation"`
}

type OrientationState int64

const (
	Unknown OrientationState = iota
	Portrait
	PortraitUpsideDown
	LandscapeRight
	LandscapeLeft
)

const (
	SpringBoardServiceName = "com.apple.springboardservices"
)

func NewSpringBoardClient(innerConn InnerConn) *SpringBoardClient {
	return &SpringBoardClient{
		newServicePacketClient(innerConn),
	}
}

type SpringBoardClient struct {
	client *servicePacketClient
}

func (c *SpringBoardClient) InnerConn() InnerConn {
	return c.client.innerConn
}

func (c *SpringBoardClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *SpringBoardClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *SpringBoardClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}

func (c *SpringBoardClient) NewBinaryPacket(req interface{}) (Packet, error) {
	return c.client.NewBinaryPacket(req)
}
