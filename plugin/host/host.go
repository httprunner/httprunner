package host

import (
	"os"
	"os/exec"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/httprunner/hrp/plugin/shared"
)

var client *plugin.Client

func Init(path string) (shared.FuncCaller, error) {
	// launch the plugin process
	client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			shared.Name: &shared.HashicorpPlugin{},
		},
		Cmd: exec.Command(path),
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   shared.Name,
			Output: os.Stdout,
			Level:  hclog.Info,
		}),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(shared.Name)
	if err != nil {
		return nil, err
	}

	// We should have a Function now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	function := raw.(shared.FuncCaller)
	return function, nil
}

func Quit() {
	client.Kill()
}
