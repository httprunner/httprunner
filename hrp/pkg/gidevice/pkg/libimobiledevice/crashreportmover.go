package libimobiledevice

const (
	CrashReportMoverServiceName      = "com.apple.crashreportmover"
	CrashReportCopyMobileServiceName = "com.apple.crashreportcopymobile"
)

func NewCrashReportMoverClient(innerConn InnerConn) *CrashReportMoverClient {
	return &CrashReportMoverClient{
		newServicePacketClient(innerConn),
	}
}

type CrashReportMoverClient struct {
	client *servicePacketClient
}

func (c *CrashReportMoverClient) InnerConn() InnerConn {
	return c.client.innerConn
}
