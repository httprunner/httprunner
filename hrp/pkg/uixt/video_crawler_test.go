//go:build localtest

package uixt

import "testing"

func TestVideoCrawler(t *testing.T) {
	device, err := NewAndroidDevice()
	if err != nil {
		t.Fatal(err)
	}
	driver, err := device.NewDriver(nil)
	if err != nil {
		t.Fatal(err)
	}
	configs := &VideoCrawlerConfigs{
		AppPackageName: "com.ss.android.ugc.aweme",
	}
	err = driver.VideoCrawler(configs)
	if err != nil {
		t.Fatal(err)
	}
}
