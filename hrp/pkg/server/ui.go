package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
	"github.com/rs/zerolog/log"
)

func tapHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var tapReq TapRequest
	if err := c.ShouldBindJSON(&tapReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	var actionOptions []uixt.ActionOption
	if tapReq.Options != nil {
		actionOptions = tapReq.Options.Options()
	}

	if tapReq.Text != "" {
		err := dExt.TapByOCR(tapReq.Text, actionOptions...)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap text %s", c.HandlerName(), tapReq.Text))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	} else if tapReq.X < 1 && tapReq.Y < 1 {
		err := dExt.TapXY(tapReq.X, tapReq.Y, actionOptions...)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap %f, %f", c.HandlerName(), tapReq.X, tapReq.Y))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	} else {
		err := dExt.TapAbsXY(tapReq.X, tapReq.Y, actionOptions...)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to tap %f, %f", c.HandlerName(), tapReq.X, tapReq.Y))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func dragHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var dragReq DragRequest
	if err := c.ShouldBindJSON(&dragReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	if dragReq.FromX < 1 && dragReq.FromY < 1 && dragReq.ToX < 1 && dragReq.ToY < 1 {
		err := dExt.SwipeRelative(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, uixt.WithPressDuration(dragReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to drag from %f, %f to %f, %f", c.HandlerName(), dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	} else {
		err := dExt.Driver.SwipeFloat(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY, uixt.WithPressDuration(dragReq.Duration))
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("[%s]: failed to drag from %f, %f to %f, %f", c.HandlerName(), dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY))
			c.JSON(http.StatusInternalServerError,
				HttpResponse{
					Code:    code.GetErrorCode(err),
					Message: err.Error(),
				},
			)
			c.Abort()
			return
		}
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}

func inputHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var inputReq InputRequest
	if err := c.ShouldBindJSON(&inputReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	err = dExt.Driver.SendKeys(inputReq.Text, uixt.WithFrequency(inputReq.Frequency))
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to input text %s", c.HandlerName(), inputReq.Text))
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
