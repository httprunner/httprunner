package gidevice

import (
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ Testmanagerd = (*testmanagerd)(nil)

func newTestmanagerd(client *libimobiledevice.TestmanagerdClient, iOSVersion []int) *testmanagerd {
	return &testmanagerd{
		client:     client,
		iOSVersion: iOSVersion,
	}
}

type testmanagerd struct {
	client     *libimobiledevice.TestmanagerdClient
	iOSVersion []int
}

func (t *testmanagerd) notifyOfPublishedCapabilities() (err error) {
	_, err = t.client.Connection()
	return
}

func (t *testmanagerd) requestChannel(channel string) (id uint32, err error) {
	return t.client.MakeChannel(channel)
}

func (t *testmanagerd) newXCTestManagerDaemon() (xcTestManager XCTestManagerDaemon, err error) {
	var channelCode uint32
	if channelCode, err = t.requestChannel("dtxproxy:XCTestManager_IDEInterface:XCTestManager_DaemonConnectionInterface"); err != nil {
		return nil, err
	}
	xcTestManager = newXcTestManagerDaemon(t, channelCode)
	return
}

func (t *testmanagerd) invoke(selector string, args *libimobiledevice.AuxBuffer, channelCode uint32, expectsReply bool) (
	result *libimobiledevice.DTXMessageResult, err error) {
	return t.client.Invoke(selector, args, channelCode, expectsReply)
}

func (t *testmanagerd) registerCallback(obj string, cb func(m libimobiledevice.DTXMessageResult)) {
	t.client.RegisterCallback(obj, cb)
}

func (t *testmanagerd) close() {
	t.client.Close()
}
