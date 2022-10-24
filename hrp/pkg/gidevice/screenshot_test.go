//go:build localtest

package gidevice

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

var screenshotSrv Screenshot

func setupScreenshotSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	if screenshotSrv, err = lockdownSrv.ScreenshotService(); err != nil {
		t.Fatal(err)
	}
}

func Test_screenshot_Take(t *testing.T) {
	setupScreenshotSrv(t)

	// raw, err := dev.Screenshot()
	raw, err := screenshotSrv.Take()
	if err != nil {
		t.Fatal(err)
	}
	_ = raw

	img, format, err := image.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	userHomeDir, _ := os.UserHomeDir()
	file, err := os.Create(userHomeDir + "/Desktop/s1." + format)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	switch format {
	case "png":
		err = png.Encode(file, img)
	case "jpeg":
		err = jpeg.Encode(file, img, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Log(file.Name())
}
