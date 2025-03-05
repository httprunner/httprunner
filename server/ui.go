package server

import (
	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (r *Router) tapHandler(c *gin.Context) {
	var tapReq TapRequest
	if err := c.ShouldBindJSON(&tapReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	if tapReq.Duration > 0 {
		err = driver.Drag(tapReq.X, tapReq.Y, tapReq.X, tapReq.Y,
			option.WithDuration(tapReq.Duration))
	} else {
		err = driver.TapXY(tapReq.X, tapReq.Y)
	}
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) rightClickHandler(c *gin.Context) {
	var rightClickReq TapRequest
	if err := c.ShouldBindJSON(&rightClickReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.IDriver.(*uixt.BrowserDriver).
		RightClick(rightClickReq.X, rightClickReq.Y)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) uploadHandler(c *gin.Context) {
	var uploadRequest uploadRequest
	if err := c.ShouldBindJSON(&uploadRequest); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}
	err = driver.IDriver.(*uixt.BrowserDriver).
		UploadFile(uploadRequest.X, uploadRequest.Y,
			uploadRequest.FileUrl, uploadRequest.FileFormat)
	if err != nil {
		c.Abort()
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) hoverHandler(c *gin.Context) {
	var hoverReq HoverRequest
	if err := c.ShouldBindJSON(&hoverReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}

	err = driver.IDriver.(*uixt.BrowserDriver).
		Hover(hoverReq.X, hoverReq.Y)

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) scrollHandler(c *gin.Context) {
	var scrollReq ScrollRequest
	if err := c.ShouldBindJSON(&scrollReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}

	err = driver.IDriver.(*uixt.BrowserDriver).
		Scroll(scrollReq.Delta)

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) doubleTapHandler(c *gin.Context) {
	var tapReq TapRequest
	if err := c.ShouldBindJSON(&tapReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	err = driver.DoubleTap(tapReq.X, tapReq.Y)
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) dragHandler(c *gin.Context) {
	var dragReq DragRequest
	if err := c.ShouldBindJSON(&dragReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	if dragReq.Duration == 0 {
		dragReq.Duration = 1
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	err = driver.Drag(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY,
		option.WithDuration(dragReq.Duration),
		option.WithPressDuration(dragReq.PressDuration))
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) inputHandler(c *gin.Context) {
	var inputReq InputRequest
	if err := c.ShouldBindJSON(&inputReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.Input(inputReq.Text, option.WithFrequency(inputReq.Frequency))
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}
