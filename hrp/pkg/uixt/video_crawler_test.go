//go:build localtest

package uixt

import "testing"

func TestVideoCrawler(t *testing.T) {
	setupAndroid(t)

	configs := &VideoCrawlerConfigs{
		AppPackageName: "com.ss.android.ugc.aweme",
		TargetCount: VideoStat{
			FeedCount: 5,
			LiveCount: 3,
		},
	}
	err := driverExt.VideoCrawler(configs)
	checkErr(t, err)
}
