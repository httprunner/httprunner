package boomer

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"

	"github.com/httprunner/httprunner/v4/hrp/pkg/boomer/data"
	"github.com/httprunner/httprunner/v4/hrp/pkg/boomer/grpc/messager"
)

type grpcClient struct {
	messager.MessageClient
	masterHost string
	masterPort int
	identity   string // nodeID

	config *grpcClientConfig

	fromMaster       chan *genericMessage
	toMaster         chan *genericMessage
	disconnectedChan chan bool
	shutdownChan     chan bool

	failCount int32
}

type grpcClientConfig struct {
	// ctx is used for the lifetime of the stream that may need to be canceled
	// on client shutdown.
	ctx       context.Context
	ctxCancel context.CancelFunc
	conn      *grpc.ClientConn
	biStream  messager.Message_BidirectionalStreamingMessageClient

	mutex sync.RWMutex
}

const token = "httprunner-secret-token"

// unaryInterceptor is an example unary interceptor.
func unaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var credsConfigured bool
	for _, o := range opts {
		_, ok := o.(grpc.PerRPCCredsCallOption)
		if ok {
			credsConfigured = true
			break
		}
	}
	if !credsConfigured {
		opts = append(opts, grpc.PerRPCCredentials(oauth.NewOauthAccess(&oauth2.Token{
			AccessToken: token,
		})))
	}
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	end := time.Now()
	logger("RPC: %s, start time: %s, end time: %s, err: %v", method, start.Format("Basic"), end.Format(time.RFC3339), err)
	return err
}

// wrappedStream  wraps around the embedded grpc.ClientStream, and intercepts the RecvMsg and
// SendMsg method call.
type wrappedStream struct {
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	logger("Receive a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	logger("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

func extractToken(ctx context.Context) (tkn string, ok bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md[token]) == 0 {
		return "", false
	}

	return md[token][0], true
}

// streamInterceptor is an example stream interceptor.
func streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	var credsConfigured bool
	for _, o := range opts {
		_, ok := o.(*grpc.PerRPCCredsCallOption)
		if ok {
			credsConfigured = true
			break
		}
	}
	if !credsConfigured {
		opts = append(opts, grpc.PerRPCCredentials(oauth.NewOauthAccess(&oauth2.Token{
			AccessToken: token,
		})))
	}
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
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
		masterHost:       masterHost,
		masterPort:       masterPort,
		identity:         identity,
		fromMaster:       make(chan *genericMessage, 100),
		toMaster:         make(chan *genericMessage, 100),
		disconnectedChan: make(chan bool),
		shutdownChan:     make(chan bool),
		config: &grpcClientConfig{
			ctx:       ctx,
			ctxCancel: cancel,
			mutex:     sync.RWMutex{},
		},
	}
	return client
}

func (c *grpcClient) start() (err error) {
	addr := fmt.Sprintf("%v:%v", c.masterHost, c.masterPort)
	// Create tls based credential.
	creds, err := credentials.NewClientTLSFromFile(data.Path("x509/ca_cert.pem"), "www.httprunner.com")
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to load credentials: %v", err))
	}
	opts := []grpc.DialOption{
		// oauth.NewOauthAccess requires the configuration of transport
		// credentials.
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32 * 10e9)),
		grpc.WithUnaryInterceptor(unaryInterceptor),
		grpc.WithStreamInterceptor(streamInterceptor),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,
				Multiplier: 1.2,
				MaxDelay:   3 * time.Second,
			},
			MinConnectTimeout: 3 * time.Second,
		}),
	}
	c.config.conn, err = grpc.Dial(addr, opts...)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect")
		return err
	}
	c.MessageClient = messager.NewMessageClient(c.config.conn)
	return nil
}

func (c *grpcClient) register(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	res, err := c.Register(ctx, &messager.RegisterRequest{NodeID: c.identity, Os: runtime.GOOS, Arch: runtime.GOARCH})
	if err != nil {
		return err
	}
	if res.Code != "0" {
		return errors.New(res.Message)
	}
	return nil
}

func (c *grpcClient) signOut(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	res, err := c.SignOut(ctx, &messager.SignOutRequest{NodeID: c.identity})
	if err != nil {
		return err
	}
	if res.Code != "0" {
		return errors.New(res.Message)
	}
	return nil
}

func (c *grpcClient) newBiStreamClient() (err error) {
	md := metadata.New(map[string]string{token: c.identity})
	ctx := metadata.NewOutgoingContext(c.config.ctx, md)
	biStream, err := c.BidirectionalStreamingMessage(ctx)
	if err != nil {
		return err
	}
	// reset failCount
	atomic.StoreInt32(&c.failCount, 0)
	// set bidirectional stream client
	c.config.setBiStreamClient(biStream)
	println("successful to establish bidirectional stream with master, press Ctrl+c to quit.")
	return nil
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
			if c.config.getBiStreamClient() == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			msg, err := c.config.getBiStreamClient().Recv()
			if err != nil {
				time.Sleep(1 * time.Second)
				// log.Error().Err(err).Msg("failed to get message")
				continue
			}
			if msg == nil {
				continue
			}

			if msg.NodeID != c.identity {
				log.Info().
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
	for {
		select {
		case <-c.shutdownChan:
			return
		case msg := <-c.toMaster:
			c.sendMessage(msg)

			// We may send genericMessage to master.
			switch msg.Type {
			case "quit":
				c.disconnectedChan <- true
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
	if err == nil {
		atomic.StoreInt32(&c.failCount, 0)
		return
	}
	// log.Error().Err(err).Interface("genericMessage", *msg).Msg("failed to send message")
	if msg.Type == "heartbeat" {
		atomic.AddInt32(&c.failCount, 1)
	}
}

func (c *grpcClient) disconnectedChannel() chan bool {
	return c.disconnectedChan
}

func (c *grpcClient) close() {
	close(c.shutdownChan)
	c.config.ctxCancel()
	if c.config.conn != nil {
		c.config.conn.Close()
	}
}
