package pluginInternal

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/httprunner/hrp/internal/json"
	"github.com/httprunner/hrp/plugin/go/proto"
)

// functionGRPCClient runs on the host side, it implements FuncCaller interface
type functionGRPCClient struct {
	client proto.DebugTalkClient
}

func (m *functionGRPCClient) GetNames() ([]string, error) {
	log.Info().Msg("function GetNames called on host side")
	resp, err := m.client.GetNames(context.Background(), &proto.Empty{})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call GetNames() failed")
		return nil, err
	}
	return resp.Names, err
}

func (m *functionGRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	log.Info().Str("funcName", funcName).Interface("funcArgs", funcArgs).Msg("call function via gRPC")

	funcArgBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal Call() funcArgs")
	}
	req := &proto.CallRequest{
		Name: funcName,
		Args: funcArgBytes,
	}

	response, err := m.client.Call(context.Background(), req)
	if err != nil {
		log.Error().Err(err).
			Str("funcName", funcName).Interface("funcArgs", funcArgs).
			Msg("gRPC Call() failed")
		return nil, err
	}

	var resp interface{}
	err = json.Unmarshal(response.Value, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() response")
	}
	return resp, nil
}

// Here is the gRPC server that functionGRPCClient talks to.
type functionGRPCServer struct {
	proto.UnimplementedDebugTalkServer
	Impl FuncCaller
}

func (m *functionGRPCServer) GetNames(ctx context.Context, req *proto.Empty) (*proto.GetNamesResponse, error) {
	log.Info().Interface("req", req).Msg("gRPC GetNames() called on plugin side")
	v, err := m.Impl.GetNames()
	if err != nil {
		log.Error().Err(err).Msg("gRPC GetNames() execution failed")
		return nil, err
	}
	return &proto.GetNamesResponse{Names: v}, err
}

func (m *functionGRPCServer) Call(ctx context.Context, req *proto.CallRequest) (*proto.CallResponse, error) {
	var funcArgs []interface{}
	if err := json.Unmarshal(req.Args, &funcArgs); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal Call() funcArgs")
	}

	log.Info().Interface("req", req).Msg("gRPC Call() called on plugin side")

	v, err := m.Impl.Call(req.Name, funcArgs...)
	if err != nil {
		log.Error().Err(err).Interface("req", req).Msg("gRPC Call() execution failed")
		return nil, err
	}

	value, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal Call() response")
	}
	return &proto.CallResponse{Value: value}, err
}

// HRPPlugin implements hashicorp's plugin.GRPCPlugin.
type GRPCPlugin struct {
	plugin.Plugin
	Impl FuncCaller
}

func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterDebugTalkServer(s, &functionGRPCServer{Impl: p.Impl})
	return nil
}

func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &functionGRPCClient{client: proto.NewDebugTalkClient(c)}, nil
}
