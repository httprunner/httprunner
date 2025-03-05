package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/ai"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (r *Router) GetDriver(c *gin.Context) (driverExt *uixt.XTDriver, err error) {
	var device uixt.IDevice
	var driver uixt.IDriver
	deviceObj, exists := c.Get("device")
	if !exists {
		device, err = r.GetDevice(c)
		if err != nil {
			return nil, err
		}
	} else {
		device = deviceObj.(uixt.IDevice)
	}

	driver, err = device.NewDriver()
	if err != nil {
		RenderErrorInitDriver(c, err)
		return
	}

	driverExt = uixt.NewXTDriver(driver,
		ai.WithCVService(ai.CVServiceTypeVEDEM))
	c.Set("driver", driverExt)
	return driverExt, nil
}

func (r *Router) GetDevice(c *gin.Context) (device uixt.IDevice, err error) {
	platform := c.Param("platform")
	serial := c.Param("serial")
	if serial == "" {
		RenderErrorInitDriver(c, err)
		return
	}
	switch strings.ToLower(platform) {
	case "android":
		device, err = uixt.NewAndroidDevice(
			option.WithSerialNumber(serial))
		if err != nil {
			RenderErrorInitDriver(c, err)
			return
		}
	case "ios":
		device, err = uixt.NewIOSDevice(
			option.WithUDID(serial),
			option.WithWDAPort(8700),
			option.WithWDAMjpegPort(8800),
			option.WithResetHomeOnStartup(false),
		)
		if err != nil {
			RenderErrorInitDriver(c, err)
			return
		}
	case "browser":
		device, err = uixt.NewBrowserDevice(option.WithBrowserID(serial))
		if err != nil {
			RenderErrorInitDriver(c, err)
			return
		}
	default:
		err = fmt.Errorf("[%s]: invalid platform", c.HandlerName())
		return
	}
	err = device.Setup()
	if err != nil {
		log.Error().Err(err).Msg("setup device failed")
	}
	c.Set("device", device)
	return device, nil
}

func RenderSuccess(c *gin.Context, result interface{}) {
	c.JSON(http.StatusOK, HttpResponse{
		Code:    code.Success,
		Message: "success",
		Result:  result,
	})
}

func RenderError(c *gin.Context, err error) {
	log.Error().Err(err).Msgf("failed to %s", c.HandlerName())
	c.JSON(http.StatusInternalServerError,
		HttpResponse{
			Code:    code.GetErrorCode(err),
			Message: "grey " + err.Error(),
		},
	)
	c.Abort()
}

func RenderErrorInitDriver(c *gin.Context, err error) {
	log.Error().Err(err).Msg("init device driver failed")
	errCode := code.GetErrorCode(err)
	if errCode == code.GeneralFail {
		errCode = code.GetErrorCode(code.MobileUIDriverError)
	}
	c.JSON(http.StatusInternalServerError,
		HttpResponse{
			Code:    errCode,
			Message: "grey init driver failed",
		},
	)
	c.Abort()
}

func RenderErrorValidateRequest(c *gin.Context, err error) {
	log.Error().Err(err).Msg("validate request failed")
	c.JSON(http.StatusBadRequest, HttpResponse{
		Code:    code.GetErrorCode(code.InvalidParamError),
		Message: fmt.Sprintf("grey validate request param failed: %s", err.Error()),
	})
	c.Abort()
}
