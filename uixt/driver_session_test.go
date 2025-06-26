package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverSession_buildURL(t *testing.T) {
	tests := []struct {
		name    string
		urlStr  string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty url",
			urlStr:  "",
			wantErr: true,
			errMsg:  "URL cannot be empty",
		},
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
		{
			name:    "invalid absolute url",
			urlStr:  "http://[invalid-url",
			wantErr: true,
			errMsg:  "failed to parse URL",
		},
		{
			name:   "relative path",
			urlStr: "/api/users",
			want:   "/api/users",
		},
		{
			name:   "relative path with query params",
			urlStr: "/api/users?id=1&name=test",
			want:   "/api/users?id=1&name=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDriverSession()
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

func TestDriverSession_WithBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
	}{
		{
			name:    "base url with root path",
			baseURL: "http://localhost:8080",
			path:    "/",
			want:    "http://localhost:8080/",
		},
		{
			name:    "base url with api path",
			baseURL: "http://localhost:8080",
			path:    "/api/users",
			want:    "http://localhost:8080/api/users",
		},
		{
			name:    "base url with path and query params",
			baseURL: "http://localhost:8080",
			path:    "/api/users?id=1&name=test",
			want:    "http://localhost:8080/api/users?id=1&name=test",
		},
		{
			name:    "base url with trailing slash and path",
			baseURL: "http://localhost:8080/",
			path:    "api/users",
			want:    "http://localhost:8080/api/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDriverSession()

			// Test that the URL concatenation works as expected
			fullURL := tt.baseURL + tt.path
			assert.Equal(t, tt.want, fullURL)

			// Test that buildURL handles the resulting full URL correctly
			if fullURL != "" {
				got, err := s.buildURL(fullURL)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDriverSession(t *testing.T) {
	session := NewDriverSession()

	// Test backward compatibility with SetBaseURL (should not error)
	session.SetBaseURL("https://postman-echo.com")

	// Test GET with full URL
	resp, err := session.GET("https://postman-echo.com/get")
	assert.Nil(t, err)
	t.Log(resp)

	// Test GET with full URL and query params
	resp, err = session.GET("https://postman-echo.com/get?a=1&b=2")
	assert.Nil(t, err)
	t.Log(resp)

	// Test GETWithBaseURL
	baseURL := "https://postman-echo.com"
	resp, err = session.GETWithBaseURL(baseURL, "/get")
	assert.Nil(t, err)
	t.Log(resp)

	driverRequests := session.History()
	assert.Equal(t, 3, len(driverRequests))

	session.Reset()
	driverRequests = session.History()
	assert.Equal(t, 0, len(driverRequests))

	// Test POST with base URL and path
	resp, err = session.POSTWithBaseURL(nil, baseURL, "/post")
	assert.Nil(t, err)
	t.Log(resp)

	// Verify one request was made after reset
	driverRequests = session.History()
	assert.Equal(t, 1, len(driverRequests))
}
