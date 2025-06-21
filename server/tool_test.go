package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeToolHandler(t *testing.T) {
	router := NewRouter()
	router.InitMCPHost("../internal/mcp/testdata/test.mcp.json")

	tests := []struct {
		name       string
		path       string
		toolReq    ToolRequest
		wantStatus int
	}{
		{
			name: "invoke tool",
			path: "/api/v1/tool/invoke",
			toolReq: ToolRequest{
				ServerName: "weather",
				ToolName:   "get_alerts",
				Args:       map[string]interface{}{"state": "CA"},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.toolReq)
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got HttpResponse
			err := json.Unmarshal(w.Body.Bytes(), &got)
			assert.NoError(t, err)
		})
	}
}
