package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/code"
	"github.com/httprunner/httprunner/v4/hrp/pkg/gadb"
)

func listDeviceHandler(c *gin.Context) {
	platform := c.Param("platform")
	switch strings.ToLower(platform) {
	case "android":
		{
			client, err := gadb.NewClient()
			if err != nil {
				log.Err(err).Msg("failed to init adb client")
				c.JSON(http.StatusInternalServerError,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			devices, err := client.DeviceList()
			if err != nil && strings.Contains(err.Error(), "no android device found") {
				c.JSON(http.StatusOK, HttpResponse{Result: nil})
				return
			} else if err != nil {
				log.Err(err).Msg("failed to list devices")
				c.JSON(http.StatusInternalServerError,
					HttpResponse{
						Code:    code.GetErrorCode(err),
						Message: err.Error(),
					},
				)
				c.Abort()
				return
			}
			var deviceList []interface{}
			for _, device := range devices {
				brand, err := device.Brand()
				if err != nil {
					log.Err(err).Msg("failed to get device brand")
					c.JSON(http.StatusInternalServerError,
						HttpResponse{
							Code:    code.GetErrorCode(err),
							Message: err.Error(),
						},
					)
					c.Abort()
					return
				}
				model, err := device.Model()
				if err != nil {
					log.Err(err).Msg("failed to get device model")
					c.JSON(http.StatusInternalServerError,
						HttpResponse{
							Code:    code.GetErrorCode(err),
							Message: err.Error(),
						},
					)
					c.Abort()
					return
				}
				deviceInfo := map[string]interface{}{
					"serial":   device.Serial(),
					"brand":    brand,
					"model":    model,
					"platform": "android",
				}
				deviceList = append(deviceList, deviceInfo)
			}
			c.JSON(http.StatusOK, HttpResponse{Result: deviceList})
			return
		}
	default:
		{
			c.JSON(http.StatusBadRequest, HttpResponse{
				Code:    code.GetErrorCode(code.InvalidParamError),
				Message: fmt.Sprintf("unsupported platform %s", platform),
			})
			c.Abort()
			return
		}
	}
}

func loginHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	info, err := dExt.Driver.LoginNoneUI(loginReq.PackageName, loginReq.PhoneNumber, loginReq.Captcha, loginReq.Password)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success", Result: info})
}

func logoutHandler(c *gin.Context) {
	dExt, err := getContextDriver(c)
	if err != nil {
		return
	}

	var logoutReq LogoutRequest
	if err := c.ShouldBindJSON(&logoutReq); err != nil {
		handlerValidateRequestFailedContext(c, err)
		return
	}

	err = dExt.Driver.LogoutNoneUI(logoutReq.PackageName)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("[%s]: failed to login", c.HandlerName()))
		c.JSON(http.StatusInternalServerError,
			HttpResponse{
				Code:    code.GetErrorCode(err),
				Message: err.Error(),
			},
		)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, HttpResponse{Code: 0, Message: "success"})
}
