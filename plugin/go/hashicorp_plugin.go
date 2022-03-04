package pluginInternal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var client *plugin.Client

// HashicorpPlugin implements hashicorp/go-plugin
type HashicorpPlugin struct {
	logOn bool // turn on plugin log
	FuncCaller
	cachedFunctions map[string]bool // cache loaded functions to improve performance
}

func (p *HashicorpPlugin) Init(path string) error {
	var pluginName string
	if IsRPCPluginType() {
		pluginName = RPCPluginName
	} else {
		pluginName = GRPCPluginName
	}

	// logger
	loggerOptions := &hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
	}
	if p.logOn {
		loggerOptions.Level = hclog.Debug
	} else {
		loggerOptions.Level = hclog.Info
	}

	// cmd
	cmd := exec.Command(path)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", hrpPluginTypeEnvName, hrpPluginType))

	// launch the plugin process
	client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			RPCPluginName:  &RPCPlugin{},
			GRPCPluginName: &GRPCPlugin{},
		},
		Cmd:    cmd,
		Logger: hclog.New(loggerOptions),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
			plugin.ProtocolGRPC,
		},
	})

	// Connect via RPC/gRPC
	rpcClient, err := client.Client()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("connect %s plugin failed", hrpPluginType))
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("request %s plugin failed", hrpPluginType))
	}

	// We should have a Function now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	p.FuncCaller = raw.(FuncCaller)

	p.cachedFunctions = make(map[string]bool)
	log.Info().Str("path", path).Msg("load hashicorp go plugin success")
	return nil
}

func (p *HashicorpPlugin) Has(funcName string) bool {
	flag, ok := p.cachedFunctions[funcName]
	if ok {
		return flag
	}

	funcNames, err := p.GetNames()
	if err != nil {
		return false
	}

	for _, name := range funcNames {
		if name == funcName {
			p.cachedFunctions[funcName] = true // cache as exists
			return true
		}
	}

	p.cachedFunctions[funcName] = false // cache as not exists
	return false
}

func (p *HashicorpPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	return p.FuncCaller.Call(funcName, args...)
}

func (p *HashicorpPlugin) Quit() error {
	// kill hashicorp plugin process
	log.Info().Msg("quit hashicorp plugin process")
	client.Kill()
	return nil
}
