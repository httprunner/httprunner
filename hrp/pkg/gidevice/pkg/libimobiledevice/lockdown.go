package libimobiledevice

const ProtocolVersion = "2"

const LockdownPort = 62078

type RequestType string

const (
	RequestTypeQueryType     RequestType = "QueryType"
	RequestTypeSetValue      RequestType = "SetValue"
	RequestTypeGetValue      RequestType = "GetValue"
	RequestTypePair          RequestType = "Pair"
	RequestTypeEnterRecovery RequestType = "EnterRecovery"
	RequestTypeStartSession  RequestType = "StartSession"
	RequestTypeStopSession   RequestType = "StopSession"
	RequestTypeStartService  RequestType = "StartService"
)

type LockdownType struct {
	Type string `plist:"Type"`
}

func NewLockdownClient(innerConn InnerConn) *LockdownClient {
	return &LockdownClient{
		client: newServicePacketClient(innerConn),
	}
}

type LockdownClient struct {
	client *servicePacketClient
}

func (c *LockdownClient) NewBasicRequest(reqType RequestType) *LockdownBasicRequest {
	return &LockdownBasicRequest{
		Label:           BundleID,
		ProtocolVersion: ProtocolVersion,
		Request:         reqType,
	}
}

func (c *LockdownClient) NewGetValueRequest(domain, key string) *LockdownValueRequest {
	return &LockdownValueRequest{
		LockdownBasicRequest: *c.NewBasicRequest(RequestTypeGetValue),
		Domain:               domain,
		Key:                  key,
	}
}

func (c *LockdownClient) NewSetValueRequest(domain, key string, value interface{}) *LockdownValueRequest {
	return &LockdownValueRequest{
		LockdownBasicRequest: *c.NewBasicRequest(RequestTypeSetValue),
		Domain:               domain,
		Key:                  key,
		Value:                value,
	}
}

func (c *LockdownClient) NewEnterRecoveryRequest() *LockdownBasicRequest {
	return c.NewBasicRequest(RequestTypeEnterRecovery)
}

func (c *LockdownClient) NewPairRequest(pairRecord *PairRecord) *LockdownPairRequest {
	return &LockdownPairRequest{
		LockdownBasicRequest: *c.NewBasicRequest(RequestTypePair),
		PairRecord:           pairRecord,
		PairingOptions: map[string]interface{}{
			"ExtendedPairingErrors": true,
		},
	}
}

func (c *LockdownClient) NewStartSessionRequest(buid, hostID string) *LockdownStartSessionRequest {
	return &LockdownStartSessionRequest{
		LockdownBasicRequest: *c.NewBasicRequest(RequestTypeStartSession),
		SystemBUID:           buid,
		HostID:               hostID,
	}
}

func (c *LockdownClient) NewStopSessionRequest(sessionID string) *LockdownStopSessionRequest {
	return &LockdownStopSessionRequest{
		LockdownBasicRequest: *c.NewBasicRequest(RequestTypeStopSession),
		SessionID:            sessionID,
	}
}

func (c *LockdownClient) NewStartServiceRequest(service string) *LockdownStartServiceRequest {
	return &LockdownStartServiceRequest{
		LockdownBasicRequest: *c.NewBasicRequest(RequestTypeStartService),
		Service:              service,
	}
}

func (c *LockdownClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *LockdownClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *LockdownClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}

func (c *LockdownClient) EnableSSL(version []int, pairRecord *PairRecord) (err error) {
	return c.client.innerConn.Handshake(version, pairRecord)
}

type (
	LockdownBasicRequest struct {
		Label           string      `plist:"Label"`
		ProtocolVersion string      `plist:"ProtocolVersion"`
		Request         RequestType `plist:"Request"`
	}

	LockdownValueRequest struct {
		LockdownBasicRequest
		Domain string      `plist:"Domain,omitempty"`
		Key    string      `plist:"Key,omitempty"`
		Value  interface{} `plist:"Value,omitempty"`
	}

	LockdownPairRequest struct {
		LockdownBasicRequest
		PairRecord     *PairRecord            `plist:"PairRecord"`
		PairingOptions map[string]interface{} `plist:"PairingOptions"`
	}

	LockdownStartSessionRequest struct {
		LockdownBasicRequest
		SystemBUID string `plist:"SystemBUID"`
		HostID     string `plist:"HostID"`
	}

	LockdownStopSessionRequest struct {
		LockdownBasicRequest
		SessionID string `plist:"SessionID"`
	}

	LockdownStartServiceRequest struct {
		LockdownBasicRequest
		Service   string `plist:"Service"`
		EscrowBag []byte `plist:"EscrowBag,omitempty"`
	}
)

type (
	LockdownBasicResponse struct {
		Request string `plist:"Request"`
		Error   string `plist:"Error"`
	}

	LockdownTypeResponse struct {
		LockdownBasicResponse
		Type string `plist:"Type"`
	}

	LockdownValueResponse struct {
		LockdownBasicResponse
		Key   string      `plist:"Key"`
		Value interface{} `plist:"Value"`
	}

	LockdownPairResponse struct {
		LockdownBasicResponse
		EscrowBag []byte `plist:"EscrowBag"`
	}

	LockdownStartSessionResponse struct {
		LockdownBasicResponse
		EnableSessionSSL bool   `plist:"EnableSessionSSL"`
		SessionID        string `plist:"SessionID"`
	}

	LockdownStartServiceResponse struct {
		LockdownBasicResponse
		EnableServiceSSL bool   `plist:"EnableServiceSSL"`
		Port             int    `plist:"Port"`
		Service          string `plist:"Service"`
	}
)
