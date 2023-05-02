package uixt

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type VideoStat struct {
	configs *VideoCrawlerConfigs

	FeedCount int `json:"feed_count"`
	LiveCount int `json:"live_count"`
}

func (s *VideoStat) isFeedTargetAchieved() bool {
	log.Info().
		Int("count", s.FeedCount).
		Int("target", s.configs.Feed.TargetCount).
		Msg("current feed count")

	return s.FeedCount >= s.configs.Feed.TargetCount
}

func (s *VideoStat) isLiveTargetAchieved() bool {
	log.Info().
		Int("count", s.LiveCount).
		Int("target", s.configs.Live.TargetCount).
		Msg("current live count")

	return s.LiveCount >= s.configs.Live.TargetCount
}

func (s *VideoStat) isTargetAchieved() bool {
	return s.isFeedTargetAchieved() && s.isLiveTargetAchieved()
}

type FeedConfig struct {
	TargetCount int           `json:"target_count"`
	SleepRandom []interface{} `json:"sleep_random"`
}

type LiveConfig struct {
	TargetCount int           `json:"target_count"`
	SleepRandom []interface{} `json:"sleep_random"`
}

type VideoCrawlerConfigs struct {
	AppPackageName string `json:"app_package_name"`

	Feed FeedConfig `json:"feed"`
	Live LiveConfig `json:"live"`
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
		return points[0].Center(), true
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

	for !l.currentStat.isLiveTargetAchieved() {
		// check if live room
		if err := l.driver.assertActivity(l.configs.AppPackageName, "live"); err != nil {
			return err
		}

		// take screenshot and get screen texts by OCR
		_, err := l.driver.GetScreenTextsByOCR()
		if err != nil {
			log.Error().Err(err).Msg("OCR GetTexts failed")
			continue
		}

		// TODO: check live type

		// swipe to next live video
		err = l.driver.SwipeUp()
		if err != nil {
			log.Error().Err(err).Msg("swipe up failed")
			// TODO: retry maximum 3 times
			continue
		}

		// sleep custom random time
		if err := sleepRandom(l.configs.Live.SleepRandom); err != nil {
			log.Error().Err(err).Msg("sleep random failed")
		}

		// TODO: check live type

		l.currentStat.LiveCount++
	}

	log.Info().Msg("live count achieved, exit live room")

	return l.exitLiveRoom()
}

func (l *LiveCrawler) exitLiveRoom() error {
	for i := 0; i < 3; i++ {
		l.driver.SwipeRelative(0.1, 0.5, 0.9, 0.5)
		time.Sleep(2 * time.Second)

		// check if back to feed page
		if err := l.driver.assertActivity(l.configs.AppPackageName, "feed"); err == nil {
			return nil
		}
	}

	// exit live room failed, while video count achieved
	if l.currentStat.isTargetAchieved() {
		return nil
	}

	return errors.New("exit live room failed")
}

func (dExt *DriverExt) VideoCrawler(configs *VideoCrawlerConfigs) (err error) {
	currVideoStat := &VideoStat{
		configs: configs,
	}
	defer func() {
		dExt.cacheStepData.VideoStat = currVideoStat
	}()

	// launch app
	if err = dExt.Driver.AppLaunch(configs.AppPackageName); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

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
			continue
		}

		// automatic handling of pop-up windows
		if err := dExt.autoPopupHandler(texts); err != nil {
			log.Error().Err(err).Msg("auto handle popup failed")
			return err
		}

		// check if live video && run live crawler
		if enterPoint, isLive := liveCrawler.checkLiveVideo(texts); isLive {
			log.Info().Msg("live video found")
			if !liveCrawler.currentStat.isLiveTargetAchieved() {
				if err := liveCrawler.Run(dExt, enterPoint); err != nil {
					log.Error().Err(err).Msg("run live crawler failed, continue")
					continue
				}
			}
		}

		// TODO: check feed type

		currVideoStat.FeedCount++
		// sleep custom random time
		if err := sleepRandom(configs.Feed.SleepRandom); err != nil {
			log.Error().Err(err).Msg("sleep random failed")
		}

		// check if target count achieved
		if currVideoStat.isTargetAchieved() {
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

func (dExt *DriverExt) assertActivity(packageName, activityType string) error {
	log.Debug().Str("pacakge_name", packageName).
		Str("activity_type", activityType).Msg("assert activity")
	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Error().Err(err).Msg("get foreground app failed")
		return err
	}

	if app.PackageName != packageName {
		return fmt.Errorf("app %s is not in foreground", packageName)
	}

	if activities, ok := androidActivities[app.PackageName]; ok {
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
			point := points[1].Center()
			if err := dExt.TapAbsXY(point.X, point.Y); err != nil {
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
