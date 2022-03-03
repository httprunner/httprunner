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

func Test_convertJmespath(t *testing.T) {
	exprs := []struct {
		before string
		after  string
	}{
		// normal check expression
		{"a.b.c", "a.b.c"},
		{"headers.\"Content-Type\"", "headers.\"Content-Type\""},
		// check expression using regex
		{"covering (.*) testing,", "covering (.*) testing,"},
		{" (.*) a-b-c", " (.*) a-b-c"},
		// abnormal check expression
		{"-", "\"-\""},
		{"b-c", "\"b-c\""},
		{"a.b-c.d", "a.\"b-c\".d"},
		{"a-b.c-d", "\"a-b\".\"c-d\""},
		{"\"a-b\".c-d", "\"a-b\".\"c-d\""},
		{"headers.Content-Type", "headers.\"Content-Type\""},
		{"body.I-am-a-Key.name", "body.\"I-am-a-Key\".name"},
	}
	for _, expr := range exprs {
		if !assert.Equal(t, convertJmespath(expr.before), expr.after) {
			t.Fail()
		}
	}
}
