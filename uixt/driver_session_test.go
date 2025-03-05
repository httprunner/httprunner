package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverSession_concatURL(t *testing.T) {
	tests := []struct {
		name    string
		baseUrl string
		urlStr  string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty url with empty base url",
			baseUrl: "",
			urlStr:  "",
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "empty url with valid base url",
			baseUrl: "http://localhost:8080",
			urlStr:  "",
			want:    "http://localhost:8080",
		},
		{
			name:    "root path with empty base url",
			baseUrl: "",
			urlStr:  "/",
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "root path with valid base url",
			baseUrl: "http://localhost:8080",
			urlStr:  "/",
			want:    "http://localhost:8080",
		},
		{
			name:    "absolute http url",
			baseUrl: "http://localhost:8080",
			urlStr:  "http://example.com/api",
			want:    "http://example.com/api",
		},
		{
			name:    "absolute https url",
			baseUrl: "http://localhost:8080",
			urlStr:  "https://example.com/api",
			want:    "https://example.com/api",
		},
		{
			name:    "invalid absolute url",
			baseUrl: "http://localhost:8080",
			urlStr:  "http://[invalid-url",
			wantErr: true,
			errMsg:  "failed to parse URL",
		},
		{
			name:    "relative path with empty base url",
			baseUrl: "",
			urlStr:  "api/users",
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "relative path with invalid base url",
			baseUrl: "http://[invalid-url",
			urlStr:  "api/users",
			wantErr: true,
			errMsg:  "failed to parse base URL",
		},
		{
			name:    "relative path with valid base url",
			baseUrl: "http://localhost:8080",
			urlStr:  "api/users",
			want:    "http://localhost:8080/api/users",
		},
		{
			name:    "relative path with query params",
			baseUrl: "http://localhost:8080",
			urlStr:  "api/users?id=1&name=test",
			want:    "http://localhost:8080/api/users?id=1&name=test",
		},
		{
			name:    "base url with query params",
			baseUrl: "http://localhost:8080?token=123",
			urlStr:  "api/users?id=1",
			want:    "http://localhost:8080/api/users?id=1",
		},
		{
			name:    "invalid query params",
			baseUrl: "http://localhost:8080",
			urlStr:  "api/users?id=%invalid",
			wantErr: true,
			errMsg:  "failed to parse query params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DriverSession{baseUrl: tt.baseUrl}
			got, err := s.concatURL(tt.urlStr)

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
	resp, err := session.GET("/get")
	assert.Nil(t, err)
	t.Log(resp)

	resp, err = session.GET("/get?a=1&b=2")
	assert.Nil(t, err)
	t.Log(resp)

	driverRequests := session.History()
	assert.Equal(t, 2, len(driverRequests))

	session.Reset()
	driverRequests = session.History()
	assert.Equal(t, 0, len(driverRequests))

	resp, err = session.GET("https://postman-echo.com/get")
	assert.Nil(t, err)
	t.Log(resp)
}
