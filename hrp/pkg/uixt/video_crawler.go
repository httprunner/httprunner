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
	lastVideo   *Video

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

func (vc *VideoCrawler) checkLiveVideo(video *Video) (enterPoint PointF, yes bool) {
	if video.Type != VideoType_PreviewLive {
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

			liveRoom, err := vc.getCurrentVideo()
			if err != nil {
				if vc.failedCount >= 5 {
					// failed 5 consecutive times
					return errors.New("get current live event trackings failed 5 consecutive times")
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
			screenResult.Video = liveRoom
			vc.LiveCount++
			log.Info().Strs("tags", screenResult.Tags).
				Interface("live", screenResult.Video).
				Msg("found live success")

			// get simulation watch duration
			if screenResult.Video.SimulationPlayDuration != 0 {
				screenResult.Video.PlayDuration = screenResult.Video.SimulationPlayDuration
			} else {
				screenResult.Video.RandomPlayDuration = getSimulationDuration(vc.configs.Live.SleepRandom)
				screenResult.Video.PlayDuration = screenResult.Video.RandomPlayDuration
			}
			// simulation watch live video
			sleepStrict(swipeFinishTime, screenResult.Video.PlayDuration)

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
		lastVideo:   &Video{},

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
			feedVideo, err := crawler.getCurrentVideo()
			if err != nil || feedVideo.Type == "" {
				if crawler.failedCount >= 10 {
					// failed 10 consecutive times
					return errors.New("get current feed video failed 10 consecutive times")
				}
				log.Warn().Interface("feedVideo", feedVideo).Msg("get current feed video failed")
				// retry
				crawler.failedCount++
				continue
			}

			if feedVideo.VideoID == crawler.lastVideo.VideoID {
				// app event tracking not changed
				// check and handle popups
				log.Warn().Msg("feed video event tracking not changed")
				if err = crawler.driverExt.ClosePopupsHandler(WithMaxRetryTimes(1)); err != nil {
					return err
				}
			}
			crawler.lastVideo = feedVideo

			screenResult := &ScreenResult{
				Resolution: dExt.windowSize,
			}
			dExt.cacheStepData.screenResults[time.Now().String()] = screenResult

			// check if live video && run live crawler
			if enterPoint, isLive := crawler.checkLiveVideo(feedVideo); isLive {
				// 直播预览流
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
				screenResult.Video = feedVideo
				crawler.FeedCount++
				log.Info().
					Strs("tags", screenResult.Tags).
					Interface("feed", screenResult.Video).
					Msg("found feed success")

				// get simulation play duration
				if screenResult.Video.SimulationPlayDuration != 0 {
					screenResult.Video.PlayDuration = screenResult.Video.SimulationPlayDuration
				} else {
					screenResult.Video.RandomPlayDuration = getSimulationDuration(crawler.configs.Feed.SleepRandom)
					screenResult.Video.PlayDuration = screenResult.Video.RandomPlayDuration
				}

				// simulation watch feed video
				sleepStrict(swipeFinishTime, screenResult.Video.PlayDuration)
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

type VideoType string

const (
	VideoType_Feed        VideoType = "FEED"
	VideoType_PreviewLive VideoType = "PREVIEW-LIVE" // 直播预览流
	VideoType_Live        VideoType = "LIVE"
	VideoType_Image       VideoType = "IMAGE"
)

type Video struct {
	Type VideoType `json:"type" required:"true"` // 视频类型, feed/preview-live/live/image

	// Feed 视频基础数据
	CacheKey string `json:"cache_key,omitempty"` // cachekey
	VideoID  string `json:"video_id,omitempty"`  // 视频 video ID
	URL      string `json:"feed_url,omitempty"`  // 实际播放的视频 url
	UserName string `json:"user_name"`           // 视频作者
	Duration int64  `json:"duration,omitempty"`  // 视频时长(ms)
	Caption  string `json:"caption,omitempty"`   // 视频文案
	// 视频热度数据
	ViewCount    int64 `json:"view_count,omitempty"`    // feed 观看数
	LikeCount    int64 `json:"like_count,omitempty"`    // feed 点赞数
	CommentCount int64 `json:"comment_count,omitempty"` // feed 评论数
	CollectCount int64 `json:"collect_count,omitempty"` // feed 收藏数
	ForwardCount int64 `json:"forward_count,omitempty"` // feed 转发数
	ShareCount   int64 `json:"share_count,omitempty"`   // feed 分享数

	// timelines
	PublishTimestamp int64 `json:"publish_timestamp,omitempty"` // feed 发布时间戳
	PreloadTimestamp int64 `json:"preload_timestamp,omitempty"` // feed 预加载时间戳

	// Live 视频基础数据
	LiveStreamID  string `json:"live_stream_id,omitempty"`  // 直播流 ID
	LiveStreamURL string `json:"live_stream_url,omitempty"` // 直播流地址
	LiveType      string `json:"live_type,omitempty"`       // 直播间类型
	// 网络数据
	ThroughputKbps int64 `json:"throughput_kbps,omitempty"` // 网速
	// 视频热度数据
	AudienceCount int64 `json:"audience_count,omitempty"` // 直播间人数

	// Image
	// ImageURLs []string

	// 记录仿真决策信息
	PlayDuration           int64   `json:"play_duration"`            // 播放时长(ms)，取自 Simulation/Random
	SimulationPlayProgress float64 `json:"simulation_play_progress"` // 仿真播放比例（完播率）
	SimulationPlayDuration int64   `json:"simulation_play_duration"` // 仿真播放时长(ms)
	RandomPlayDuration     int64   `json:"random_play_duration"`     // 随机播放时长(ms)
}

func (vc *VideoCrawler) getCurrentVideo() (video *Video, err error) {
	if !vc.driverExt.plugin.Has("GetCurrentVideo") {
		return nil, errors.New("plugin missing GetCurrentVideo method")
	}

	resp, err := vc.driverExt.plugin.Call("GetCurrentVideo")
	if err != nil {
		return nil, errors.Wrap(err, "call plugin GetCurrentVideo failed")
	}

	if resp == nil {
		return nil, errors.New("video not found")
	}

	feedBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.New("json marshal video info failed")
	}

	video = &Video{}
	err = json.Unmarshal(feedBytes, video)
	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal video info failed")
	}

	log.Info().
		Interface("videoCaption", video.Caption).
		Msg("get current video success")
	return video, nil
}
