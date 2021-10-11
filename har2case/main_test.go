package har2case

import (
	"log"
	"os"
	"testing"
)

var harPath string

func TestMain(m *testing.M) {
	harPath = "demo.har"

	// run all tests
	code := m.Run()
	defer os.Exit(code)

	// teardown
}

func TestGenJSON(t *testing.T) {
	jsonPath, err := NewHAR(harPath).GenJSON()
	log.Printf("jsonPath: %v, err: %v", jsonPath, err)
}
