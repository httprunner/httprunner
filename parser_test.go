package httpboomer

import (
	"testing"
)

func TestBuildURL(t *testing.T) {
	var url string
	url = buildURL("https://postman-echo.com", "/get")
	if url != "https://postman-echo.com/get" {
		t.Fatalf("buildURL error, %s != 'https://postman-echo.com/get'", url)
	}

	url = buildURL("https://postman-echo.com/abc/", "/get?a=1&b=2")
	if url != "https://postman-echo.com/get?a=1&b=2" {
		t.Fatalf("buildURL error, %s != 'https://postman-echo.com/get'", url)
	}

	url = buildURL("", "https://postman-echo.com/get")
	if url != "https://postman-echo.com/get" {
		t.Fatalf("buildURL error, %s != 'https://postman-echo.com/get'", url)
	}

	// notice: step request url > config base url
	url = buildURL("https://postman-echo.com", "https://httpbin.org/get")
	if url != "https://httpbin.org/get" {
		t.Fatalf("buildURL error, %s != 'https://httpbin.org/get'", url)
	}
}
