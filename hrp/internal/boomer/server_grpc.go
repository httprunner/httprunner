package boomer

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/httprunner/httprunner/v4/hrp/internal/boomer/data"
	"github.com/httprunner/httprunner/v4/hrp/internal/boomer/grpc/messager"
	"github.com/rs/zerolog/log"
)

type WorkerNode struct {
	ID                string  `json:"id"`
	State             int32   `json:"state"`
	Heartbeat         int32   `json:"heartbeat"`
	SpawnCount        int64   `json:"spawn_count"`
	CPUUsage          float64 `json:"cpu_usage"`
	CPUWarningEmitted bool    `json:"cpu_warning_emitted"`
	MemoryUsage       float64 `json:"memory_usage"`
	stream            chan *messager.StreamResponse
	mutex             sync.RWMutex
	disconnectedChan  chan bool
}

func newWorkerNode(id string) *WorkerNode {
	stream := make(chan *messager.StreamResponse, 100)
	return &WorkerNode{State: StateInit, ID: id, Heartbeat: 3, stream: stream, disconnectedChan: make(chan bool)}
}

func (w *WorkerNode) getState() int32 {
	return atomic.LoadInt32(&w.State)
}

func (w *WorkerNode) setState(state int32) {
	atomic.StoreInt32(&w.State, state)
}

func (w *WorkerNode) updateHeartbeat(heartbeat int32) {
	atomic.StoreInt32(&w.Heartbeat, heartbeat)
}

func (w *WorkerNode) getHeartbeat() int32 {
	return atomic.LoadInt32(&w.Heartbeat)
}

func (w *WorkerNode) updateSpawnCount(spawnCount int64) {
	atomic.StoreInt64(&w.SpawnCount, spawnCount)
}

func (w *WorkerNode) getSpawnCount() int64 {
	return atomic.LoadInt64(&w.SpawnCount)
}

func (w *WorkerNode) updateCPUUsage(cpuUsage float64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.CPUUsage = cpuUsage
}

func (w *WorkerNode) getCPUUsage() float64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.CPUUsage
}

func (w *WorkerNode) updateCPUWarningEmitted(cpuWarningEmitted bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.CPUWarningEmitted = cpuWarningEmitted
}

func (w *WorkerNode) getCPUWarningEmitted() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.CPUWarningEmitted
}

func (w *WorkerNode) updateMemoryUsage(memoryUsage float64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.MemoryUsage = memoryUsage
}

func (w *WorkerNode) getMemoryUsage() float64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.MemoryUsage
}

func (w *WorkerNode) setStream(stream chan *messager.StreamResponse) {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	w.stream = stream
}

func (w *WorkerNode) getStream() chan *messager.StreamResponse {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.stream
}

func (w *WorkerNode) getWorkerInfo() WorkerNode {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return WorkerNode{
		ID:                w.ID,
		State:             w.getState(),
		Heartbeat:         w.getHeartbeat(),
		SpawnCount:        w.getSpawnCount(),
		CPUUsage:          w.getCPUUsage(),
		CPUWarningEmitted: w.getCPUWarningEmitted(),
		MemoryUsage:       w.getMemoryUsage(),
	}
}

type grpcServer struct {
	messager.UnimplementedMessageServer
	masterHost string
	masterPort int
	server     *grpc.Server
	secure     bool
	clients    *sync.Map

	fromWorker       chan *genericMessage
	disconnectedChan chan bool
	shutdownChan     chan bool
	wg               *sync.WaitGroup
}

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

func logger(format string, a ...interface{}) {
	log.Info().Msg(fmt.Sprintf(format, a...))
}

// valid validates the authorization.
func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == "httprunner-secret-token"
}

func serverUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// authentication (token verification)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}
	if !valid(md["authorization"]) {
		return nil, errInvalidToken
	}
	m, err := handler(ctx, req)
	if err != nil {
		logger("RPC failed with error %v", err)
	}
	return m, err
}

// serverWrappedStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type serverWrappedStream struct {
	grpc.ServerStream
}

func (w *serverWrappedStream) RecvMsg(m interface{}) error {
	logger("Receive a message (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *serverWrappedStream) SendMsg(m interface{}) error {
	logger("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

func newServerWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &serverWrappedStream{s}
}

func serverStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// authentication (token verification)
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return errMissingMetadata
	}
	if !valid(md["authorization"]) {
		return errInvalidToken
	}

	err := handler(srv, newServerWrappedStream(ss))
	if err != nil {
		logger("RPC failed with error %v", err)
	}
	return err
}

func newServer(masterHost string, masterPort int) (server *grpcServer) {
	log.Info().Msg("Boomer is built with grpc support.")
	server = &grpcServer{
		masterHost:       masterHost,
		masterPort:       masterPort,
		clients:          &sync.Map{},
		fromWorker:       make(chan *genericMessage, 100),
		disconnectedChan: make(chan bool),
		shutdownChan:     make(chan bool),
	}
	return server
}

func (s *grpcServer) start() (err error) {
	addr := fmt.Sprintf("%v:%v", s.masterHost, s.masterPort)
	// Create tls based credential.
	creds, err := credentials.NewServerTLSFromFile(data.Path("x509/server_cert.pem"), data.Path("x509/server_key.pem"))
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to load key pair: %s", err))
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(serverUnaryInterceptor),
		grpc.StreamInterceptor(serverStreamInterceptor),
		// Enable TLS for all incoming connections.
		grpc.Creds(creds),
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error().Err(err).Msg("failed to listen")
		return
	}
	// create gRPC server
	s.server = grpc.NewServer(opts...)
	// register message server
	messager.RegisterMessageServer(s.server, s)
	reflection.Register(s.server)
	// start grpc server
	go func() {
		err = s.server.Serve(lis)
		if err != nil {
			log.Error().Err(err).Msg("failed to serve")
			return
		}
	}()
	return nil
}

func (s *grpcServer) Register(_ context.Context, req *messager.RegisterRequest) (*messager.RegisterResponse, error) {
	// store worker information
	wn := newWorkerNode(req.NodeID)
	s.clients.Store(req.NodeID, wn)
	log.Warn().Str("worker id", req.NodeID).Msg("worker joined")
	return &messager.RegisterResponse{Code: "0", Message: "register successfully"}, nil
}

func (s *grpcServer) SignOut(_ context.Context, req *messager.SignOutRequest) (*messager.SignOutResponse, error) {
	// delete worker information
	s.clients.Delete(req.NodeID)
	log.Warn().Str("worker id", req.NodeID).Msg("worker quited")
	return &messager.SignOutResponse{Code: "0", Message: "sign out successfully"}, nil
}

func (s *grpcServer) valid(token string) (isValid bool) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.ID == token {
				isValid = true
			}
		}
		return true
	})
	return
}

func (s *grpcServer) BidirectionalStreamingMessage(srv messager.Message_BidirectionalStreamingMessageServer) error {
	token, ok := extractToken(srv.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing token header")
	}

	ok = s.valid(token)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	go s.sendMsg(srv, token)
FOR:
	for {
		msg, err := srv.Recv()
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.OK:
				s.fromWorker <- newGenericMessage(msg.Type, msg.Data, msg.NodeID)
				log.Info().
					Str("nodeID", msg.NodeID).
					Str("type", msg.Type).
					Interface("data", msg.Data).
					Msg("receive data from worker")
			case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
				s.fromWorker <- newQuitMessage(token)
				break FOR
			default:
				log.Error().Err(err).Msg("failed to get stream from client")
				break FOR
			}
		}
	}
	// disconnected to worker
	select {
	case <-srv.Context().Done():
		return srv.Context().Err()
	case <-s.disconnectedChan:
	}
	log.Warn().Str("worker id", token).Msg("worker quited")
	return nil
}

func (s *grpcServer) sendMsg(srv messager.Message_BidirectionalStreamingMessageServer, id string) {
	stream := s.getWorkersByID(id).getStream()
	for {
		select {
		case <-srv.Context().Done():
			return
		case res := <-stream:
			if s, ok := status.FromError(srv.Send(res)); ok {
				switch s.Code() {
				case codes.OK:
					log.Info().
						Str("nodeID", res.NodeID).
						Str("type", res.Type).
						Interface("data", res.Data).
						Interface("profile", res.Profile).
						Msg("send data to worker")
				case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
					log.Warn().Msg(fmt.Sprintf("client (%s) terminated connection", id))
					return
				default:
					log.Warn().Msg(fmt.Sprintf("failed to send to client (%s): %v", id, s.Err()))
					return
				}
			}
		}
	}
}

func (s *grpcServer) sendBroadcasts(msg *genericMessage) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.getState() == StateQuitting || workerInfo.getState() == StateMissing {
				return true
			}
			workerInfo.getStream() <- &messager.StreamResponse{
				Type:    msg.Type,
				Profile: msg.Profile,
				Data:    msg.Data,
				NodeID:  workerInfo.ID,
				Tasks:   msg.Tasks,
			}
		}
		return true
	})
}

func (s *grpcServer) stopServer(ctx context.Context) {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		// close listeners to stop accepting new connections,
		// will block on any existing transports
		s.server.GracefulStop()
	}()

	// wait until all pending RPCs are finished
	select {
	case <-ch:
	case <-ctx.Done():
		// took too long, manually close open transports
		// e.g. watch streams
		s.server.Stop()

		// concurrent GracefulStop should be interrupted
		<-ch
	}
}

func (s *grpcServer) close() {
	// close client requests with request timeout
	timeout := 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	s.stopServer(ctx)
	cancel()
}

func (s *grpcServer) recvChannel() chan *genericMessage {
	return s.fromWorker
}

func (s *grpcServer) shutdownChannel() chan bool {
	return s.shutdownChan
}

func (s *grpcServer) disconnectedChannel() chan bool {
	return s.disconnectedChan
}

func (s *grpcServer) getWorkersByState(state int32) (wns []*WorkerNode) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.getState() == state {
				wns = append(wns, workerInfo)
			}
		}
		return true
	})
	return wns
}

func (s *grpcServer) getWorkersByID(id string) (wn *WorkerNode) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.ID == id {
				wn = workerInfo
			}
		}
		return true
	})
	return wn
}

func (s *grpcServer) getWorkersLengthByState(state int32) (l int) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.getState() == state {
				l++
			}
		}
		return true
	})
	return
}

func (s *grpcServer) getAllWorkers() (wns []WorkerNode) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			wns = append(wns, workerInfo.getWorkerInfo())
		}
		return true
	})
	return wns
}

func (s *grpcServer) getClients() *sync.Map {
	return s.clients
}

func (s *grpcServer) getClientsLength() (l int) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.getState() != StateQuitting && workerInfo.getState() != StateMissing {
				l++
			}
		}
		return true
	})
	return
}
