package uixt

import (
	"math"
	"math/rand"
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

func (vc *VideoCrawler) exitLiveRoom() error {
	log.Info().Msg("press back to exit live room")
	return vc.driverExt.Driver.PressBack()
}

const (
	FOUND_FEED_SUCCESS = "found feed success"
	FOUND_LIVE_SUCCESS = "found live success"
)

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
		FeedCount:   0,
		FeedStat:    make(map[string]int),
		LiveCount:   0,
		LiveStat:    make(map[string]int),
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
			if err = crawler.clearCurrentVideo(); err != nil {
				log.Error().Err(err).Msg("clear cache failed")
			}

			// swipe to next feed video
			log.Info().Msg("swipe to next feed video")
			swipeStartTime := time.Now()
			if err = dExt.SwipeRelative(0.9, 0.8, 0.9, 0.1, WithOffsetRandomRange(-10, 10)); err != nil {
				log.Error().Err(err).Msg("feed swipe up failed")
				return err
			}
			swipeFinishTime := time.Now()

			// get app event trackings
			// retry 10 times if get feed failed, abort if fail 10 consecutive times
			currentVideo, err := crawler.getCurrentVideo()
			if err != nil || currentVideo.Type == "" {
				crawler.failedCount++
				if crawler.failedCount >= 10 {
					// failed 10 consecutive times
					return errors.Wrap(code.TrackingGetError,
						"get current feed video failed 10 consecutive times")
				}
				log.Warn().
					Int64("failedCount", crawler.failedCount).
					Msg("get current feed video failed")

				// check and handle popups
				if err := crawler.driverExt.ClosePopupsHandler(); err != nil {
					return err
				}

				// retry
				continue
			}

			switch currentVideo.Type {
			case VideoType_PreviewLive:
				// 直播预览流
				var skipEnterLive bool
				if crawler.isLiveTargetAchieved() {
					log.Info().Interface("video", currentVideo).
						Msg("live count achieved, skip entering live room")
					skipEnterLive = true
				} else if rand.Float64() <= 0.10 {
					// 10% chance skip entering live room
					log.Info().Msg("skip entering preview live by 10% chance")
					skipEnterLive = true
				}

				if !skipEnterLive {
					time.Sleep(1 * time.Second)
					// enter live room
					entryPoint := PointF{
						X: float64(dExt.windowSize.Width / 2),
						Y: float64(dExt.windowSize.Height / 2),
					}

					log.Info().Msg("tap screen center to enter live room")
					if err := crawler.driverExt.TapAbsXY(entryPoint.X, entryPoint.Y,
						WithOffsetRandomRange(-20, 20)); err != nil {
						log.Error().Err(err).Msg("tap live video failed")
						continue
					}
				} else {
					// skip entering live room
					// only mock simulation play duration
					sleepTime := math.Min(float64(currentVideo.SimulationPlayDuration), float64(currentVideo.RandomPlayDuration))
					currentVideo.PlayDuration = int64(sleepTime)
				}

				fallthrough

			case VideoType_Live:
				// 直播
				crawler.LiveCount++
				log.Info().Interface("video", currentVideo).Msg(FOUND_LIVE_SUCCESS)

				// wait 3s for live loading
				time.Sleep(3 * time.Second)
				// take screenshot and get screen texts by OCR
				screenResult, err := crawler.driverExt.GetScreenResult(
					WithScreenShotOCR(true),
					WithScreenShotUpload(true),
					WithScreenShotLiveType(true),
				)
				if err != nil {
					log.Error().Err(err).Msg("get screen result failed")
					time.Sleep(3 * time.Second)
					continue
				}

				// add live type
				if screenResult.imageResult != nil &&
					screenResult.imageResult.LiveType != "" &&
					screenResult.imageResult.LiveType != "NoLive" {
					currentVideo.LiveType = screenResult.imageResult.LiveType
				}

				// simulation watch feed video
				sleepStrict(swipeFinishTime, currentVideo.PlayDuration)

				screenResult.Video = currentVideo
				screenResult.Resolution = dExt.windowSize
				screenResult.SwipeStartTime = swipeStartTime.UnixMilli()
				screenResult.SwipeFinishTime = swipeFinishTime.UnixMilli()
				screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()

				var exitLive bool
				if crawler.isLiveTargetAchieved() {
					log.Info().Interface("live", currentVideo).
						Msg("live count achieved, exit live room")
					exitLive = true
				} else if rand.Float64() <= 0.10 {
					// 10% chance exit live room
					log.Info().Msg("exit live room by 10% chance")
					exitLive = true
				}
				if exitLive && currentVideo.Type == VideoType_Live {
					err = crawler.exitLiveRoom()
					if err != nil {
						if errors.Is(err, code.TimeoutError) || errors.Is(err, code.InterruptError) {
							return err
						}
						log.Error().Err(err).Msg("run live crawler failed, continue")
					}
				}

			default:
				// 点播 || 图文 || 广告 || etc.
				crawler.FeedCount++
				log.Info().Interface("video", currentVideo).Msg(FOUND_FEED_SUCCESS)

				screenResult := &ScreenResult{
					Resolution: dExt.windowSize,
					Video:      currentVideo,

					// log swipe timelines
					SwipeStartTime:  swipeStartTime.UnixMilli(),
					SwipeFinishTime: swipeFinishTime.UnixMilli(),
				}
				dExt.cacheStepData.screenResults[time.Now().String()] = screenResult

				// simulation watch feed video
				sleepStrict(swipeFinishTime, currentVideo.PlayDuration)
				screenResult.TotalElapsed = time.Since(swipeFinishTime).Milliseconds()
			}

			// check if target count achieved
			if crawler.isTargetAchieved() {
				log.Info().Msg("target count achieved, exit crawler")
				return nil
			}

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
	Type     VideoType `json:"type" required:"true"` // 视频类型, feed/preview-live/live/image
	DataType string    `json:"data_type"`            // 数据源对应的事件名称

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

	// 图文数据
	ImageUrls []string `json:"image_urls,omitempty"` // 图片对应的 url 列表

	// 记录仿真决策信息
	PlayDuration           int64   `json:"play_duration"`            // 播放时长(ms)，取自 Simulation/Random
	SimulationPlayProgress float64 `json:"simulation_play_progress"` // 仿真播放比例（完播率）
	SimulationPlayDuration int64   `json:"simulation_play_duration"` // 仿真播放时长(ms)
	RandomPlayDuration     int64   `json:"random_play_duration"`     // 随机播放时长(ms)
}

func (vc *VideoCrawler) clearCurrentVideo() error {
	if !vc.driverExt.plugin.Has("ClearCurrentVideo") {
		return errors.New("plugin missing ClearCurrentVideo method")
	}

	_, err := vc.driverExt.plugin.Call("ClearCurrentVideo")
	if err != nil {
		return errors.Wrap(err, "call plugin ClearCurrentVideo failed")
	}

	return nil
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

	if video.Type == VideoType_Live || video.Type == VideoType_PreviewLive {
		video.RandomPlayDuration = getSimulationDuration(vc.configs.Live.SleepRandom)
	} else {
		video.RandomPlayDuration = getSimulationDuration(vc.configs.Feed.SleepRandom)
	}

	// get simulation play duration
	if video.SimulationPlayDuration != 0 {
		video.PlayDuration = video.SimulationPlayDuration
	} else {
		video.PlayDuration = video.RandomPlayDuration
	}

	log.Info().
		Str("type", string(video.Type)).
		Str("dataType", video.DataType).
		Msg("get current video success")
	return video, nil
}
