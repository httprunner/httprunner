package server

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/uixt"
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
	var appInfoReq AppInfoRequest
	if err := c.ShouldBindQuery(&appInfoReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	device, err := r.GetDevice(c)
	if err != nil {
		return
	}
	if androidDevice, ok := device.(*uixt.AndroidDevice); ok {
		appInfo, err := androidDevice.GetAppInfo(appInfoReq.PackageName)
		if err != nil {
			RenderError(c, err)
			return
		}
		RenderSuccess(c, appInfo)
		return
	} else if iOSDevice, ok := device.(*uixt.IOSDevice); ok {
		appInfo, err := iOSDevice.GetAppInfo(appInfoReq.PackageName)
		if err != nil {
			RenderError(c, err)
			return
		}
		RenderSuccess(c, appInfo)
		return
	}
}

func (r *Router) clearAppHandler(c *gin.Context) {
	var appClearReq AppClearRequest
	if err := c.ShouldBindJSON(&appClearReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.AppClear(appClearReq.PackageName)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) launchAppHandler(c *gin.Context) {
	var appLaunchReq AppLaunchRequest
	if err := c.ShouldBindJSON(&appLaunchReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.AppLaunch(appLaunchReq.PackageName)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) terminalAppHandler(c *gin.Context) {
	var appTerminalReq AppTerminalRequest
	if err := c.ShouldBindJSON(&appTerminalReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	_, err = driver.AppTerminate(appTerminalReq.PackageName)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) uninstallAppHandler(c *gin.Context) {
	var appUninstallReq AppUninstallRequest
	if err := c.ShouldBindJSON(&appUninstallReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.GetDevice().Uninstall(appUninstallReq.PackageName)
	if err != nil {
		log.Err(err).Msg("failed to uninstall app")
	}
	RenderSuccess(c, true)
}
