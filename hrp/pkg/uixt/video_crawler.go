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

	FeedCount int            `json:"feed_count"`
	FeedStat  map[string]int `json:"feed_stat"` // 分类统计 feed 数量：视频/图文/广告/特效/模板/购物
	LiveCount int            `json:"live_count"`
	LiveStat  map[string]int `json:"live_stat"` // 分类统计 live 数量：秀场/游戏/电商/多人
}

func (s *VideoStat) isFeedTargetAchieved() bool {
	log.Info().
		Int("count", s.FeedCount).
		Interface("stat", s.FeedStat).
		Int("target", s.configs.Feed.TargetCount).
		Msg("current feed count")

	return s.FeedCount >= s.configs.Feed.TargetCount
}

func (s *VideoStat) isLiveTargetAchieved() bool {
	log.Info().
		Int("count", s.LiveCount).
		Interface("stat", s.FeedStat).
		Int("target", s.configs.Live.TargetCount).
		Msg("current live count")

	return s.LiveCount >= s.configs.Live.TargetCount
}

func (s *VideoStat) isTargetAchieved() bool {
	return s.isFeedTargetAchieved() && s.isLiveTargetAchieved()
}

// incrFeed increases feed count and feed stat
func (s *VideoStat) incrFeed(texts OCRTexts, driverExt *DriverExt) error {
	// feed author
	actionOptions := []ActionOption{
		WithRegex(true),
		driverExt.GenAbsScope(0, 0.5, 1, 1).Option(),
	}
	if ocrText, err := texts.FindText("^@", actionOptions...); err == nil {
		log.Info().Str("author", ocrText.Text).Msg("found feed author")
	}

	for _, targetLabel := range s.configs.Feed.TargetLabels {
		scope := targetLabel.Scope
		actionOptions := []ActionOption{
			WithRegex(targetLabel.Regex),
			driverExt.GenAbsScope(scope[0], scope[1], scope[2], scope[3]).Option(),
		}
		if ocrText, err := texts.FindText(targetLabel.Text, actionOptions...); err == nil {
			log.Info().Str("label", targetLabel.Text).
				Str("text", ocrText.Text).Msg("found feed success")

			key := targetLabel.Text
			if _, ok := s.FeedStat[key]; !ok {
				s.FeedStat[key] = 0
			}
			s.FeedStat[key]++
		}
	}

	s.FeedCount++
	return nil
}

type TargetLabel struct {
	Text  string `json:"text"`
	Scope Scope  `json:"scope"`
	Regex bool   `json:"regex"`
}

type FeedConfig struct {
	TargetCount  int           `json:"target_count"`
	TargetLabels []TargetLabel `json:"target_labels"`
	SleepRandom  []interface{} `json:"sleep_random"`
}

type LiveConfig struct {
	TargetCount  int           `json:"target_count"`
	TargetLabels []TargetLabel `json:"target_labels"`
	SleepRandom  []interface{} `json:"sleep_random"`
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
	// set default sleep random strategy if not set
	if configs.Feed.SleepRandom == nil {
		configs.Feed.SleepRandom = []interface{}{1, 5}
	}
	if configs.Live.SleepRandom == nil {
		configs.Live.SleepRandom = []interface{}{10, 15}
	}

	currVideoStat := &VideoStat{
		configs: configs,

		FeedCount: 0,
		FeedStat:  make(map[string]int),
		LiveCount: 0,
		LiveStat:  make(map[string]int),
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

		// check feed type and incr feed count
		if err := currVideoStat.incrFeed(texts, dExt); err != nil {
			log.Error().Err(err).Msg("incr feed failed")
		}

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
		time.Sleep(1 * time.Second)
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
