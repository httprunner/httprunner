//go:build localtest

package uixt

import (
	"path/filepath"
	"testing"

	"github.com/httprunner/httprunner/v5/internal/config"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func TestGetScreenShot(t *testing.T) {
	setupAndroidAdbDriver(t)

	imagePath := filepath.Join(config.ScreenShotsPath, "test_screenshot")
	_, err := driverExt.ScreenShot(option.WithScreenShotFileName(imagePath))
	if err != nil {
		t.Fatalf("GetScreenShot failed: %v", err)
	}

	t.Logf("screenshot saved at: %s", imagePath)
}
