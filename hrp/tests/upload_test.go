package tests

import (
	"testing"

	"github.com/httprunner/httprunner/v4/hrp"
)

func TestCaseUploadFile(t *testing.T) {
	testcase := &hrp.TestCase{
		Config: hrp.NewConfig("test upload file to httpbin").
			SetBaseURL("https://httpbin.org").
			WithVariables(map[string]interface{}{"upload_file": "test.env"}),
		TestSteps: []hrp.IStep{
			hrp.NewStep("upload file").
				WithVariables(map[string]interface{}{
					"m_encoder": "${multipart_encoder($m_upload)}",
					"m_upload":  map[string]interface{}{"file": "$upload_file"},
				}).
				POST("/post").
				WithHeaders(map[string]string{"Content-Type": "${multipart_content_type($m_encoder)}"}).
				WithBody("$m_encoder").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertStartsWith("body.files.file", "UserName=test", "check uploaded file"),
			hrp.NewStep("upload file with keyword").
				POST("/post").
				WithUpload(map[string]interface{}{"file": "$upload_file"}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertStartsWith("body.files.file", "UserName=test", "check uploaded file"),
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
