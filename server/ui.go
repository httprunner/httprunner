package server

import (
	"github.com/gin-gonic/gin"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

func tapHandler(c *gin.Context) {
	var tapReq TapRequest
	if err := c.ShouldBindJSON(&tapReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := GetDriver(c)
	if err != nil {
		return
	}
	if tapReq.Duration > 0 {
		err = driver.Drag(tapReq.X, tapReq.Y, tapReq.X, tapReq.Y,
			option.WithDuration(tapReq.Duration),
			option.WithAbsoluteCoordinate(true))
	} else {
		err = driver.TapAbsXY(tapReq.X, tapReq.Y)
	}
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func uploadHandler(c *gin.Context) {
	var uploadRequest uploadRequest
	if err := c.ShouldBindJSON(&uploadRequest); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}
	err = driver.IDriver.(*uixt.BrowserWebDriver).UploadFile(uploadRequest.X, uploadRequest.Y, uploadRequest.FileUrl, uploadRequest.FileFormat)
	if err != nil {
		c.Abort()
		return
	}
	RenderSuccess(c, true)
}
func hoverHandler(c *gin.Context) {
	var hoverReq HoverRequest
	if err := c.ShouldBindJSON(&hoverReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}

	err = driver.IDriver.(*uixt.BrowserWebDriver).Hover(hoverReq.X, hoverReq.Y)

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func scrollHandler(c *gin.Context) {
	var scrollReq ScrollRequest
	if err := c.ShouldBindJSON(&scrollReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := GetDriver(c)
	if err != nil {
		RenderError(c, err)
		return
	}

	err = driver.IDriver.(*uixt.BrowserWebDriver).Scroll(scrollReq.Delta)

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func doubleTapHandler(c *gin.Context) {
	var tapReq TapRequest
	if err := c.ShouldBindJSON(&tapReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}

	driver, err := GetDriver(c)
	if err != nil {
		return
	}

	if tapReq.X < 1 && tapReq.Y < 1 {
		err = driver.DoubleTapXY(tapReq.X, tapReq.Y)
	} else {
		err = driver.DoubleTapXY(tapReq.X, tapReq.Y,
			option.WithAbsoluteCoordinate(true))
	}

	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func dragHandler(c *gin.Context) {
	var dragReq DragRequest
	if err := c.ShouldBindJSON(&dragReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	if dragReq.Duration == 0 {
		dragReq.Duration = 1
	}
	driver, err := GetDriver(c)
	if err != nil {
		return
	}

	err = driver.Drag(dragReq.FromX, dragReq.FromY, dragReq.ToX, dragReq.ToY,
		option.WithDuration(dragReq.Duration), option.WithPressDuration(dragReq.PressDuration),
		option.WithAbsoluteCoordinate(true))
	if err != nil {
		RenderError(c, err)
		return
	}
	RenderSuccess(c, true)
}

func inputHandler(c *gin.Context) {
	var inputReq InputRequest
	if err := c.ShouldBindJSON(&inputReq); err != nil {
		RenderErrorValidateRequest(c, err)
		return
	}
	driver, err := GetDriver(c)
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
