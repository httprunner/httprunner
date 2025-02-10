package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v5/code"
	"github.com/httprunner/httprunner/v5/pkg/uixt"
	"github.com/httprunner/httprunner/v5/pkg/uixt/option"
)

var uiClients = make(map[string]uixt.IDriverExt) // UI automation clients for iOS and Android, key is udid/serial

func handleDeviceContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		serial := c.Param("serial")
		if serial == "" {
			log.Error().Str("platform", platform).Msg(fmt.Sprintf("[%s]: serial is empty", c.HandlerName()))
			c.JSON(http.StatusBadRequest, HttpResponse{
				Code:    code.GetErrorCode(code.InvalidParamError),
				Message: "serial is empty",
			})
			c.Abort()
			return
		}
		// get cached driver
		if driver, ok := uiClients[serial]; ok {
			c.Set("driver", driver)
			c.Next()
			return
		}

		switch strings.ToLower(platform) {
		case "android":
			device, err := uixt.NewAndroidDevice(
				option.WithSerialNumber(serial))
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).
					Msg("device not found")
				c.JSON(http.StatusBadRequest,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			device.Setup()

			driver, err := device.NewDriver()
			if err != nil {
				log.Error().Err(err).Str("platform", platform).Str("serial", serial).
					Msg("failed to init driver")
				c.JSON(http.StatusInternalServerError,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			c.Set("driver", driver)
			// cache driver
			uiClients[serial] = driver
		default:
			c.JSON(http.StatusBadRequest, HttpResponse{
				Code:    code.GetErrorCode(code.InvalidParamError),
				Message: fmt.Sprintf("unsupported platform %s", platform),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getContextDriver(c *gin.Context) (*uixt.DriverExt, error) {
	driverObj, exists := c.Get("driver")
	if !exists {
		handlerInitDeviceDriverFailedContext(c)
		return nil, fmt.Errorf("driver not found")
	}
	dExt := driverObj.(*uixt.DriverExt)
	return dExt, nil
}

func handlerInitDeviceDriverFailedContext(c *gin.Context) {
	log.Error().Msg("init device driver failed")
	c.JSON(http.StatusInternalServerError,
		HttpResponse{
			Code:    code.GetErrorCode(code.MobileUIDriverError),
			Message: "init driver failed",
		},
	)
	c.Abort()
}

func handlerValidateRequestFailedContext(c *gin.Context, err error) {
	log.Error().Err(err).Msg("validate request failed")
	c.JSON(http.StatusBadRequest, HttpResponse{
		Code:    code.GetErrorCode(code.InvalidParamError),
		Message: fmt.Sprintf("validate request param failed: %s", err.Error()),
	})
	c.Abort()
}
