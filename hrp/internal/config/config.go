package config

import (
	"os"
	"path/filepath"
	"time"
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
}
