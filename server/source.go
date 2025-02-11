package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func screenshotHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	raw, err := dExt.ScreenShot()
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
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	var screenReq ScreenRequest
	if err := c.ShouldBindJSON(&screenReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	var actionOptions []option.ActionOption
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

func adbSourceHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	source, err := dExt.Source()
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
