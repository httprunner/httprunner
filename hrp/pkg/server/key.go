package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/hrp/code"
	"github.com/httprunner/httprunner/v5/hrp/pkg/uixt"
	"github.com/rs/zerolog/log"
)

func unlockHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	err = dExt.Driver.Unlock()
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

func homeHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	err = dExt.Driver.Homescreen()
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to enter homescreen", c.HandlerName()))
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

func keycodeHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var keycodeReq KeycodeRequest
	if err := c.ShouldBindJSON(&keycodeReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	err = dExt.Driver.PressKeyCode(uixt.KeyCode(keycodeReq.Keycode))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to input keycode %d", c.HandlerName(), keycodeReq.Keycode))
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
