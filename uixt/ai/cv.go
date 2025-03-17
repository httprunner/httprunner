package ai

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"regexp"

	"github.com/pkg/errors"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/internal/builtin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/httprunner/httprunner/v5/uixt/types"
)

type ICVService interface {
	// returns CV result including ocr texts, uploaded image url, etc
	ReadFromBuffer(imageBuf *bytes.Buffer, opts ...option.ActionOption) (*CVResult, error)
	ReadFromPath(imagePath string, opts ...option.ActionOption) (*CVResult, error)
}

type CVResult struct {
	URL       string     `json:"url,omitempty"`       // image uploaded url
	OCRResult OCRResults `json:"ocrResult,omitempty"` // OCR texts
	// NoLive（非直播间）
	// Shop（电商）
	// LifeService（生活服务）
	// Show（秀场）
	// Game（游戏）
	// People（多人）
	// PK（PK）
	// Media（媒体）
	// Chat（语音）
	// Event（赛事）
	LiveType          string             `json:"liveType,omitempty"`       // 直播间类型
	LivePopularity    int64              `json:"livePopularity,omitempty"` // 直播间热度
	UIResult          UIResultMap        `json:"uiResult,omitempty"`       // 图标检测
	ClosePopupsResult *ClosePopupsResult `json:"closeResult,omitempty"`    // 弹窗按钮检测
}

type OCRResult struct {
	Text   string   `json:"text"`
	Points []PointF `json:"points"`
}

type OCRResults []OCRResult

func (o OCRResults) ToOCRTexts() (ocrTexts OCRTexts) {
	for _, ocrResult := range o {
		rect := image.Rectangle{
			// ocrResult.Points 顺序：左上 -> 右上 -> 右下 -> 左下
			Min: image.Point{
				X: int(ocrResult.Points[0].X),
				Y: int(ocrResult.Points[0].Y),
			},
			Max: image.Point{
				X: int(ocrResult.Points[2].X),
				Y: int(ocrResult.Points[2].Y),
			},
		}
		rectStr := fmt.Sprintf("%d,%d,%d,%d",
			rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
		ocrText := OCRText{
			Text:    ocrResult.Text,
			Rect:    rect,
			RectStr: rectStr,
		}
		ocrTexts = append(ocrTexts, ocrText)
	}
	return
}

type OCRText struct {
	Text    string          `json:"text"`
	RectStr string          `json:"rect"`
	Rect    image.Rectangle `json:"-"`
}

func (t OCRText) Size() types.Size {
	return types.Size{
		Width:  t.Rect.Dx(),
		Height: t.Rect.Dy(),
	}
}

func (t OCRText) Center() PointF {
	rect := t.Rect
	x, y := float64(rect.Min.X), float64(rect.Min.Y)
	width, height := float64(rect.Dx()), float64(rect.Dy())
	point := PointF{
		X: x + width*0.5,
		Y: y + height*0.5,
	}
	return point
}

func (t OCRText) RandomPoint() PointF {
	rect := t.Rect
	x, y := float64(rect.Min.X), float64(rect.Min.Y)
	width, height := float64(rect.Dx()), float64(rect.Dy())
	point := PointF{
		X: x + width*rand.Float64(),
		Y: y + height*rand.Float64(),
	}
	return point
}

type OCRTexts []OCRText

func (t OCRTexts) texts() (texts []string) {
	for _, text := range t {
		texts = append(texts, text.Text)
	}
	return texts
}

func (t OCRTexts) FilterScope(scope option.AbsScope) (results OCRTexts) {
	for _, ocrText := range t {
		rect := ocrText.Rect

		// check if text in scope
		if len(scope) == 4 {
			if rect.Min.X < scope[0] ||
				rect.Min.Y < scope[1] ||
				rect.Max.X > scope[2] ||
				rect.Max.Y > scope[3] {
				// not in scope
				continue
			}
		}

		results = append(results, ocrText)
	}
	return
}

// FindText returns matched text with options
// Notice: filter scope should be specified with WithAbsScope
func (t OCRTexts) FindText(text string, opts ...option.ActionOption) (result OCRText, err error) {
	options := option.NewActionOptions(opts...)

	var results []OCRText
	for _, ocrText := range t.FilterScope(options.AbsScope) {
		if options.Regex {
			// regex on, check if match regex
			if !regexp.MustCompile(text).MatchString(ocrText.Text) {
				continue
			}
		} else {
			// regex off, check if match exactly
			if ocrText.Text != text {
				continue
			}
		}

		results = append(results, ocrText)

		// return the first one matched exactly when index not specified
		if ocrText.Text == text && options.Index == 0 {
			return ocrText, nil
		}
	}

	if len(results) == 0 {
		return OCRText{}, errors.Wrap(code.CVResultNotFoundError,
			fmt.Sprintf("text %s not found in %v", text, t.texts()))
	}

	// get index
	idx := options.Index
	if idx < 0 {
		idx = len(results) + idx
	}

	// index out of range
	if idx >= len(results) || idx < 0 {
		return OCRText{}, errors.Wrap(code.CVResultNotFoundError,
			fmt.Sprintf("text %s found %d, index %d out of range", text, len(results), idx))
	}

	return results[idx], nil
}

func (t OCRTexts) FindTexts(texts []string, opts ...option.ActionOption) (results OCRTexts, err error) {
	options := option.NewActionOptions(opts...)
	for _, text := range texts {
		ocrText, err := t.FindText(text, opts...)
		if err != nil {
			continue
		}
		results = append(results, ocrText)

		// found one, skip searching and return
		if options.MatchOne {
			return results, nil
		}
	}

	if len(results) == len(texts) {
		return results, nil
	}
	return nil, errors.Wrap(code.CVResultNotFoundError,
		fmt.Sprintf("texts %s not found in %v", texts, t.texts()))
}

type UIResultMap map[string]UIResults

// FilterUIResults filters ui icons, the former the uiTypes, the higher the priority
func (u UIResultMap) FilterUIResults(uiTypes []string) (uiResults UIResults, err error) {
	var ok bool
	for _, uiType := range uiTypes {
		uiResults, ok = u[uiType]
		if ok && len(uiResults) != 0 {
			return
		}
	}
	err = errors.Wrap(code.CVResultNotFoundError, fmt.Sprintf("UI types %v not detected", uiTypes))
	return
}

type UIResult struct {
	Box
}

type Box struct {
	Point  PointF  `json:"point"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (box Box) IsEmpty() bool {
	return builtin.IsZeroFloat64(box.Width) && builtin.IsZeroFloat64(box.Height)
}

func (box Box) IsIdentical(box2 Box) bool {
	// set the coordinate precision to 1 pixel
	return box.Point.IsIdentical(box2.Point) &&
		builtin.IsZeroFloat64(math.Abs(box.Width-box2.Width)) &&
		builtin.IsZeroFloat64(math.Abs(box.Height-box2.Height))
}

func (box Box) Center() PointF {
	return PointF{
		X: box.Point.X + box.Width*0.5,
		Y: box.Point.Y + box.Height*0.5,
	}
}

type UIResults []UIResult

func (u UIResults) FilterScope(scope option.AbsScope) (results UIResults) {
	for _, uiResult := range u {
		rect := image.Rectangle{
			Min: image.Point{
				X: int(uiResult.Point.X),
				Y: int(uiResult.Point.Y),
			},
			Max: image.Point{
				X: int(uiResult.Point.X + uiResult.Width),
				Y: int(uiResult.Point.Y + uiResult.Height),
			},
		}

		// check if ui result in scope
		if len(scope) == 4 {
			if rect.Min.X < scope[0] ||
				rect.Min.Y < scope[1] ||
				rect.Max.X > scope[2] ||
				rect.Max.Y > scope[3] {
				// not in scope
				continue
			}
		}
		results = append(results, uiResult)
	}
	return
}

func (u UIResults) GetUIResult(opts ...option.ActionOption) (UIResult, error) {
	options := option.NewActionOptions(opts...)
	uiResults := u.FilterScope(options.AbsScope)
	if len(uiResults) == 0 {
		return UIResult{}, errors.Wrap(code.CVResultNotFoundError,
			"ui types not found in scope")
	}
	// get index
	idx := options.Index
	if idx < 0 {
		idx = len(uiResults) + idx
	}

	// index out of range
	if idx >= len(uiResults) || idx < 0 {
		return UIResult{}, errors.Wrap(code.CVResultNotFoundError,
			fmt.Sprintf("ui types index %d out of range", idx))
	}
	return uiResults[idx], nil
}

// ClosePopupsResult represents the result of recognized popup to close
type ClosePopupsResult struct {
	Type      string `json:"type"`
	PopupArea Box    `json:"popupArea"`
	CloseArea Box    `json:"closeArea"`
	Text      string `json:"text"`
}

func (c ClosePopupsResult) IsEmpty() bool {
	return c.PopupArea.IsEmpty() && c.CloseArea.IsEmpty()
}

type Point struct {
	X int `json:"x"` // upper left X coordinate of selected element
	Y int `json:"y"` // upper left Y coordinate of selected element
}

type PointF struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p PointF) IsIdentical(p2 PointF) bool {
	// set the coordinate precision to 1 pixel
	return math.Abs(p.X-p2.X) < 1 && math.Abs(p.Y-p2.Y) < 1
}
