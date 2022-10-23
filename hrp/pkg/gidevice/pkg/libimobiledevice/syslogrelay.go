package libimobiledevice

const SyslogRelayServiceName = "com.apple.syslog_relay"

func NewSyslogRelayClient(innerConn InnerConn) *SyslogRelayClient {
	return &SyslogRelayClient{
		newServicePacketClient(innerConn),
	}
}

type SyslogRelayClient struct {
	client *servicePacketClient
}

func (c *SyslogRelayClient) InnerConn() InnerConn {
	return c.client.innerConn
}

func (c *SyslogRelayClient) Close() {
	c.client.innerConn.Close()
}
