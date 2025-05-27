package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/stretchr/testify/assert"
)

func TestTapHandler(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name       string
		path       string
		req        option.UnifiedActionRequest
		wantStatus int
		wantResp   HttpResponse
	}{
		{
			name: "tap abs xy",
			path: fmt.Sprintf("/api/v1/android/%s/ui/tap", "4622ca24"),
			req: option.UnifiedActionRequest{
				X:        &[]float64{500}[0],
				Y:        &[]float64{800}[0],
				Duration: &[]float64{0}[0],
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
			req: option.UnifiedActionRequest{
				X:        &[]float64{0.5}[0],
				Y:        &[]float64{0.6}[0],
				Duration: &[]float64{0}[0],
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
			reqBody, _ := json.Marshal(tt.req)
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
