package uixt

// newVEDEMUIService return image service for
func newVEDEMUIService() (*veDEMImageService, error) {
	return newVEDEMImageService("ui", "upload")
}
