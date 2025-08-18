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
	// results directory names
	ResultsDirName     = "results"     // $PWD/results/
	DownloadsDirName   = "downloads"   // $PWD/results/20060102150405/downloads/
	ScreenshotsDirName = "screenshots" // $PWD/results/20060102150405/screenshots/
	ActionLogDirName   = "action_log"  // $PWD/results/20060102150405/action_log/

	// results file names
	SummaryFileName = "hrp_summary.json" // $PWD/results/20060102150405/hrp_summary.json
	LogFileName     = "hrp.log"          // $PWD/results/20060102150405/hrp.log
	ReportFileName  = "report.html"      // $PWD/results/20060102150405/report.html
	CaseFileName    = "case.json"        // $PWD/results/20060102150405/case.json

	// mobile device path
	DeviceActionLogFilePath = "/storage/emulated/0/Download/"
)

type Config struct {
	StartTime        time.Time
	RootDir          string
	resultsPath      string
	downloadsPath    string
	screenShotsPath  string
	summaryFilePath  string
	logFilePath      string
	reportFilePath   string
	caseFilePath     string
	actionLogDirPath string
	mu               sync.Mutex
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
		cfg.downloadsPath = filepath.Join(cfg.RootDir, filepath.Join(DownloadsDirName))
		cfg.screenShotsPath = filepath.Join(cfg.resultsPath, ScreenshotsDirName)
		cfg.actionLogDirPath = filepath.Join(resultsDir, ActionLogDirName)
		globalConfig = cfg
	})

	return globalConfig
}

// resultsPathUnlocked returns the results path and creates the directory if it doesn't exist (internal use, no lock)
func (c *Config) resultsPathUnlocked() string {
	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(c.resultsPath); os.IsNotExist(err) {
		if err := builtin.EnsureFolderExists(c.resultsPath); err != nil {
			log.Fatal().Err(err).Str("path", c.resultsPath).Msg("failed to create results directory")
		} else {
			log.Info().Str("path", c.resultsPath).Msg("created results folder")
		}
	}
	return c.resultsPath
}

// ResultsPath returns the results path and creates the directory if it doesn't exist
func (c *Config) ResultsPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.resultsPathUnlocked()
}

// DownloadsPath returns the downloads path and creates the directory if it doesn't exist
func (c *Config) DownloadsPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(c.downloadsPath); os.IsNotExist(err) {
		if err := builtin.EnsureFolderExists(c.downloadsPath); err != nil {
			log.Fatal().Err(err).Str("path", c.downloadsPath).Msg("failed to create downloads directory")
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
			log.Fatal().Err(err).Str("path", c.screenShotsPath).Msg("failed to create screenshots directory")
		} else {
			log.Info().Str("path", c.screenShotsPath).Msg("created screenshots folder")
		}
	}
	return c.screenShotsPath
}

// $PWD/results/20060102150405/action_log/
func (c *Config) ActionLogDirPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if directory exists, create if it doesn't
	if _, err := os.Stat(c.actionLogDirPath); os.IsNotExist(err) {
		if err := builtin.EnsureFolderExists(c.actionLogDirPath); err != nil {
			log.Fatal().Err(err).Str("path", c.actionLogDirPath).Msg("failed to create action log directory")
		} else {
			log.Info().Str("path", c.actionLogDirPath).Msg("created action log folder")
		}
	}
	return c.actionLogDirPath
}

// $PWD/results/20060102150405/hrp_summary.json
func (c *Config) SummaryFilePath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.summaryFilePath != "" {
		return c.summaryFilePath
	}

	// Ensure directory creation and set cached path
	c.summaryFilePath = filepath.Join(c.resultsPathUnlocked(), SummaryFileName)
	return c.summaryFilePath
}

// $PWD/results/20060102150405/hrp.log
func (c *Config) LogFilePath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.logFilePath != "" {
		return c.logFilePath
	}

	// Ensure directory creation and set cached path
	c.logFilePath = filepath.Join(c.resultsPathUnlocked(), LogFileName)
	return c.logFilePath
}

// $PWD/results/20060102150405/report.html
func (c *Config) ReportFilePath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.reportFilePath != "" {
		return c.reportFilePath
	}

	// Ensure directory creation and set cached path
	c.reportFilePath = filepath.Join(c.resultsPathUnlocked(), ReportFileName)
	return c.reportFilePath
}

// $PWD/results/20060102150405/case.json
func (c *Config) CaseFilePath() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.caseFilePath != "" {
		return c.caseFilePath
	}

	// Ensure directory creation and set cached path
	c.caseFilePath = filepath.Join(c.resultsPathUnlocked(), CaseFileName)
	return c.caseFilePath
}
