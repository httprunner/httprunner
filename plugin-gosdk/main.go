package plugin

import (
	"os"
	"os/exec"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

const Name = "debugtalk"

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HttpRunnerPlus",
	MagicCookieValue: Name,
}

// Create an hclog.Logger
var logger = hclog.New(&hclog.LoggerOptions{
	Name:   Name,
	Output: os.Stdout,
	Level:  hclog.Debug,
})

var client *plugin.Client

func Init(path string) (FuncCaller, error) {
	// launch the plugin process
	client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			Name: &hashicorpPlugin{},
		},
		Cmd:    exec.Command(path),
		Logger: logger,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(Name)
	if err != nil {
		return nil, err
	}

	// We should have a Function now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	function := raw.(FuncCaller)
	return function, nil
}

func Quit() {
	client.Kill()
}
