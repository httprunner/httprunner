package uixt

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/httprunner/funplugin/myexec"
	"github.com/rs/zerolog/log"
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
	file      *os.File
	resultDir string
	UUID      string    `json:"uuid"`
	AppName   string    `json:"appName"`
	BundleID  string    `json:"bundleID"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	Interval  int       `json:"interval"` // seconds
	Duration  int       `json:"duration"` // seconds
	Timelines []timeLog `json:"timelines"`
	PerfFile  string    `json:"perf"`
}

func (dExt *DriverExt) CollectEndToEndDelay(options ...ActionOption) {
	dataOptions := NewActionOptions(options...)
	var err error
	startTime := time.Now()
	resultDir := filepath.Join("endtoenddelay", startTime.Format("2006-01-02 15:04:05"))

	if err = os.MkdirAll(filepath.Join(resultDir, "screenshot"), 0o755); err != nil {
		log.Fatal().Err(err).Msg("failed to create result dir")
	}

	filename := filepath.Join(resultDir, "log.txt")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o755)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open file")
	}
	// write title
	f.WriteString("utc_time\tutc_timestamp\tlive_time\tlive_seconds\n")

	if dataOptions.Interval == 0 {
		dataOptions.Interval = 5
	}
	if dataOptions.Timeout == 0 {
		dataOptions.Timeout = 60
	}

	endToEndDelay := &EndToEndDelay{
		driver:    dExt,
		file:      f,
		resultDir: resultDir,
		Duration:  int(dataOptions.Timeout),
		Interval:  int(dataOptions.Interval),
		StartTime: startTime.Format("2006-01-02 15:04:05"),
	}

	SntpCheckTime()

	endToEndDelay.Start()

	dExt.cacheStepData.e2eDelay = endToEndDelay.Timelines
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
		if strings.HasPrefix(ocrText.Text, "16") &&
			len(ocrText.Text) > 8 &&
			!strings.Contains(ocrText.Text, ":") {
			liveTimeTexts = append(liveTimeTexts, ocrText.Text)
		}
	}

	var liveTimeText string
	if len(liveTimeTexts) != 0 {
		liveTimeText = liveTimeTexts[0]
	} else {
		log.Warn().Msg("no time text found")
		return nil
	}

	if len(liveTimeText) < 13 {
		for (13 - len(liveTimeText)) > 0 {
			liveTimeText += "0"
		}
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
	line := fmt.Sprintf("%s\t%d\t%s\t%s\n",
		utcTimeStr, utcTime.UnixMicro(), liveTimeStr, liveTimeText)
	log.Info().
		Str("utcTime", utcTimeStr).
		Int64("utcTimeInt", utcTime.UnixMilli()).
		Str("liveTime", liveTimeStr).
		Int64("liveTimeInt", int64(liveTimeInt)).
		Float64("delay", float64(utcTime.UnixMilli()-int64(liveTimeInt))/1000).
		Msg("log live time")
	if _, err := ete.file.WriteString(line); err != nil {
		log.Error().Err(err).Str("line", line).Msg("write timeseries failed")
	}
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
			return
		case <-c:
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

func SntpCheckTime() {
	err := myexec.RunCommand("sudo", "sntp", "-sS", "time.asia.apple.com")
	if err != nil {
		log.Error().Err(err).Msg("failed to synchronize time using sntp")
	}
}
