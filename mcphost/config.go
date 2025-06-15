package mcphost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const (
	transportStdio = "stdio"
	transportSSE   = "sse"
)

// MCPConfig represents the configuration for MCP servers
type MCPConfig struct {
	ConfigPath string                         `json:"-"`
	MCPServers map[string]ServerConfigWrapper `json:"mcpServers"`
}

// ServerConfig is an interface for different types of server configurations
type ServerConfig interface {
	GetType() string
	IsDisabled() bool
}

// STDIOServerConfig represents configuration for a STDIO-based server
type STDIOServerConfig struct {
	Command  string            `json:"command"`
	Args     []string          `json:"args"`
	Env      map[string]string `json:"env,omitempty"`
	Disabled bool              `json:"disabled,omitempty"`
}

func (s STDIOServerConfig) GetType() string {
	return transportStdio
}

func (s STDIOServerConfig) IsDisabled() bool {
	return s.Disabled
}

// SSEServerConfig represents configuration for an SSE-based server
type SSEServerConfig struct {
	Url      string   `json:"url"`
	Headers  []string `json:"headers,omitempty"`
	Disabled bool     `json:"disabled,omitempty"`
}

func (s SSEServerConfig) GetType() string {
	return transportSSE
}

func (s SSEServerConfig) IsDisabled() bool {
	return s.Disabled
}

// ServerConfigWrapper is a wrapper for different types of server configurations
type ServerConfigWrapper struct {
	Config ServerConfig
}

func (w *ServerConfigWrapper) UnmarshalJSON(data []byte) error {
	var typeField struct {
		Url string `json:"url"`
	}

	if err := json.Unmarshal(data, &typeField); err != nil {
		return err
	}
	if typeField.Url != "" {
		// If the URL field is present, treat it as an SSE server
		var sse SSEServerConfig
		if err := json.Unmarshal(data, &sse); err != nil {
			return err
		}
		w.Config = sse
	} else {
		// Otherwise, treat it as a STDIOServerConfig
		var stdio STDIOServerConfig
		if err := json.Unmarshal(data, &stdio); err != nil {
			return err
		}
		w.Config = stdio
	}

	return nil
}

func (w ServerConfigWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Config)
}

// LoadMCPConfig loads the MCP configuration from the specified path or default location
func LoadMCPConfig(configPath string) (*MCPConfig, error) {
	log.Debug().Str("configPath", configPath).Msg("Loading MCP config")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".mcp.json")
	}
	configPath = os.ExpandEnv(configPath)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Read existing config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading config file %s: %w",
			configPath,
			err,
		)
	}

	var config MCPConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	config.ConfigPath = configPath
	log.Debug().Str("configPath", configPath).
		Interface("config", config).Msg("Loaded MCP config")
	return &config, nil
}
