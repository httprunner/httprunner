package boomer

import (
	"context"
	"fmt"
	"io"

	pb "github.com/httprunner/httprunner/hrp/internal/grpc/messager"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type client interface {
	connect() (err error)
	close()
	recvChannel() chan *genericMessage
	sendChannel() chan *genericMessage
	disconnectedChannel() chan bool
}

type grpcClient struct {
	masterHost string
	masterPort int
	identity   string // nodeID

	config *grpcClientConfig

	fromMaster             chan *genericMessage
	toMaster               chan *genericMessage
	disconnectedFromMaster chan bool
	shutdownChan           chan bool
}

type grpcClientConfig struct {
	ctx      context.Context
	cancel   context.CancelFunc // use cancel() to stop client
	conn     *grpc.ClientConn
	client   pb.MessageClient
	biStream pb.Message_BidirectionalStreamingMessageClient
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

	c.config.client = pb.NewMessageClient(c.config.conn)
	c.config.biStream, err = c.config.client.BidirectionalStreamingMessage(c.config.ctx)
	if err != nil {
		log.Error().Err(err).Msg("call bidirectional streaming message err")
		return err
	}

	log.Info().Msg(fmt.Sprintf("Boomer is connected to master(%s) press Ctrl+c to quit.\n", addr))
	return nil
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
	for {
		select {
		case <-c.shutdownChan:
			return
		default:
			msg, err := c.config.biStream.Recv()
			switch err {
			case nil:
				break
			case io.EOF:
				return
			default:
				if err.Error() == status.Error(codes.Canceled, context.Canceled.Error()).Error() {
					return
				}
				log.Error().Err(err).Msg("failed to get stream from server")
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

			c.fromMaster <- &genericMessage{Type: msg.Type, Data: msg.Data, NodeID: msg.NodeID}
			log.Info().
				Str("nodeID", msg.NodeID).
				Str("type", msg.Type).
				Interface("data", msg.Data).
				Msg("receive data from master")
		}
	}
}

func (c *grpcClient) sendChannel() chan *genericMessage {
	return c.toMaster
}

func (c *grpcClient) send() {
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
	err := c.config.biStream.Send(&pb.StreamRequest{Type: msg.Type, Data: msg.Data, NodeID: msg.NodeID})
	switch err {
	case nil:
		break
	case io.EOF:
		return
	default:
		if err.Error() == status.Error(codes.Canceled, context.Canceled.Error()).Error() {
			return
		}
		log.Error().Err(err).Msg("failed to send message")
	}
}

func (c *grpcClient) disconnectedChannel() chan bool {
	return c.disconnectedFromMaster
}
