//go:build localtest

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
			hrp.NewStep("upload file explicitly").
				WithVariables(map[string]interface{}{
					"m_encoder": "${multipart_encoder($m_upload)}",
					"m_upload":  map[string]interface{}{"file": "@$upload_file"},
				}).
				POST("/post").
				WithHeaders(map[string]string{"Content-Type": "${multipart_content_type($m_encoder)}"}).
				WithBody("$m_encoder").
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertStartsWith("body.files.file", "UserName=test", "check uploaded file"),
			hrp.NewStep("upload both text and file").
				POST("/post").
				WithUpload(map[string]interface{}{
					"foo1":  "\"bar1\"",
					"foo2":  "\"@$upload_file\"",
					"foo3":  "\"\"@$upload_file\"\"",
					"file1": "@\"$upload_file\"",
					"file2": "@$upload_file",
				}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.form.foo1", "bar1", "check foo1 in form").
				AssertEqual("body.form.foo2", "@$upload_file", "check foo2 in form").
				AssertEqual("body.form.foo3", "\"@$upload_file\"", "check foo3 in form").
				AssertStartsWith("body.files.file1", "UserName=test", "check uploaded file1").
				AssertStartsWith("body.files.file2", "UserName=test", "check uploaded file2"),
			hrp.NewStep("upload empty field").
				POST("/post").
				WithUpload(map[string]interface{}{
					"foo1":  "",
					"foo2":  "\"\"",
					"foo3":  "\"\";",
					"dummy": ";filename=empty",
				}).
				Validate().
				AssertEqual("status_code", 200, "check status code").
				AssertEqual("body.form.foo1", "", "check foo1 in form").
				AssertEqual("body.form.foo2", "", "check foo2 in form").
				AssertEqual("body.files.dummy", "", "check dummy file in files"),
		},
	}

	err := hrp.NewRunner(t).Run(testcase)
	if err != nil {
		t.Fatalf("run testcase error: %v", err)
	}
}
