package boomer

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/httprunner/httprunner/v4/hrp/internal/grpc/messager"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type grpcClient struct {
	masterHost string
	masterPort int
	identity   string // nodeID

	config *grpcClientConfig

	fromMaster             chan *genericMessage
	toMaster               chan *genericMessage
	disconnectedFromMaster chan bool
	shutdownChan           chan bool

	failCount int32

	wg sync.WaitGroup
}

type grpcClientConfig struct {
	ctx      context.Context
	cancel   context.CancelFunc // use cancel() to stop client
	conn     *grpc.ClientConn
	biStream messager.Message_BidirectionalStreamingMessageClient

	mutex sync.RWMutex
}

func (c *grpcClientConfig) getBiStreamClient() messager.Message_BidirectionalStreamingMessageClient {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.biStream
}

func (c *grpcClientConfig) setBiStreamClient(s messager.Message_BidirectionalStreamingMessageClient) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.biStream = s
}

func newClient(masterHost string, masterPort int, identity string) (client *grpcClient) {
	log.Info().Msg("Boomer is built with grpc support.")
	// Initiate the stream with a context that supports cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	client = &grpcClient{
		masterHost:             masterHost,
		masterPort:             masterPort,
		identity:               identity,
		fromMaster:             make(chan *genericMessage, 100),
		toMaster:               make(chan *genericMessage, 100),
		disconnectedFromMaster: make(chan bool),
		shutdownChan:           make(chan bool),
		config: &grpcClientConfig{
			ctx:    ctx,
			cancel: cancel,
			mutex:  sync.RWMutex{},
		},
	}
	return client
}

func (c *grpcClient) connect() (err error) {
	addr := fmt.Sprintf("%v:%v", c.masterHost, c.masterPort)
	c.config.conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Error().Err(err).Msg("failed to connect")
		return err
	}

	go c.recv()
	go c.send()

	biStream, err := messager.NewMessageClient(c.config.conn).BidirectionalStreamingMessage(c.config.ctx)
	if err != nil {
		log.Error().Err(err).Msg("call bidirectional streaming message err")
		return err
	}
	c.config.setBiStreamClient(biStream)
	log.Info().Msg(fmt.Sprintf("Boomer is connected to master(%s) press Ctrl+c to quit.\n", addr))

	return nil
}

func (c *grpcClient) reConnect() (err error) {
	biStream, err := messager.NewMessageClient(c.config.conn).BidirectionalStreamingMessage(c.config.ctx)
	if err != nil {
		return
	}
	c.config.setBiStreamClient(biStream)

	// register worker information to master
	c.sendChannel() <- newGenericMessage("register", nil, c.identity)
	//// tell master, I'm ready
	//log.Info().Msg("send client ready signal")
	//c.sendChannel() <- newClientReadyMessageToMaster(c.identity)
	log.Info().Msg(fmt.Sprintf("Boomer is reConnected to master press Ctrl+c to quit.\n"))
	return
}

func (c *grpcClient) close() {
	close(c.shutdownChan)
	c.config.cancel()
	if c.config.conn != nil {
		c.config.conn.Close()
	}
}

func (c *grpcClient) recvChannel() chan *genericMessage {
	return c.fromMaster
}

func (c *grpcClient) recv() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case <-c.shutdownChan:
			return
		default:
			if c.config.getBiStreamClient() == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			msg, err := c.config.getBiStreamClient().Recv()
			if err != nil {
				time.Sleep(1 * time.Second)
				//log.Error().Err(err).Msg("failed to get message")
				continue
			}
			if msg == nil {
				continue
			}

			if msg.NodeID != c.identity {
				log.Warn().
					Str("nodeID", msg.NodeID).
					Str("type", msg.Type).
					Interface("data", msg.Data).
					Msg(fmt.Sprintf("not for me(%s)", c.identity))
				continue
			}

			c.fromMaster <- &genericMessage{
				Type:    msg.Type,
				Profile: msg.Profile,
				Data:    msg.Data,
				NodeID:  msg.NodeID,
				Tasks:   msg.Tasks,
			}

			log.Info().
				Str("nodeID", msg.NodeID).
				Str("type", msg.Type).
				Interface("data", msg.Data).
				Interface("tasks", msg.Tasks).
				Msg("receive data from master")
		}
	}
}

func (c *grpcClient) sendChannel() chan *genericMessage {
	return c.toMaster
}

func (c *grpcClient) send() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		select {
		case <-c.shutdownChan:
			return
		case msg := <-c.toMaster:
			c.sendMessage(msg)

			// We may send genericMessage to master.
			switch msg.Type {
			case "quit":
				c.disconnectedFromMaster <- true
			}
		}
	}
}

func (c *grpcClient) sendMessage(msg *genericMessage) {
	log.Info().
		Str("nodeID", msg.NodeID).
		Str("type", msg.Type).
		Interface("data", msg.Data).
		Msg("send data to server")
	if c.config.getBiStreamClient() == nil {
		atomic.AddInt32(&c.failCount, 1)
		return
	}
	err := c.config.getBiStreamClient().Send(&messager.StreamRequest{Type: msg.Type, Data: msg.Data, NodeID: msg.NodeID})
	switch err {
	case nil:
		atomic.StoreInt32(&c.failCount, 0)
		break
	case io.EOF:
		fallthrough
	default:
		//log.Error().Err(err).Interface("genericMessage", *msg).Msg("failed to send message")
		atomic.AddInt32(&c.failCount, 1)
	}
}

func (c *grpcClient) disconnectedChannel() chan bool {
	return c.disconnectedFromMaster
}
