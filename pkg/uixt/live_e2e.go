package uixt

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

type timeLog struct {
	UTCTimeStr  string  `json:"utc_time_str"`
	UTCTime     int64   `json:"utc_time"`
	LiveTimeStr string  `json:"live_time_str"`
	LiveTime    int64   `json:"live_time"`
	Delay       float64 `json:"delay"`
}

type EndToEndDelay struct {
	driver    *DriverExt
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	Interval  int       `json:"interval"` // seconds
	Duration  int       `json:"duration"` // seconds
	Timelines []timeLog `json:"timelines"`
}

func CollectEndToEndDelay(dExt *DriverExt, opts ...option.ActionOption) {
	dataOptions := option.NewActionOptions(opts...)
	startTime := time.Now()

	if dataOptions.Interval == 0 {
		dataOptions.Interval = 5
	}
	if dataOptions.Timeout == 0 {
		dataOptions.Timeout = 60
	}

	endToEndDelay := &EndToEndDelay{
		driver:    dExt,
		Duration:  int(dataOptions.Timeout),
		Interval:  int(dataOptions.Interval),
		StartTime: startTime.Format("2006-01-02 15:04:05"),
	}

	endToEndDelay.Start()

	// TODO: remove
	dExt.Driver.GetSession().e2eDelay = endToEndDelay.Timelines
}

func (ete *EndToEndDelay) getCurrentLiveTime(utcTime time.Time) error {
	utcTimeStr := utcTime.Format("2006-01-02 15:04:05")
	ocrTexts, err := ete.driver.GetScreenTexts()
	if err != nil {
		log.Error().Err(err).Msg("get ocr texts failed")
		return err
	}

	// filter ocr texts with time format
	var liveTimeTexts []string
	for _, ocrText := range ocrTexts {
		if len(ocrText.Text) < 13 || strings.Contains(ocrText.Text, ":") {
			continue
		}
		// exclude digit(s) recognized as letter(s)
		_, errParseInt := strconv.ParseInt(ocrText.Text[:13], 10, 64)
		if errParseInt != nil {
			continue
		}
		liveTimeTexts = append(liveTimeTexts, ocrText.Text)
	}

	var liveTimeText string
	if len(liveTimeTexts) != 0 {
		liveTimeText = liveTimeTexts[0]
	} else {
		log.Warn().Msg("no time text found")
		return nil
	}

	liveTimeInt, err := strconv.Atoi(liveTimeText)
	if err != nil {
		liveTimeInt = 0
	}
	liveTimeSInt, err := strconv.Atoi(liveTimeText[:10])
	if err != nil {
		liveTimeSInt = 0
	}
	liveTimeNSInt, err := strconv.Atoi(liveTimeText[10:13])
	if err != nil {
		liveTimeNSInt = 0
	}
	liveTimeStr := time.Unix(int64(liveTimeSInt), int64(liveTimeNSInt*1000*1000)).Format("2006-01-02 15:04:05")
	log.Info().
		Str("utcTime", utcTimeStr).
		Int64("utcTimeInt", utcTime.UnixMilli()).
		Str("liveTime", liveTimeStr).
		Int64("liveTimeInt", int64(liveTimeInt)).
		Float64("delay", float64(utcTime.UnixMilli()-int64(liveTimeInt))/1000).
		Msg("log live time")
	ete.Timelines = append(ete.Timelines, timeLog{
		UTCTimeStr:  utcTimeStr,
		UTCTime:     utcTime.UnixMilli(),
		LiveTimeStr: liveTimeStr,
		LiveTime:    int64(liveTimeInt),
		Delay:       float64(utcTime.UnixMilli()-int64(liveTimeInt)) / 1000,
	})
	return nil
}

func (ete *EndToEndDelay) Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	timer := time.NewTimer(time.Duration(ete.Duration) * time.Second)
	for {
		select {
		case <-timer.C:
			ete.EndTime = time.Now().Format("2006-01-02 15:04:05")
			return
		case <-c:
			ete.EndTime = time.Now().Format("2006-01-02 15:04:05")
			return
		default:
			utcTime := time.Now()
			if utcTime.Unix()%int64(ete.Interval) == 0 {
				_ = ete.getCurrentLiveTime(utcTime)
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
