package gidevice

import "github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"

func newDiagnosticsRelay(client *libimobiledevice.DiagnosticsRelayClient) *diagnostics {
	return &diagnostics{
		client: client,
	}
}

type diagnostics struct {
	client *libimobiledevice.DiagnosticsRelayClient
}

func (d *diagnostics) Reboot() (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = d.client.NewXmlPacket(
		d.client.NewBasicRequest("Restart"),
	); err != nil {
		return
	}
	if err = d.client.SendPacket(pkt); err != nil {
		return err
	}
	return
}

func (d *diagnostics) Shutdown() (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = d.client.NewXmlPacket(
		d.client.NewBasicRequest("Shutdown"),
	); err != nil {
		return
	}
	if err = d.client.SendPacket(pkt); err != nil {
		return err
	}
	return
}
