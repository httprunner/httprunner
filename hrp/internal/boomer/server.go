package boomer

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	pb "github.com/httprunner/httprunner/hrp/internal/grpc/messager"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func (s *grpcServer) BidirectionalStreamingMessage(srv pb.Message_BidirectionalStreamingMessageServer) error {
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
	wn := &workerNode{messenger: srv, id: req.NodeID, heartbeat: 3}
	s.clients.Store(req.NodeID, wn)
	println(fmt.Sprintf("worker(%v) joined, current worker count: %v", req.NodeID, s.getClientsLength()))
	for {
		nodeID := <-s.disconnectedChannel()
		if nodeID == req.NodeID {
			s.clients.Delete(nodeID)
			println(fmt.Sprintf("worker(%v) quited, current worker count: %v", nodeID, s.getClientsLength()))
			return nil
		}
	}
}

type workerNode struct {
	id                string
	state             int32
	heartbeat         int32
	cpuUsage          float64
	cpuWarningEmitted bool
	memoryUsage       float64
	messenger         pb.Message_BidirectionalStreamingMessageServer
}

func (w *workerNode) getState() int32 {
	return atomic.LoadInt32(&w.state)
}

func (w *workerNode) setState(state int32) {
	atomic.StoreInt32(&w.state, state)
}

type grpcServer struct {
	masterHost string
	masterPort int
	server     *grpc.Server
	clients    *sync.Map

	fromWorker           chan *genericMessage
	toWorker             chan *genericMessage
	disconnectedToWorker chan string
	shutdownChan         chan bool
}

func newServer(masterHost string, masterPort int) (server *grpcServer) {
	log.Info().Msg("Boomer is built with grpc support.")
	server = &grpcServer{
		masterHost:           masterHost,
		masterPort:           masterPort,
		clients:              &sync.Map{},
		fromWorker:           make(chan *genericMessage, 100),
		toWorker:             make(chan *genericMessage, 100),
		disconnectedToWorker: make(chan string, 100),
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
	pb.RegisterMessageServer(serv, s)
	reflection.Register(serv)
	// start grpc server
	go func() {
		err = serv.Serve(lis)
		if err != nil {
			log.Error().Err(err).Msg("failed to serve")
			return
		}
	}()
	return nil
}

func (s *grpcServer) getWorkersByState(state int32) (wns []*workerNode) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*workerNode); ok {
			if workerInfo.getState() == state {
				wns = append(wns, workerInfo)
			}
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
		if workerInfo, ok := value.(*workerNode); ok {
			if workerInfo.getState() != stateQuitting && workerInfo.getState() != stateMissing {
				l++
			}
		}
		return true
	})
	return
}

func (s *grpcServer) close() {
	close(s.shutdownChan)
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*workerNode); ok {
			if workerInfo.messenger == nil {
				return true
			}
			err := workerInfo.messenger.Send(&pb.StreamResponse{Type: "stop", Data: nil, NodeID: workerInfo.id})
			if err != nil {
				log.Error().Err(err).Msg("failed to serve")
			}
		}
		return true
	})
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
				if workerInfo, ok := value.(*workerNode); ok {
					if workerInfo.getState() == stateQuitting || workerInfo.getState() == stateMissing {
						return true
					}
					msg, err := workerInfo.messenger.Recv()
					switch err {
					case nil:
						if msg == nil {
							return true
						}
						s.fromWorker <- &genericMessage{
							Type:   msg.Type,
							Data:   msg.Data,
							NodeID: msg.NodeID,
						}
						log.Info().
							Str("nodeID", msg.NodeID).
							Str("type", msg.Type).
							Interface("data", msg.Data).
							Msg("receive data from worker")
					case io.EOF:
						s.fromWorker <- &genericMessage{
							Type:   "quit",
							NodeID: workerInfo.id,
						}
					default:
						if err.Error() == status.Error(codes.Canceled, context.Canceled.Error()).Error() {
							s.fromWorker <- &genericMessage{
								Type:   "quit",
								NodeID: workerInfo.id,
							}
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
				s.disconnectedToWorker <- msg.NodeID
			}
		}
	}
}

func (s *grpcServer) sendMessage(msg *genericMessage) {
	s.clients.Range(func(key, value interface{}) bool {
		if workerInfo, ok := value.(*workerNode); ok {
			if workerInfo.getState() == stateQuitting || workerInfo.getState() == stateMissing {
				return true
			}
			err := workerInfo.messenger.Send(&pb.StreamResponse{Type: msg.Type, Data: msg.Data, NodeID: workerInfo.id})
			switch err {
			case nil:
				break
			case io.EOF:
				return true
			default:
				if err.Error() == status.Error(codes.Canceled, context.Canceled.Error()).Error() {
					s.fromWorker <- &genericMessage{
						Type:   "quit",
						NodeID: workerInfo.id,
					}
					return true
				}
				log.Error().Err(err).Msg("failed to send message")
				return true
			}
			log.Info().
				Str("nodeID", workerInfo.id).
				Str("type", msg.Type).
				Interface("data", msg.Data).
				Int32("state", workerInfo.getState()).
				Msg("send data to worker")
		}
		return true
	})
}

func (s *grpcServer) disconnectedChannel() chan string {
	return s.disconnectedToWorker
}
