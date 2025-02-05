package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/hrp/code"
	"github.com/httprunner/httprunner/v5/hrp/pkg/uixt"
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
			log.Err(err).Str("text", tapReq.Text).Msg("tap text failed")
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
			log.Err(err).Float64("x", tapReq.X).Float64("y", tapReq.Y).Msg("tap relative xy failed")
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
			log.Err(err).Float64("x", tapReq.X).Float64("y", tapReq.Y).Msg("tap abs xy failed")
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

	var actionOptions []uixt.ActionOption
	if dragReq.Options != nil {
		actionOptions = dragReq.Options.Options()
	}

	if dragReq.FromX < 1 && dragReq.FromY < 1 && dragReq.ToX < 1 && dragReq.ToY < 1 {
		err := dExt.SwipeRelative(
			dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY,
			actionOptions...)
		if err != nil {
			log.Err(err).
				Float64("from_x", dragReq.FromX).Float64("from_y", dragReq.FromY).
				Float64("to_x", dragReq.ToX).Float64("to_y", dragReq.ToY).
				Msg("swipe relative failed")
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
		err := dExt.Driver.Swipe(
			dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY,
			actionOptions...)
		if err != nil {
			log.Err(err).
				Float64("from_x", dragReq.FromX).Float64("from_y", dragReq.FromY).
				Float64("to_x", dragReq.ToX).Float64("to_y", dragReq.ToY).
				Msg("swipe absolute failed")
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
