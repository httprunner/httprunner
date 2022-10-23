package libimobiledevice

const (
	InstrumentsServiceName            = "com.apple.instruments.remoteserver"
	InstrumentsSecureProxyServiceName = "com.apple.instruments.remoteserver.DVTSecureSocketProxy"
)

func NewInstrumentsClient(innerConn InnerConn) *InstrumentsClient {
	return &InstrumentsClient{
		client: newDtxMessageClient(innerConn),
	}
}

type InstrumentsClient struct {
	client *dtxMessageClient
}

func (c *InstrumentsClient) NotifyOfPublishedCapabilities() (publishedChannels map[string]int32, err error) {
	return c.client.Connection()
}

func (c *InstrumentsClient) RequestChannel(channel string) (id uint32, err error) {
	return c.client.MakeChannel(channel)
}

func (c *InstrumentsClient) Invoke(selector string, args *AuxBuffer, channelCode uint32, expectsReply bool) (result *DTXMessageResult, err error) {
	var msgID uint32
	if msgID, err = c.client.SendDTXMessage(selector, args.Bytes(), channelCode, expectsReply); err != nil {
		return nil, err
	}
	if expectsReply {
		if result, err = c.client.GetResult(msgID); err != nil {
			return nil, err
		}
	}
	return
}

func (c *InstrumentsClient) RegisterCallback(obj string, cb func(m DTXMessageResult)) {
	c.client.RegisterCallback(obj, cb)
}
