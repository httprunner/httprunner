package uixt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriverSession(t *testing.T) {
	session := NewDriverSession()
	session.SetBaseURL("https://postman-echo.com")
	resp, err := session.GET("/get")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)

	resp, err = session.GET("/get?a=1&b=2")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)

	driverRequests := session.History()
	assert.Equal(t, 2, len(driverRequests))

	session.Reset()
	driverRequests = session.History()
	assert.Equal(t, 0, len(driverRequests))
}
