package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
)

const (
	ResultsDirName     = "results"
	ScreenshotsDirName = "screenshots"
	ActionLogDireName  = "action_log"
)

var (
	RootDir                 string
	ResultsDir              string
	ResultsPath             string
	ScreenShotsPath         string
	StartTime               = time.Now()
	StartTimeStr            = StartTime.Format("20060102150405")
	ActionLogFilePath       string
	DeviceActionLogFilePath string
)

func init() {
	var err error
	RootDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	ResultsDir = filepath.Join(ResultsDirName, StartTimeStr)
	ResultsPath = filepath.Join(RootDir, ResultsDir)
	ScreenShotsPath = filepath.Join(ResultsPath, ScreenshotsDirName)
	ActionLogFilePath = filepath.Join(ResultsDir, ActionLogDireName)
	DeviceActionLogFilePath = "/sdcard/Android/data/io.appium.uiautomator2.server/files/hodor"

	// create results directory
	if err := builtin.EnsureFolderExists(ResultsPath); err != nil {
		log.Fatal().Err(err).Msg("create results directory failed")
	}
	if err := builtin.EnsureFolderExists(ScreenShotsPath); err != nil {
		log.Fatal().Err(err).Msg("create screenshots directory failed")
	}
}
