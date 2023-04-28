package uixt

import (
	"time"

	"github.com/rs/zerolog/log"
)

type VideoCrawlerConfigs struct {
	AppPackageName string `json:"app_package_name"`

	TargetFeedCount int `json:"target_feed_count"`
	TargetLiveCount int `json:"target_live_count"`
}

func (dExt *DriverExt) VideoCrawler(configs *VideoCrawlerConfigs) (err error) {
	// launch app
	if configs.AppPackageName != "" {
		if err = dExt.Driver.AppLaunch(configs.AppPackageName); err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}

	// loop until target count achieved
	for {
		// take screenshot and get screen texts by OCR
		texts, err := dExt.GetScreenTextsByOCR()
		if err != nil {
			log.Error().Err(err).Msg("OCR GetTexts failed")
			return err
		}

		// check if text popup exists
		if isTextPopup(texts) {
			log.Info().Msg("text popup found")
		}

		// check if live video
		if isLiveVideo(texts) {
			log.Info().Msg("live video found")
		}

		// assert feed video type

		// swipe to next video
		if err = dExt.SwipeUp(); err != nil {
			log.Error().Err(err).Msg("swipe up failed")
			return err
		}

		time.Sleep(5 * time.Second)
	}

	// return nil
}

func isTextPopup(texts OCRTexts) bool {
	texts.FindTexts([]string{"确定", "取消"})
	return false
}

func isLiveVideo(texts OCRTexts) bool {
	_, err := texts.FindTexts([]string{"点击进入直播间", "直播中"})
	return err == nil
}
