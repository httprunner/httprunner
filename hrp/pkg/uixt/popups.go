package uixt

import (
	"math/rand"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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
	*ClosePopupsResult
	CloseStatus string   `json:"close_status"`           // found/success/fail
	ClosePoints []PointF `json:"close_points,omitempty"` // CV 识别的所有关闭按钮（仅关闭按钮，可能存在多个）
}

func (p *PopupInfo) getClosePoint(lastPopup *PopupInfo) (*PointF, error) {
	closeResult := p.ClosePopupsResult
	if closeResult == nil {
		return nil, nil
	}

	// 弹框不存在 && 关闭按钮不存在
	if closeResult.PopupArea.IsEmpty() && closeResult.CloseArea.IsEmpty() {
		if p.ClosePoints == nil {
			// 关闭图标不存在 => 100% 确定不存在弹窗
			return nil, nil
		}

		// 存在关闭按钮，结合上一次的 popup 进行判断
		if lastPopup == nil || lastPopup.ClosePoints == nil {
			// 当前关闭图标为首次出现，确定是弹窗关闭按钮的概率较小
			log.Debug().Interface("closePoints", p.ClosePoints).
				Msg("skip close points for the first time")
			return nil, nil
		}

		// 连续两次都存在关闭图标
		if p.ClosePoints[0].IsIdentical(lastPopup.ClosePoints[0]) {
			// 连续两次图标位置相同 => 存在弹窗 => 点击关闭
			log.Warn().
				Interface("closePoint", p.ClosePoints[0]).
				Interface("lastClosePoints", lastPopup.ClosePoints).
				Msg("popup close point detected")
			return getRandomClosePoint(p.ClosePoints), nil
		}

		// 连续两次图标位置不同 => 可能不是弹窗 => skip
		log.Debug().Interface("closePoints", p.ClosePoints).
			Interface("lastClosePoints", lastPopup.ClosePoints).
			Msg("skip close points for not sure")
		return nil, nil
	}

	// 弹窗存在 && 关闭按钮不存在
	if !closeResult.PopupArea.IsEmpty() && closeResult.CloseArea.IsEmpty() {
		if p.ClosePoints == nil {
			// 关闭图标不存在 => 无法处理，抛异常
			log.Error().Interface("popup", p).Msg("popup close area not found")
			return nil, errors.Wrap(code.MobileUIPopupError, "popup close area not found")
		}

		// 使用关闭图标作为关闭按钮（随机选择一个）
		return getRandomClosePoint(p.ClosePoints), nil
	}

	// 关闭按钮存在 && (弹框存在 || 不存在)
	if closeResult.Type != "" || p.ClosePoints == nil {
		// 弹窗类型存在 || 关闭图标不存在 => 基于关闭按钮关闭弹窗
		closePoint := closeResult.CloseArea.Center()
		return &closePoint, nil
	} else {
		// 弹窗类型不存在 && 关闭图标存在，使用关闭图标作为关闭按钮（随机选择一个）
		return getRandomClosePoint(p.ClosePoints), nil
	}
}

func getRandomClosePoint(closePoints []PointF) *PointF {
	if len(closePoints) == 1 {
		return &closePoints[0]
	}
	return &closePoints[rand.Intn(len(closePoints))]
}

func (dExt *DriverExt) ClosePopupsHandler() (err error) {
	log.Info().Msg("try to find and close popups")

	screenResult, err := dExt.GetScreenResult(
		WithScreenShotUpload(true),
		WithScreenShotClosePopups(true), // get popup area and close area
		WithScreenShotUITypes("close"),  // get all close buttons
	)
	if err != nil {
		log.Error().Err(err).Msg("get screen result failed for popup handler")
		return err
	}

	popup := screenResult.Popup

	defer func() {
		dExt.lastPopup = popup
	}()

	closePoint, err := popup.getClosePoint(dExt.lastPopup)
	if err != nil {
		return err
	}

	if closePoint == nil {
		// close point not found
		log.Debug().Msg("close point not found")
		return nil
	}

	popup.CloseStatus = CloseStatusFound
	log.Info().
		Interface("closePoint", closePoint).
		Interface("popup", popup).
		Msg("tap to close popup")
	if err := dExt.TapAbsXY(closePoint.X, closePoint.Y); err != nil {
		log.Error().Err(err).Msg("tap popup failed")
		return errors.Wrap(code.MobileUIPopupError, err.Error())
	}

	// wait 1s and check if popup still exists
	log.Info().Msg("tap close point success, check if popup still exists")
	return nil
}
