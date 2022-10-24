package libimobiledevice

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const ImageMounterServiceName = "com.apple.mobile.mobile_image_mounter"

var ErrDeviceLocked = errors.New("device locked")

type CommandType string

const (
	CommandTypeLookupImage  CommandType = "LookupImage"
	CommandTypeReceiveBytes CommandType = "ReceiveBytes"
	CommandTypeMountImage   CommandType = "MountImage"
)

func NewImageMounterClient(innerConn InnerConn) *ImageMounterClient {
	return &ImageMounterClient{
		client: newServicePacketClient(innerConn),
	}
}

type ImageMounterClient struct {
	client *servicePacketClient
}

func (c *ImageMounterClient) NewBasicRequest(cmdType CommandType, imgType string) *ImageMounterBasicRequest {
	return &ImageMounterBasicRequest{
		Command:   cmdType,
		ImageType: imgType,
	}
}

func (c *ImageMounterClient) NewReceiveBytesRequest(imgType string, imgSize uint32, imgSignature []byte) *ImageMounterReceiveBytesRequest {
	return &ImageMounterReceiveBytesRequest{
		ImageMounterBasicRequest: *c.NewBasicRequest(CommandTypeReceiveBytes, imgType),
		ImageSize:                imgSize,
		ImageSignature:           imgSignature,
	}
}

func (c *ImageMounterClient) NewMountImageRequest(imgType, imgPath string, imgSignature []byte) *ImageMounterMountImageRequest {
	return &ImageMounterMountImageRequest{
		ImageMounterBasicRequest: *c.NewBasicRequest(CommandTypeMountImage, imgType),
		ImagePath:                imgPath,
		ImageSignature:           imgSignature,
	}
}

func (c *ImageMounterClient) NewXmlPacket(req interface{}) (Packet, error) {
	return c.client.NewXmlPacket(req)
}

func (c *ImageMounterClient) SendPacket(pkt Packet) (err error) {
	return c.client.SendPacket(pkt)
}

func (c *ImageMounterClient) ReceivePacket() (respPkt Packet, err error) {
	respPkt, err = c.client.ReceivePacket()
	if err != nil {
		if strings.Contains(err.Error(), io.EOF.Error()) {
			return nil, ErrDeviceLocked
		}
	}
	return
}

func (c *ImageMounterClient) SendDmg(data []byte) (err error) {
	debugLog(fmt.Sprintf("--> ...DmgData...\n"))
	return c.client.innerConn.Write(data)
}

type (
	ImageMounterBasicRequest struct {
		Command   CommandType `plist:"Command"`
		ImageType string      `plist:"ImageType"`
	}

	ImageMounterReceiveBytesRequest struct {
		ImageMounterBasicRequest
		ImageSignature []byte `plist:"ImageSignature"`
		ImageSize      uint32 `plist:"ImageSize"`
	}

	ImageMounterMountImageRequest struct {
		ImageMounterBasicRequest
		ImagePath      string `plist:"ImagePath"`
		ImageSignature []byte `plist:"ImageSignature"`
	}
)

type (
	ImageMounterBasicResponse struct {
		LockdownBasicResponse
		Status string `plist:"Status"`
	}

	ImageMounterLookupImageResponse struct {
		ImageMounterBasicResponse
		ImageSignature [][]byte `plist:"ImageSignature"`
	}
)
