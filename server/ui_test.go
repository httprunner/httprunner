package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTapHandler(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name       string
		path       string
		tapReq     TapRequest
		wantStatus int
		wantResp   HttpResponse
	}{
		{
			name: "tap abs xy",
			path: fmt.Sprintf("/api/v1/android/%s/ui/tap", "4622ca24"),
			tapReq: TapRequest{
				X:        500,
				Y:        800,
				Duration: 0,
			},
			wantStatus: http.StatusOK,
			wantResp: HttpResponse{
				Code:    0,
				Message: "success",
				Result:  true,
			},
		},
		{
			name: "tap relative xy",
			path: fmt.Sprintf("/api/v1/android/%s/ui/tap", "4622ca24"),
			tapReq: TapRequest{
				X:        0.5,
				Y:        0.6,
				Duration: 0,
			},
			wantStatus: http.StatusOK,
			wantResp: HttpResponse{
				Code:    0,
				Message: "success",
				Result:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.tapReq)
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got HttpResponse
			err := json.Unmarshal(w.Body.Bytes(), &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantResp, got)
		})
	}
}

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
