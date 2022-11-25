package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gidevice"
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

func initIOSDevice() uixt.Device {
	perfOptions := []gidevice.PerfOption{}
	for _, p := range perf {
		switch p {
		case "sys_cpu":
			perfOptions = append(perfOptions, hrp.WithPerfSystemCPU(true))
		case "sys_mem":
			perfOptions = append(perfOptions, hrp.WithPerfSystemMem(true))
		case "sys_net":
			perfOptions = append(perfOptions, hrp.WithPerfSystemNetwork(true))
		case "sys_disk":
			perfOptions = append(perfOptions, hrp.WithPerfSystemDisk(true))
		case "network":
			perfOptions = append(perfOptions, hrp.WithPerfNetwork(true))
		case "fps":
			perfOptions = append(perfOptions, hrp.WithPerfFPS(true))
		case "gpu":
			perfOptions = append(perfOptions, hrp.WithPerfGPU(true))
		}
	}

	device, err := uixt.NewIOSDevice(
		uixt.WithUDID(uuid),
		uixt.WithWDAPort(8700), uixt.WithWDAMjpegPort(8800),
		uixt.WithResetHomeOnStartup(false), // not reset home on startup
		uixt.WithPerfOptions(perfOptions...),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init ios device")
	}
	return device
}

func initAndroidDevice() uixt.Device {
	device, err := uixt.NewAndroidDevice(uixt.WithSerialNumber(uuid))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init android device")
	}
	return device
}

type timeLog struct {
	UTCTime         int64  `json:"utc_time"`
	LiveTime        string `json:"live_time"`
	LiveTimeSeconds int    `json:"live_time_seconds"`
}

type WorldCupLive struct {
	driver    *uixt.DriverExt
	done      chan bool
	file      *os.File
	resultDir string
	UUID      string    `json:"uuid"`
	MatchName string    `json:"matchName"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	Interval  int       `json:"interval"` // seconds
	Duration  int       `json:"duration"` // seconds
	Summary   []timeLog `json:"summary"`
	PerfData  []string  `json:"perfData"`
}

func NewWorldCupLive(matchName, osType string, duration, interval int) *WorldCupLive {
	var device uixt.Device
	log.Info().Str("osType", osType).Msg("init device")
	if osType == "ios" {
		device = initIOSDevice()
	} else {
		device = initAndroidDevice()
	}

	driverExt, err := device.NewDriver(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init driver")
	}

	if matchName == "" {
		matchName = "unknown-match"
	}

	startTime := time.Now()
	matchName = fmt.Sprintf("%s-%s", startTime.Format("2006-01-02"), matchName)
	resultDir := filepath.Join("worldcup-archives", matchName, startTime.Format("15:04:05"))

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
	ocrTexts, err := wc.driver.GetTextsByOCR(uixt.WithScreenShot(fileName))
	if err != nil {
		log.Error().Err(err).Msg("get ocr texts failed")
		return err
	}

	var liveTimeSeconds int
	for _, ocrText := range ocrTexts {
		seconds, err := convertTimeToSeconds(ocrText.Text)
		if err == nil {
			liveTimeSeconds = seconds
			line := fmt.Sprintf("%s\t%d\t%s\t%d\n",
				utcTimeStr, utcTime.Unix(), ocrText.Text, liveTimeSeconds)
			log.Info().Str("utcTime", utcTimeStr).Str("liveTime", ocrText.Text).Msg("log live time")
			if _, err := wc.file.WriteString(line); err != nil {
				log.Error().Err(err).Str("line", line).Msg("write timeseries failed")
			}
			wc.Summary = append(wc.Summary, timeLog{
				UTCTime:         utcTime.Unix(),
				LiveTime:        ocrText.Text,
				LiveTimeSeconds: liveTimeSeconds,
			})
			break
		}
	}
	return nil
}

func (wc *WorldCupLive) Start() {
	wc.done = make(chan bool)
	go func() {
		for {
			select {
			case <-wc.done:
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
	}()
	time.Sleep(time.Duration(wc.Duration) * time.Second)
	wc.Stop()
}

func (wc *WorldCupLive) Stop() {
	wc.EndTime = time.Now().Format("2006-01-02 15:04:05")
	wc.done <- true
}

func (wc *WorldCupLive) DumpResult() error {
	// init json encoder
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	wc.PerfData = wc.driver.GetPerfData()
	err := encoder.Encode(wc)
	if err != nil {
		return err
	}

	filename := filepath.Join(wc.resultDir, "summary.json")
	err = os.WriteFile(filename, buffer.Bytes(), 0o755)
	if err != nil {
		log.Error().Err(err).Msg("dump json path failed")
		return err
	}
	return nil
}
