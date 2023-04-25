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

func (dExt *DriverExt) TapByOCR(ocrText string, options ...DataOption) error {
	dataOptions := NewDataOptions(options...)

	point, err := dExt.FindScreenTextByOCR(ocrText, options...)
	if err != nil {
		if dataOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	return dExt.TapAbsXY(point.X, point.Y, options...)
}

func (dExt *DriverExt) TapByCV(imagePath string, options ...DataOption) error {
	dataOptions := NewDataOptions(options...)

	point, err := dExt.FindImageRectInUIKit(imagePath, options...)
	if err != nil {
		if dataOptions.IgnoreNotFoundError {
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
	dataOptions := NewDataOptions(options...)

	point, err := dExt.FindUIRectInUIKit(param, options...)
	if err != nil {
		if dataOptions.IgnoreNotFoundError {
			return nil
		}
		return err
	}

	// FIXME: handle offset
	return dExt.TapAbsXY(point.X+xOffset, point.Y+yOffset, options...)
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
	point, err := dExt.FindUIRectInUIKit(param)
	if err != nil {
		return err
	}

	// FIXME: handle offset
	return dExt.Driver.DoubleTapFloat(point.X+xOffset, point.Y+yOffset)
}
