package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// MCPSettings represents the main configuration structure
type MCPSettings struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig represents configuration for a single MCP server
type ServerConfig struct {
	TransportType string        `json:"transportType,omitempty"` // "sse" or "stdio"
	AutoApprove   []string      `json:"autoApprove,omitempty"`
	Disabled      bool          `json:"disabled,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`

	// SSE specific config
	URL string `json:"url,omitempty"`

	// Stdio specific config
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

const (
	DefaultMCPTimeoutSeconds = 30
	MinMCPTimeoutSeconds     = 5
)

// GetTimeoutDuration converts timeout seconds to time.Duration
func (c *ServerConfig) GetTimeoutDuration() time.Duration {
	if c.Timeout == 0 {
		return time.Duration(DefaultMCPTimeoutSeconds) * time.Second
	}
	return c.Timeout
}

// LoadSettings loads MCP settings from the config file
func LoadSettings(path string) (*MCPSettings, error) {
	log.Info().Str("path", path).Msg("load MCP settings")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings MCPSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	if err := validateSettings(&settings); err != nil {
		return nil, fmt.Errorf("invalid settings: %w", err)
	}

	return &settings, nil
}

// validateSettings validates the MCP settings
func validateSettings(settings *MCPSettings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	for name, server := range settings.MCPServers {
		if server.Timeout > 0 && server.Timeout < time.Duration(MinMCPTimeoutSeconds)*time.Second {
			return fmt.Errorf("server %s: timeout must be at least %d seconds", name, MinMCPTimeoutSeconds)
		}

		switch server.TransportType {
		case "sse":
			if server.URL == "" {
				return fmt.Errorf("server %s: URL is required for SSE transport", name)
			}
		case "stdio", "":
			if server.Command == "" {
				return fmt.Errorf("server %s: command is required for stdio transport", name)
			}
		default:
			return fmt.Errorf("server %s: unsupported transport type: %s", name, server.TransportType)
		}
	}

	return nil
}
