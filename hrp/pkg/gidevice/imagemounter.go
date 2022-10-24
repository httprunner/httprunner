package gidevice

import (
	"fmt"
	"os"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

var _ ImageMounter = (*imageMounter)(nil)

func newImageMounter(client *libimobiledevice.ImageMounterClient) *imageMounter {
	return &imageMounter{
		client: client,
	}
}

type imageMounter struct {
	client *libimobiledevice.ImageMounterClient
}

func (m *imageMounter) Images(imgType string) (imageSignatures [][]byte, err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = m.client.NewXmlPacket(
		m.client.NewBasicRequest(libimobiledevice.CommandTypeLookupImage, imgType),
	); err != nil {
		return nil, err
	}

	if err = m.client.SendPacket(pkt); err != nil {
		return nil, err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = m.client.ReceivePacket(); err != nil {
		return nil, err
	}

	var reply libimobiledevice.ImageMounterLookupImageResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return nil, err
	}

	imageSignatures = reply.ImageSignature
	return
}

func (m *imageMounter) UploadImage(imgType, dmgPath string, signatureData []byte) (err error) {
	var dmgFileInfo os.FileInfo
	if dmgFileInfo, err = os.Stat(dmgPath); err != nil {
		return err
	}

	var pkt libimobiledevice.Packet
	if pkt, err = m.client.NewXmlPacket(
		m.client.NewReceiveBytesRequest(imgType, uint32(dmgFileInfo.Size()), signatureData),
	); err != nil {
		return err
	}

	if err = m.client.SendPacket(pkt); err != nil {
		return err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = m.client.ReceivePacket(); err != nil {
		return err
	}

	var reply libimobiledevice.ImageMounterBasicResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return err
	}

	if reply.Status != "ReceiveBytesAck" {
		return fmt.Errorf("image mounter 'ReceiveBytes' status: %s", reply.Status)
	}

	var dmgData []byte
	if dmgData, err = os.ReadFile(dmgPath); err != nil {
		return err
	}

	if err = m.client.SendDmg(dmgData); err != nil {
		return err
	}
	if respPkt, err = m.client.ReceivePacket(); err != nil {
		return err
	}

	if err = respPkt.Unmarshal(&reply); err != nil {
		return err
	}

	if reply.Status != "Complete" {
		return fmt.Errorf("image mounter 'SendDmg' status: %s", reply.Status)
	}

	return
}

func (m *imageMounter) Mount(imgType, devImgPath string, signatureData []byte) (err error) {
	var pkt libimobiledevice.Packet
	if pkt, err = m.client.NewXmlPacket(
		m.client.NewMountImageRequest(imgType, devImgPath, signatureData),
	); err != nil {
		return err
	}

	if err = m.client.SendPacket(pkt); err != nil {
		return err
	}

	var respPkt libimobiledevice.Packet
	if respPkt, err = m.client.ReceivePacket(); err != nil {
		return err
	}

	var reply libimobiledevice.ImageMounterBasicResponse
	if err = respPkt.Unmarshal(&reply); err != nil {
		return err
	}

	if reply.Status != "Complete" {
		return fmt.Errorf("image mounter 'MountImage' status: %s", reply.Status)
	}
	return
}

func (m *imageMounter) UploadImageAndMount(imgType, devImgPath, dmgPath, signaturePath string) (err error) {
	var signatureData []byte
	if signatureData, err = os.ReadFile(signaturePath); err != nil {
		return err
	}
	if err = m.UploadImage(imgType, dmgPath, signatureData); err != nil {
		return err
	}
	if err = m.Mount(imgType, devImgPath, signatureData); err != nil {
		return err
	}
	return
}
