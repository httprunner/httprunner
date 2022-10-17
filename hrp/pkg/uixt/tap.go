package uixt

import (
	"fmt"
)

func (dExt *DriverExt) TapAbsXY(x, y float64, options ...DataOption) error {
	// tap on absolute coordinate [x, y]
	return dExt.Driver.TapFloat(x, y, options...)
}

func (dExt *DriverExt) TapXY(x, y float64, options ...DataOption) error {
	// tap on [x, y] percent of window size
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be < 1, got x=%v, y=%v", x, y)
	}

	x = x * float64(dExt.windowSize.Width)
	y = y * float64(dExt.windowSize.Height)

	return dExt.TapAbsXY(x, y, options...)
}

func (dExt *DriverExt) GetTextXY(ocrText string, options ...DataOption) (point PointF, err error) {
	x, y, width, height, err := dExt.FindTextByOCR(ocrText, options...)
	if err != nil {
		return PointF{}, err
	}

	point = PointF{
		X: x + width*0.5,
		Y: y + height*0.5,
	}
	return point, nil
}

func (dExt *DriverExt) GetTextXYs(ocrText []string, options ...DataOption) (points []PointF, err error) {
	ps, err := dExt.FindTextsByOCR(ocrText, options...)
	if err != nil {
		return nil, err
	}

	for _, point := range ps {
		pointF := PointF{
			X: point[0] + point[2]*0.5,
			Y: point[1] + point[3]*0.5,
		}
		points = append(points, pointF)
	}

	return points, nil
}

func (dExt *DriverExt) GetImageXY(imagePath string, options ...DataOption) (point PointF, err error) {
	x, y, width, height, err := dExt.FindImageRectInUIKit(imagePath, options...)
	if err != nil {
		return PointF{}, err
	}

	point = PointF{
		X: x + width*0.5,
		Y: y + height*0.5,
	}
	return point, nil
}

func (dExt *DriverExt) TapByOCR(ocrText string, options ...DataOption) error {
	data := NewData(options...)

	point, err := dExt.GetTextXY(ocrText, options...)
	if err != nil {
		if data.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, options...)
}

func (dExt *DriverExt) TapByCV(imagePath string, options ...DataOption) error {
	data := NewData(options...)

	point, err := dExt.GetImageXY(imagePath, options...)
	if err != nil {
		if data.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, options...)
}

func (dExt *DriverExt) Tap(param string, options ...DataOption) error {
	return dExt.TapOffset(param, 0.5, 0.5, options...)
}

func (dExt *DriverExt) TapOffset(param string, xOffset, yOffset float64, options ...DataOption) (err error) {
	// click on element, find by name attribute
	ele, err := dExt.FindUIElement(param)
	if err == nil {
		return ele.Click()
	}

	data := NewData(options...)

	x, y, width, height, err := dExt.FindUIRectInUIKit(param, options...)
	if err != nil {
		if data.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(x+width*xOffset, y+height*yOffset, options...)
}

func (dExt *DriverExt) DoubleTapXY(x, y float64) error {
	// double tap on coordinate: [x, y] should be relative
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be < 1, got x=%v, y=%v", x, y)
	}

	x = x * float64(dExt.windowSize.Width)
	y = y * float64(dExt.windowSize.Height)
	return dExt.Driver.DoubleTapFloat(x, y)
}

func (dExt *DriverExt) DoubleTap(param string) (err error) {
	return dExt.DoubleTapOffset(param, 0.5, 0.5)
}

func (dExt *DriverExt) DoubleTapOffset(param string, xOffset, yOffset float64) (err error) {
	// click on element, find by name attribute
	ele, err := dExt.FindUIElement(param)
	if err == nil {
		return ele.DoubleTap()
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(param); err != nil {
		return err
	}

	return dExt.Driver.DoubleTapFloat(x+width*xOffset, y+height*yOffset)
}

// TapWithNumber sends one or more taps
func (dExt *DriverExt) TapWithNumber(param string, numberOfTaps int) (err error) {
	return dExt.TapWithNumberOffset(param, numberOfTaps, 0.5, 0.5)
}

func (dExt *DriverExt) TapWithNumberOffset(param string, numberOfTaps int, xOffset, yOffset float64) (err error) {
	if numberOfTaps <= 0 {
		numberOfTaps = 1
	}
	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(param); err != nil {
		return err
	}

	x = x + width*xOffset
	y = y + height*yOffset

	touchActions := NewTouchActions().Tap(NewTouchActionTap().WithXYFloat(x, y).WithCount(numberOfTaps))
	return dExt.PerformTouchActions(touchActions)
}
