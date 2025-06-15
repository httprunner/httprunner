package server

import (
	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/uixt/option"
	"github.com/rs/zerolog/log"
)

// exec a single uixt action
func (r *Router) uixtActionHandler(c *gin.Context) {
	dExt, err := r.GetDriver(c)
	if err != nil {
		return
	}

	var req option.MobileAction
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	if _, err = dExt.ExecuteAction(c.Request.Context(), req); err != nil {
		log.Err(err).Interface("action", req).
			Msg("exec uixt action failed")
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

// exec multiple uixt actions
func (r *Router) uixtActionsHandler(c *gin.Context) {
	dExt, err := r.GetDriver(c)
	if err != nil {
		return
	}

	var actions []option.MobileAction
	if err := c.ShouldBindJSON(&actions); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	for _, action := range actions {
		if _, err = dExt.ExecuteAction(c.Request.Context(), action); err != nil {
			log.Err(err).Interface("action", action).
				Msg("exec uixt action failed")
			RenderError(c, err)
			return
		}
	}
	RenderSuccess(c, true)
}
