package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/uixt"
	"github.com/httprunner/httprunner/v5/uixt/option"
)

func (r *Router) GetDevice(c *gin.Context) (uixt.IDevice, error) {
	platform := c.Param("platform")
	switch strings.ToLower(platform) {
	case "android":
		serial := c.Param("serial")
		if serial == "" {
			err := fmt.Errorf("[%s]: serial is empty", c.HandlerName())
			log.Error().Err(err).Str("platform", platform).Msg(err.Error())
			RenderError(c, err)
			return nil, err
		}
		device, err := uixt.NewAndroidDevice(option.WithSerialNumber(serial))
		if err != nil {
			time.Sleep(5 * time.Second)
			device, err = uixt.NewAndroidDevice(option.WithSerialNumber(serial))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).
					Msg(fmt.Sprintf("[%s]: Device Not Found; %s", c.HandlerName(), err.Error()))
				RenderErrorInitDevice(c, err)
				return nil, err
			}
		}
		c.Set("device", device)
		return device, nil

	case "ios":
		serial := c.Param("serial")
		if serial == "" {
			err := fmt.Errorf("[%s]: serial is empty", c.HandlerName())
			log.Error().Err(err).Str("platform", platform).Msg(err.Error())
			RenderError(c, err)
			return nil, err
		}
		device, err := uixt.NewIOSDevice(
			option.WithUDID(serial),
			option.WithWDAPort(8700),
			option.WithWDAMjpegPort(8800),
			option.WithResetHomeOnStartup(false))
		if err != nil {
			log.Error().Err(err).Str("platform", platform).Str("serial", serial).
				Msg(fmt.Sprintf("[%s]: Device Not Found", c.HandlerName()))
			RenderErrorInitDevice(c, err)
			return nil, err
		}
		c.Set("device", device)
		return device, nil

	case "browser":
		serial := c.Param("serial")
		if serial == "" {
			err := fmt.Errorf("[%s]: serial is empty", c.HandlerName())
			log.Error().Err(err).Str("platform", platform).Msg(err.Error())
			RenderError(c, err)
			return nil, err
		}
		device, err := uixt.NewBrowserDevice(option.WithBrowserID(serial))
		if err != nil {
			RenderErrorInitDevice(c, err)
			return nil, err
		}
		c.Set("device", device)
		return device, nil

	default:
		err := fmt.Errorf("[%s]: invalid platform", c.HandlerName())
		RenderError(c, err)
		return nil, err
	}
}

func (r *Router) GetDriver(c *gin.Context) (*uixt.XTDriver, error) {
	platform := c.Param("platform")

	// Try to get existing device from context
	deviceObj, exists := c.Get("device")
	var device uixt.IDevice
	var err error

	if !exists {
		device, err = r.GetDevice(c)
		if err != nil {
			return nil, err
		}
	} else {
		device = deviceObj.(uixt.IDevice)
	}

	// Create driver
	driver, err := device.NewDriver()
	if err != nil {
		log.Error().Err(err).Str("platform", platform).Str("serial", device.UUID()).
			Msg(fmt.Sprintf("[%s]: Failed New Driver", c.HandlerName()))
		RenderErrorInitDriver(c, err)
		return nil, err
	}

	// Create XTDriver wrapper
	xtDriver, err := uixt.NewXTDriver(driver)
	if err != nil {
		RenderErrorInitDriver(c, err)
		return nil, err
	}

	c.Set("driver", xtDriver)
	return xtDriver, nil
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
	c.JSON(http.StatusOK,
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
	c.JSON(http.StatusOK,
		HttpResponse{
			Code:    errCode,
			Message: "grey init driver failed",
		},
	)
	c.Abort()
}

func RenderErrorInitDevice(c *gin.Context, err error) {
	log.Error().Err(err).Msg("init device failed")
	errCode := code.GetErrorCode(err)
	if errCode == code.GeneralFail {
		errCode = code.GetErrorCode(code.DeviceConnectionError)
	}
	c.JSON(http.StatusInternalServerError,
		HttpResponse{
			Code:    errCode,
			Message: "grey init device failed",
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
