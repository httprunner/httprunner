package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverSession_concatURL(t *testing.T) {
	tests := []struct {
		name    string
		baseUrl string
		elem    []string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty elements with empty base url",
			baseUrl: "",
			elem:    []string{},
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "empty elements with valid base url",
			baseUrl: "http://localhost:8080",
			elem:    []string{},
			want:    "http://localhost:8080",
		},
		{
			name:    "absolute url in first element",
			baseUrl: "http://localhost:8080",
			elem:    []string{"https://example.com/api", "users"},
			want:    "https://example.com/api/users",
		},
		{
			name:    "invalid absolute url",
			baseUrl: "http://localhost:8080",
			elem:    []string{"http://[invalid-url", "users"},
			wantErr: true,
			errMsg:  "failed to parse URL",
		},
		{
			name:    "relative path with empty base url",
			baseUrl: "",
			elem:    []string{"api", "users"},
			wantErr: true,
			errMsg:  "base URL is empty",
		},
		{
			name:    "relative path with invalid base url",
			baseUrl: "http://[invalid-url",
			elem:    []string{"api", "users"},
			wantErr: true,
			errMsg:  "failed to parse base URL",
		},
		{
			name:    "relative path with query params",
			baseUrl: "http://localhost:8080",
			elem:    []string{"api", "users?id=1&name=test"},
			want:    "http://localhost:8080/api/users?id=1&name=test",
		},
		{
			name:    "base url with query params",
			baseUrl: "http://localhost:8080?token=123",
			elem:    []string{"api", "users?id=1"},
			want:    "http://localhost:8080/api/users?id=1&token=123",
		},
		{
			name:    "invalid query params",
			baseUrl: "http://localhost:8080",
			elem:    []string{"api", "users?id=%invalid"},
			wantErr: true,
			errMsg:  "failed to parse query params",
		},
		{
			name:    "multiple path segments",
			baseUrl: "http://localhost:8080",
			elem:    []string{"api", "v1", "users", "profile"},
			want:    "http://localhost:8080/api/v1/users/profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &DriverSession{baseUrl: tt.baseUrl}
			got, err := s.concatURL(tt.elem...)

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
