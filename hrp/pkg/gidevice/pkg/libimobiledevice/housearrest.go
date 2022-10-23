package libimobiledevice

const HouseArrestServiceName = "com.apple.mobile.house_arrest"

const (
	CommandTypeVendDocuments CommandType = "VendDocuments"
	CommandTypeVendContainer CommandType = "VendContainer"
)

func NewHouseArrestClient(innerConn InnerConn) *HouseArrestClient {
	return &HouseArrestClient{
		newServicePacketClient(innerConn),
	}
}

type HouseArrestClient struct {
	client *servicePacketClient
}

func (c *HouseArrestClient) NewBasicRequest(cmdType CommandType, bundleID string) *HouseArrestBasicRequest {
	return &HouseArrestBasicRequest{
		Command:    cmdType,
		Identifier: bundleID,
	}
}

func (c *HouseArrestClient) NewDocumentsRequest(bundleID string) *HouseArrestBasicRequest {
	return c.NewBasicRequest(CommandTypeVendDocuments, bundleID)
}

func (c *HouseArrestClient) NewContainerRequest(bundleID string) *HouseArrestBasicRequest {
	return c.NewBasicRequest(CommandTypeVendContainer, bundleID)
}

func (c *HouseArrestClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *HouseArrestClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *HouseArrestClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}

func (c *HouseArrestClient) InnerConn() InnerConn {
	return c.client.innerConn
}

type (
	HouseArrestBasicRequest struct {
		Command    CommandType `plist:"Command"`
		Identifier string      `plist:"Identifier"`
	}
)
