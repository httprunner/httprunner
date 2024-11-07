package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/pkg/uixt"
	"github.com/rs/zerolog/log"
)

func uixtActionHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var req uixt.MobileAction
	if err := c.ShouldBindJSON(&req); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	if err = dExt.DoAction(req); err != nil {
		log.Err(err).Interface("action", req).
			Msg("exec uixt action failed")
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
