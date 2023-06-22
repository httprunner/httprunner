//go:build opencv

package uixt

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"os"

	"github.com/pkg/errors"
	"gocv.io/x/gocv"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
)

const (
	// TmCcoeffNormed maps to TM_CCOEFF_NORMED
	TmCcoeffNormed TemplateMatchMode = iota
	// TmSqdiff maps to TM_SQDIFF
	TmSqdiff
	// TmSqdiffNormed maps to TM_SQDIFF_NORMED
	TmSqdiffNormed
	// TmCcorr maps to TM_CCORR
	TmCcorr
	// TmCcorrNormed maps to TM_CCORR_NORMED
	TmCcorrNormed
	// TmCcoeff maps to TM_CCOEFF
	TmCcoeff
)

type DebugMode int

const (
	// DmOff no output
	DmOff DebugMode = iota
	// DmEachMatch output matched and mismatched values
	DmEachMatch
	// DmNotMatch output only values that do not match
	DmNotMatch
)

// extendCV 获得扩展后的 Driver，
// 并指定匹配阀值，
// 获取当前设备的 Scale，
// 默认匹配模式为 TmCcoeffNormed，
// 默认关闭 OpenCV 匹配值计算后的输出
func (dExt *DriverExt) extendCV(options ...CVOption) (err error) {
	for _, option := range options {
		option(&dExt.CVArgs)
	}

	if dExt.threshold == 0 {
		dExt.threshold = 0.95 // default threshold
	}
	if dExt.matchMode == 0 {
		dExt.matchMode = TmCcoeffNormed // default match mode
	}
	Debug(DebugMode(DmOff))
	return
}

func (dExt *DriverExt) Debug(dm DebugMode) {
	Debug(DebugMode(dm))
}

func (dExt *DriverExt) OnlyOnceThreshold(threshold float64) (newExt *DriverExt) {
	newExt = new(DriverExt)
	newExt.Driver = dExt.Driver
	newExt.matchMode = dExt.matchMode
	newExt.threshold = threshold
	return
}

func (dExt *DriverExt) OnlyOnceMatchMode(matchMode TemplateMatchMode) (newExt *DriverExt) {
	newExt = new(DriverExt)
	newExt.Driver = dExt.Driver
	newExt.matchMode = matchMode
	newExt.threshold = dExt.threshold
	return
}

// func (sExt *DriverExt) findImgRect(search string) (rect image.Rectangle, err error) {
// 	pathSource := filepath.Join(sExt.pathname, GenFilename())
// 	if err = sExt.driver.ScreenshotToDisk(pathSource); err != nil {
// 		return image.Rectangle{}, err
// 	}
//
// 	if rect, err = FindImageRectFromDisk(pathSource, search, float32(sExt.Threshold), TemplateMatchMode(sExt.MatchMode)); err != nil {
// 		return image.Rectangle{}, err
// 	}
// 	return
// }

func (dExt *DriverExt) FindAllImageRect(search string) (rects []image.Rectangle, err error) {
	var bufSource, bufSearch *bytes.Buffer
	if bufSearch, err = getBufFromDisk(search); err != nil {
		return nil, err
	}
	if bufSource, _, err = dExt.takeScreenShot(builtin.GenNameWithTimestamp("%d_cv")); err != nil {
		return nil, err
	}

	if rects, err = FindAllImageRectsFromRaw(bufSource, bufSearch, float32(dExt.threshold), TemplateMatchMode(dExt.matchMode)); err != nil {
		return nil, err
	}
	return
}

func (dExt *DriverExt) FindImageRectInUIKit(imagePath string, options ...ActionOption) (point PointF, err error) {
	var bufSource, bufSearch *bytes.Buffer
	if bufSearch, err = getBufFromDisk(imagePath); err != nil {
		return PointF{}, err
	}

	var rect image.Rectangle
	if rect, err = FindImageRectFromRaw(bufSource, bufSearch, float32(dExt.threshold), TemplateMatchMode(dExt.matchMode)); err != nil {
		return PointF{}, err
	}

	point = getRectangleCenterPoint(rect)
	return point, nil
}

func getBufFromDisk(name string) (*bytes.Buffer, error) {
	var f *os.File
	var err error
	if f, err = os.Open(name); err != nil {
		return nil, err
	}
	var all []byte
	if all, err = ioutil.ReadAll(f); err != nil {
		return nil, err
	}
	return bytes.NewBuffer(all), nil
}

var debug = DmOff

const dmOutputMsg = `[DEBUG] The current value is '%.4f', the expected value is '%.4f'`

func Debug(dm DebugMode) {
	debug = dm
}

const DefaultMatchMode = TmCcoeffNormed

var fillColor = color.RGBA{R: 255, G: 255, B: 255, A: 0}

func FindImageLocationFromRaw(source, search *bytes.Buffer, threshold float32, matchMode ...TemplateMatchMode) (loc image.Point, err error) {
	if len(matchMode) == 0 {
		matchMode = []TemplateMatchMode{DefaultMatchMode}
	}
	var matImage, matTpl gocv.Mat
	if matImage, matTpl, err = getMatsFromRaw(source, search, gocv.IMReadGrayScale); err != nil {
		return image.Point{}, err
	}
	defer func() {
		_ = matImage.Close()
		_ = matTpl.Close()
	}()
	return getMatchingLocation(matImage, matTpl, threshold, matchMode[0])
}

func FindImageLocationFromDisk(source, search string, threshold float32, matchMode ...TemplateMatchMode) (loc image.Point, err error) {
	if len(matchMode) == 0 {
		matchMode = []TemplateMatchMode{DefaultMatchMode}
	}
	var matImage, matTpl gocv.Mat
	if matImage, matTpl, err = getMatsFromDisk(source, search, gocv.IMReadGrayScale); err != nil {
		return image.Point{}, err
	}
	defer func() {
		_ = matImage.Close()
		_ = matTpl.Close()
	}()

	return getMatchingLocation(matImage, matTpl, threshold, matchMode[0])
}

func FindAllImageLocationsFromDisk(source, search string, threshold float32, matchMode ...TemplateMatchMode) (locs []image.Point, err error) {
	if len(matchMode) == 0 {
		matchMode = []TemplateMatchMode{DefaultMatchMode}
	}
	var matImage, matTpl gocv.Mat
	if matImage, matTpl, err = getMatsFromDisk(source, search, gocv.IMReadGrayScale); err != nil {
		return nil, err
	}
	defer func() {
		_ = matImage.Close()
		_ = matTpl.Close()
	}()

	var loc image.Point
	if loc, err = getMatchingLocation(matImage, matTpl, threshold, matchMode[0]); err != nil {
		return nil, err
	}
	widthTpl := matTpl.Cols()
	heightTpl := matTpl.Rows()

	locs = make([]image.Point, 0, 9)
	locs = append(locs, loc)

	gocv.FillPoly(&matImage, gocv.NewPointsVectorFromPoints(getPts(loc, widthTpl, heightTpl)), fillColor)

	loc, err = getMatchingLocation(matImage, matTpl, threshold, matchMode[0])
	for ; err == nil; loc, err = getMatchingLocation(matImage, matTpl, threshold, matchMode[0]) {
		locs = append(locs, loc)
		gocv.FillPoly(&matImage, gocv.NewPointsVectorFromPoints(getPts(loc, widthTpl, heightTpl)), fillColor)
	}

	return locs, nil
}

func FindAllImageLocationsFromRaw(source, search *bytes.Buffer, threshold float32, matchMode ...TemplateMatchMode) (locs []image.Point, err error) {
	if len(matchMode) == 0 {
		matchMode = []TemplateMatchMode{DefaultMatchMode}
	}
	var matImage, matTpl gocv.Mat
	if matImage, matTpl, err = getMatsFromRaw(source, search, gocv.IMReadGrayScale); err != nil {
		return nil, err
	}
	defer func() {
		_ = matImage.Close()
		_ = matTpl.Close()
	}()

	var loc image.Point
	if loc, err = getMatchingLocation(matImage, matTpl, threshold, matchMode[0]); err != nil {
		return nil, err
	}
	widthTpl := matTpl.Cols()
	heightTpl := matTpl.Rows()

	locs = make([]image.Point, 0, 9)
	locs = append(locs, loc)

	gocv.FillPoly(&matImage, gocv.NewPointsVectorFromPoints(getPts(loc, widthTpl, heightTpl)), fillColor)

	loc, err = getMatchingLocation(matImage, matTpl, threshold, matchMode[0])
	for ; err == nil; loc, err = getMatchingLocation(matImage, matTpl, threshold, matchMode[0]) {
		locs = append(locs, loc)
		gocv.FillPoly(&matImage, gocv.NewPointsVectorFromPoints(getPts(loc, widthTpl, heightTpl)), fillColor)
	}

	return locs, nil
}

// getPts 根据图片坐标和宽高，获取填充区域
func getPts(loc image.Point, width, height int) [][]image.Point {
	return [][]image.Point{
		{
			image.Pt(loc.X, loc.Y),
			image.Pt(loc.X, loc.Y+height),
			image.Pt(loc.X+width, loc.Y+height),
			image.Pt(loc.X+width, loc.Y),
		},
	}
}

func FindImageRectFromDisk(source, search string, threshold float32, matchMode ...TemplateMatchMode) (rect image.Rectangle, err error) {
	var matTpl gocv.Mat
	if _, matTpl, err = getMatsFromDisk(source, search, gocv.IMReadGrayScale); err != nil {
		return image.Rectangle{}, err
	}
	defer func() {
		_ = matTpl.Close()
	}()

	var loc image.Point
	if loc, err = FindImageLocationFromDisk(source, search, threshold, matchMode...); err != nil {
		return image.Rectangle{}, err
	}
	rect = image.Rect(loc.X, loc.Y, loc.X+matTpl.Cols(), loc.Y+matTpl.Rows())
	return
}

func FindAllImageRectsFromDisk(source, search string, threshold float32, matchMode ...TemplateMatchMode) (rects []image.Rectangle, err error) {
	var matTpl gocv.Mat
	if _, matTpl, err = getMatsFromDisk(source, search, gocv.IMReadGrayScale); err != nil {
		return nil, err
	}
	defer func() {
		_ = matTpl.Close()
	}()

	var locs []image.Point
	if locs, err = FindAllImageLocationsFromDisk(source, search, threshold, matchMode...); err != nil {
		return nil, err
	}

	rects = make([]image.Rectangle, 0, len(locs))
	for i := range locs {
		r := image.Rect(locs[i].X, locs[i].Y, locs[i].X+matTpl.Cols(), locs[i].Y+matTpl.Rows())
		rects = append(rects, r)
	}
	return
}

func FindImageRectFromRaw(source, search *bytes.Buffer, threshold float32, matchMode ...TemplateMatchMode) (rect image.Rectangle, err error) {
	var matTpl gocv.Mat
	if _, matTpl, err = getMatsFromRaw(source, search, gocv.IMReadGrayScale); err != nil {
		return image.Rectangle{}, err
	}
	defer func() {
		_ = matTpl.Close()
	}()

	var loc image.Point
	if loc, err = FindImageLocationFromRaw(source, search, threshold, matchMode...); err != nil {
		return image.Rectangle{}, err
	}
	rect = image.Rect(loc.X, loc.Y, loc.X+matTpl.Cols(), loc.Y+matTpl.Rows())
	return
}

func FindAllImageRectsFromRaw(source, search *bytes.Buffer, threshold float32, matchMode ...TemplateMatchMode) (rects []image.Rectangle, err error) {
	var matTpl gocv.Mat
	if _, matTpl, err = getMatsFromRaw(source, search, gocv.IMReadGrayScale); err != nil {
		return nil, err
	}
	defer func() {
		_ = matTpl.Close()
	}()

	var locs []image.Point
	if locs, err = FindAllImageLocationsFromRaw(source, search, threshold, matchMode...); err != nil {
		return nil, err
	}

	rects = make([]image.Rectangle, 0, len(locs))
	for i := range locs {
		r := image.Rect(locs[i].X, locs[i].Y, locs[i].X+matTpl.Cols(), locs[i].Y+matTpl.Rows())
		rects = append(rects, r)
	}
	return
}

// getMatsFromDisk 从指定路径获取原图和目标图的 `gocv.Mat`
func getMatsFromDisk(nameImage, nameTpl string, flags gocv.IMReadFlag) (matImage, matTpl gocv.Mat, err error) {
	matImage = gocv.IMRead(nameImage, flags)
	if matImage.Empty() {
		return gocv.Mat{}, gocv.Mat{}, fmt.Errorf("invalid read %s", nameImage)
	}
	matTpl = gocv.IMRead(nameTpl, flags)
	if matTpl.Empty() {
		return gocv.Mat{}, gocv.Mat{}, fmt.Errorf("invalid read %s", nameTpl)
	}
	return
}

// func getBufsFromDisk(nameImage, nameTpl string) (bufImage, bufTpl *bytes.Buffer, err error) {
// 	getBuf := func(_name string) (*bytes.Buffer, error) {
// 		var f *os.File
// 		var e error
// 		if f, e = os.Open(_name); e != nil {
// 			return nil, e
// 		}
// 		var all []byte
// 		if all, e = ioutil.ReadAll(f); e != nil {
// 			return nil, e
// 		}
// 		return bytes.NewBuffer(all), nil
// 	}
// 	if bufImage, err = getBuf(nameImage); err != nil {
// 		return nil, nil, err
// 	}
// 	if bufTpl, err = getBuf(nameTpl); err != nil {
// 		return nil, nil, err
// 	}
// 	return
// }

// getMatsFromRaw 获取原图和目标图的 `gocv.Mat`
func getMatsFromRaw(bufImage, bufTpl *bytes.Buffer, flags gocv.IMReadFlag) (matImage, matTpl gocv.Mat, err error) {
	if matImage, err = gocv.IMDecode(bufImage.Bytes(), flags); err != nil {
		return gocv.Mat{}, gocv.Mat{}, fmt.Errorf("invalid read %w", err)
	}
	if matImage.Empty() {
		return gocv.Mat{}, gocv.Mat{}, errors.New("invalid read [source]")
	}
	if matTpl, err = gocv.IMDecode(bufTpl.Bytes(), flags); err != nil {
		return gocv.Mat{}, gocv.Mat{}, fmt.Errorf("invalid read %w", err)
	}
	if matTpl.Empty() {
		return gocv.Mat{}, gocv.Mat{}, errors.New("invalid read [template]")
	}
	return
}

// getMatchingLocation 获取匹配的图片位置
func getMatchingLocation(matImage gocv.Mat, matTpl gocv.Mat, threshold float32, matchMode TemplateMatchMode) (loc image.Point, err error) {
	if threshold > 1 {
		threshold = 1.0
	}
	// TM_SQDIFF：该方法使用平方差进行匹配，最好匹配为 0。值越大匹配结果越差。
	// TM_SQDIFF_NORMED：该方法使用归一化的平方差进行匹配，最佳匹配也在结果为0处。
	// TmCcoeff 将模版对其均值的相对值与图像对其均值的相关值进行匹配,1表示完美匹配,-1表示糟糕的匹配,0表示没有任何相关性(随机序列)。
	minVal, maxVal, minLoc, maxLoc := getMatchingResult(matImage, matTpl, matchMode)

	// fmt.Println(matchMode[0], "\t", minVal, maxVal, "\t", minLoc, maxLoc)
	// fmt.Printf("%s\t %.10f \t %.10f \t %v \t %v \n", matchMode[0], minVal, maxVal, minLoc, maxLoc)

	var val float32
	val, loc = getValLoc(minVal, maxVal, minLoc, maxLoc, matchMode)

	if debug == DmEachMatch {
		log.Println(fmt.Sprintf(dmOutputMsg, val, threshold))
	}

	if val >= threshold {
		return loc, nil
	} else {
		if debug == DmNotMatch {
			log.Println(fmt.Sprintf(dmOutputMsg, val, threshold))
		}
		return image.Point{}, errors.New("no such target search image")
	}
}

// getMatchingResult 匹配图片并返回匹配值和位置
func getMatchingResult(matImage gocv.Mat, matTpl gocv.Mat, matchMode TemplateMatchMode) (minVal float32, maxVal float32, minLoc image.Point, maxLoc image.Point) {
	matResult, tmpMask := gocv.NewMat(), gocv.NewMat()
	defer func() {
		_ = matResult.Close()
		_ = tmpMask.Close()
	}()
	gocv.MatchTemplate(matImage, matTpl, &matResult, gocv.TemplateMatchMode(matchMode), tmpMask)
	minVal, maxVal, minLoc, maxLoc = gocv.MinMaxLoc(matResult)
	return
}

// getValLoc 根据不同的匹配模式返回匹配值和位置
func getValLoc(minVal float32, maxVal float32, minLoc image.Point, maxLoc image.Point, matchMode TemplateMatchMode) (val float32, loc image.Point) {
	val, loc = maxVal, maxLoc

	switch matchMode {
	case TmSqdiff, TmSqdiffNormed:
		// 平方差，最佳匹配为 0
		val = minVal
		// minVal = 8
		// for val >= 1 {
		// 	val -= 1
		// }
		if val >= 1 {
			val = float32(math.Mod(float64(val), 1))
		}
		val = 1 - val
		loc = minLoc
	case TmCcoeff:
		// TmCcoeff 将模版对其均值的相对值与图像对其均值的相关值进行匹配,1表示完美匹配,-1表示糟糕的匹配,0表示没有任何相关性(随机序列)。
		// maxVal = 5064792.5000000000
		_, frac := math.Modf(float64(val))
		val = float32(frac)
	case TmCcorr:
		// maxVal = 50553512.0000000000
		_, frac := math.Modf(float64(val))
		val = float32(frac)
	}
	// fmt.Println("匹配度", val)
	return
}
