package uixt

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

const (
	CloseStatusFound   = "found"
	CloseStatusSuccess = "success"
	CloseStatusFail    = "fail"
)

// ClosePopupsResult represents the result of recognized popup to close
type ClosePopupsResult struct {
	Type      string `json:"type"`
	PopupArea Box    `json:"popupArea"`
	CloseArea Box    `json:"closeArea"`
	Text      string `json:"text"`
}

type PopupInfo struct {
	CloseStatus string `json:"close_status"` // found/success/fail
	Type        string `json:"type"`
	Text        string `json:"text"`
	RetryCount  int    `json:"retry_count"`
	PicName     string `json:"pic_name"`
	PicURL      string `json:"pic_url"`
	PopupArea   Box    `json:"popup_area"`
	CloseArea   Box    `json:"close_area"`
}

func (dExt *DriverExt) ClosePopups(options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)

	// default to retry 5 times
	if actionOptions.MaxRetryTimes == 0 {
		options = append(options, WithMaxRetryTimes(5))
	}
	// set default swipe interval to 1 second
	if builtin.IsZeroFloat64(actionOptions.Interval) {
		options = append(options, WithInterval(1))
	}
	return dExt.ClosePopupsHandler(options...)
}

func (dExt *DriverExt) ClosePopupsHandler(options ...ActionOption) error {
	actionOptions := NewActionOptions(options...)
	log.Info().Interface("actionOptions", actionOptions).Msg("try to find and close popups")
	maxRetryTimes := actionOptions.MaxRetryTimes
	interval := actionOptions.Interval

	for retryCount := 0; retryCount < maxRetryTimes; retryCount++ {
		screenResult, err := dExt.GetScreenResult(
			WithScreenShotClosePopups(true), WithScreenShotUpload(true))
		if err != nil {
			log.Error().Err(err).Msg("get screen result failed for popup handler")
			continue
		}
		screenResult.Popup.RetryCount = retryCount

		// 1. there are no popups here (fast return normally)
		// 2. failed to close popup （maybe tap error, return error）
		// 3. successful to close popup (sleep and wait for next retry if existed)
		if screenResult.Popup == nil {
			log.Debug().Msg("no popup found")
			break
		}

		// popup found
		if !screenResult.Popup.PopupArea.IsEmpty() {
			screenResult.Popup.CloseStatus = CloseStatusFound
		}
		if screenResult.Popup.CloseArea.IsEmpty() {
			log.Warn().Msg("popup close area not found")
			break
		}
		screenResult.Popup.CloseStatus = CloseStatusFound

		if err = dExt.tapPopupHandler(screenResult.Popup); err != nil {
			return err
		}

		log.Info().Interface("popup", screenResult.Popup).Msg("close popup success")
		// sleep for another popup (if existed) to pop
		time.Sleep(time.Duration(1000*interval) * time.Millisecond)
	}
	return nil
}

func (dExt *DriverExt) tapPopupHandler(popup *PopupInfo) error {
	if popup == nil {
		return nil
	}
	if popup.CloseArea.IsEmpty() {
		log.Warn().Msg("popup close area not found")
		return nil
	}

	closePoint := popup.CloseArea.Center()
	log.Info().
		Interface("popup", popup).
		Interface("closePoint", closePoint).
		Msg("tap to close popup")
	if err := dExt.TapAbsXY(closePoint.X, closePoint.Y); err != nil {
		log.Error().Err(err).Msg("tap popup failed")
		return errors.Wrap(code.MobileUIPopupError, err.Error())
	}
	// tap popup success
	return nil
}
