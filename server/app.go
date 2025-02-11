package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/rs/zerolog/log"
)

func foregroundAppHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	appInfo, err := dExt.GetDriver().GetForegroundApp()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: appInfo})
}

func clearAppHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	var appClearReq AppClearRequest
	if err := c.ShouldBindJSON(&appClearReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	err = dExt.Driver.AppClear(appClearReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to unlick screen", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func launchAppHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	var appLaunchReq AppLaunchRequest
	if err := c.ShouldBindJSON(&appLaunchReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	err = dExt.GetDriver().AppLaunch(appLaunchReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to launch app %s", c.HandlerName(), appLaunchReq.PackageName))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func terminalAppHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	var appTerminalReq AppTerminalRequest
	if err := c.ShouldBindJSON(&appTerminalReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	success, err := dExt.GetDriver().AppTerminate(appTerminalReq.PackageName)
	if !success {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to launch app %s", c.HandlerName(), appTerminalReq.PackageName))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}
