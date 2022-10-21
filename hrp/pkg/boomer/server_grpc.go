package boomer

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/httprunner/httprunner/v4/hrp/pkg/boomer/data"
	"github.com/httprunner/httprunner/v4/hrp/pkg/boomer/grpc/messager"
)

type WorkerNode struct {
	ID                string  `json:"id"`
	IP                string  `json:"ip"`
	OS                string  `json:"os"`
	Arch              string  `json:"arch"`
	State             int32   `json:"state"`
	Heartbeat         int32   `json:"heartbeat"`
	UserCount         int64   `json:"user_count"`
	WorkerCPUUsage    float64 `json:"worker_cpu_usage"`
	CPUUsage          float64 `json:"cpu_usage"`
	CPUWarningEmitted bool    `json:"cpu_warning_emitted"`
	WorkerMemoryUsage float64 `json:"worker_memory_usage"`
	MemoryUsage       float64 `json:"memory_usage"`
	stream            chan *messager.StreamResponse
	mutex             sync.RWMutex
	disconnectedChan  chan bool
}

func newWorkerNode(id, ip, os, arch string) *WorkerNode {
	stream := make(chan *messager.StreamResponse, 100)
	return &WorkerNode{State: StateInit, ID: id, IP: ip, OS: os, Arch: arch, Heartbeat: 3, stream: stream, disconnectedChan: make(chan bool)}
}

func (w *WorkerNode) getState() int32 {
	return atomic.LoadInt32(&w.State)
}

func (w *WorkerNode) setState(state int32) {
	atomic.StoreInt32(&w.State, state)
}

func (w *WorkerNode) isStarting() bool {
	return w.getState() == StateRunning || w.getState() == StateSpawning
}

func (w *WorkerNode) isStopping() bool {
	return w.getState() == StateStopping
}

func (w *WorkerNode) isAvailable() bool {
	state := w.getState()
	return state != StateMissing && state != StateQuitting
}

func (w *WorkerNode) isReady() bool {
	state := w.getState()
	return state == StateInit || state == StateStopped
}

func (w *WorkerNode) updateHeartbeat(heartbeat int32) {
	atomic.StoreInt32(&w.Heartbeat, heartbeat)
}

func (w *WorkerNode) getHeartbeat() int32 {
	return atomic.LoadInt32(&w.Heartbeat)
}

func (w *WorkerNode) updateUserCount(spawnCount int64) {
	atomic.StoreInt64(&w.UserCount, spawnCount)
}

func (w *WorkerNode) getUserCount() int64 {
	return atomic.LoadInt64(&w.UserCount)
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

func (w *WorkerNode) updateWorkerCPUUsage(workerCPUUsage float64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.WorkerCPUUsage = workerCPUUsage
}

func (w *WorkerNode) getWorkerCPUUsage() float64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.WorkerCPUUsage
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

func (w *WorkerNode) updateWorkerMemoryUsage(workerMemoryUsage float64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.WorkerMemoryUsage = workerMemoryUsage
}

func (w *WorkerNode) getWorkerMemoryUsage() float64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.WorkerMemoryUsage
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
	w.mutex.Lock()
	defer w.mutex.Unlock()
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
		IP:                w.IP,
		OS:                w.OS,
		Arch:              w.Arch,
		State:             w.getState(),
		Heartbeat:         w.getHeartbeat(),
		UserCount:         w.getUserCount(),
		WorkerCPUUsage:    w.getWorkerCPUUsage(),
		CPUUsage:          w.getCPUUsage(),
		CPUWarningEmitted: w.getCPUWarningEmitted(),
		WorkerMemoryUsage: w.getWorkerMemoryUsage(),
		MemoryUsage:       w.getMemoryUsage(),
	}
}

type grpcServer struct {
	messager.UnimplementedMessageServer
	masterHost string
	masterPort int
	server     *grpc.Server
	clients    *sync.Map

	fromWorker       chan *genericMessage
	disconnectedChan chan bool
	shutdownChan     chan bool

	wg sync.WaitGroup
}

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

func logger(format string, a ...interface{}) {
	// FIXME: support server-side and client-side logging to files
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
		wg:               sync.WaitGroup{},
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

func (s *grpcServer) Register(ctx context.Context, req *messager.RegisterRequest) (*messager.RegisterResponse, error) {
	// get client ip
	p, _ := peer.FromContext(ctx)
	clientIp := strings.Split(p.Addr.String(), ":")[0]
	// store worker information
	wn := newWorkerNode(req.NodeID, clientIp, req.Os, req.Arch)
	s.clients.Store(req.NodeID, wn)
	log.Warn().Str("worker id", req.NodeID).Msg("worker joined")
	return &messager.RegisterResponse{Code: "0", Message: "register successful"}, nil
}

func (s *grpcServer) SignOut(_ context.Context, req *messager.SignOutRequest) (*messager.SignOutResponse, error) {
	// delete worker information
	s.clients.Delete(req.NodeID)
	log.Warn().Str("worker id", req.NodeID).Msg("worker quited")
	return &messager.SignOutResponse{Code: "0", Message: "sign out successful"}, nil
}

func (s *grpcServer) validClientToken(token string) bool {
	_, ok := s.clients.Load(token)
	return ok
}

func (s *grpcServer) BidirectionalStreamingMessage(srv messager.Message_BidirectionalStreamingMessageServer) error {
	s.wg.Add(1)
	defer s.wg.Done()
	token, ok := extractToken(srv.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing token header")
	}

	ok = s.validClientToken(token)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	go s.sendMsg(srv, token)
FOR:
	for {
		select {
		case <-srv.Context().Done():
			break FOR
		case <-s.disconnectedChannel():
			break FOR
		default:
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
	}

	log.Info().Str("worker id", token).Msg("bidirectional stream closed")
	return nil
}

func (s *grpcServer) sendMsg(srv messager.Message_BidirectionalStreamingMessageServer, id string) {
	stream := s.getWorkersByID(id).getStream()
	for {
		select {
		case <-srv.Context().Done():
			return
		case <-s.disconnectedChannel():
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
			if !workerInfo.isAvailable() {
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

	// disconnecting workers
	close(s.disconnectedChan)

	// waiting to close bidirectional stream
	s.wg.Wait()
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

func (s *grpcServer) getAvailableClientsLength() (l int) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.isAvailable() {
				l++
			}
		}
		return true
	})
	return
}

func (s *grpcServer) getReadyClientsLength() (l int) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.isReady() {
				l++
			}
		}
		return true
	})
	return
}

func (s *grpcServer) getStartingClientsLength() (l int) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.isStarting() {
				l++
			}
		}
		return true
	})
	return
}

func (s *grpcServer) getCurrentUsers() (l int) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.isStarting() {
				l += int(workerInfo.getUserCount())
			}
		}
		return true
	})
	return
}
