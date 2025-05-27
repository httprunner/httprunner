package server

import (
	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

// processUnifiedRequest is a helper function to handle common request processing
func (r *Router) processUnifiedRequest(c *gin.Context, actionType option.ActionMethod) (*option.UnifiedActionRequest, error) {
	var req option.UnifiedActionRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderErrorValidateRequest(c, err)
		return nil, err
	}

	// Set platform and serial from URL parameters
	setRequestContextFromURL(c, &req)

	// Validate for HTTP API usage
	if err := req.ValidateForHTTPAPI(actionType); err != nil {
		RenderErrorValidateRequest(c, err)
		return nil, err
	}

	return &req, nil
}

// setRequestContextFromURL sets platform and serial from URL parameters
func setRequestContextFromURL(c *gin.Context, req *option.UnifiedActionRequest) {
	if req.Platform == "" {
		req.Platform = c.Param("platform")
	}
	if req.Serial == "" {
		req.Serial = c.Param("serial")
	}
}

func (r *Router) tapHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_Tap)
	if err != nil {
		return // Error already handled in processUnifiedRequest
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	// Use UnifiedActionRequest directly
	if req.GetDuration() > 0 {
		err = driver.Drag(req.GetX(), req.GetY(), req.GetX(), req.GetY(),
			option.WithDuration(req.GetDuration()))
	} else {
		err = driver.TapXY(req.GetX(), req.GetY())
	}
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) rightClickHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_RightClick)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.IDriver.(*uixt.BrowserDriver).
		SecondaryClick(req.GetX(), req.GetY())
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
	req, err := r.processUnifiedRequest(c, option.ACTION_Hover)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}

	err = driver.IDriver.(*uixt.BrowserDriver).
		Hover(req.GetX(), req.GetY())

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) scrollHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_Scroll)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}

	err = driver.IDriver.(*uixt.BrowserDriver).
		Scroll(req.GetDelta())

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) doubleTapHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_DoubleTap)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	err = driver.DoubleTap(req.GetX(), req.GetY())
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) dragHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_Drag)
	if err != nil {
		return
	}

	duration := req.GetDuration()
	if duration == 0 {
		duration = 1
	}
	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}

	err = driver.Drag(req.GetFromX(), req.GetFromY(), req.GetToX(), req.GetToY(),
		option.WithDuration(duration),
		option.WithPressDuration(req.GetPressDuration()))
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func (r *Router) inputHandler(c *gin.Context) {
	req, err := r.processUnifiedRequest(c, option.ACTION_Input)
	if err != nil {
		return
	}

	driver, err := r.GetDriver(c)
	if err != nil {
		return
	}
	err = driver.Input(req.Text, option.WithFrequency(req.GetFrequency()))
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}
