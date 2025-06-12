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
	resultsPath             string
	downloadsPath           string
	screenShotsPath         string
	StartTime               time.Time
	ActionLogFilePath       string
	DeviceActionLogFilePath string
	mu                      sync.Mutex
}

var (
	globalConfig  *Config
	getConfigOnce sync.Once
)

func GetConfig() *Config {
	getConfigOnce.Do(func() {
		cfg := &Config{
			StartTime: time.Now(),
		}

		var err error
		cfg.RootDir, err = os.Getwd()
		if err != nil {
			log.Fatal().Err(err).Msg("get current working directory failed")
		}

		startTimeStr := cfg.StartTime.Format("20060102150405")
		resultsDir := filepath.Join(ResultsDirName, startTimeStr)
		cfg.resultsPath = filepath.Join(cfg.RootDir, resultsDir)
		cfg.downloadsPath = filepath.Join(cfg.RootDir, filepath.Join(DownloadsDirName, startTimeStr))
		cfg.screenShotsPath = filepath.Join(cfg.resultsPath, ScreenshotsDirName)
		cfg.ActionLogFilePath = filepath.Join(resultsDir, ActionLogDirName)
		cfg.DeviceActionLogFilePath = "/sdcard/Android/data/io.appium.uiautomator2.server/files/hodor"

		globalConfig = cfg
	})

	return globalConfig
}

// ResultsPath returns the results path and creates the directory if it doesn't exist
func (c *Config) ResultsPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(c.resultsPath); os.IsNotExist(err) {
		if err := builtin.EnsureFolderExists(c.resultsPath); err != nil {
			log.Error().Err(err).Str("path", c.resultsPath).Msg("failed to create results directory")
		} else {
			log.Info().Str("path", c.resultsPath).Msg("created results folder")
		}
	}
	return c.resultsPath
}

// DownloadsPath returns the downloads path and creates the directory if it doesn't exist
func (c *Config) DownloadsPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(c.downloadsPath); os.IsNotExist(err) {
		if err := builtin.EnsureFolderExists(c.downloadsPath); err != nil {
			log.Error().Err(err).Str("path", c.downloadsPath).Msg("failed to create downloads directory")
		} else {
			log.Info().Str("path", c.downloadsPath).Msg("created downloads folder")
		}
	}
	return c.downloadsPath
}

// ScreenShotsPath returns the screenshots path and creates the directory if it doesn't exist
func (c *Config) ScreenShotsPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(c.screenShotsPath); os.IsNotExist(err) {
		if err := builtin.EnsureFolderExists(c.screenShotsPath); err != nil {
			log.Error().Err(err).Str("path", c.screenShotsPath).Msg("failed to create screenshots directory")
		} else {
			log.Info().Str("path", c.screenShotsPath).Msg("created screenshots folder")
		}
	}
	return c.screenShotsPath
}
