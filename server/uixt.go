package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/rs/zerolog/log"
)

// exec a single uixt action
func uixtActionHandler(c *gin.Context) {
	dExt, err := GetDriver(c)
	if err != nil {
		return
	}

	var req uixt.MobileAction
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderErrorValidateRequest(c, err)
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
	RenderSuccess(c, true)
}

// exec multiple uixt actions
func uixtActionsHandler(c *gin.Context) {
	dExt, err := GetDriver(c)
	if err != nil {
		return
	}

	var actions []uixt.MobileAction
	if err := c.ShouldBindJSON(&actions); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	for _, action := range actions {
		if err = dExt.DoAction(action); err != nil {
			log.Err(err).Interface("action", action).
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
	}
	RenderSuccess(c, true)
}
