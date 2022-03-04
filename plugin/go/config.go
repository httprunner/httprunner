package pluginInternal

import (
	"os"
	"strings"

	"github.com/hashicorp/go-plugin"
)

const PluginName = "debugtalk"
const RPCPluginName = PluginName + "_rpc"
const GRPCPluginName = PluginName + "_grpc"

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HttpRunnerPlus",
	MagicCookieValue: PluginName,
}

const hrpPluginTypeEnvName = "HRP_PLUGIN_TYPE"

var hrpPluginType string

func init() {
	hrpPluginType = strings.ToLower(os.Getenv(hrpPluginTypeEnvName))
	if hrpPluginType == "" {
		hrpPluginType = "grpc" // default
	}
}

func IsRPCPluginType() bool {
	return hrpPluginType == "rpc"
}
