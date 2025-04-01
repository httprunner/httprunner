package server

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type ToolRequest struct {
	ServerName string                 `json:"server_name"`
	ToolName   string                 `json:"tool_name"`
	Args       map[string]interface{} `json:"args"`
}

func (r *Router) invokeToolHandler(c *gin.Context) {
	if r.mcpHub == nil {
		RenderError(c, errors.New("mcp hub not initialized"))
		return
	}

	var req ToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, err)
		return
	}

	result, err := r.mcpHub.InvokeTool(c.Request.Context(),
		req.ServerName, req.ToolName, req.Args)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, result)
}
