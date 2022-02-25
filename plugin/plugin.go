package plugin

import (
	"fmt"
	"os"
	"reflect"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	pluginInternal "github.com/httprunner/hrp/plugin/inner"
)

// functionsMap stores plugin functions
type functionsMap map[string]reflect.Value

// functionPlugin implements the FuncCaller interface
type functionPlugin struct {
	logger    hclog.Logger
	functions functionsMap
}

func (p *functionPlugin) GetNames() ([]string, error) {
	var names []string
	for name := range p.functions {
		names = append(names, name)
	}
	return names, nil
}

func (p *functionPlugin) Call(funcName string, args ...interface{}) (interface{}, error) {
	p.logger.Info("call function", "funcName", funcName, "args", args)

	fn, ok := p.functions[funcName]
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	return pluginInternal.CallFunc(fn, args...)
}

var functions = make(functionsMap)

// Register registers a plugin function.
// Every plugin function must be registered before Serve() is called.
func Register(funcName string, fn interface{}) {
	if _, ok := functions[funcName]; ok {
		return
	}
	functions[funcName] = reflect.ValueOf(fn)
}

// Serve starts a plugin server process.
func Serve() {
	funcPlugin := &functionPlugin{
		logger: hclog.New(&hclog.LoggerOptions{
			Name:   pluginInternal.PluginName,
			Output: os.Stdout,
			Level:  hclog.Info,
		}),
		functions: functions,
	}
	var pluginMap = map[string]plugin.Plugin{
		pluginInternal.PluginName: &pluginInternal.HRPPlugin{Impl: funcPlugin},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginInternal.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
