package server

import (
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func (r *Router) screenshotHandler(c *gin.Context) {
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	raw, err := driver.ScreenShot()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, base64.StdEncoding.EncodeToString(raw.Bytes()))
}

func (r *Router) screenResultHandler(c *gin.Context) {
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	var screenReq ScreenRequest
	if err := c.ShouldBindJSON(&screenReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	var actionOptions []option.ActionOption
	if screenReq.Options != nil {
		actionOptions = screenReq.Options.Options()
	}

	screenResult, err := driver.GetScreenResult(actionOptions...)
	if err != nil {
		log.Err(err).Msg("get screen result failed")
		RenderError(c, err)
		return
	}
	RenderSuccess(c, screenResult)
}

func (r *Router) adbSourceHandler(c *gin.Context) {
	dExt, err := r.GetDriver(c)
	if err != nil {
		return
	}

	source, err := dExt.Source()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, source)
}
