package uixt

import (
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

	// used to help checking if swipe success
	failedCount int64
	lastFeed    *FeedVideo
	lastLive    *LiveRoom

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

func (vc *VideoCrawler) checkLiveVideo(feedVideo *FeedVideo) (enterPoint PointF, yes bool) {
	if !feedVideo.IsLivePreview {
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

			liveRoom, err := vc.getCurrentLiveRoom()
			if err != nil {
				if vc.failedCount >= 3 {
					// failed 3 consecutive times
					return errors.New("get current live event trackings failed 3 consecutive times")
				}
				// retry
				vc.failedCount++
				log.Warn().Int64("failedCount", vc.failedCount).Msg("get current live room failed")
				continue
			}

			// take screenshot and get screen texts by OCR
			screenResult, err := vc.driverExt.GetScreenResult(
				WithScreenShotOCR(true),
				WithScreenShotUpload(true),
				WithScreenShotLiveType(true),
			)
			if err != nil {
				log.Error().Err(err).Msg("OCR GetTexts failed")
				time.Sleep(3 * time.Second)
				continue
			}

			// add live type
			if screenResult.imageResult != nil &&
				screenResult.imageResult.LiveType != "" &&
				screenResult.imageResult.LiveType != "NoLive" {
				liveRoom.LiveType = screenResult.imageResult.LiveType
			}

			// incr live count
			screenResult.VideoType = "live"
			screenResult.Live = liveRoom
			vc.LiveCount++
			log.Info().Strs("tags", screenResult.Tags).
				Interface("live", screenResult.Live).
				Msg("found live success")

			// get simulation watch duration
			if screenResult.Live.SimulationWatchDuration != 0 {
				screenResult.Live.WatchDuration = screenResult.Live.SimulationWatchDuration
			} else {
				screenResult.Live.RandomWatchDuration = getSimulationDuration(vc.configs.Live.SleepRandom)
				screenResult.Live.WatchDuration = screenResult.Live.RandomWatchDuration
			}
			// simulation watch live video
			sleepStrict(swipeFinishTime, screenResult.Live.WatchDuration)

			// log swipe timelines
			screenResult.SwipeStartTime = swipeStartTime.UnixMilli()
			screenResult.SwipeFinishTime = swipeFinishTime.UnixMilli()
			screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()

			// reset failed count
			vc.failedCount = 0
		}
	}

	log.Info().Msg("live count achieved, exit live room")

	return vc.exitLiveRoom()
}

func (vc *VideoCrawler) exitLiveRoom() error {
	log.Info().Msg("exit live room")
	// swipe right twice to exit live room
	for i := 0; i < 2; i++ {
		vc.driverExt.SwipeRelative(0.1, 0.5, 0.9, 0.5)
		time.Sleep(2 * time.Second)
	}
	// TODO: check exit live room success
	return nil
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

		failedCount: 0,
		lastFeed:    &FeedVideo{},
		lastLive:    &LiveRoom{},

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

			// get app event trackings
			// retry 3 times if get feed failed, abort if fail 3 consecutive times
			feedVideo, err := crawler.getCurrentFeedVideo()
			if err != nil || feedVideo.Type == "" {
				if crawler.failedCount >= 3 {
					// failed 3 consecutive times
					return errors.New("get current feed video failed 3 consecutive times")
				}
				log.Warn().Interface("feedVideo", feedVideo).Msg("get current feed video failed")
				// retry
				crawler.failedCount++
				continue
			}

			if feedVideo.VideoID == crawler.lastFeed.VideoID {
				// app event tracking not changed
				// check and handle popups
				log.Warn().Msg("feed video event tracking not changed")
				if err = crawler.driverExt.ClosePopupsHandler(WithMaxRetryTimes(1)); err != nil {
					return err
				}
			}
			crawler.lastFeed = feedVideo

			screenResult := &ScreenResult{
				Resolution: dExt.windowSize,
			}
			dExt.cacheStepData.screenResults[time.Now().String()] = screenResult

			// check if live video && run live crawler
			if enterPoint, isLive := crawler.checkLiveVideo(feedVideo); isLive {
				// 直播预览流
				screenResult.VideoType = "live-preview"
				// TODO
				// screenResult.Live = feedVideo
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
				screenResult.VideoType = "feed"
				screenResult.Feed = feedVideo
				crawler.FeedCount++
				log.Info().
					Strs("tags", screenResult.Tags).
					Interface("feed", screenResult.Feed).
					Msg("found feed success")

				// get simulation play duration
				if screenResult.Feed.SimulationPlayDuration != 0 {
					screenResult.Feed.PlayDuration = screenResult.Feed.SimulationPlayDuration
				} else {
					screenResult.Feed.RandomPlayDuration = getSimulationDuration(crawler.configs.Feed.SleepRandom)
					screenResult.Feed.PlayDuration = screenResult.Feed.RandomPlayDuration
				}

				// simulation watch feed video
				sleepStrict(swipeFinishTime, screenResult.Feed.PlayDuration)
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

			// reset failed count
			crawler.failedCount = 0
		}
	}
}

type FeedVideo struct {
	// 视频基础数据
	VideoID       string `json:"video_id"`        // 视频 video ID
	URL           string `json:"url"`             // 视频 url
	UserName      string `json:"user_name"`       // 视频作者
	Duration      int64  `json:"duration"`        // 视频时长(ms)
	Caption       string `json:"caption"`         // 视频文案
	Type          string `json:"type"`            // 视频类型
	IsLivePreview bool   `json:"is_live_preview"` // 是否直播预览流

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
	LiveStreamID  string `json:"live_stream_id"`  // 直播流 ID
	LiveStreamURL string `json:"live_stream_url"` // 直播流地址
	UserName      string `json:"user_name"`       // 主播名称（无法获取？）
	LiveType      string `json:"live_type"`       // 直播类型

	// 视频热度数据
	AudienceCount int64 `json:"audience_count"` // 直播间人数

	// 网络数据
	ThroughputKbps int64 `json:"throughput_kbps"` // 网速

	// 记录仿真决策信息
	WatchDuration           int64 `json:"watch_duration"`            // 观播时长(ms)，取自 Simulation/Random
	SimulationWatchDuration int64 `json:"simulation_watch_duration"` // 仿真观播时长(ms)
	RandomWatchDuration     int64 `json:"random_watch_duration"`     // 随机观播时长(ms)
}

func (vc *VideoCrawler) getCurrentFeedVideo() (feedVideo *FeedVideo, err error) {
	if !vc.driverExt.plugin.Has("GetCurrentFeedVideo") {
		return nil, errors.New("plugin missing GetCurrentFeedVideo method")
	}

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

	log.Info().
		Interface("feedVideoCaption", feedVideo.Caption).
		Msg("get current feed video success")
	return feedVideo, nil
}

func (vc *VideoCrawler) getCurrentLiveRoom() (liveRoom *LiveRoom, err error) {
	if !vc.driverExt.plugin.Has("GetCurrentLiveRoom") {
		return nil, errors.New("plugin missing GetCurrentLiveRoom method")
	}

	resp, err := vc.driverExt.plugin.Call("GetCurrentLiveRoom")
	if err != nil {
		return nil, errors.Wrap(err, "call plugin GetCurrentLiveRoom failed")
	}

	if resp == nil {
		return nil, errors.New("live not found")
	}

	liveBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.New("json marshal live room info failed")
	}

	liveRoom = &LiveRoom{}
	err = json.Unmarshal(liveBytes, liveRoom)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal live room info failed")
	}

	log.Info().
		Interface("liveRoomUserName", liveRoom.UserName).
		Msg("get current live room success")
	return liveRoom, nil
}
