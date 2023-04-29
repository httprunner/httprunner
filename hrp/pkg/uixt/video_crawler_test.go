//go:build localtest

package uixt

import "testing"

func TestVideoCrawler(t *testing.T) {
	setupAndroid(t)

	configs := &VideoCrawlerConfigs{
		AppPackageName: "com.ss.android.ugc.aweme",

		Feed: FeedConfig{
			TargetCount: 5,
			SleepRandom: []interface{}{0, 5, 0.7, 5, 10, 0.3},
		},
		Live: LiveConfig{
			TargetCount: 3,
			SleepRandom: []interface{}{15, 20},
		},
	}
	err := driverExt.VideoCrawler(configs)
	checkErr(t, err)
}
