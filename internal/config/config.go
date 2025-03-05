package config

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/internal/builtin"
)

const (
	ResultsDirName     = "results"
	DownloadsDirName   = "downloads"
	ScreenshotsDirName = "screenshots"
	ActionLogDirName   = "action_log"
)

type Config struct {
	RootDir                 string
	ResultsDir              string
	ResultsPath             string
	DownloadsPath           string
	ScreenShotsPath         string
	StartTime               time.Time
	ActionLogFilePath       string
	DeviceActionLogFilePath string
}

var (
	globalConfig *Config
	once         sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg := &Config{
			StartTime: time.Now(),
		}

		var err error
		cfg.RootDir, err = os.Getwd()
		if err != nil {
			panic(err)
		}

		startTimeStr := cfg.StartTime.Format("20060102150405")
		cfg.ResultsDir = filepath.Join(ResultsDirName, startTimeStr)
		cfg.ResultsPath = filepath.Join(cfg.RootDir, cfg.ResultsDir)
		cfg.DownloadsPath = filepath.Join(cfg.RootDir, filepath.Join(DownloadsDirName, startTimeStr))
		cfg.ScreenShotsPath = filepath.Join(cfg.ResultsPath, ScreenshotsDirName)
		cfg.ActionLogFilePath = filepath.Join(cfg.ResultsDir, ActionLogDirName)
		cfg.DeviceActionLogFilePath = "/sdcard/Android/data/io.appium.uiautomator2.server/files/hodor"

		// create results directory
		if err := builtin.EnsureFolderExists(cfg.ResultsPath); err != nil {
			log.Fatal().Err(err).Msg("create results directory failed")
		}
		if err := builtin.EnsureFolderExists(cfg.DownloadsPath); err != nil {
			log.Fatal().Err(err).Msg("create downloads directory failed")
		}
		if err := builtin.EnsureFolderExists(cfg.ScreenShotsPath); err != nil {
			log.Fatal().Err(err).Msg("create screenshots directory failed")
		}

		globalConfig = cfg
	})

	return globalConfig
}
