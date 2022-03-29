package boomer

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/httprunner/httprunner/v4/hrp/internal/grpc/messager"
	"github.com/rs/zerolog/log"
)

func (s *grpcServer) BidirectionalStreamingMessage(srv messager.Message_BidirectionalStreamingMessageServer) error {
	s.wg.Add(1)
	defer s.wg.Done()
	req, err := srv.Recv()
	switch err {
	case nil:
		break
	case io.EOF:
		return nil
	default:
		if err.Error() == status.Error(codes.Canceled, context.Canceled.Error()).Error() {
			return nil
		}
		log.Error().Err(err).Msg("failed to get stream from client")
		return err
	}
	wn := &WorkerNode{messenger: srv, ID: req.NodeID, Heartbeat: 3}
	s.clients.Store(req.NodeID, wn)
	println(fmt.Sprintf("worker(%v) joined, current worker count: %v", req.NodeID, s.getClientsLength()))
	<-s.disconnectedChannel()
	s.clients.Delete(req.NodeID)
	println(fmt.Sprintf("worker(%v) quited, current worker count: %v", req.NodeID, s.getClientsLength()))
	return nil
}

type WorkerNode struct {
	ID                string  `json:"id"`
	State             int32   `json:"state"`
	Heartbeat         int32   `json:"heartbeat"`
	SpawnCount        int64   `json:"spawn_count"`
	CPUUsage          float64 `json:"cpu_usage"`
	CPUWarningEmitted bool    `json:"cpu_warning_emitted"`
	MemoryUsage       float64 `json:"memory_usage"`
	messenger         messager.Message_BidirectionalStreamingMessageServer
	mutex             sync.RWMutex
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
	clients    *sync.Map

	fromWorker           chan *genericMessage
	toWorker             chan *genericMessage
	disconnectedToWorker chan bool
	shutdownChan         chan bool
	wg                   sync.WaitGroup
}

func newServer(masterHost string, masterPort int) (server *grpcServer) {
	log.Info().Msg("Boomer is built with grpc support.")
	server = &grpcServer{
		masterHost:           masterHost,
		masterPort:           masterPort,
		clients:              &sync.Map{},
		fromWorker:           make(chan *genericMessage, 100),
		toWorker:             make(chan *genericMessage, 100),
		disconnectedToWorker: make(chan bool),
		shutdownChan:         make(chan bool),
	}
	return server
}

func (s *grpcServer) start() (err error) {
	addr := fmt.Sprintf("%v:%v", s.masterHost, s.masterPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error().Err(err).Msg("failed to listen")
		return
	}
	// create gRPC server
	serv := grpc.NewServer()
	// register message server
	messager.RegisterMessageServer(serv, s)
	reflection.Register(serv)
	// start grpc server
	go func() {
		err = serv.Serve(lis)
		if err != nil {
			log.Error().Err(err).Msg("failed to serve")
			return
		}
	}()

	go s.recv()
	go s.send()

	return nil
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

func (s *grpcServer) close() {
	close(s.shutdownChan)
}

func (s *grpcServer) recvChannel() chan *genericMessage {
	return s.fromWorker
}

func (s *grpcServer) shutdownChannel() chan bool {
	return s.shutdownChan
}

func (s *grpcServer) recv() {
	for {
		select {
		case <-s.shutdownChan:
			return
		default:
			s.clients.Range(func(key, value interface{}) bool {
				if workerInfo, ok := value.(*WorkerNode); ok {
					if workerInfo.getState() == StateQuitting || workerInfo.getState() == StateMissing {
						return true
					}
					msg, err := workerInfo.messenger.Recv()
					switch err {
					case nil:
						if msg == nil {
							return true
						}
						s.fromWorker <- newGenericMessage(msg.Type, msg.Data, msg.NodeID)
						log.Info().
							Str("nodeID", msg.NodeID).
							Str("type", msg.Type).
							Interface("data", msg.Data).
							Msg("receive data from worker")
					case io.EOF:
						s.fromWorker <- newQuitMessage(workerInfo.ID)
					default:
						if err.Error() == status.Error(codes.Canceled, context.Canceled.Error()).Error() {
							s.fromWorker <- newQuitMessage(workerInfo.ID)
							return true
						}
						log.Error().Err(err).Msg("failed to get stream from client")
					}
				}
				return true
			})
		}
	}
}

func (s *grpcServer) sendChannel() chan *genericMessage {
	return s.toWorker
}

func (s *grpcServer) send() {
	for {
		select {
		case <-s.shutdownChan:
			return
		case msg := <-s.toWorker:
			s.sendMessage(msg)

			// We may send genericMessage to Worker.
			if msg.Type == "quit" {
				close(s.disconnectedToWorker)
			}
		}
	}
}

func (s *grpcServer) sendMessage(msg *genericMessage) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*WorkerNode); ok {
			if workerInfo.getState() == StateQuitting || workerInfo.getState() == StateMissing {
				return true
			}
			err := workerInfo.messenger.Send(
				&messager.StreamResponse{
					Type:   msg.Type,
					Data:   msg.Data,
					NodeID: workerInfo.ID,
					Tasks:  msg.Tasks},
			)
			switch err {
			case nil:
				break
			case io.EOF:
				fallthrough
			default:
				s.fromWorker <- newQuitMessage(workerInfo.ID)
				log.Error().Err(err).Msg("failed to send message")
				return true
			}
			log.Info().
				Str("nodeID", workerInfo.ID).
				Str("type", msg.Type).
				Interface("data", msg.Data).
				Int32("state", workerInfo.getState()).
				Msg("send data to worker")
		}
		return true
	})
}

func (s *grpcServer) disconnectedChannel() chan bool {
	return s.disconnectedToWorker
}
