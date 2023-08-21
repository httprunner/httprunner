package uixt

import (
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/code"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

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
	Timeout int `json:"timeout"` // seconds

	Feed FeedConfig `json:"feed"`
	Live LiveConfig `json:"live"`
}

type VideoCrawler struct {
	driverExt *DriverExt
	configs   *VideoCrawlerConfigs
	timer     *time.Timer

	FeedCount int            `json:"feed_count"`
	FeedStat  map[string]int `json:"feed_stat"` // 分类统计 feed 数量：视频/图文/广告/特效/模板/购物
	LiveCount int            `json:"live_count"`
	LiveStat  map[string]int `json:"live_stat"` // 分类统计 live 数量：秀场/游戏/电商/多人
}

func (vc *VideoCrawler) isFeedTargetAchieved() bool {
	targetStat := make(map[string]int)
	for _, targetLabel := range vc.configs.Feed.TargetLabels {
		targetStat[targetLabel.Text] = targetLabel.Target
	}

	log.Info().
		Int("current_total", vc.FeedCount).
		Interface("current_stat", vc.FeedStat).
		Int("target_total", vc.configs.Feed.TargetCount).
		Interface("target_stat", targetStat).
		Msg("display feed crawler progress")

	// check total feed count
	if vc.FeedCount < vc.configs.Feed.TargetCount {
		return false
	}

	// check each feed type's count
	for _, targetLabel := range vc.configs.Feed.TargetLabels {
		if vc.FeedStat[targetLabel.Text] < targetLabel.Target {
			return false
		}
	}

	return true
}

func (vc *VideoCrawler) isLiveTargetAchieved() bool {
	targetStat := make(map[string]int)
	for _, targetLabel := range vc.configs.Live.TargetLabels {
		targetStat[targetLabel.Text] = targetLabel.Target
	}

	log.Info().
		Int("current_total", vc.LiveCount).
		Interface("current_stat", vc.LiveStat).
		Int("target_total", vc.configs.Live.TargetCount).
		Interface("target_stat", targetStat).
		Msg("display live crawler progress")

	// check total live count
	if vc.LiveCount < vc.configs.Live.TargetCount {
		return false
	}

	// check each live type's count
	for _, targetLabel := range vc.configs.Live.TargetLabels {
		if vc.LiveStat[targetLabel.Text] < targetLabel.Target {
			return false
		}
	}

	return true
}

func (vc *VideoCrawler) isTargetAchieved() bool {
	return vc.isFeedTargetAchieved() && vc.isLiveTargetAchieved()
}

// incrFeed increases feed count and feed stat
func (vc *VideoCrawler) incrFeed(screenResult *ScreenResult) error {
	screenResult.VideoType = "feed"

	var author string
	if screenResult.Texts != nil {
		// handle screenshot
		// find feed author
		actionOptions := []ActionOption{
			WithRegex(true),
			vc.driverExt.GenAbsScope(0, 0.5, 1, 1).Option(),
		}
		ocrText, err := screenResult.Texts.FindText("^@", actionOptions...)
		if err != nil {
			return errors.Wrap(err, "find feed author failed")
		}
		author = fmt.Sprintf("@%s", removeNonAlphanumeric(ocrText.Text))
		log.Info().Str("author", author).Msg("found feed author by OCR")

		// find target labels
		for _, targetLabel := range vc.configs.Feed.TargetLabels {
			scope := targetLabel.Scope
			actionOptions := []ActionOption{
				WithRegex(targetLabel.Regex),
				vc.driverExt.GenAbsScope(scope[0], scope[1], scope[2], scope[3]).Option(),
			}
			if _, err := screenResult.Texts.FindText(targetLabel.Text, actionOptions...); err == nil {
				key := targetLabel.Text
				if _, ok := vc.FeedStat[key]; !ok {
					vc.FeedStat[key] = 0
				}
				vc.FeedStat[key]++
				screenResult.Tags = append(screenResult.Tags, key)
			}
		}
	}

	if screenResult.Feed == nil {
		// get feed trackings by author
		if vc.driverExt.plugin != nil {
			feedVideo, err := vc.getFeedVideo(author)
			if err != nil {
				return errors.Wrap(err, "get feed video from plugin failed")
			}
			screenResult.Feed = feedVideo
		} else {
			screenResult.Feed = &FeedVideo{}
		}
	}

	// get simulation play duration
	if screenResult.Feed.SimulationPlayDuration != 0 {
		screenResult.Feed.PlayDuration = screenResult.Feed.SimulationPlayDuration
	} else {
		screenResult.Feed.RandomPlayDuration = getSimulationDuration(vc.configs.Feed.SleepRandom)
		screenResult.Feed.PlayDuration = screenResult.Feed.RandomPlayDuration
	}

	log.Info().Strs("tags", screenResult.Tags).
		Interface("feed", screenResult.Feed).
		Msg("found feed success")
	vc.FeedCount++
	return nil
}

// incrLive increases live count and live stat
func (vc *VideoCrawler) incrLive(screenResult *ScreenResult) error {
	screenResult.VideoType = "live"
	// TODO: check live type

	if screenResult.Live == nil {
		screenResult.Live = &LiveRoom{}
	}

	// TODO: add popularity data for live

	screenResult.Live.SimulationWatchDuration = getSimulationDuration(vc.configs.Live.SleepRandom)

	log.Info().Strs("tags", screenResult.Tags).
		Interface("live", screenResult.Live).
		Msg("found live success")
	vc.LiveCount++
	return nil
}

func (vc *VideoCrawler) checkLiveVideo(feedVideo *FeedVideo) (enterPoint PointF, yes bool) {
	// TODO: check if preview-live from feedVideo
	if feedVideo.Type != "live" {
		return PointF{}, false
	}

	// take screenshot and get OCR texts via image service
	texts, err := vc.driverExt.GetScreenTexts()
	if err != nil {
		return PointF{}, false
	}

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
func (vc *VideoCrawler) startLiveCrawler(enterPoint PointF) error {
	log.Info().Msg("enter live room")
	if err := vc.driverExt.TapAbsXY(enterPoint.X, enterPoint.Y); err != nil {
		log.Error().Err(err).Msg("tap live video failed")
		return err
	}
	time.Sleep(5 * time.Second)
	for !vc.isLiveTargetAchieved() {
		select {
		case <-vc.timer.C:
			log.Warn().Msg("timeout in live crawler")
			return errors.Wrap(code.TimeoutError, "live crawler timeout")
		case <-vc.driverExt.interruptSignal:
			log.Warn().Msg("interrupted in live crawler")
			return errors.Wrap(code.InterruptError, "live crawler interrupted")
		default:
			// swipe to next live video
			swipeStartTime := time.Now()
			if err := vc.driverExt.SwipeUp(); err != nil {
				log.Error().Err(err).Msg("live swipe up failed")
				return err
			}
			swipeFinishTime := time.Now()

			// wait for live video loading
			time.Sleep(5 * time.Second)

			// take screenshot and get screen texts by OCR
			screenResult, err := vc.driverExt.GetScreenResult()
			if err != nil {
				log.Error().Err(err).Msg("OCR GetTexts failed")
				time.Sleep(3 * time.Second)
				continue
			}

			// check live type and incr live count
			if err := vc.incrLive(screenResult); err != nil {
				log.Error().Err(err).Msg("incr live failed")
			}

			// simulation watch live video
			sleepStrict(swipeFinishTime, screenResult.Live.SimulationWatchDuration)

			// log swipe timelines
			screenResult.SwipeStartTime = swipeStartTime.UnixMilli()
			screenResult.SwipeFinishTime = swipeFinishTime.UnixMilli()
			screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()
		}
	}

	log.Info().Msg("live count achieved, exit live room")

	return vc.exitLiveRoom()
}

func (vc *VideoCrawler) exitLiveRoom() error {
	for i := 0; i < 3; i++ {
		vc.driverExt.SwipeRelative(0.1, 0.5, 0.9, 0.5)
		time.Sleep(2 * time.Second)
	}

	// exit live room failed, while video count achieved
	if vc.isTargetAchieved() {
		return nil
	}

	// click X button on upper-right corner
	if err := vc.driverExt.TapXY(0.95, 0.05); err == nil {
		log.Info().Msg("tap X button on upper-right corner to exit live room")
		time.Sleep(2 * time.Second)
	}

	return errors.New("exit live room failed")
}

func (dExt *DriverExt) VideoCrawler(configs *VideoCrawlerConfigs) (err error) {
	if dExt.plugin == nil {
		return errors.New("miss plugin for video crawler")
	}

	// set default sleep random strategy if not set
	if configs.Feed.SleepRandom == nil {
		configs.Feed.SleepRandom = []interface{}{1, 5}
	}
	if configs.Live.SleepRandom == nil {
		configs.Live.SleepRandom = []interface{}{10, 15}
	}

	crawler := &VideoCrawler{
		driverExt: dExt,
		configs:   configs,

		FeedCount: 0,
		FeedStat:  make(map[string]int),
		LiveCount: 0,
		LiveStat:  make(map[string]int),
	}
	defer func() {
		dExt.cacheStepData.videoCrawler = crawler
	}()

	// loop until target count achieved or timeout
	// the main loop is feed crawler
	crawler.timer = time.NewTimer(time.Duration(configs.Timeout) * time.Second)
	for {
		select {
		case <-crawler.timer.C:
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

			var screenResult *ScreenResult
			// get app event trackings
			feedVideo, err := crawler.getCurrentFeedVideo()
			if err != nil {
				return errors.Wrap(err, "get app event trackings failed")
			}
			screenResult = &ScreenResult{
				Feed:  feedVideo,
				Texts: nil,
				Tags:  nil,
			}
			dExt.cacheStepData.screenResults[time.Now().String()] = screenResult

			// check if live video && run live crawler
			if enterPoint, isLive := crawler.checkLiveVideo(feedVideo); isLive {
				// 直播预览流
				screenResult.VideoType = "live-preview"
				log.Info().Msg("live video found")
				if !crawler.isLiveTargetAchieved() {
					if err := crawler.startLiveCrawler(enterPoint); err != nil {
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
				err := crawler.incrFeed(screenResult)
				if err != nil {
					log.Warn().Err(err).Msg("incr feed failed")
				} else {
					// simulation watch feed video
					sleepStrict(swipeFinishTime, screenResult.Feed.PlayDuration)
				}
			}

			// check if target count achieved
			if crawler.isTargetAchieved() {
				log.Info().Msg("target count achieved, exit crawler")
				return nil
			}

			// log swipe timelines
			screenResult.SwipeStartTime = swipeStartTime.UnixMilli()
			screenResult.SwipeFinishTime = swipeFinishTime.UnixMilli()
			screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()
		}
	}
}

type FeedVideo struct {
	// 视频基础数据
	CacheKey string `json:"cache_key"` // 视频 CacheKey
	UserName string `json:"user_name"` // 视频作者
	Duration int64  `json:"duration"`  // 视频时长(ms)
	Caption  string `json:"caption"`   // 视频文案
	Type     string `json:"type"`      // 视频类型, feed/live

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
	LiveStreamID string `json:"live_stream_id"` // 直播流 ID
	UserName     string `json:"user_name"`      // 视频作者
	Caption      string `json:"caption"`        // 视频文案
	LiveType     string `json:"live_type"`      // 直播间类型

	// 视频热度数据
	AudienceCount string `json:"audience_count"` // 直播间人数
	LikeCount     int64  `json:"like_count"`     // 点赞数

	// 记录仿真决策信息
	SimulationWatchDuration int64 `json:"simulation_watch_duration"` // 仿真观播时长(ms)

	// timelines
	PreloadTimestamp int64 `json:"preload_timestamp"` // feed 预加载时间戳
}

func (vc *VideoCrawler) getFeedVideo(authorName string) (feedVideo *FeedVideo, err error) {
	if !vc.driverExt.plugin.Has("GetFeedVideo") {
		return nil, errors.New("plugin missing GetFeedVideo method")
	}

	resp, err := vc.driverExt.plugin.Call("GetFeedVideo", authorName)
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

	feedVideo = &FeedVideo{}
	err = json.Unmarshal(feedBytes, feedVideo)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal feed video info failed")
	}

	log.Info().Interface("feedVideo", feedVideo).Msg("get feed video success")
	return feedVideo, nil
}

func (vc *VideoCrawler) getCurrentFeedVideo() (feedVideo *FeedVideo, err error) {
	if !vc.driverExt.plugin.Has("GetCurrentFeedVideo") {
		return nil, errors.New("plugin missing GetCurrentFeedVideo method")
	}

	// FIXME: wait for cache update
	time.Sleep(2000 * time.Millisecond)

	// TODO: retry 3 times if get failed, abort if fail more than 3 times
	resp, err := vc.driverExt.plugin.Call("GetCurrentFeedVideo")
	if err != nil {
		return nil, errors.Wrap(err, "call plugin GetCurrentFeedVideo failed")
	}

	if resp == nil {
		return nil, errors.New("feed not found")
	}

	feedBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.New("json marshal feed video info failed")
	}

	feedVideo = &FeedVideo{}
	err = json.Unmarshal(feedBytes, feedVideo)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal feed video info failed")
	}

	// TODO: check if app event trackings changed
	// TODO: check and handle popups if event trackings not changed
	log.Info().
		Interface("feedVideoCaption", feedVideo.Caption).
		Msg("get current feed video success")
	return feedVideo, nil
}

func (vc *VideoCrawler) getCurrentLiveRoom() (liveVideo *LiveRoom, err error) {
	// TODO
	return
}

func removeNonAlphanumeric(input string) string {
	// 使用正则表达式匹配中英文字符以外的内容
	re := regexp.MustCompile(`[^\p{L}\p{N}]+`)
	// 删除匹配到的非中英文字符
	processed := re.ReplaceAllString(input, "")
	return processed
}
