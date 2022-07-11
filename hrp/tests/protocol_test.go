package tests

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestHTTPProtocol(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with HTTP/1.1 and HTTP/2").
			SetBaseURL("https://postman-echo.com"),
		TestSteps: []hrp.IStep{
			hrp.NewStep("HTTP/1.1 get").
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/1.1", "check protocol type").
				AssertLengthEqual("body.args.foo1", 4, "check param foo1"),
			hrp.NewStep("HTTP/1.1 post").
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				WithBody(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/1.1", "check protocol type").
				AssertLengthEqual("body.json.foo1", 4, "check body foo1"),
			hrp.NewStep("HTTP/2 get").
				HTTP2().
				GET("/get").
				WithParams(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/2.0", "check protocol type").
				AssertLengthEqual("body.args.foo1", 4, "check param foo1"),
			hrp.NewStep("HTTP/2 post").
				HTTP2().
				POST("/post").
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				WithBody(map[string]interface{}{"foo1": "foo1", "foo2": "foo2"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("proto", "HTTP/2.0", "check protocol type").
				AssertLengthEqual("body.json.foo1", 4, "check body foo1"),
		},
	}
	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestWebSocketProtocol(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with WebSocket protocol").
			WithVariables(map[string]interface{}{
				"n":            5,
				"a":            12.3,
				"b":            3.45,
				"file":         "./demo_file_load_ws_message.txt",
				"wsEchoURL":    "ws://echo.websocket.events",
				"wsPostmanURL": "wss://ws.postman-echo.com/raw",
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("open connection").
				WebSocket().
				OpenConnection("$wsEchoURL"). // absolute url specified, disable base url anyway
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 101, "check open status code").
				AssertEqual("headers.Connection", "Upgrade", "check headers"),
			hrp.NewStep("ping pong test").
				WebSocket().
				PingPong("$wsEchoURL").
				WithTimeout(5000),
			hrp.NewStep("read sponsor info").
				WebSocket().
				Read("$wsEchoURL").
				WithTimeout(5000).
				Validate().
				AssertContains("body", "Lob.com", "check sponsor message"),
			hrp.NewStep("write json").
				WebSocket().
				Write("$wsEchoURL").
				WithTextMessage(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}),
			hrp.NewStep("read json").
				WebSocket().
				Read("$wsEchoURL").
				Extract().
				WithJmesPath("body.foo1", "varFoo1").
				Validate().
				AssertLengthEqual("body.foo1", 5, "check json foo1").
				AssertEqual("body.foo2", 12.3, "check json foo2"),
			hrp.NewStep("write and read text").
				WebSocket().
				WriteAndRead("$wsEchoURL").
				WithTextMessage("$varFoo1").
				Validate().
				AssertLengthEqual("body", 5, "check length equal"),
			hrp.NewStep("write and read binary file").
				WebSocket().
				WriteAndRead("$wsEchoURL").
				WithBinaryMessage("${load_ws_message($file)}"),
			hrp.NewStep("write something redundant").
				WebSocket().
				Write("$wsEchoURL").
				WithTextMessage("have a nice day!"),
			hrp.NewStep("write something redundant").
				WebSocket().
				Write("$wsEchoURL").
				WithTextMessage("balabala ..."),
			hrp.NewStep("close connection").
				WebSocket().
				CloseConnection("$wsEchoURL").
				WithTimeout(30000).
				WithCloseStatus(1000).
				Validate().
				AssertEqual("status_code", 1000, "check close status code"),
			hrp.NewStep("[postman-echo] open connection").
				WebSocket().
				OpenConnection("$wsPostmanURL").
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 101, "check open status code").
				AssertEqualFold("headers.Connection", "Upgrade", "check headers").
				AssertEqualFold("headers.Server", "nginx", "check server").
				AssertEqualFold("headers.Upgrade", "websocket", "checkout upgrade"),
			hrp.NewStep("[postman-echo] write json").
				WebSocket().
				Write("$wsPostmanURL").
				WithTextMessage(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}),
			hrp.NewStep("[postman-echo] read json").
				WebSocket().
				Read("$wsPostmanURL").
				Validate().
				AssertLengthEqual("body.foo1", 5, "check json foo1").
				AssertEqual("body.foo2", 12.3, "check json foo2"),
			hrp.NewStep("[postman-echo] write and read text").
				WebSocket().
				WriteAndRead("$wsPostmanURL").
				WithTextMessage("$varFoo1").
				Validate().
				AssertLengthEqual("body", 5, "check length equal"),
			hrp.NewStep("[postman-echo] close connection").
				WebSocket().
				CloseConnection("$wsPostmanURL").
				WithTimeout(30000).
				WithCloseStatus(1000).
				Validate().
				AssertEqual("status_code", 1000, "check close status code"),
		},
	}
	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestWebSocketProtocolUsingRelativeURL(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with WebSocket protocol").
			SetBaseURL("wss://ws.postman-echo.com").
			WithVariables(map[string]interface{}{
				"n":    5,
				"a":    12.3,
				"b":    3.45,
				"file": "./demo_file_load_ws_message.txt",
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("open connection").
				WebSocket().
				OpenConnection("/raw"). // relative url specified ==> wss://ws.postman-echo.com/raw
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 101, "check open status code").
				AssertEqualFold("headers.Connection", "Upgrade", "check headers").
				AssertEqualFold("headers.Server", "nginx", "check server").
				AssertEqualFold("headers.Upgrade", "websocket", "checkout upgrade"),
			hrp.NewStep("write json").
				WebSocket().
				Write("/raw").
				WithTextMessage(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}),
			hrp.NewStep("read json").
				WebSocket().
				Read("/raw").
				Extract().
				WithJmesPath("body.foo1", "varFoo1").
				Validate().
				AssertLengthEqual("body.foo1", 5, "check json foo1").
				AssertEqual("body.foo2", 12.3, "check json foo2"),
			hrp.NewStep("write and read text").
				WebSocket().
				WriteAndRead("/raw").
				WithTextMessage("$varFoo1").
				Validate().
				AssertLengthEqual("body", 5, "check length equal"),
			hrp.NewStep("close connection").
				WebSocket().
				CloseConnection("/raw").
				WithTimeout(30000).
				WithCloseStatus(1000).
				Validate().
				AssertEqual("status_code", 1000, "check close status code"),
		},
	}
	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}

func TestWebSocketProtocolUsingBaseURL(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("run request with WebSocket protocol").
			SetBaseURL("wss://ws.postman-echo.com/raw").
			WithVariables(map[string]interface{}{
				"n":    5,
				"a":    12.3,
				"b":    3.45,
				"file": "./demo_file_load_ws_message.txt",
			}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("open connection").
				WebSocket().
				OpenConnection(). // no url specified, using base url instead
				WithHeaders(map[string]string{"User-Agent": "HttpRunnerPlus"}).
				Validate().
				AssertEqual("status_code", 101, "check open status code").
				AssertEqualFold("headers.Connection", "Upgrade", "check headers").
				AssertEqualFold("headers.Server", "nginx", "check server").
				AssertEqualFold("headers.Upgrade", "websocket", "checkout upgrade"),
			hrp.NewStep("write json").
				WebSocket().
				Write().
				WithTextMessage(map[string]interface{}{"foo1": "${gen_random_string($n)}", "foo2": "${max($a, $b)}"}),
			hrp.NewStep("read json").
				WebSocket().
				Read().
				Extract().
				WithJmesPath("body.foo1", "varFoo1").
				Validate().
				AssertLengthEqual("body.foo1", 5, "check json foo1").
				AssertEqual("body.foo2", 12.3, "check json foo2"),
			hrp.NewStep("write and read text").
				WebSocket().
				WriteAndRead().
				WithTextMessage("$varFoo1").
				Validate().
				AssertLengthEqual("body", 5, "check length equal"),
			hrp.NewStep("close connection").
				WebSocket().
				CloseConnection().
				WithTimeout(30000).
				WithCloseStatus(1000).
				Validate().
				AssertEqual("status_code", 1000, "check close status code"),
		},
	}
	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
