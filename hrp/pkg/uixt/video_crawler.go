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

func (s *VideoStat) isFeedTargetAchieved(target *VideoStat) bool {
	log.Info().
		Int("count", s.FeedCount).
		Int("target", target.FeedCount).
		Msg("current feed count")

	return s.FeedCount >= target.FeedCount
}

func (s *VideoStat) isLiveTargetAchieved(target *VideoStat) bool {
	log.Info().
		Int("count", s.LiveCount).
		Int("target", target.LiveCount).
		Msg("current live count")

	return s.LiveCount >= target.LiveCount
}

func (s *VideoStat) isTargetAchieved(target *VideoStat) bool {
	return s.isFeedTargetAchieved(target) && s.isLiveTargetAchieved(target)
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
	// TODO: SPH, XHS
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

	for !l.currentStat.isLiveTargetAchieved(&l.configs.TargetCount) {
		// check if live room
		if err := l.driver.assertActivity(l.configs.AppPackageName, "live"); err != nil {
			return err
		}

		// swipe to next live video
		err := l.driver.SwipeUp()
		// TODO: sleep custom random time
		time.Sleep(15 * time.Second)
		if err != nil {
			log.Error().Err(err).Msg("swipe up failed")
			// TODO: retry maximum 3 times
			continue
		}

		// TODO: check live type

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

	// exit live room failed, while video count achieved
	if l.currentStat.isTargetAchieved(&l.configs.TargetCount) {
		return nil
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
			if !liveCrawler.currentStat.isLiveTargetAchieved(&configs.TargetCount) {
				if err := liveCrawler.Run(dExt, enterPoint); err != nil {
					log.Error().Err(err).Msg("run live crawler failed, continue")
					continue
				}
			}
		}

		// TODO: check feed type

		currVideoStat.FeedCount++
		// TODO: sleep custom random time
		time.Sleep(5 * time.Second)

		// check if target count achieved
		if currVideoStat.isTargetAchieved(&configs.TargetCount) {
			log.Info().Msg("target count achieved, exit crawler")
			break
		}

		// swipe to next feed video
		log.Info().Msg("swipe to next feed video")
		if err = dExt.SwipeUp(); err != nil {
			log.Error().Err(err).Msg("swipe up failed")
			return err
		}
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

// TODO: add more popup texts
var popups = [][]string{
	{"青少年", "我知道了"}, // 青少年弹窗
	{"允许", "拒绝"},
	{"确定", "取消"},
}

func (dExt *DriverExt) autoPopupHandler(texts OCRTexts) error {
	for _, popup := range popups {
		if len(popup) != 2 {
			continue
		}

		points, err := texts.FindTexts([]string{"确定", "取消"})
		if err == nil {
			log.Warn().Msg("text popup found")
			if err := dExt.TapAbsXY(points[1].X, points[1].Y); err != nil {
				log.Error().Err(err).Msg("tap popup failed")
				return err
			}
			// tap popup success
			return nil
		}
	}

	// no popup found
	return nil
}
