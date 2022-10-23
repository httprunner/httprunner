package libimobiledevice

const (
	TestmanagerdSecureServiceName = "com.apple.testmanagerd.lockdown.secure"
	TestmanagerdServiceName       = "com.apple.testmanagerd.lockdown"
)

func NewTestmanagerdClient(innerConn InnerConn) *TestmanagerdClient {
	return &TestmanagerdClient{
		client: newDtxMessageClient(innerConn),
	}
}

type TestmanagerdClient struct {
	client *dtxMessageClient
}

func (t *TestmanagerdClient) Connection() (publishedChannels map[string]int32, err error) {
	return t.client.Connection()
}

func (t *TestmanagerdClient) MakeChannel(channel string) (id uint32, err error) {
	return t.client.MakeChannel(channel)
}

func (t *TestmanagerdClient) Invoke(selector string, args *AuxBuffer, channelCode uint32, expectsReply bool) (result *DTXMessageResult, err error) {
	var msgID uint32
	if msgID, err = t.client.SendDTXMessage(selector, args.Bytes(), channelCode, expectsReply); err != nil {
		return nil, err
	}
	if expectsReply {
		if result, err = t.client.GetResult(msgID); err != nil {
			return nil, err
		}
	}
	return
}

func (t *TestmanagerdClient) RegisterCallback(obj string, cb func(m DTXMessageResult)) {
	t.client.RegisterCallback(obj, cb)
}

func (t *TestmanagerdClient) Close() {
	t.client.Close()
}
