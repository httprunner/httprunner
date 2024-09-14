//go:build localtest

package uixt

import (
	"testing"
)

func TestGetScreenShot(t *testing.T) {
	setupAndroidAdbDriver(t)

	fileName := "test_screenshot"
	_, path, err := driverExt.GetScreenShot(fileName)
	if err != nil {
		t.Fatalf("GetScreenShot failed: %v", err)
	}

	if path == "" {
		t.Fatal("screenshot path is empty")
	}

	t.Logf("screenshot saved at: %s", path)
}
