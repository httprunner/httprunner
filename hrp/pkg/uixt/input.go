package uixt

func (dExt *DriverExt) Input(text string) (err error) {
	return dExt.Driver.Input(text)
}
