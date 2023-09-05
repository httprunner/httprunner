package uixt

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/code"
)

// TODO: add more popup texts
var popups = [][]string{
	{".*青少年.*", "我知道了"}, // 青少年弹窗
	{".*个人信息保护.*", "同意"},
	{".*通讯录.*", "拒绝"},
	{".*更新.*", "以后再说|稍后|取消"},
	{".*升级.*", "以后再说|稍后|取消"},
	{".*定位.*", "仅.*允许"},
	{".*拍照.*", "仅.*允许"},
	{".*录音.*", "仅.*允许"},
	{".*位置.*", "仅.*允许"},
	{".*权限.*", "仅.*允许|始终允许"},
	{".*允许.*", "仅.*允许|始终允许"},
	{".*风险.*", "继续使用"},
	{"管理使用时间", ".*忽略.*"},
}

const (
	CloseStatusFound   = "found"
	CloseStatusSuccess = "success"
	CloseStatusFail    = "fail"
)

func findTextPopup(screenTexts OCRTexts) (closePoint *OCRText) {
	for _, popup := range popups {
		if len(popup) != 2 {
			continue
		}

		points, err := screenTexts.FindTexts([]string{popup[0], popup[1]}, WithRegex(true))
		if err == nil {
			log.Warn().Interface("popup", popup).
				Interface("texts", screenTexts).Msg("text popup found")
			closePoint = &points[1]
			break
		}
	}
	return
}

func (dExt *DriverExt) handleTextPopup(screenTexts OCRTexts) error {
	closePoint := findTextPopup(screenTexts)
	if closePoint == nil {
		// no popup found
		return nil
	}

	log.Info().Str("text", closePoint.Text).Msg("close popup")
	pointCenter := closePoint.Center()
	if err := dExt.TapAbsXY(pointCenter.X, pointCenter.Y); err != nil {
		log.Error().Err(err).Msg("tap popup failed")
		return errors.Wrap(code.MobileUIPopupError, err.Error())
	}
	// tap popup success
	return nil
}

func (dExt *DriverExt) AutoPopupHandler() error {
	// TODO: check popup by activity type

	// check popup by screenshot
	screenResult, err := dExt.GetScreenResult(
		WithScreenShotOCR(true), WithScreenShotUpload(true))
	if err != nil {
		return errors.Wrap(err, "get screen result failed for popup handler")
	}

	return dExt.handleTextPopup(screenResult.Texts)
}

// ClosePopupsResult represents the result of recognized popup to close
type ClosePopupsResult struct {
	Type      string `json:"type"`
	PopupArea Box    `json:"popupArea"`
	CloseArea Box    `json:"closeArea"`
	Text      string `json:"text"`
}

type PopupInfo struct {
	CloseStatus string `json:"close_status"`
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
	log.Info().Msg("try to find and close popups")
	actionOptions := NewActionOptions(options...)
	maxRetryTimes := actionOptions.MaxRetryTimes
	interval := actionOptions.Interval

	for retryCount := 0; retryCount < maxRetryTimes; retryCount++ {
		screenResult, err := dExt.GetScreenResult(
			WithScreenShotClosePopups(true), WithScreenShotUpload(true))
		if err != nil {
			log.Error().Err(err).Msg("get screen result failed for popup handler")
			continue
		}
		// 1. there are no popups here (fast return normally)
		// 2. failed to close popup （maybe tap error, return error）
		// 3. successful to close popup (sleep and wait for next retry if existed)
		if screenResult.Popup == nil {
			break
		}
		screenResult.Popup.RetryCount = retryCount
		if !screenResult.Popup.PopupArea.IsEmpty() {
			screenResult.Popup.CloseStatus = CloseStatusFound
		}
		if screenResult.Popup.CloseArea.IsEmpty() {
			break
		}
		screenResult.Popup.CloseStatus = CloseStatusFound

		if err = dExt.tapPopupHandler(screenResult.Popup); err != nil {
			return err
		}
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
		return nil
	}
	log.Info().Str("type", popup.Type).Str("text", popup.Text).Msg("close popup")
	popupCenter := popup.CloseArea.Center()
	if err := dExt.TapAbsXY(popupCenter.X, popupCenter.Y); err != nil {
		log.Error().Err(err).Msg("tap popup failed")
		return errors.Wrap(code.MobileUIPopupError, err.Error())
	}
	// tap popup success
	return nil
}
