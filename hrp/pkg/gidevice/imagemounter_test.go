//go:build localtest

package gidevice

import (
	"encoding/base64"
	"testing"
)

var imageMounterSrv ImageMounter

func setupImageMounterSrv(t *testing.T) {
	setupLockdownSrv(t)

	var err error
	if lockdownSrv, err = dev.lockdownService(); err != nil {
		t.Fatal(err)
	}

	// Once
	// dev.Images()
	if imageMounterSrv, err = lockdownSrv.ImageMounterService(); err != nil {
		t.Fatal(err)
	}
}

func Test_imageMounter_Images(t *testing.T) {
	setupImageMounterSrv(t)

	// imageSignatures, err := dev.Images()
	imageSignatures, err := imageMounterSrv.Images("Developer")
	if err != nil {
		t.Fatal(err)
	}

	for i, imgSign := range imageSignatures {
		t.Logf("%2d, %s", i+1, base64.StdEncoding.EncodeToString(imgSign))
	}
}

func Test_imageMounter_UploadImageAndMount(t *testing.T) {
	setupImageMounterSrv(t)

	devImgPath := "/private/var/mobile/Media/PublicStaging/staging.dimage"
	dmgPath := "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/DeviceSupport/14.4/DeveloperDiskImage.dmg"
	signaturePath := "/Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/DeviceSupport/14.4/DeveloperDiskImage.dmg.signature"

	if err := imageMounterSrv.UploadImageAndMount("Developer", devImgPath, dmgPath, signaturePath); err != nil {
		t.Fatal(err)
	}
}
