package libimobiledevice

const (
	DiagnosticsRelayServiceName = "com.apple.mobile.diagnostics_relay"
)

type DiagnosticsRelayBasicRequest struct {
	Request string `plist:"Request"`
	Label   string `plist:"Label"`
}

func NewDiagnosticsRelayClient(innerConn InnerConn) *DiagnosticsRelayClient {
	return &DiagnosticsRelayClient{
		newServicePacketClient(innerConn),
	}
}

type DiagnosticsRelayClient struct {
	client *servicePacketClient
}

func (c *DiagnosticsRelayClient) InnerConn() InnerConn {
	return c.client.innerConn
}

func (c *DiagnosticsRelayClient) NewBasicRequest(relayType string) *DiagnosticsRelayBasicRequest {
	return &DiagnosticsRelayBasicRequest{
		Request: relayType,
		Label:   BundleID,
	}
}

func (c *DiagnosticsRelayClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *DiagnosticsRelayClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}
