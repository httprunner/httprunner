package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
)

func unlockHandler(c *gin.Context) {
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	err = dExt.Unlock()
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
	dExt, err := GetContextDriver(c)
	if err != nil {
		return
	}

	err = dExt.Home()
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
