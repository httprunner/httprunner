package pluginInternal

import (
	"encoding/gob"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog/log"
)

func init() {
	gob.Register(new(funcData))
}

// funcData is used to transfer between plugin and host via RPC.
type funcData struct {
	Name string        // function name
	Args []interface{} // function arguments
}

// functionRPCClient runs on the host side, it implements FuncCaller interface
type functionRPCClient struct {
	client *rpc.Client
}

func (g *functionRPCClient) GetNames() ([]string, error) {
	var resp []string
	err := g.client.Call("Plugin.GetNames", new(interface{}), &resp)
	if err != nil {
		log.Error().Err(err).Msg("rpc call GetNames() failed")
		return nil, err
	}
	return resp, nil
}

// host -> plugin
func (g *functionRPCClient) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
	log.Info().Str("funcName", funcName).Interface("funcArgs", funcArgs).Msg("call function via RPC")
	f := funcData{
		Name: funcName,
		Args: funcArgs,
	}

	var args interface{} = f
	var resp interface{}
	err := g.client.Call("Plugin.Call", &args, &resp)
	if err != nil {
		log.Error().Err(err).
			Str("funcName", funcName).Interface("funcArgs", funcArgs).
			Msg("rpc Call() failed")
		return nil, err
	}
	return resp, nil
}

// functionRPCServer runs on the plugin side, executing the user custom function.
type functionRPCServer struct {
	Impl FuncCaller
}

// plugin execution
func (s *functionRPCServer) GetNames(args interface{}, resp *[]string) error {
	log.Info().Interface("args", args).Msg("rpc GetNames() called on plugin side")
	var err error
	*resp, err = s.Impl.GetNames()
	if err != nil {
		log.Error().Err(err).Msg("rpc GetNames() execution failed")
		return err
	}
	return nil
}

// plugin execution
func (s *functionRPCServer) Call(args interface{}, resp *interface{}) error {
	log.Info().Interface("args", args).Msg("rpc Call() called on plugin side")
	f := args.(*funcData)
	var err error
	*resp, err = s.Impl.Call(f.Name, f.Args...)
	if err != nil {
		log.Error().Err(err).Interface("args", args).Msg("rpc Call() execution failed")
		return err
	}
	return nil
}

// RPCPlugin implements hashicorp's plugin.Plugin.
type RPCPlugin struct {
	Impl FuncCaller
}

func (p *RPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &functionRPCServer{Impl: p.Impl}, nil
}

func (RPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &functionRPCClient{client: c}, nil
}
