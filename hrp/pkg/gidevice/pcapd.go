package gidevice

import (
	"log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

type pcapdClient struct {
	stop chan struct{}
	c    *libimobiledevice.PcapdClient
}

func newPcapdClient(c *libimobiledevice.PcapdClient) *pcapdClient {
	return &pcapdClient{
		stop: make(chan struct{}),
		c:    c,
	}
}

func (c *pcapdClient) Packet() <-chan []byte {
	packetCh := make(chan []byte, 10)
	go func() {
		for {
			select {
			case <-c.stop:
				return
			default:
				pkt, err := c.c.ReceivePacket()
				if err != nil {
					close(packetCh)
					return
				}
				var payload []byte
				_ = pkt.Unmarshal(&payload)
				raw, err := c.c.GetPacket(payload)
				if err != nil {
					close(packetCh)
					return
				}
				res, err := c.c.CreatePacket(raw)
				if err != nil {
					log.Println("failed to create packet")
					return
				}
				packetCh <- res
			}
		}
	}()
	return packetCh
}

func (c *pcapdClient) Stop() {
	select {
	case <-c.stop:
	default:
		close(c.stop)
	}
	c.c.Close()
}
