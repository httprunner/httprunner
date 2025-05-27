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
	req, err := r.processUnifiedRequest(c, option.ACTION_Backspace)
	if err != nil {
		return
	}

	count := req.GetCount()
	if count == 0 {
		count = 20
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.Backspace(count)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) keycodeHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_KeyCode)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	// TODO FIXME
	err = driver.IDriver.(*uixt.ADBDriver).
		PressKeyCode(uixt.KeyCode(req.GetKeycode()), uixt.KMEmpty)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}
