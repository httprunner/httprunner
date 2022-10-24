package gidevice

import (
	"fmt"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ InstallationProxy = (*installationProxy)(nil)

func newInstallationProxy(client *libimobiledevice.InstallationProxyClient) *installationProxy {
	return &installationProxy{
		client: client,
	}
}

type installationProxy struct {
	client *libimobiledevice.InstallationProxyClient
}

func (p *installationProxy) Browse(opts ...InstallationProxyOption) (currentList []interface{}, err error) {
	opt := new(installationProxyOption)
	if len(opts) == 0 {
		opt = nil
	} else {
		for _, optFunc := range opts {
			optFunc(opt)
		}
	}

	var pkt libimobiledevice.Packet
	if pkt, err = p.client.NewXmlPacket(
		p.client.NewBasicRequest(libimobiledevice.CommandTypeBrowse, opt),
	); err != nil {
		return nil, err
	}

	if err = p.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = p.client.ReceivePacket(); err != nil {
		return nil, err
	}

	var reply libimobiledevice.InstallationProxyBrowseResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, err
	}

	for reply.Status != "Complete" {
		if respPkt, err = p.client.ReceivePacket(); err != nil {
			return nil, err
		}
		if err = respPkt.Unmarshal(&reply); err != nil {
			return nil, err
		}
	}

	currentList = reply.CurrentList
	return
}

func (p *installationProxy) Lookup(opts ...InstallationProxyOption) (lookupResult interface{}, err error) {
	opt := new(installationProxyOption)
	if len(opts) == 0 {
		opt = nil
	} else {
		for _, optFunc := range opts {
			optFunc(opt)
		}
	}

	var pkt libimobiledevice.Packet
	if pkt, err = p.client.NewXmlPacket(
		p.client.NewBasicRequest(libimobiledevice.CommandTypeLookup, opt),
	); err != nil {
		return nil, err
	}

	if err = p.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = p.client.ReceivePacket(); err != nil {
		return nil, err
	}

	var reply libimobiledevice.InstallationProxyLookupResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, err
	}
	if reply.Status != "Complete" {
		return nil, fmt.Errorf("installation proxy 'Lookup' status: %s", reply.Status)
	}

	lookupResult = reply.LookupResult

	return
}

func (p *installationProxy) Install(bundleID, packagePath string) (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = p.client.NewXmlPacket(
		p.client.NewInstallRequest(bundleID, packagePath),
	); err != nil {
		return err
	}

	if err = p.client.SendPacket(pkt); err != nil {
		return err
	}

	var reply libimobiledevice.InstallationProxyInstallResponse
	for len(reply.Error) == 0 {
		var respPkt libimobiledevice.Packet
		if respPkt, err = p.client.ReceivePacket(); err != nil {
			return err
		}
		if err = respPkt.Unmarshal(&reply); err != nil {
			return err
		}
		if reply.Status == "Complete" {
			break
		}
	}

	if len(reply.Error) != 0 {
		return fmt.Errorf("installation proxy 'Install' status: %s (err: %s, desc: %s)", reply.Status, reply.Error, reply.ErrorDescription)
	}

	return
}

func (p *installationProxy) Uninstall(bundleID string) (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = p.client.NewXmlPacket(
		p.client.NewUninstallRequest(bundleID),
	); err != nil {
		return err
	}

	if err = p.client.SendPacket(pkt); err != nil {
		return err
	}

	var reply libimobiledevice.InstallationProxyInstallResponse
	for len(reply.Error) == 0 {
		var respPkt libimobiledevice.Packet
		if respPkt, err = p.client.ReceivePacket(); err != nil {
			return err
		}
		if err = respPkt.Unmarshal(&reply); err != nil {
			return err
		}
		if reply.Status == "Complete" {
			break
		}
	}

	if len(reply.Error) != 0 {
		return fmt.Errorf("installation proxy 'Uninstall' status: %s (err: %s, desc: %s)", reply.Status, reply.Error, reply.ErrorDescription)
	}
	return
}
