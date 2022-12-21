package gidevice

import (
	"log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice/pkg/libimobiledevice"
)

type PcapOptions struct {
	All      bool   // capture all packets
	Pid      int    // capture packets from specific PID
	ProcName string // capture packets from specific process name
	BundleID string // convert to PID first, then capture packets from the PID
}

type PcapOption func(*PcapOptions)

func WithPcapAll(all bool) PcapOption {
	return func(opt *PcapOptions) {
		opt.All = all
	}
}

func WithPcapProcName(procName string) PcapOption {
	return func(opt *PcapOptions) {
		opt.ProcName = procName
	}
}

func WithPcapPID(pid int) PcapOption {
	return func(opt *PcapOptions) {
		opt.Pid = pid
	}
}

func WithPcapBundleID(bundleID string) PcapOption {
	return func(opt *PcapOptions) {
		opt.BundleID = bundleID
	}
}

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
				if raw == nil {
					// filtered packet
					continue
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
