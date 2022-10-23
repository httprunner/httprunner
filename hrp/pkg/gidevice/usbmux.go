package gidevice

import (
	"context"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ Usbmux = (*usbmux)(nil)

func NewUsbmux() (Usbmux, error) {
	umClient, err := libimobiledevice.NewUsbmuxClient()
	if err != nil {
		return nil, err
	}
	return &usbmux{client: umClient}, nil
}

func newUsbmux(client *libimobiledevice.UsbmuxClient) *usbmux {
	return &usbmux{client: client}
}

type usbmux struct {
	client *libimobiledevice.UsbmuxClient
}

func (um *usbmux) Devices() (devices []Device, err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = um.client.NewPlistPacket(
		um.client.NewBasicRequest(libimobiledevice.MessageTypeDeviceList),
	); err != nil {
		return nil, err
	}

	if err = um.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = um.client.ReceivePacket(); err != nil {
		return nil, err
	}

	reply := struct {
		DeviceList []libimobiledevice.BaseDevice `plist:"DeviceList"`
	}{}
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, err
	}

	devices = make([]Device, len(reply.DeviceList))
	for i := range reply.DeviceList {
		dev := reply.DeviceList[i]
		devices[i] = newDevice(um.client, dev.Properties)
	}

	return
}

func (um *usbmux) ReadBUID() (buid string, err error) {
	var pktReadBUID libimobiledevice.Packet
	if pktReadBUID, err = um.client.NewPlistPacket(
		um.client.NewBasicRequest(libimobiledevice.MessageTypeReadBUID),
	); err != nil {
		return "", err
	}

	if err = um.client.SendPacket(pktReadBUID); err != nil {
		return "", err
	}

	respPkt, err := um.client.ReceivePacket()
	if err != nil {
		return "", err
	}

	reply := struct {
		BUID string `plist:"BUID"`
	}{}
	if err = respPkt.Unmarshal(&reply); err != nil {
		return "", err
	}

	buid = reply.BUID

	return
}

func (um *usbmux) Listen(devNotifier chan Device) (context.CancelFunc, error) {
	baseDevNotifier := make(chan libimobiledevice.BaseDevice)
	ctx, cancelFunc, err := um.listen(baseDevNotifier)
	go func(ctx context.Context) {
		defer close(devNotifier)
		for {
			select {
			case <-ctx.Done():
				return
			case baseDev := <-baseDevNotifier:
				if baseDev.MessageType != libimobiledevice.MessageTypeDeviceAdd {
					baseDev.Properties.DeviceID = baseDev.DeviceID
				}
				client, err := libimobiledevice.NewUsbmuxClient()
				if err != nil {
					continue
				}
				devNotifier <- newDevice(client, baseDev.Properties)
			}
		}
	}(ctx)
	return cancelFunc, err
}

func (um *usbmux) listen(devNotifier chan libimobiledevice.BaseDevice) (ctx context.Context, cancelFunc context.CancelFunc, err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = um.client.NewPlistPacket(
		um.client.NewBasicRequest(libimobiledevice.MessageTypeListen),
	); err != nil {
		return nil, nil, err
	}

	if err = um.client.SendPacket(pkt); err != nil {
		return nil, nil, err
	}

	ctx, cancelFunc = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		defer close(devNotifier)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var respPkt libimobiledevice.Packet
				if respPkt, err = um.client.ReceivePacket(); err != nil {
					break
				}

				var replyDevice libimobiledevice.BaseDevice
				if err = respPkt.Unmarshal(&replyDevice); err != nil {
					break
				}
				if replyDevice.MessageType == libimobiledevice.MessageTypeResult {
					break
				}

				devNotifier <- replyDevice
			}
		}
	}(ctx)

	return ctx, cancelFunc, nil
}
