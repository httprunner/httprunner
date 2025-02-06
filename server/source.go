package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/options"
	"github.com/rs/zerolog/log"
)

func screenshotHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	raw, err := dExt.Driver.Screenshot()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get screenshot", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK,
		HttpResponse{
			Code:    code.Success,
			Message: "success",
			Result:  base64.StdEncoding.EncodeToString(raw.Bytes()),
		},
	)
}

func screenResultHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var screenReq ScreenRequest
	if err := c.ShouldBindJSON(&screenReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	var actionOptions []options.ActionOption
	if screenReq.Options != nil {
		actionOptions = screenReq.Options.Options()
	}

	screenResult, err := dExt.GetScreenResult(actionOptions...)
	if err != nil {
		log.Err(err).Msg("get screen result failed")
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK,
		HttpResponse{
			Code:    code.Success,
			Message: "success",
			Result:  screenResult,
		},
	)
}

func sourceHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	app, err := dExt.Driver.GetForegroundApp()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get foreground app", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	source, err := dExt.Driver.Source(uixt.NewSourceOption().WithProcessName(app.PackageName))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get source %s", c.HandlerName(), app.PackageName))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, HttpResponse{Result: source})
}

func adbSourceHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	source, err := dExt.Driver.Source()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to get adb source", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Result: source})
}
