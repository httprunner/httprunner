package uixt

import (
	"fmt"

	"github.com/electricbubble/gwda"
)

func (dExt *DriverExt) TapXY(x, y float64) error {
	// tap on coordinate: [x, y] should be relative
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be < 1, got x=%v, y=%v", x, y)
	}

	x = x * float64(dExt.windowSize.Width)
	y = y * float64(dExt.windowSize.Height)
	return dExt.WebDriver.TapFloat(x, y)
}

func (dExt *DriverExt) TapByOCR(ocrText string) error {
	x, y, width, height, err := dExt.FindTextByOCR(ocrText)
	if err != nil {
		return err
	}

	return dExt.WebDriver.TapFloat(x+width*0.5, y+height*0.5)
}

func (dExt *DriverExt) TapByCV(imagePath string) error {
	x, y, width, height, err := dExt.FindImageRectInUIKit(imagePath)
	if err != nil {
		return err
	}

	return dExt.WebDriver.TapFloat(x+width*0.5, y+height*0.5)
}

func (dExt *DriverExt) Tap(param string) error {
	return dExt.TapOffset(param, 0.5, 0.5)
}

func (dExt *DriverExt) TapOffset(param string, xOffset, yOffset float64) (err error) {
	// click on element, find by name attribute
	ele, err := dExt.FindUIElement(param)
	if err == nil {
		return ele.Click()
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(param); err != nil {
		return err
	}

	return dExt.WebDriver.TapFloat(x+width*xOffset, y+height*yOffset)
}

func (dExt *DriverExt) DoubleTapXY(x, y float64) error {
	// double tap on coordinate: [x, y] should be relative
	if x > 1 || y > 1 {
		return fmt.Errorf("x, y percentage should be < 1, got x=%v, y=%v", x, y)
	}

	x = x * float64(dExt.windowSize.Width)
	y = y * float64(dExt.windowSize.Height)
	return dExt.WebDriver.DoubleTapFloat(x, y)
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

	return dExt.WebDriver.DoubleTapFloat(x+width*xOffset, y+height*yOffset)
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

	touchActions := gwda.NewTouchActions().Tap(gwda.NewTouchActionTap().WithXYFloat(x, y).WithCount(numberOfTaps))
	return dExt.PerformTouchActions(touchActions)
}
