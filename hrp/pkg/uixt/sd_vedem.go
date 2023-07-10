package uixt

// newVEDEMSDService return image service for scenario detection
func newVEDEMSDService() (*veDEMImageService, error) {
	return newVEDEMImageService("liveType", "upload")
}
