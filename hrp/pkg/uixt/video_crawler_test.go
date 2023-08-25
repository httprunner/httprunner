//go:build localtest

package uixt

import (
	"testing"
)

func TestVideoCrawler(t *testing.T) {
	setupAndroid(t)

	driverExt.Driver.AppLaunch("com.ss.android.ugc.aweme")
	configs := &VideoCrawlerConfigs{
		Timeout: 600,
		Feed: FeedConfig{
			TargetCount: 5,
			TargetLabels: []TargetLabel{
				{Text: `^广告$`, Scope: Scope{0, 0.5, 1, 1}, Regex: true},
				{Text: `^图文$`, Scope: Scope{0, 0.5, 1, 1}, Regex: true, Target: 2},
				{Text: `^特效\|`, Scope: Scope{0, 0.5, 1, 1}, Regex: true},
				{Text: `^模板\|`, Scope: Scope{0, 0.5, 1, 1}, Regex: true},
				{Text: `^购物\|`, Scope: Scope{0, 0.5, 1, 1}, Regex: true},
			},
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
