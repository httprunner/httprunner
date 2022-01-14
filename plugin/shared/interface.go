package shared

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

// FuncCaller is the interface that we're exposing as a plugin.
type FuncCaller interface {
	GetNames() ([]string, error)                                    // get all plugin function names list
	Call(funcName string, args ...interface{}) (interface{}, error) // call plugin function
}

// functionRPC runs on the host side.
type functionRPC struct {
	client *rpc.Client
}

func (g *functionRPC) GetNames() ([]string, error) {
	var resp []string
	err := g.client.Call("Plugin.GetNames", new(interface{}), &resp)
	if err != nil {
		log.Error().Err(err).Msg("rpc call GetNames() failed")
		return nil, err
	}
	return resp, nil
}

// host -> plugin
func (g *functionRPC) Call(funcName string, funcArgs ...interface{}) (interface{}, error) {
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
			Msg("rpc call Call() failed")
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
	log.Info().Interface("args", args).Msg("GetNames called on plugin side")
	var err error
	*resp, err = s.Impl.GetNames()
	if err != nil {
		log.Error().Err(err).Msg("GetNames execution failed")
		return err
	}
	return nil
}

// plugin execution
func (s *functionRPCServer) Call(args interface{}, resp *interface{}) error {
	log.Info().Interface("args", args).Msg("function called on plugin side")
	f := args.(*funcData)
	var err error
	*resp, err = s.Impl.Call(f.Name, f.Args...)
	if err != nil {
		log.Error().Err(err).Interface("args", args).Msg("function execution failed")
		return err
	}
	return nil
}

// HashicorpPlugin implements hashicorp's plugin.Plugin.
type HashicorpPlugin struct {
	Impl FuncCaller
}

func (p *HashicorpPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &functionRPCServer{Impl: p.Impl}, nil
}

func (HashicorpPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &functionRPC{client: c}, nil
}
