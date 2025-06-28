package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverSession_buildURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		urlStr  string
		want    string
		wantErr bool
		errMsg  string
	}{
		// Error cases
		{
			name:    "empty url without baseUrl",
			urlStr:  "",
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "relative path without baseUrl",
			urlStr:  "/api/users",
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "invalid absolute url",
			urlStr:  "http://[invalid-url",
			wantErr: true,
			errMsg:  "failed to parse URL",
		},

		// Absolute URLs (no baseURL needed)
		{
			name:   "absolute http url",
			urlStr: "http://example.com/api",
			want:   "http://example.com/api",
		},
		{
			name:   "absolute https url",
			urlStr: "https://example.com/api",
			want:   "https://example.com/api",
		},

		// Empty/root path with baseURL
		{
			name:    "empty url with baseUrl",
			baseURL: "http://localhost:8080",
			urlStr:  "",
			want:    "http://localhost:8080",
		},
		{
			name:    "root path with baseUrl",
			baseURL: "http://localhost:8080",
			urlStr:  "/",
			want:    "http://localhost:8080",
		},

		// Relative paths with baseURL
		{
			name:    "relative path with leading slash",
			baseURL: "http://localhost:8080",
			urlStr:  "/api/users",
			want:    "http://localhost:8080/api/users",
		},
		{
			name:    "relative path without leading slash",
			baseURL: "http://localhost:8080/api",
			urlStr:  "users",
			want:    "http://localhost:8080/users",
		},
		{
			name:    "relative path with query params",
			baseURL: "http://localhost:8080",
			urlStr:  "/api/users?id=1&name=test",
			want:    "http://localhost:8080/api/users?id=1&name=test",
		},

		// BaseURL with path scenarios
		{
			name:    "baseUrl with path + relative path",
			baseURL: "http://localhost:8080/api/v1",
			urlStr:  "users",
			want:    "http://localhost:8080/api/users",
		},
		{
			name:    "baseUrl with trailing slash + relative path",
			baseURL: "http://localhost:8080/api/",
			urlStr:  "users",
			want:    "http://localhost:8080/api/users",
		},
		// Reproduction case: base URL with path + absolute path (leading slash)
		{
			name:    "baseUrl with path + absolute path should preserve base path",
			baseURL: "http://forward-to-38153:6790/wd/hub",
			urlStr:  "/session",
			want:    "http://forward-to-38153:6790/wd/hub/session",
		},
		// Additional test cases for comprehensive coverage
		{
			name:    "baseUrl with path + absolute path with query params",
			baseURL: "http://localhost:8080/api/v1",
			urlStr:  "/session?timeout=30",
			want:    "http://localhost:8080/api/v1/session?timeout=30",
		},
		{
			name:    "baseUrl with path + absolute path with query params and fragment",
			baseURL: "http://localhost:8080/wd/hub",
			urlStr:  "/session/123?param=value#fragment",
			want:    "http://localhost:8080/wd/hub/session/123?param=value#fragment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDriverSession()
			if tt.baseURL != "" {
				s.SetBaseURL(tt.baseURL)
			}

			got, err := s.buildURL(tt.urlStr)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDriverSession(t *testing.T) {
	session := NewDriverSession()
	session.SetBaseURL("https://postman-echo.com")

	// Test GET with full URL
	resp, err := session.GET("https://postman-echo.com/get")
	assert.Nil(t, err)
	t.Log(resp)

	// Test GET with full URL and query params
	resp, err = session.GET("https://postman-echo.com/get?a=1&b=2")
	assert.Nil(t, err)
	t.Log(resp)

	// Test GET with relative path (using baseURL)
	resp, err = session.GET("/get")
	assert.Nil(t, err)
	t.Log(resp)

	// Verify request history
	driverRequests := session.History()
	assert.Equal(t, 3, len(driverRequests))

	// Test reset functionality
	session.Reset()
	driverRequests = session.History()
	assert.Equal(t, 0, len(driverRequests))

	// Test POST with relative path
	resp, err = session.POST(nil, "/post")
	assert.Nil(t, err)
	t.Log(resp)

	// Verify one request was made after reset
	driverRequests = session.History()
	assert.Equal(t, 1, len(driverRequests))
}
