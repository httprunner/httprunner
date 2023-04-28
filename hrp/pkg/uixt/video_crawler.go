package uixt

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type VideoStat struct {
	FeedCount int `json:"feed_count"`
	LiveCount int `json:"live_count"`
}

func (s *VideoStat) IsTargetAchieved(target *VideoStat) bool {
	log.Info().
		Interface("current", s).
		Interface("target", target).
		Msg("current video stat")
	if s.FeedCount < target.FeedCount {
		return false
	}
	if s.LiveCount < target.LiveCount {
		return false
	}
	return true
}

type VideoCrawlerConfigs struct {
	AppPackageName string `json:"app_package_name"`

	TargetCount VideoStat `json:"target_count"`
}

var androidActivities = map[string]map[string]string{
	// DY
	"com.ss.android.ugc.aweme": {
		"feed": ".splash.SplashActivity",
		"live": ".live.LivePlayActivity",
	},
	// KS
	"com.smile.gifmaker": {
		"feed": "com.yxcorp.gifshow.HomeActivity",
		"live": "com.kuaishou.live.core.basic.activity.LiveSlideActivity",
	},
}

type LiveCrawler struct {
	driver      *DriverExt
	configs     *VideoCrawlerConfigs // target video count
	currentStat *VideoStat           // current video stat
}

func (l *LiveCrawler) checkLiveVideo(texts OCRTexts) (enterPoint PointF, yes bool) {
	// 预览流入口
	points, err := texts.FindTexts([]string{"点击进入直播间", "直播中"})
	if err == nil {
		return points[0], true
	}

	// TODO: 头像入口

	return PointF{}, false
}

// run live video crawler
func (l *LiveCrawler) Run(driver *DriverExt, enterPoint PointF) error {
	log.Info().Msg("enter live room")
	if err := driver.TapAbsXY(enterPoint.X, enterPoint.Y); err != nil {
		log.Error().Err(err).Msg("tap live video failed")
		return err
	}
	time.Sleep(5 * time.Second)

	for l.currentStat.LiveCount < l.configs.TargetCount.LiveCount {
		// check if entered live room
		if err := l.driver.assertActivity(l.configs.AppPackageName, "live"); err != nil {
			return err
		}

		log.Info().
			Int("count", l.currentStat.LiveCount).
			Int("target", l.configs.TargetCount.LiveCount).
			Msg("current live count")

		// swipe to next live video
		err := l.driver.SwipeUp()
		if err != nil {
			log.Error().Err(err).Msg("swipe up failed")
			return err
		}
		time.Sleep(2 * time.Second)
		l.currentStat.LiveCount++
	}

	log.Info().Msg("live count achieved, exit live room")

	return l.exitLiveRoom()
}

func (l *LiveCrawler) exitLiveRoom() error {
	// FIXME: exit live room
	for i := 0; i < 5; i++ {
		l.driver.SwipeRelative(0, 0.5, 0.5, 0.5)
		time.Sleep(2 * time.Second)

		// check if back to feed page
		if err := l.driver.assertActivity(l.configs.AppPackageName, "feed"); err == nil {
			return nil
		}
	}
	return errors.New("exit live room failed")
}

func (dExt *DriverExt) VideoCrawler(configs *VideoCrawlerConfigs) (err error) {
	// launch app
	if err = dExt.Driver.AppLaunch(configs.AppPackageName); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

	currVideoStat := &VideoStat{}
	liveCrawler := LiveCrawler{
		driver:      dExt,
		configs:     configs,
		currentStat: currVideoStat,
	}

	// loop until target count achieved
	// the main loop is feed crawler
	for {
		// check if feed page
		if err := dExt.assertActivity(configs.AppPackageName, "feed"); err != nil {
			return err
		}

		// take screenshot and get screen texts by OCR
		texts, err := dExt.GetScreenTextsByOCR()
		if err != nil {
			log.Error().Err(err).Msg("OCR GetTexts failed")
			return err
		}

		// automatic handling of pop-up windows
		if err := dExt.autoPopupHandler(texts); err != nil {
			log.Error().Err(err).Msg("auto handle popup failed")
			return err
		}

		// check if live video && run live crawler
		if enterPoint, isLive := liveCrawler.checkLiveVideo(texts); isLive {
			log.Info().Msg("live video found")
			if liveCrawler.currentStat.LiveCount < configs.TargetCount.LiveCount {
				if err := liveCrawler.Run(dExt, enterPoint); err != nil {
					return err
				}
			}
		}

		// check if target count achieved
		if currVideoStat.IsTargetAchieved(&configs.TargetCount) {
			log.Info().Msg("target count achieved, exit crawler")
			break
		}

		// swipe to next feed video
		log.Info().Msg("swipe to next feed video")
		if err = dExt.SwipeUp(); err != nil {
			log.Error().Err(err).Msg("swipe up failed")
			return err
		}
		time.Sleep(5 * time.Second)
		currVideoStat.FeedCount++
	}

	return nil
}

func (dExt *DriverExt) assertActivity(pacakgeName, activityType string) error {
	log.Debug().Str("pacakge_name", pacakgeName).
		Str("activity_type", activityType).Msg("assert activity")
	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Error().Err(err).Msg("get foreground app failed")
		return err
	}

	if app.BundleId != pacakgeName {
		return fmt.Errorf("app %s is not in foreground", pacakgeName)
	}

	if activities, ok := androidActivities[app.BundleId]; ok {
		if activity, ok := activities[activityType]; ok {
			if strings.HasSuffix(app.Activity, activity) {
				return nil
			}
		}
	}

	log.Error().Interface("app", app.AppBaseInfo).Msg("app activity not match")
	return fmt.Errorf("%s activity is not in foreground", activityType)
}

func (dExt *DriverExt) autoPopupHandler(texts OCRTexts) error {
	texts.FindTexts([]string{"确定", "取消"})

	// log.Warn().Msg("text popup found")
	return nil
}
