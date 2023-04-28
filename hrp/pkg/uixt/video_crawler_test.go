//go:build localtest

package uixt

import "testing"

func TestVideoCrawler(t *testing.T) {
	setupAndroid(t)

	configs := &VideoCrawlerConfigs{
		AppPackageName: "com.ss.android.ugc.aweme",
	}
	err := driverExt.VideoCrawler(configs)
	if err != nil {
		t.Fatal(err)
	}
}
