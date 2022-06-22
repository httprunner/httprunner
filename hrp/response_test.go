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
	testText := `
	<ul class="nav navbar-nav navbar-right">
	<li><a href="/order/addToCart" style="color: white"><i class="fa fa-shopping-cart fa-2x"></i><span class="badge">0</span></a></li>
	<li class="dropdown">
	  <a class="dropdown-toggle" data-toggle="dropdown" href="#" style="color: white">
		Leo   <i class="fa fa-cog fa-2x"></i><span class="caret"></span></a>
	  <ul class="dropdown-menu">
		<li><a href="/user/changePassword">Change Password</a></li>
		<li><a href="/user/addAddress">Shipping</a></li>
		<li><a href="/user/addCard">Payment</a></li>
		<li><a href="/order/orderHistory">Order History</a></li>
		<li><a href="/user/signOut">Sign Out</a></li>
	  </ul>
	</li>

	<li>&nbsp;&nbsp;&nbsp;</li>
	<li><a href="/user/signOut" style="color: white"><i class="fa fa-sign-out fa-2x"></i>
	  Sign Out</a></li>
  </ul>
`
	testData := []struct {
		raw      string
		expected string
	}{
		{"/user/signOut\">(.*)</a></li>", "Sign Out"},
		{"<li><a href=\"/user/(.*)\" style", "signOut"},
		{"		(.*)   <i class=\"fa fa-cog fa-2x\"></i>", "Leo"},
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
