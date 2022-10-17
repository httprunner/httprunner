package uixt

func (dExt *DriverExt) Drag(pathname string, toX, toY int, pressForDuration ...float64) (err error) {
	return dExt.DragFloat(pathname, float64(toX), float64(toY), pressForDuration...)
}

func (dExt *DriverExt) DragFloat(pathname string, toX, toY float64, pressForDuration ...float64) (err error) {
	return dExt.DragOffsetFloat(pathname, toX, toY, 0.5, 0.5, pressForDuration...)
}

func (dExt *DriverExt) DragOffset(pathname string, toX, toY int, xOffset, yOffset float64, pressForDuration ...float64) (err error) {
	return dExt.DragOffsetFloat(pathname, float64(toX), float64(toY), xOffset, yOffset, pressForDuration...)
}

func (dExt *DriverExt) DragOffsetFloat(pathname string, toX, toY, xOffset, yOffset float64, pressForDuration ...float64) (err error) {
	if len(pressForDuration) == 0 {
		pressForDuration = []float64{1.0}
	}

	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	fromX := x + width*xOffset
	fromY := y + height*yOffset

	return dExt.Driver.DragFloat(fromX, fromY, toX, toY,
		WithDataPressDuration(pressForDuration[0]))
}
