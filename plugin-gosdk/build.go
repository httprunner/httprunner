package plugin

import (
	"fmt"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// Here is a real implementation of funcCaller
type functionPlugin struct {
	logger    hclog.Logger
	functions map[string]func(args ...interface{}) (interface{}, error)
}

func (p *functionPlugin) GetNames() ([]string, error) {
	var names []string
	for name := range p.functions {
		names = append(names, name)
	}
	return names, nil
}

func (p *functionPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	p.logger.Info("Call function", "funcName", funcName, "args", args)

	f, ok := p.functions[funcName]
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	return f(args...)
}

var functions = make(map[string]func(args ...interface{}) (interface{}, error))

func Register(funcName string, fn func(args ...interface{}) (interface{}, error)) {
	functions[funcName] = fn
}

func Serve() {
	funcPlugin := &functionPlugin{
		logger:    logger,
		functions: functions,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		Name: &hashicorpPlugin{Impl: funcPlugin},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
