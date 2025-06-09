package ai

import (
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/mcp"
)

// ConvertMCPToolToEinoToolInfo converts an MCP tool to eino ToolInfo
func ConvertMCPToolToEinoToolInfo(mcpTool mcp.Tool, namePrefix string) *schema.ToolInfo {
	// Create eino ToolInfo from MCP tool
	toolName := mcpTool.Name
	if namePrefix != "" {
		toolName = fmt.Sprintf("%s__%s", namePrefix, mcpTool.Name)
	}

	toolInfo := &schema.ToolInfo{
		Name: toolName,
		Desc: mcpTool.Description,
	}

	// Convert input schema
	if mcpTool.InputSchema.Properties != nil {
		params := make(map[string]*schema.ParameterInfo)

		for propName, propValue := range mcpTool.InputSchema.Properties {
			if propMap, ok := propValue.(map[string]interface{}); ok {
				paramInfo := &schema.ParameterInfo{}

				if propType, exists := propMap["type"]; exists {
					if typeStr, ok := propType.(string); ok {
						switch typeStr {
						case "string":
							paramInfo.Type = schema.String
						case "number":
							paramInfo.Type = schema.Number
						case "integer":
							paramInfo.Type = schema.Integer
						case "boolean":
							paramInfo.Type = schema.Boolean
						case "array":
							paramInfo.Type = schema.Array
						case "object":
							paramInfo.Type = schema.Object
						default:
							paramInfo.Type = schema.String // default to string
						}
					}
				}

				if description, exists := propMap["description"]; exists {
					if descStr, ok := description.(string); ok {
						paramInfo.Desc = descStr
					}
				}

				if enum, exists := propMap["enum"]; exists {
					if enumSlice, ok := enum.([]interface{}); ok {
						var enumStrings []string
						for _, enumVal := range enumSlice {
							if enumStr, ok := enumVal.(string); ok {
								enumStrings = append(enumStrings, enumStr)
							}
						}
						paramInfo.Enum = enumStrings
					}
				}

				// Check if this parameter is required
				for _, requiredField := range mcpTool.InputSchema.Required {
					if requiredField == propName {
						paramInfo.Required = true
						break
					}
				}

				params[propName] = paramInfo
			}
		}

		if len(params) > 0 {
			toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(params)
		}
	}

	return toolInfo
}

// ConvertMCPToolsToEinoToolInfos converts multiple MCP tools to eino ToolInfos
func ConvertMCPToolsToEinoToolInfos(mcpTools []mcp.Tool, namePrefix string) []*schema.ToolInfo {
	var einoTools []*schema.ToolInfo
	for _, mcpTool := range mcpTools {
		einoTool := ConvertMCPToolToEinoToolInfo(mcpTool, namePrefix)
		if einoTool != nil {
			einoTools = append(einoTools, einoTool)
		}
	}
	return einoTools
}
