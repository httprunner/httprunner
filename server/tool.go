package server

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type ToolRequest struct {
	ServerName string                 `json:"mcp_server"`
	ToolName   string                 `json:"tool_name"`
	Args       map[string]interface{} `json:"args"`
}

func (r *Router) invokeToolHandler(c *gin.Context) {
	if r.mcpHost == nil {
		RenderError(c, errors.New("mcphost not initialized"))
		return
	}

	var req ToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, err)
		return
	}

	// add platform and serial to tool invoke args
	req.Args["platform"] = c.Param("platform")
	req.Args["serial"] = c.Param("serial")

	result, err := r.mcpHost.InvokeTool(c.Request.Context(),
		req.ServerName, req.ToolName, req.Args)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, result)
}
