package uixt

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
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
	*ClosePopupsResult
	ClosePoints []PointF `json:"close_points,omitempty"` // CV 识别的所有关闭按钮（仅关闭按钮，可能存在多个）
	PicName     string   `json:"pic_name"`
	PicURL      string   `json:"pic_url"`
}

func (p *PopupInfo) getClosePoint() *PointF {
	closeResult := p.ClosePopupsResult
	if closeResult == nil {
		return nil
	}

	// 弹窗关闭按钮不存在
	if closeResult.CloseArea.IsEmpty() {
		return nil
	}

	closePoint := closeResult.CloseArea.Center()
	return &closePoint
}

func (dExt *DriverExt) ClosePopupsHandler() (err error) {
	log.Info().Msg("try to find and close popups")

	screenResult, err := dExt.GetScreenResult(
		WithScreenShotUpload(true),
		WithScreenShotClosePopups(true), // get popup area and close area
	)
	if err != nil {
		log.Error().Err(err).Msg("get screen result failed for popup handler")
		return err
	}

	popup := screenResult.Popup
	closePoint := popup.getClosePoint()
	if closePoint == nil {
		// close point not found
		log.Debug().Msg("close point not found")
		return nil
	}

	log.Info().
		Interface("closePoint", closePoint).
		Interface("popup", popup).
		Msg("tap to close popup")
	if err := dExt.TapAbsXY(closePoint.X, closePoint.Y); err != nil {
		log.Error().Err(err).Msg("tap popup failed")
		return errors.Wrap(code.MobileUIPopupError, err.Error())
	}

	return nil
}
