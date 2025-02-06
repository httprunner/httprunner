package uixt

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
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

		points, err := screenTexts.FindTexts([]string{popup[0], popup[1]}, option.WithRegex(true))
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
		option.WithScreenShotOCR(true),
		option.WithScreenShotUpload(true),
		option.WithScreenShotFileName("check_popup"),
	)
	if err != nil {
		return errors.Wrap(err, "get screen result failed for popup handler")
	}

	return dExt.handleTextPopup(screenResult.Texts)
}

type PopupInfo struct {
	*ClosePopupsResult
	ClosePoints []PointF `json:"close_points,omitempty"` // CV 识别的所有关闭按钮（仅关闭按钮，可能存在多个）
	PicName     string   `json:"pic_name"`
	PicURL      string   `json:"pic_url"`
}

func (p *PopupInfo) ClosePoint() *PointF {
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

func (dExt *DriverExt) CheckPopup() (popup *PopupInfo, err error) {
	screenResult, err := dExt.GetScreenResult(
		option.WithScreenShotUpload(true),
		option.WithScreenShotClosePopups(true), // get popup area and close area
		option.WithScreenShotFileName("check_popup"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "get screen result failed for popup handler")
	}
	popup = screenResult.Popup
	if popup == nil {
		// popup not found
		log.Debug().Msg("check popup, no found")
		return nil, nil
	}
	closePoint := popup.ClosePoint()
	if closePoint == nil {
		// close point not found
		return nil, errors.Wrap(code.MobileUIPopupError, "popup close point not found")
	}
	log.Info().Interface("popup", popup).Msg("found popup")
	return popup, nil
}

func (dExt *DriverExt) ClosePopupsHandler() (err error) {
	log.Info().Msg("try to find and close popups")

	popup, err := dExt.CheckPopup()
	if err != nil {
		// check popup failed
		return err
	} else if popup == nil {
		// no popup found
		return nil
	}

	// found popup
	closePoint := popup.ClosePoint()

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
