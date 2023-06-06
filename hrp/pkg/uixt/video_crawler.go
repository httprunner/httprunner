package uixt

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

type VideoStat struct {
	configs *VideoCrawlerConfigs
	timer   *time.Timer

	FeedCount int            `json:"feed_count"`
	FeedStat  map[string]int `json:"feed_stat"` // 分类统计 feed 数量：视频/图文/广告/特效/模板/购物
	LiveCount int            `json:"live_count"`
	LiveStat  map[string]int `json:"live_stat"` // 分类统计 live 数量：秀场/游戏/电商/多人
}

func (s *VideoStat) isFeedTargetAchieved() bool {
	targetStat := make(map[string]int)
	for _, targetLabel := range s.configs.Feed.TargetLabels {
		targetStat[targetLabel.Text] = targetLabel.Target
	}

	log.Info().
		Int("current_total", s.FeedCount).
		Interface("current_stat", s.FeedStat).
		Int("target_total", s.configs.Feed.TargetCount).
		Interface("target_stat", targetStat).
		Msg("display feed crawler progress")

	// check total feed count
	if s.FeedCount < s.configs.Feed.TargetCount {
		return false
	}

	// check each feed type's count
	for _, targetLabel := range s.configs.Feed.TargetLabels {
		if s.FeedStat[targetLabel.Text] < targetLabel.Target {
			return false
		}
	}

	return true
}

func (s *VideoStat) isLiveTargetAchieved() bool {
	targetStat := make(map[string]int)
	for _, targetLabel := range s.configs.Live.TargetLabels {
		targetStat[targetLabel.Text] = targetLabel.Target
	}

	log.Info().
		Int("current_total", s.LiveCount).
		Interface("current_stat", s.LiveStat).
		Int("target_total", s.configs.Live.TargetCount).
		Interface("target_stat", targetStat).
		Msg("display live crawler progress")

	// check total live count
	if s.LiveCount < s.configs.Live.TargetCount {
		return false
	}

	// check each live type's count
	for _, targetLabel := range s.configs.Live.TargetLabels {
		if s.LiveStat[targetLabel.Text] < targetLabel.Target {
			return false
		}
	}

	return true
}

func (s *VideoStat) isTargetAchieved() bool {
	return s.isFeedTargetAchieved() && s.isLiveTargetAchieved()
}

// incrFeed increases feed count and feed stat
func (s *VideoStat) incrFeed(screenResult *ScreenResult, driverExt *DriverExt) error {
	// feed author
	actionOptions := []ActionOption{
		WithRegex(true),
		driverExt.GenAbsScope(0, 0.5, 1, 1).Option(),
	}
	if ocrText, err := screenResult.Texts.FindText("^@", actionOptions...); err == nil {
		log.Debug().Str("author", ocrText.Text).Msg("found feed author")
		screenResult.Tags = append(screenResult.Tags, ocrText.Text)
	}

	for _, targetLabel := range s.configs.Feed.TargetLabels {
		scope := targetLabel.Scope
		actionOptions := []ActionOption{
			WithRegex(targetLabel.Regex),
			driverExt.GenAbsScope(scope[0], scope[1], scope[2], scope[3]).Option(),
		}
		if _, err := screenResult.Texts.FindText(targetLabel.Text, actionOptions...); err == nil {
			key := targetLabel.Text
			if _, ok := s.FeedStat[key]; !ok {
				s.FeedStat[key] = 0
			}
			s.FeedStat[key]++
			screenResult.Tags = append(screenResult.Tags, key)
		}
	}

	// add popularity data for feed
	popularityData := screenResult.Texts.FilterScope(driverExt.GenAbsScope(0.8, 0.5, 1, 0.8))
	if len(popularityData) != 4 {
		log.Warn().Interface("popularity", popularityData).Msg("get feed popularity data failed")
	} else {
		screenResult.Popularity = Popularity{
			Stars:     popularityData[0].Text,
			Comments:  popularityData[1].Text,
			Favorites: popularityData[2].Text,
			Shares:    popularityData[3].Text,
		}
	}

	log.Info().Strs("tags", screenResult.Tags).
		Interface("popularity", screenResult.Popularity).
		Msg("found feed success")
	s.FeedCount++
	return nil
}

// incrLive increases live count and live stat
func (s *VideoStat) incrLive(screenResult *ScreenResult, driverExt *DriverExt) error {
	// TODO: check live type

	// add popularity data for live
	popularityData := screenResult.Texts.FilterScope(driverExt.GenAbsScope(0.7, 0.05, 1, 0.15))
	if len(popularityData) != 1 {
		log.Warn().Interface("popularity", popularityData).Msg("get live popularity data failed")
	} else {
		screenResult.Popularity = Popularity{
			LiveUsers: popularityData[0].Text,
		}
	}

	log.Info().Strs("tags", screenResult.Tags).
		Interface("popularity", screenResult.Popularity).
		Msg("found live success")
	s.LiveCount++
	return nil
}

type TargetLabel struct {
	Text   string `json:"text"`
	Scope  Scope  `json:"scope"`
	Regex  bool   `json:"regex"`
	Target int    `json:"target"` // target count for current label
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
	Timeout        int    `json:"timeout"` // seconds

	Feed FeedConfig `json:"feed"`
	Live LiveConfig `json:"live"`
}

var androidActivities = map[string]map[string]string{
	// DY
	"com.ss.android.ugc.aweme": {
		"feed": ".splash.SplashActivity",
		"live": ".live.LivePlayActivity",
	},
	// DY lite
	"com.ss.android.ugc.aweme.lite": {
		"feed": ".splash.SplashActivity",
		"live": ".live.LivePlayActivity",
	},
	// KS
	"com.smile.gifmaker": {
		"feed": "com.yxcorp.gifshow.HomeActivity",
		"live": "com.kuaishou.live.core.basic.activity.LiveSlideActivity",
	},
	// KS lite
	"com.kuaishou.nebula": {
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
	// 预览流入口：DY/KS
	// 标签文案：点击进入直播间|进入直播间领金币
	points, err := texts.FindTexts([]string{".*进入直播间.*"}, WithScope(0, 0.3, 1, 0.8), WithRegex(true))
	if err == nil {
		return points[0].Center(), true
	}
	// 标签文案：直播中|直播卖货|直播团购
	points, err = texts.FindTexts([]string{"直播中|直播卖货|直播团购"},
		WithScope(0, 0.7, 0.5, 1), WithRegex(true))
	if err == nil {
		return points[0].Center(), true
	}

	// 预览流入口：KS/KSLite
	// 评论框文案：和主播聊聊天...|聊聊天...
	points, err = texts.FindTexts([]string{".*聊聊天.*"}, WithRegex(true))
	if err == nil {
		point := points[0].Center()
		enterPoint = PointF{
			X: point.X,
			Y: point.Y - 300,
		}
		return enterPoint, true
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
		select {
		case <-l.currentStat.timer.C:
			log.Warn().Msg("timeout in live crawler")
			return errors.Wrap(code.TimeoutError, "live crawler timeout")
		case <-l.driver.interruptSignal:
			log.Warn().Msg("interrupted in live crawler")
			return errors.Wrap(code.InterruptError, "live crawler interrupted")
		default:
			// check if live room
			if err := l.driver.assertActivity(l.configs.AppPackageName, "live"); err != nil {
				return err
			}

			// swipe to next live video
			err := l.driver.SwipeUp()
			if err != nil {
				log.Error().Err(err).Msg("swipe up failed")
				// TODO: retry maximum 3 times
				continue
			}

			// sleep custom random time
			if err := sleepRandom(l.configs.Live.SleepRandom); err != nil {
				log.Error().Err(err).Msg("sleep random failed")
			}

			// take screenshot and get screen texts by OCR
			imageResult, err := l.driver.GetScreenResult()
			if err != nil {
				log.Error().Err(err).Msg("OCR GetTexts failed")
				time.Sleep(3 * time.Second)
				continue
			}
			screenResult := l.driver.cacheStepData.screenResults[imageResult.imagePath]
			screenResult.Tags = []string{"live"}
			if imageResult.LiveType != "" {
				screenResult.Tags = append(screenResult.Tags, imageResult.LiveType)
			}

			// check live type and incr live count
			if err := l.currentStat.incrLive(screenResult, l.driver); err != nil {
				log.Error().Err(err).Msg("incr live failed")
			}
		}
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

	// click X button on upper-right corner
	if err := l.driver.TapXY(0.95, 0.05); err == nil {
		log.Info().Msg("tap X button on upper-right corner to exit live room")
		time.Sleep(2 * time.Second)

		// check if back to feed page
		if err := l.driver.assertActivity(l.configs.AppPackageName, "feed"); err == nil {
			return nil
		}
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
		dExt.cacheStepData.videoStat = currVideoStat
	}()

	// launch app
	if configs.AppPackageName != "" {
		if err = dExt.Driver.AppLaunch(configs.AppPackageName); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	} else {
		app, err := dExt.Driver.GetForegroundApp()
		if err != nil {
			// ignore error when get foreground app failed
			log.Warn().Err(err).Msg("get foreground app failed, ignore")
		}
		log.Info().
			Str("packageName", app.PackageName).
			Str("activity", app.Activity).
			Msg("start to video crawler for current foreground app")
		configs.AppPackageName = app.PackageName
	}

	liveCrawler := LiveCrawler{
		driver:      dExt,
		configs:     configs,
		currentStat: currVideoStat,
	}

	// loop until target count achieved or timeout
	// the main loop is feed crawler
	currVideoStat.timer = time.NewTimer(time.Duration(configs.Timeout) * time.Second)
	for {
		select {
		case <-currVideoStat.timer.C:
			log.Warn().Msg("timeout in feed crawler")
			return errors.Wrap(code.TimeoutError, "feed crawler timeout")
		case <-dExt.interruptSignal:
			log.Warn().Msg("interrupted in feed crawler")
			return errors.Wrap(code.InterruptError, "feed crawler interrupted")
		default:
			// take screenshot and get screen texts by OCR
			imageResult, err := dExt.GetScreenResult()
			if err != nil {
				log.Error().Err(err).Msg("OCR GetTexts failed")
				time.Sleep(3 * time.Second)
				continue
			}
			screenResult := dExt.cacheStepData.screenResults[imageResult.imagePath]

			// automatic handling of pop-up windows
			if err := dExt.autoPopupHandler(screenResult); err != nil {
				log.Error().Err(err).Msg("auto handle popup failed")
				return err
			}

			// check if live video && run live crawler
			texts := imageResult.OCRResult.ToOCRTexts()
			if enterPoint, isLive := liveCrawler.checkLiveVideo(texts); isLive {
				log.Info().Msg("live video found")
				if !liveCrawler.currentStat.isLiveTargetAchieved() {
					if err := liveCrawler.Run(dExt, enterPoint); err != nil {
						if errors.Is(err, code.TimeoutError) || errors.Is(err, code.InterruptError) {
							return err
						}
						log.Error().Err(err).Msg("run live crawler failed, continue")
						continue
					}
				}
				screenResult.Tags = []string{"live-preview"}
			} else {
				screenResult.Tags = []string{"feed"}

				// check feed type and incr feed count
				if err := currVideoStat.incrFeed(screenResult, dExt); err != nil {
					log.Error().Err(err).Msg("incr feed failed")
				}
			}

			// sleep custom random time
			if err := sleepRandom(configs.Feed.SleepRandom); err != nil {
				log.Error().Err(err).Msg("sleep random failed")
			}

			// check if target count achieved
			if currVideoStat.isTargetAchieved() {
				log.Info().Msg("target count achieved, exit crawler")
				return nil
			}

			// swipe to next feed video
			log.Info().Msg("swipe to next feed video")
			if err = dExt.SwipeUp(); err != nil {
				log.Error().Err(err).Msg("swipe up failed")
				return err
			}
			time.Sleep(1 * time.Second)

			// check if feed page
			if err := dExt.assertActivity(configs.AppPackageName, "feed"); err != nil {
				return err
			}
		}
	}
}

func (dExt *DriverExt) assertActivity(packageName, activityType string) error {
	log.Debug().Str("pacakge_name", packageName).
		Str("activity_type", activityType).Msg("assert activity")
	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Warn().Err(err).Msg("get foreground app failed, skip app/activity assertion")
		return nil // Notice: ignore error when get foreground app failed
	}

	if app.PackageName != packageName {
		return errors.Wrap(code.MobileUIAppNotInForegroundError,
			fmt.Sprintf("foreground app %s, expect %s", app.PackageName, packageName))
	}

	activities, ok := androidActivities[app.PackageName]
	if !ok {
		msg := fmt.Sprintf("app package %s not configured", app.PackageName)
		log.Error().Interface("app", app.AppBaseInfo).Msg(msg)
		return errors.Wrap(code.MobileUIActivityNotMatchError, msg)
	}

	expectActivity, ok := activities[activityType]
	if !ok {
		msg := fmt.Sprintf("app package %s %s not configured", app.PackageName, activityType)
		log.Error().Interface("app", app.AppBaseInfo).Msg(msg)
		return errors.Wrap(code.MobileUIActivityNotMatchError, msg)
	}

	// assert success
	if strings.HasSuffix(app.Activity, expectActivity) {
		return nil
	}

	// assert failed
	log.Error().
		Interface("app", app.AppBaseInfo).
		Str("expectActivityType", activityType).
		Str("expectActivity", expectActivity).
		Msg("assert activity failed")

	return errors.Wrap(code.MobileUIActivityNotMatchError, "assert activity failed")
}

// TODO: add more popup texts
var popups = [][]string{
	{".*青少年.*", "我知道了"}, // 青少年弹窗
	{".*个人信息保护.*", "同意"},
	{".*通讯录.*", "拒绝"},
	{".*更新.*", "以后再说|稍后"},
	{".*升级.*", "以后再说|稍后"},
	{".*定位.*", "仅.*允许"},
	{".*拍照.*", "仅.*允许"},
	{".*录音.*", "仅.*允许"},
	{".*位置.*", "仅.*允许"},
	{".*权限.*", "仅.*允许|始终允许"},
	{".*允许.*", "仅.*允许|始终允许"},
	{".*风险.*", "继续使用"},
	{"管理使用时间", ".*忽略.*"},
}

func (dExt *DriverExt) autoPopupHandler(screenResult *ScreenResult) error {
	for _, popup := range popups {
		if len(popup) != 2 {
			continue
		}

		points, err := screenResult.Texts.FindTexts([]string{popup[0], popup[1]}, WithRegex(true))
		if err == nil {
			log.Warn().Interface("popup", popup).
				Interface("texts", screenResult.Texts).Msg("text popup found")
			point := points[1].Center()
			log.Info().Str("text", points[1].Text).Msg("close popup")
			if err := dExt.TapAbsXY(point.X, point.Y); err != nil {
				log.Error().Err(err).Msg("tap popup failed")
				return errors.Wrap(code.MobileUIPopupError, err.Error())
			}
			// tap popup success
			return nil
		}
	}

	// no popup found
	return nil
}
