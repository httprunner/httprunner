package gidevice

import "github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"

var _ HouseArrest = (*houseArrest)(nil)

func newHouseArrest(client *libimobiledevice.HouseArrestClient) *houseArrest {
	return &houseArrest{
		client: client,
	}
}

type houseArrest struct {
	client *libimobiledevice.HouseArrestClient
}

func (h *houseArrest) Documents(bundleID string) (afc Afc, err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = h.client.NewXmlPacket(
		h.client.NewDocumentsRequest(bundleID),
	); err != nil {
		return nil, err
	}

	if err = h.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	if _, err = h.client.ReceivePacket(); err != nil {
		return nil, err
	}

	afcClient := libimobiledevice.NewAfcClient(h.client.InnerConn())
	afc = newAfc(afcClient)
	return
}

func (h *houseArrest) Container(bundleID string) (afc Afc, err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = h.client.NewXmlPacket(
		h.client.NewContainerRequest(bundleID),
	); err != nil {
		return nil, err
	}

	if err = h.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	if _, err = h.client.ReceivePacket(); err != nil {
		return nil, err
	}

	afcClient := libimobiledevice.NewAfcClient(h.client.InnerConn())
	afc = newAfc(afcClient)
	return
}
