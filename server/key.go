package server

import (
	"github.com/gin-gonic/gin"

	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (r *Router) unlockHandler(c *gin.Context) {
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.Unlock()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) homeHandler(c *gin.Context) {
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.Home()
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) backspaceHandler(c *gin.Context) {
	var deleteReq option.DeleteRequest
	if err := c.ShouldBindJSON(&deleteReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	if deleteReq.Count == 0 {
		deleteReq.Count = 20
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.Backspace(deleteReq.Count)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) keycodeHandler(c *gin.Context) {
	var keycodeReq option.KeycodeRequest
	if err := c.ShouldBindJSON(&keycodeReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	// TODO FIXME
	err = driver.IDriver.(*uixt.ADBDriver).
		PressKeyCode(uixt.KeyCode(keycodeReq.Keycode), uixt.KMEmpty)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}
