package hrp

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchRegexp(t *testing.T) {
	testText := `hrp aims to be a one-stop solution for HTTP(S) testing, covering API testing, load testing and digital experience monitoring (DEM).`
	testData := []struct {
		raw      string
		expected string
	}{
		{"covering (.*) testing,", "API"},
		{" (.*) to", "aims"},
		{"^(.*) aims", "hrp"},
		{".* (.*?)$", "(DEM)."},
	}
	// new response object
	resp := http.Response{}
	resp.Body = io.NopCloser(strings.NewReader(testText))
	respObj, err := newResponseObject(t, newParser(), &resp)
	if err != nil {
		t.Fail()
	}
	for _, data := range testData {
		if !assert.Equal(t, data.expected, respObj.searchRegexp(data.raw)) {
			t.Fail()
		}
	}
}
