package hrp

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchJmespath(t *testing.T) {
	testText := `{"a": {"b": "foo"}, "c": "bar", "d": {"e": [{"f": "foo"}, {"f": "bar"}]}}`
	testData := []struct {
		raw      string
		expected string
	}{
		{"body.a.b", "foo"},
		{"body.c", "bar"},
		{"body.d.e[0].f", "foo"},
		{"body.d.e[1].f", "bar"},
	}
	resp := http.Response{}
	resp.Body = io.NopCloser(strings.NewReader(testText))
	respObj, err := newHttpResponseObject(t, newParser(), &resp)
	if err != nil {
		t.Fatal()
	}
	for _, data := range testData {
		if !assert.Equal(t, data.expected, respObj.searchJmespath(data.raw)) {
			t.Fatal()
		}
	}
}

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
	respObj, err := newHttpResponseObject(t, newParser(), &resp)
	if err != nil {
		t.Fatal()
	}
	for _, data := range testData {
		if !assert.Equal(t, data.expected, respObj.searchRegexp(data.raw)) {
			t.Fatal()
		}
	}
}
