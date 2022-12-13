package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
)

func convertTimeToSeconds(timeStr string) (int, error) {
	if !strings.Contains(timeStr, ":") {
		return 0, fmt.Errorf("invalid time string: %s", timeStr)
	}

	ss := strings.Split(timeStr, ":")
	var seconds int
	for idx, s := range ss {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		seconds += i * int(math.Pow(60, float64(len(ss)-idx-1)))
	}

	return seconds, nil
}

func initIOSDevice(uuid string) uixt.Device {
	perfOptions := []uixt.IOSPerfOption{}
	for _, p := range perf {
		switch p {
		case "sys_cpu":
			perfOptions = append(perfOptions, uixt.WithIOSPerfSystemCPU(true))
		case "sys_mem":
			perfOptions = append(perfOptions, uixt.WithIOSPerfSystemMem(true))
		case "sys_net":
			perfOptions = append(perfOptions, uixt.WithIOSPerfSystemNetwork(true))
		case "sys_disk":
			perfOptions = append(perfOptions, uixt.WithIOSPerfSystemDisk(true))
		case "network":
			perfOptions = append(perfOptions, uixt.WithIOSPerfNetwork(true))
		case "fps":
			perfOptions = append(perfOptions, uixt.WithIOSPerfFPS(true))
		case "gpu":
			perfOptions = append(perfOptions, uixt.WithIOSPerfGPU(true))
		}
	}
	perfOptions = append(perfOptions, uixt.WithIOSPerfOutputInterval(interval*1000))

	device, err := uixt.NewIOSDevice(
		uixt.WithUDID(uuid),
		uixt.WithWDAPort(8700), uixt.WithWDAMjpegPort(8800),
		uixt.WithResetHomeOnStartup(false), // not reset home on startup
		uixt.WithIOSPerfOptions(perfOptions...),
		uixt.WithXCTest("com.gtf.wda.runner.xctrunner"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init ios device")
	}
	return device
}

func initAndroidDevice(uuid string) uixt.Device {
	device, err := uixt.NewAndroidDevice(uixt.WithSerialNumber(uuid))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init android device")
	}
	return device
}

type timeLog struct {
	UTCTimeStr      string `json:"utc_time_str"`
	UTCTime         int64  `json:"utc_time"`
	LiveTime        string `json:"live_time"`
	LiveTimeSeconds int    `json:"live_time_seconds"`
}

type WorldCupLive struct {
	driver    *uixt.DriverExt
	file      *os.File
	resultDir string
	UUID      string    `json:"uuid"`
	MatchName string    `json:"matchName"`
	BundleID  string    `json:"bundleID"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	Interval  int       `json:"interval"` // seconds
	Duration  int       `json:"duration"` // seconds
	Timelines []timeLog `json:"timelines"`
	PerfData  []string  `json:"perfData"`
}

func NewWorldCupLive(device uixt.Device, matchName, bundleID string, duration, interval int) *WorldCupLive {
	driverExt, err := device.NewDriver(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init driver")
	}

	if matchName == "" {
		matchName = "unknown-match"
	}

	startTime := time.Now()
	matchName = fmt.Sprintf("%s-%s", startTime.Format("2006-01-02"), matchName)
	resultDir := filepath.Join("worldcuplive", matchName, startTime.Format("15:04:05"))

	if err = os.MkdirAll(filepath.Join(resultDir, "screenshot"), 0o755); err != nil {
		log.Fatal().Err(err).Msg("failed to create result dir")
	}

	filename := filepath.Join(resultDir, "log.txt")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o755)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open file")
	}
	// write title
	f.WriteString(fmt.Sprintf("%s\t%s\t%s\n", matchName, device.UUID(), bundleID))
	f.WriteString("utc_time\tutc_timestamp\tlive_time\tlive_seconds\n")

	if interval == 0 {
		interval = 15
	}
	if duration == 0 {
		duration = 30
	}

	return &WorldCupLive{
		driver:    driverExt,
		file:      f,
		resultDir: resultDir,
		UUID:      device.UUID(),
		BundleID:  bundleID,
		Duration:  duration,
		Interval:  interval,
		StartTime: startTime.Format("2006-01-02 15:04:05"),
		MatchName: matchName,
	}
}

func (wc *WorldCupLive) getCurrentLiveTime(utcTime time.Time) error {
	utcTimeStr := utcTime.Format("15:04:05")
	fileName := filepath.Join(
		wc.resultDir, "screenshot", utcTimeStr)
	ocrTexts, err := wc.driver.GetTextsByOCR(
		uixt.WithScreenShot(fileName),
	)
	if err != nil {
		log.Error().Err(err).Msg("get ocr texts failed")
		return err
	}

	// filter ocr texts with time format
	secondsMap := map[string]int{}
	var secondsTexts []string
	for _, ocrText := range ocrTexts {
		seconds, err := convertTimeToSeconds(ocrText.Text)
		if err == nil {
			secondsTexts = append(secondsTexts, ocrText.Text)
			secondsMap[ocrText.Text] = seconds
		}
	}

	var secondsText string
	if len(secondsTexts) == 1 {
		secondsText = secondsTexts[0]
	} else if len(secondsTexts) >= 2 {
		// select the second, the first maybe mobile system time
		secondsText = secondsTexts[1]
	} else {
		log.Warn().Msg("no time text found")
		return nil
	}

	liveTimeSeconds := secondsMap[secondsText]
	line := fmt.Sprintf("%s\t%d\t%s\t%d\n",
		utcTimeStr, utcTime.Unix(), secondsText, liveTimeSeconds)
	log.Info().Str("utcTime", utcTimeStr).Str("liveTime", secondsText).Msg("log live time")
	if _, err := wc.file.WriteString(line); err != nil {
		log.Error().Err(err).Str("line", line).Msg("write timeseries failed")
	}
	wc.Timelines = append(wc.Timelines, timeLog{
		UTCTimeStr:      utcTimeStr,
		UTCTime:         utcTime.Unix(),
		LiveTime:        secondsText,
		LiveTimeSeconds: liveTimeSeconds,
	})
	return nil
}

func (wc *WorldCupLive) EnterLive(bundleID string) error {
	log.Info().Msg("enter world cup live")

	// kill app
	_, err := wc.driver.Driver.AppTerminate(bundleID)
	if err != nil {
		log.Error().Err(err).Msg("terminate app failed")
	}

	// launch app
	err = wc.driver.Driver.AppLaunch(bundleID)
	if err != nil {
		log.Error().Err(err).Msg("launch app failed")
		return err
	}
	time.Sleep(5 * time.Second)

	// 青少年弹窗处理
	if points, err := wc.driver.GetTextXYs([]string{"青少年模式", "我知道了"}); err == nil {
		_ = wc.driver.TapAbsXY(points[1].X, points[1].Y)
	}

	// 进入世界杯 tab
	if err = wc.driver.TapByOCR("世界杯"); err != nil {
		log.Error().Err(err).Msg("enter 直播中 failed")
		return err
	}
	time.Sleep(3 * time.Second)

	// 进入世界杯直播
	if err = wc.driver.TapByOCR("直播中"); err != nil {
		log.Error().Err(err).Msg("enter 直播中 failed")
		return err
	}

	return nil
}

func (wc *WorldCupLive) Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	timer := time.NewTimer(time.Duration(wc.Duration) * time.Second)
	for {
		select {
		case <-timer.C:
			wc.dumpResult()
			return
		case <-c:
			wc.dumpResult()
			return
		default:
			utcTime := time.Now()
			if utcTime.Unix()%int64(wc.Interval) == 0 {
				wc.getCurrentLiveTime(utcTime)
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func (wc *WorldCupLive) dumpResult() error {
	wc.EndTime = time.Now().Format("2006-01-02 15:04:05")

	// init json encoder
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	wc.PerfData = wc.driver.GetPerfData()
	err := encoder.Encode(wc)
	if err != nil {
		log.Error().Err(err).Msg("encode json failed")
		return err
	}

	path := filepath.Join(wc.resultDir, "summary.json")
	err = os.WriteFile(path, buffer.Bytes(), 0o755)
	if err != nil {
		log.Error().Err(err).Msg("dump json path failed")
		return err
	}
	log.Info().Str("path", path).Msg("dump summary success")
	return nil
}
