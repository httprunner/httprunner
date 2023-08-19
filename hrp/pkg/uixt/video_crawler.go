package uixt

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/httprunner/funplugin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
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
	screenResult.VideoType = "feed"
	screenResult.Feed = &FeedVideo{}

	// find feed author
	actionOptions := []ActionOption{
		WithRegex(true),
		driverExt.GenAbsScope(0, 0.5, 1, 1).Option(),
	}
	ocrText, err := screenResult.Texts.FindText("^@", actionOptions...)
	if err != nil {
		return errors.Wrap(err, "find feed author failed")
	}
	author := fmt.Sprintf("@%s", removeNonAlphanumeric(ocrText.Text))
	log.Info().Str("author", author).Msg("found feed author by OCR")
	screenResult.Feed.UserName = author

	// find target labels
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

	// get feed trackings by author
	if driverExt.plugin != nil {
		feedVideo, err := getFeedVideo(driverExt.plugin, author)
		if err != nil {
			return errors.Wrap(err, "get feed video from plugin failed")
		}
		screenResult.Feed = feedVideo
	}

	// get simulation play duration
	if screenResult.Feed.SimulationPlayDuration != 0 {
		screenResult.Feed.PlayDuration = screenResult.Feed.SimulationPlayDuration
	} else {
		screenResult.Feed.RandomPlayDuration = getSimulationDuration(s.configs.Feed.SleepRandom)
		screenResult.Feed.PlayDuration = screenResult.Feed.RandomPlayDuration
	}

	log.Info().Strs("tags", screenResult.Tags).
		Interface("feed", screenResult.Feed).
		Msg("found feed success")
	s.FeedCount++
	return nil
}

// incrLive increases live count and live stat
func (s *VideoStat) incrLive(screenResult *ScreenResult, driverExt *DriverExt) error {
	screenResult.VideoType = "live"
	// TODO: check live type

	if screenResult.Live == nil {
		screenResult.Live = &LiveRoom{}
	}

	// TODO: add popularity data for live

	screenResult.Live.WatchDuration = getSimulationDuration(s.configs.Live.SleepRandom)

	log.Info().Strs("tags", screenResult.Tags).
		Interface("live", screenResult.Live).
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
			// swipe to next live video
			swipeStartTime := time.Now()
			if err := l.driver.SwipeUp(); err != nil {
				log.Error().Err(err).Msg("live swipe up failed")
				return err
			}
			swipeFinishTime := time.Now()

			// wait for live video loading
			time.Sleep(5 * time.Second)

			// take screenshot and get screen texts by OCR
			screenResult, err := l.driver.GetScreenResult()
			if err != nil {
				log.Error().Err(err).Msg("OCR GetTexts failed")
				time.Sleep(3 * time.Second)
				continue
			}

			// check live type and incr live count
			if err := l.currentStat.incrLive(screenResult, l.driver); err != nil {
				log.Error().Err(err).Msg("incr live failed")
			}

			// simulation watch live video
			sleepStrict(swipeFinishTime, screenResult.Live.WatchDuration)

			// check if live room
			if err := l.driver.Driver.AssertForegroundApp(l.configs.AppPackageName, "live"); err != nil {
				return err
			}

			// log swipe timelines
			screenResult.SwipeStartTime = swipeStartTime.UnixMilli()
			screenResult.SwipeFinishTime = swipeFinishTime.UnixMilli()
			screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()
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
		if err := l.driver.Driver.AssertForegroundApp(l.configs.AppPackageName, "feed"); err == nil {
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
		if err := l.driver.Driver.AssertForegroundApp(l.configs.AppPackageName, "feed"); err == nil {
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
		if err != nil && !errors.Is(err, errDriverNotImplemented) {
			log.Warn().Err(err).Msg("get foreground app failed, ignore")
			return errors.Wrap(code.MobileUIAssertForegroundAppError, err.Error())
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
			// swipe to next feed video
			log.Info().Msg("swipe to next feed video")
			swipeStartTime := time.Now()
			if err = dExt.SwipeUp(); err != nil {
				log.Error().Err(err).Msg("feed swipe up failed")
				return err
			}
			swipeFinishTime := time.Now()

			// take screenshot and get screen texts by OCR
			screenResult, err := dExt.GetScreenResult()
			if err != nil {
				if strings.Contains(err.Error(), "connect: connection refused") {
					return err
				}
				log.Error().Err(err).Msg("OCR GetTexts failed")
				time.Sleep(3 * time.Second)
				continue
			}

			// automatic handling of pop-up windows
			if err := dExt.AutoPopupHandler(screenResult.Texts); err != nil {
				log.Error().Err(err).Msg("auto handle popup failed")
				return err
			}

			// check if live video && run live crawler
			if enterPoint, isLive := liveCrawler.checkLiveVideo(screenResult.Texts); isLive {
				// 直播预览流
				screenResult.VideoType = "live-preview"
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
			} else {
				// 点播
				// check feed type and incr feed count
				err := currVideoStat.incrFeed(screenResult, dExt)
				if err != nil {
					log.Warn().Err(err).Msg("incr feed failed")
				} else {
					// simulation watch feed video
					sleepStrict(swipeFinishTime, screenResult.Feed.PlayDuration)
				}
			}

			// check if target count achieved
			if currVideoStat.isTargetAchieved() {
				log.Info().Msg("target count achieved, exit crawler")
				return nil
			}

			// check if feed page
			if err := dExt.Driver.AssertForegroundApp(configs.AppPackageName, "feed"); err != nil {
				return err
			}

			// log swipe timelines
			screenResult.SwipeStartTime = swipeStartTime.UnixMilli()
			screenResult.SwipeFinishTime = swipeFinishTime.UnixMilli()
			screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()
		}
	}
}

func getFeedVideo(plugin funplugin.IPlugin, authorName string) (feedVideo *FeedVideo, err error) {
	if !plugin.Has("GetFeedVideo") {
		return nil, errors.New("plugin missing GetFeedVideo method")
	}

	resp, err := plugin.Call("GetFeedVideo", authorName)
	if err != nil {
		return nil, errors.Wrap(err, "call plugin GetFeedVideo failed")
	}

	if resp == nil {
		return nil, errors.New("feed not found")
	}

	feedBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.New("json marshal feed video info failed")
	}

	err = json.Unmarshal(feedBytes, &feedVideo)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal feed video info failed")
	}

	log.Info().Interface("feedVideo", feedVideo).Msg("get feed video success")
	return feedVideo, nil
}

func removeNonAlphanumeric(input string) string {
	// 使用正则表达式匹配中英文字符以外的内容
	re := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	// 删除匹配到的非中英文字符
	processed := re.ReplaceAllString(input, "")
	return processed
}

type FeedVideo struct {
	// 视频基础数据
	UserName string `json:"user_name"` // 视频作者
	Duration int64  `json:"duration"`  // 视频时长(ms)
	Caption  string `json:"caption"`   // 视频文案

	// 视频热度数据
	ViewCount    int64 `json:"view_count"`    // feed 观看数
	LikeCount    int64 `json:"like_count"`    // feed 点赞数
	CommentCount int64 `json:"comment_count"` // feed 评论数
	CollectCount int64 `json:"collect_count"` // feed 收藏数
	ForwardCount int64 `json:"forward_count"` // feed 转发数
	ShareCount   int64 `json:"share_count"`   // feed 分享数

	// 记录仿真决策信息
	PlayDuration           int64   `json:"play_duration"`            // 播放时长(ms)，取自 Simulation/Random
	SimulationPlayProgress float64 `json:"simulation_play_progress"` // 仿真播放比例（完播率）
	SimulationPlayDuration int64   `json:"simulation_play_duration"` // 仿真播放时长(ms)
	RandomPlayDuration     int64   `json:"random_play_duration"`     // 随机播放时长(ms)

	// timelines
	PublishTimestamp int64 `json:"publish_timestamp"` // feed 发布时间戳
	PreloadTimestamp int64 `json:"preload_timestamp"` // feed 预加载时间戳
}

type LiveRoom struct {
	// 视频基础数据
	UserName string `json:"user_name"` // 主播名
	LiveType string `json:"live_type"` // 直播间类型
	// 直播热度数据
	LiveUsers string `json:"live_users"` // 直播间人数
	// 记录仿真决策信息
	WatchDuration int64 `json:"watch_duration"` // 观看时长(ms)
}
