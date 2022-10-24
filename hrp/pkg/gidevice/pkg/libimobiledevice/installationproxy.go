package libimobiledevice

const InstallationProxyServiceName = "com.apple.mobile.installation_proxy"

const (
	CommandTypeBrowse    CommandType = "Browse"
	CommandTypeLookup    CommandType = "Lookup"
	CommandTypeInstall   CommandType = "Install"
	CommandTypeUninstall CommandType = "Uninstall"
)

type ApplicationType string

const (
	ApplicationTypeSystem   ApplicationType = "System"
	ApplicationTypeUser     ApplicationType = "User"
	ApplicationTypeInternal ApplicationType = "internal"
	ApplicationTypeAny      ApplicationType = "Any"
)

func NewInstallationProxyClient(innerConn InnerConn) *InstallationProxyClient {
	return &InstallationProxyClient{
		client: newServicePacketClient(innerConn),
	}
}

type InstallationProxyClient struct {
	client *servicePacketClient
}

func (c *InstallationProxyClient) NewBasicRequest(cmdType CommandType, opt *InstallationProxyOption) *InstallationProxyBasicRequest {
	req := &InstallationProxyBasicRequest{Command: cmdType}
	if opt != nil {
		req.ClientOptions = opt
	}
	return req
}

func (c *InstallationProxyClient) NewInstallRequest(bundleID, packagePath string) *InstallationProxyInstallRequest {
	opt := &InstallationProxyOption{
		BundleID: bundleID,
	}
	req := &InstallationProxyInstallRequest{
		Command:       CommandTypeInstall,
		ClientOptions: opt,
		PackagePath:   packagePath,
	}
	return req
}

func (c *InstallationProxyClient) NewUninstallRequest(bundleID string) *InstallationProxyUninstallRequest {
	req := &InstallationProxyUninstallRequest{
		Command:  CommandTypeUninstall,
		BundleID: bundleID,
	}
	return req
}

func (c *InstallationProxyClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *InstallationProxyClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *InstallationProxyClient) ReceivePacket() (respPkt Packet, err error) {
	return c.client.ReceivePacket()
}

type InstallationProxyOption struct {
	ApplicationType  ApplicationType `plist:"ApplicationType,omitempty"`
	ReturnAttributes []string        `plist:"ReturnAttributes,omitempty"`
	MetaData         bool            `plist:"com.apple.mobile_installation.metadata,omitempty"`
	BundleIDs        []string        `plist:"BundleIDs,omitempty"`          // for Lookup
	BundleID         string          `plist:"CFBundleIdentifier,omitempty"` // for Install
}

type (
	InstallationProxyBasicRequest struct {
		Command       CommandType              `plist:"Command"`
		ClientOptions *InstallationProxyOption `plist:"ClientOptions,omitempty"`
	}

	InstallationProxyInstallRequest struct {
		Command       CommandType              `plist:"Command"`
		ClientOptions *InstallationProxyOption `plist:"ClientOptions"`
		PackagePath   string                   `plist:"PackagePath"`
	}

	InstallationProxyUninstallRequest struct {
		Command  CommandType `plist:"Command"`
		BundleID string      `plist:"ApplicationIdentifier"`
	}
)

type (
	InstallationProxyBasicResponse struct {
		Status string `plist:"Status"`
	}

	InstallationProxyLookupResponse struct {
		InstallationProxyBasicResponse
		LookupResult interface{} `plist:"LookupResult"`
	}

	InstallationProxyBrowseResponse struct {
		InstallationProxyBasicResponse
		CurrentAmount int           `plist:"CurrentAmount"`
		CurrentIndex  int           `plist:"CurrentIndex"`
		CurrentList   []interface{} `plist:"CurrentList"`
	}

	InstallationProxyInstallResponse struct {
		InstallationProxyBasicResponse
		Error            string `plist:"Error"`
		ErrorDescription string `plist:"ErrorDescription"`
	}
)
