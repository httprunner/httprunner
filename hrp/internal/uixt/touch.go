package uixt

func (dExt *DriverExt) ForceTouch(pathname string, pressure float64, duration ...float64) (err error) {
	return dExt.ForceTouchOffset(pathname, pressure, 0.5, 0.5, duration...)
}

func (dExt *DriverExt) ForceTouchOffset(pathname string, pressure, xOffset, yOffset float64, duration ...float64) (err error) {
	if len(duration) == 0 {
		duration = []float64{1.0}
	}
	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	return dExt.ForceTouchFloat(x+width*xOffset, y+height*yOffset, pressure, duration[0])
}

func (dExt *DriverExt) TouchAndHold(pathname string, duration ...float64) (err error) {
	return dExt.TouchAndHoldOffset(pathname, 0.5, 0.5, duration...)
}

func (dExt *DriverExt) TouchAndHoldOffset(pathname string, xOffset, yOffset float64, duration ...float64) (err error) {
	if len(duration) == 0 {
		duration = []float64{1.0}
	}
	var x, y, width, height float64
	if x, y, width, height, err = dExt.FindUIRectInUIKit(pathname); err != nil {
		return err
	}

	return dExt.TouchAndHoldFloat(x+width*xOffset, y+height*yOffset, duration[0])
}
