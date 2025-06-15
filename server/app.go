package server

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (r *Router) foregroundAppHandler(c *gin.Context) {
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	appInfo, err := driver.ForegroundInfo()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, appInfo)
}

func (r *Router) appInfoHandler(c *gin.Context) {
	var req option.ActionOptions
	if err := c.ShouldBindQuery(&req); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	// Set platform and serial from URL parameters
	setRequestContextFromURL(c, &req)

	// Validate for HTTP API usage
	if err := req.ValidateForHTTPAPI(option.ACTION_AppInfo); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	device, err := r.GetDevice(c)
	if err != nil {
		return
	}
	if androidDevice, ok := device.(*uixt.AndroidDevice); ok {
		appInfo, err := androidDevice.GetAppInfo(req.PackageName)
		if err != nil {
			RenderError(c, err)
			return
		}
		RenderSuccess(c, appInfo)
		return
	} else if iOSDevice, ok := device.(*uixt.IOSDevice); ok {
		appInfo, err := iOSDevice.GetAppInfo(req.PackageName)
		if err != nil {
			RenderError(c, err)
			return
		}
		RenderSuccess(c, appInfo)
		return
	}
}

func (r *Router) clearAppHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_AppClear)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.AppClear(req.PackageName)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) launchAppHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_AppLaunch)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.AppLaunch(req.PackageName)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) terminalAppHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_AppTerminate)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	_, err = driver.AppTerminate(req.PackageName)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) uninstallAppHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_AppUninstall)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.GetDevice().Uninstall(req.PackageName)
	if err != nil {
		log.Err(err).Msg("failed to uninstall app")
	}
	RenderSuccess(c, true)
}
